/*=================================
@Author :tix_hjq
@Date   :2020/10/18 下午3:15
@File   :sds.go
@email  :hjq1922451756@gmail.com or 1922451756@qq.com
@version:1.15.2
=================================*/

package main

import (
	"fmt"
	"runtime"
	"unsafe"
)

/*
	SDS:[Simple Dynamic Strings]
	SDS struct use len to avoid "\0"
		e.g["hello \0 world"]==>"\0" to stop

	Note:
		about flag pack:
			In fact,ori tony_redis sds use __pack__ to merge flag&buf to reduce free size.
			but i don't know how to achieve this op in golang ==> not to achieve.
			use flag=buf[0] to replace it.
			it maybe to more waste size.
		about gc&point:
			I try to use ori tony_redis template,save shiftPoint to point buffer,but because
			gc,if not used point item ==> free memory.
		about interface:
			in fact,use interface make struct==>only struct,but it's reduce effective.Unless no way to achieve,
			can't to use it
		about single struct func:
			it's not elegant.through now isn't enough elegant
		final struct:
			direly use point to memory,"dynamic type" skip struct
*/

type (
	sdsDataType = byte
	sdshdr5     struct {
		/*
			flags:(1 sdsDataType)[0-2:(sds_type_sigmoid),3-7:(free)]
				max sdsType==> max flag==4 flag only need 3bit
			about pack position:
				because i can't pack position,so ==> flag insert into buf
												 ==> buf[0]=flag
			about why i not to use buf[0]=flag
				because i don't know how to shift buf[1]==>buf[0],final to format{*[]buf},
				direct to pos type==buf[0:....] is very dangerous.
			buf: data save address
		*/
		//flags sdsDataType
		buf []sdsDataType
	}

	/*
		len:sds all len
		alloc:already use size
	*/
	sdshdr8 struct {
		//flags sdsDataType
		len   uint8
		alloc uint8
		buf   []sdsDataType
	}

	sdshdr16 struct {
		//flags sdsDataType
		len   uint16
		alloc uint16
		buf   []sdsDataType
	}

	sdshdr32 struct {
		//flags sdsDataType
		len   uint32
		alloc uint32
		buf   []sdsDataType
	}

	sdshdr64 struct {
		//flags sdsDataType
		len   uint64
		alloc uint64
		buf   []sdsDataType
	}

	sdshdr struct {
		//direly use point to memory,"dynamic type" skip struct
		len   unsafe.Pointer
		alloc unsafe.Pointer
		buf   *[]sdsDataType
	}
)

//sds_type_sigmoid
const (
	SDS_TYPE_5 sdsDataType = iota
	SDS_TYPE_8
	SDS_TYPE_16
	SDS_TYPE_32
	SDS_TYPE_64
	SDS_TYPE_LEN   sdsDataType = 3
	SDS_MAX_EXPAND int         = 1024 * 1024
	SDS_MAX_TYPE_5             = 31
)

func JudgeSdsType(initLen int) byte {
	/*
		ori tony_redis function name is ["sdsReqType"]
		func:judge len to pick up SdsType
	*/

	if initLen < 1<<5 {
		return SDS_TYPE_5
	} else if initLen < 1<<8 {
		return SDS_TYPE_8
	} else if initLen < 1<<16 {
		return SDS_TYPE_16
	} else if initLen < 1<<32 {
		return SDS_TYPE_32
	} else {
		return SDS_TYPE_64
	}
}

//func NewSds(initLen int)*[]sdsDataType{
func NewSds(initLen int, initData []sdsDataType) *sdshdr {
	/*
		ori tony_redis function name is ["sdsnewlen"]
		init New Sds
		return: bufPoint
	*/

	//new empty init Type5 high prob to expand,to waste time==>default new init use type_8
	//already exist Type5 low prob to expand,fix ori type
	sdsType := JudgeSdsType(initLen)
	if sdsType == SDS_TYPE_5 && initLen == 0 {
		sdsType = SDS_TYPE_8
	}

	//give up version:[init buffer&type,[+1+1]=='\0'&flag],because gc
	//init buffer,[+1]=='\0'
	initAlloc := 0
	buf := make([]sdsDataType, initLen+1)
	if len(initData) != 0 {
		copy(buf[1:], initData)
		initAlloc = len(initData)
	}
	buf[0] = sdsType
	if sdsType == SDS_TYPE_5 {
		buf[0] = sdsType | sdsDataType(initLen)<<SDS_TYPE_LEN
	}

	switch sdsType {
	case SDS_TYPE_5:
		sds := sdshdr5{buf: buf}
		return &sdshdr{unsafe.Pointer(&sds.buf[0]), unsafe.Pointer(&sds.buf[0]), &sds.buf}
		//return &sds
	case SDS_TYPE_8:
		sds := sdshdr8{len: uint8(initLen), buf: buf, alloc: uint8(initAlloc)}
		return &sdshdr{unsafe.Pointer(&sds.len), unsafe.Pointer(&sds.alloc), &sds.buf}
		//return &dataP
	case SDS_TYPE_16:
		sds := sdshdr16{len: uint16(initLen), buf: buf, alloc: uint16(initAlloc)}
		//fmt.Println(sds)
		//dataP:=sds.buf
		return &sdshdr{unsafe.Pointer(&sds.len), unsafe.Pointer(&sds.alloc), &sds.buf}
		//return &dataP
	case SDS_TYPE_32:
		sds := sdshdr32{len: uint32(initLen), buf: buf, alloc: uint32(initAlloc)}
		//dataP:=sds.buf
		return &sdshdr{unsafe.Pointer(&sds.len), unsafe.Pointer(&sds.alloc), &sds.buf}
		//return &dataP
	case SDS_TYPE_64:
		sds := sdshdr64{len: uint64(initLen), buf: buf, alloc: uint64(initAlloc)}
		//dataP:=sds.buf
		return &sdshdr{unsafe.Pointer(&sds.len), unsafe.Pointer(&sds.alloc), &sds.buf}
		//return &dataP
	default:
		return nil
	}
}

func (sds *sdshdr) GetData() *[]sdsDataType {
	data := *sds.buf
	data = data[1:]
	return &data
}

//fail try
//func GetFlag(buf *[]sdsDataType)*sdsDataType{
//	dataPoint:=&(*buf)[0]
//	tempP:=uintptr(unsafe.Pointer(dataPoint))
//	tempP--
//	sdsType:=(*sdsDataType)(unsafe.Pointer(tempP))
//
//	return sdsType
//}

func (sds *sdshdr) GetType() sdsDataType {
	return (*sds.buf)[0] & 0b00000111
}

func (sds *sdshdr) GetCapacity(sdsType sdsDataType) (int, int) {
	//structPoint := GetStructPoint(sds, sdsType)
	switch sdsType {
	case SDS_TYPE_5:
		sdsLen := int((*sds.buf)[0]&0b11111000) >> SDS_TYPE_LEN
		return sdsLen, sdsLen

	case SDS_TYPE_8:
		sds := (*sdshdr8)(sds.len)
		return int(sds.len), int(sds.alloc)

	case SDS_TYPE_16:
		sds := (*sdshdr16)(sds.len)
		return int(sds.len), int(sds.alloc)

	case SDS_TYPE_32:
		sds := (*sdshdr32)(sds.len)
		return int(sds.len), int(sds.alloc)

	case SDS_TYPE_64:
		sds := (*sdshdr64)(sds.len)
		return int(sds.len), int(sds.alloc)

	default:
		return -1, -1
	}
}

//fail try
//func GetStructPoint(buf *[]sdsDataType,sdsType sdsDataType)unsafe.Pointer{
//	switch sdsType{
//	case SDS_TYPE_5:
//		return unsafe.Pointer(buf)
//	case SDS_TYPE_8:
//		return unsafe.Pointer(uintptr(unsafe.Pointer(buf))-unsafe.Offsetof(sdshdr8{}.buf))
//	case SDS_TYPE_16:
//		return unsafe.Pointer(uintptr(unsafe.Pointer(buf))-unsafe.Offsetof(sdshdr16{}.buf))
//	case SDS_TYPE_32:
//		return unsafe.Pointer(uintptr(unsafe.Pointer(buf))-unsafe.Offsetof(sdshdr32{}.buf))
//	case SDS_TYPE_64:
//		return unsafe.Pointer(uintptr(unsafe.Pointer(buf))-unsafe.Offsetof(sdshdr64{}.buf))
//	default:
//		return nil
//	}
//}

func (sds *sdshdr) SetCapacity(len int, alloc int) {
	sdsType := sds.GetType()
	//structPoint:=GetStructPoint(buf,sdsType)
	switch sdsType {
	case SDS_TYPE_5:
		(*sds.buf)[0] = sdsDataType(len) << 5

	case SDS_TYPE_8:
		sds := (*sdshdr8)(sds.len)
		sds.len = uint8(len)
		sds.alloc = uint8(alloc)
	case SDS_TYPE_16:
		sds := (*sdshdr16)(sds.len)
		sds.len = uint16(len)
		sds.alloc = uint16(alloc)
	case SDS_TYPE_32:
		sds := (*sdshdr32)(sds.len)
		sds.len = uint32(len)
		sds.alloc = uint32(alloc)
	case SDS_TYPE_64:
		sds := (*sdshdr64)(sds.len)
		sds.len = uint64(len)
		sds.alloc = uint64(alloc)
	}
}

func (sds *sdshdr) ConcatSds(aim *[]sdsDataType) *sdshdr {
	/*
		Merge sdshdr & []sdsDataType wrapper function
		return: concat new function
	*/
	sdsType := sds.GetType()
	sdsLen, sdsAlloc := sds.GetCapacity(sdsType)
	sds = sds.ExpandSds(aim, sdsLen, sdsAlloc)

	return sds
}

func (sds *sdshdr) ExpandSds(aim *[]sdsDataType, sdsLen int, sdsAlloc int) *sdshdr {
	/*
		auto choose type to expandSds sds
		rule:
			if all_expand_size==>Type 5==>Auto use Type 8 as new size
				==>all_expand_size=32
			if all_expand_size<1M ==> needSize*2
				else needSize+1M
	*/
	aimLen := len(*aim)
	sdsNeed := sdsAlloc + aimLen
	srcData := sds.GetData()

	if sdsAlloc == -1 {
		sdsNeed = aimLen + 32
	}

	if sdsNeed <= sdsLen {
		copy((*srcData)[sdsAlloc:], (*aim)[:aimLen])
		sds.SetCapacity(sdsLen, sdsNeed)
		return sds
	} else {
		if sdsNeed < SDS_MAX_EXPAND {
			sdsNeed *= 2
		} else {
			sdsNeed += SDS_MAX_EXPAND
		}
		if sdsNeed <= SDS_MAX_TYPE_5 {
			sdsNeed = 32
		}
		expandSds := NewSds(sdsNeed, []sdsDataType{})
		expandLen, _ := expandSds.GetCapacity(expandSds.GetType())
		copy((*expandSds.GetData()), (*srcData)[:sdsAlloc])
		expandSds.SetCapacity(expandLen, sdsAlloc)
		copy((*expandSds.GetData())[sdsAlloc:], (*aim)[:aimLen])
		expandSds.SetCapacity(expandLen, sdsAlloc+aimLen)

		return expandSds
	}

}

func (sds *sdshdr) DelSds() {
	/*
		clear sds,free memory
	*/
	sds.alloc = nil
	sds.buf = nil
	sds.len = nil
	runtime.GC()
}

func (sds *sdshdr) ClearSds() {
	/*
		only clear alloc,don't use to reAlloc memory.
	*/
	sdsType := sds.GetType()
	sdsLen, _ := sds.GetCapacity(sdsType)
	sds.SetCapacity(sdsLen, 0)
}

func main() {
	testData := []sdsDataType{1, 2, 3, 4}
	sds := NewSds(10, testData)
	fmt.Println(sds.GetType())
	fmt.Println(sds.GetCapacity(sds.GetType()))
	sds = sds.ConcatSds(&testData)
	sds.ClearSds()
	fmt.Println(sds.GetCapacity(sds.GetType()))
	sds.DelSds()
	fmt.Println(sds)
}
