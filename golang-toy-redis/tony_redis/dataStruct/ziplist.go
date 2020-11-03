/*=================================
@Author :tix_hjq
@Date   :2020/10/26 下午6:32
@File   :ziplist.go
@email  :hjq1922451756@gmail.com or 1922451756@qq.com
@version:1.15.2
=================== ==============*/

package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"unsafe"
)

const (
	//lasted byte=zLend("same to string final char "\0",sigmoid ZipList lasted")
	//ZIP_MAX_PREVLEN = 254
	//ZIP_ENDLEN = 255
	//
	//ZIP_STRLIST_06B = (0<<6)
	//ZIP_STRLIST_14B = (1<<6)
	//ZIP_STRLIST_32B = (2<<6)

	ZIP_INT_8B  = 0xfe
	ZIP_INT_16B = (0xc0 | 0<<4)
	ZIP_INT_32B = (0xc0 | 1<<4)
	ZIP_INT_64B = (0xc0 | 2<<4)
	ZIP_INT_24B = (0xc0 | 3<<4)
)

type (
	ZipListEntry []byte
	//ZipListEntry struct {
	//	//headerSize uint
	//	//p *ZipListEntry
	//	headerInfo []byte
	//}

	ZipList struct {
		zLBytesLen    uint32
		zLTailShift   uint32
		zLElementSize uint16
		zLTail        *ZipListEntry
		headEntryX    ZipListEntry
	}

	HeaderInfo struct {
		preRawLenSize byte
		preRawLen     []byte
		lenSize       byte
		len           uint
		headerSize    uint
		encoding      byte
		p             *byte
	}
)

func (zle *ZipListEntry) InitHeaderInfo(preRawLenSize byte, preRawLen []byte, encodingSize byte, encoding []byte, value []byte, headSize uint) {
	//init headerInfo
	newHead := make([]byte, headSize)
	newHead[0] = preRawLenSize
	copy(newHead[1:preRawLenSize+1], preRawLen)
	newHead[preRawLenSize+1] = encodingSize
	copy(newHead[preRawLenSize+2:encodingSize+preRawLenSize+2], encoding)
	copy(newHead[preRawLenSize+encodingSize+2:], value[:])
	defer func() { newHead = nil }()

	//copy data
	copy(*zle, newHead)
}

//======================warning=============================
//======================warning=============================
//======================warning=============================

func (zle *ZipListEntry) GetPreLenInfo() (byte, []byte) {
	/*

		preLen=1byte or 5byte,==>[]byte
		need to upgrade function

	*/
	preLenSize := (*zle)[0]
	preLen := (*zle)[1:2]

	if preLenSize != 1 {
		buffer := bytes.NewBuffer((*zle)[1:4])
		if err := binary.Read(buffer, binary.BigEndian, &preLen); err != nil {
			fmt.Println("Read Failed")
		}
	}

	return preLenSize, preLen
}

//======================warning=============================
//======================warning=============================
//======================warning=============================

func (zle *ZipListEntry) GetEncodingInfo() {

}

func (zle *ZipListEntry) GetHeaderInfo() *HeaderInfo {
	preLenSize, preLen := zle.GetPreLenInfo()
	Info := HeaderInfo{preRawLenSize: preLenSize, preRawLen: preLen}
	return &Info
}

//
//func ByteList2Value(byteList []byte)unsafe.Pointer{
//	values:=unsafe.Pointer(nil)
//	buffer:=bytes.NewBuffer(byteList)
//	binary.Read(buffer,binary.BigEndian,)
//	return nil
//}

func (zle *ZipListEntry) NewZipListEntry(preRawLenSize byte, preRawLen []byte, encodingSize byte, encoding []byte, headerSize uint, value []byte) {

	//init Entry && copy headerInfo
	//entry:=ZipListEntry{headerInfo:}
	entry := make(ZipListEntry, headerSize)
	entry.InitHeaderInfo(preRawLenSize, preRawLen, encodingSize, encoding, value, headerSize)

	//copy zle to ZipListPoint
	nowNodeP := (*[unsafe.Sizeof(entry)]byte)(unsafe.Pointer(zle))
	entryP := (*[unsafe.Sizeof(entry)]byte)(unsafe.Pointer(&entry))
	copy(nowNodeP[:], entryP[:])
	zle = (*ZipListEntry)(unsafe.Pointer(nowNodeP))
	defer func() { nowNodeP = nil; entryP = nil }()
}

func JudgeValuesEncoding(encoding byte, valuesLen uint) ([]byte, byte) {
	/*
		param:
			encoding:pre encoding[judge str or value]
			valuesLen:value len[judge list type]
		return:
			encoding bytes
	*/
	var aimEncodingSize byte
	var aimEncoding []byte
	if encoding == 0 {
		if valuesLen <= 63 {
			aimEncoding = make([]byte, 1)
			aimEncoding[0] = byte(valuesLen)
			aimEncodingSize = 1
		} else if float64(valuesLen) <= (math.Pow(2, 14) - 1) {
			aimEncoding = make([]byte, 2)
			buffer := bytes.NewBuffer(aimEncoding)
			if err := binary.Write(buffer, binary.BigEndian, valuesLen+1<<14); err != nil {
				fmt.Println("Read Failed")
			}
			aimEncodingSize = 2
		} else {
			aimEncoding = make([]byte, 5)
			buffer := bytes.NewBuffer(aimEncoding)
			if err := binary.Write(buffer, binary.BigEndian, valuesLen+1<<39); err != nil {
				fmt.Println("Read Failed")
			}
			aimEncodingSize = 5
		}
	} else {
		aimEncoding = make([]byte, 1)
		aimEncoding[0] = encoding
		aimEncodingSize = 1
	}
	//return aimEncoding

	return aimEncoding, aimEncodingSize
}

func TryByteList2Value(values string, valLen uint) ([]byte, byte, uint) {
	val, err := strconv.Atoi(values)
	if err != nil || valLen > 32 || valLen == 0 {
		return nil, 0, valLen
	}
	var encoding byte
	if val >= math.MinInt8 && val <= math.MaxInt8 {
		encoding = ZIP_INT_8B
		valLen = 1
	} else if val >= math.MinInt16 && val <= math.MaxInt16 {
		encoding = ZIP_INT_16B
		valLen = 2
	} else if val >= (-1<<23) && val <= (1<<23-1) {
		encoding = ZIP_INT_24B
		valLen = 3
	} else if val >= math.MinInt32 && val <= math.MaxInt32 {
		encoding = ZIP_INT_32B
		valLen = 4
	} else {
		encoding = ZIP_INT_64B
		valLen = 8
	}

	return FormatVal2ByteList(valLen, val), encoding, valLen
}

func FormatVal2ByteList(valLen uint, val int) []byte {
	buffer := bytes.NewBuffer(make([]byte, valLen))

	switch valLen {
	case 1:
		val := int8(val)
		if err := binary.Write(buffer, binary.BigEndian, val); err != nil {
			fmt.Println("I'm at conv string2Int:", err)
		}
		break

	case 2:
		val := int16(val)
		if err := binary.Write(buffer, binary.BigEndian, val); err != nil {
			fmt.Println("I'm at conv string2Int:", err)
		}
		break

	case 4:
		buffer := bytes.NewBuffer(make([]byte, valLen))
		val := int32(val)
		if err := binary.Write(buffer, binary.BigEndian, val); err != nil {
			fmt.Println("I'm at conv string2Int:", err)
		}
		break
	case 8:
		val := int64(val)
		if err := binary.Write(buffer, binary.BigEndian, val); err != nil {
			fmt.Println("I'm at conv string2Int:", err)
		}
		break
	}

	res := buffer.Bytes()

	return res[len(res)-int(valLen):]
}

func NewZipList() *ZipList {
	zl := ZipList{}
	zl.zLTail = nil
	return &zl
}

func (zl *ZipList) InsertZipListNode(nextNode *ZipListEntry, values string, valuesLen uint) *ZipListEntry {
	/*
		param:
	*/
	var nowNode *ZipListEntry
	var preRawLen []byte
	preRawLenSize := byte(0)

	//-------------------------pre Info----------------------------
	if nextNode != nil {
		//nowNode = (*ZipListEntry)(unsafe.Pointer(uintptr(unsafe.Pointer(zl.zLTail)) + unsafe.Sizeof(*zl.zLTail)))
		nowNode = &ZipListEntry{}
		preRawLenSize, preRawLen = nextNode.GetPreLenInfo()
	} else {
		nowNode = (*ZipListEntry)(unsafe.Pointer(&zl.headEntryX))
		preRawLenSize = 1
		preRawLen = make([]byte, 1)
		preRawLen[0] = 0
	}

	//-------------------------encoding Info----------------------------
	//encoding byte,valuesLen uint
	formatVal, valEncoding, valLen := TryByteList2Value(values, valuesLen)
	if valEncoding != 0 {
		valuesLen = valLen
	} else {
		formatVal = []byte(values)
	}

	encoding, encodingSize := JudgeValuesEncoding(valEncoding, valuesLen)
	headerLen := uint(preRawLenSize) + 1 + uint(encodingSize) + 1 + valuesLen

	//----------------------create entry-----------------------------
	nowNode.NewZipListEntry(preRawLenSize, preRawLen, encodingSize, encoding, headerLen, formatVal)

	//----------------------move ori entry(nextNode)----------------
	if nextNode != nil {
		dst := (*[unsafe.Sizeof(*nextNode)]byte)(unsafe.Pointer(uintptr(unsafe.Pointer(nextNode)) + unsafe.Sizeof(*nowNode)))
		src := (*[unsafe.Sizeof(*nextNode)]byte)(unsafe.Pointer(nextNode))
		copy(dst[:], src[:])
	}

	//----------------------merge entry------------------------------
	src := (*[unsafe.Sizeof(*nowNode)]byte)(unsafe.Pointer(nowNode))
	if zl.zLTail == nil {
		dst := (*[unsafe.Sizeof(*nowNode)]byte)(unsafe.Pointer(&zl.headEntryX))
		copy(dst[:], src[:])
		zl.zLTail = nowNode
	} else {
		dst := (*[unsafe.Sizeof(*nowNode)]byte)(unsafe.Pointer(nextNode))
		copy(dst[:], src[:])
	}

	/*
		format header:
			because it's nextNode address not change,so header is not to format
	*/

	//---------------------format tail-----------------------------
	if zl.zLTail == nextNode {
		zl.zLTail = (*ZipListEntry)(unsafe.Pointer(uintptr(unsafe.Pointer(nextNode)) + unsafe.Sizeof(*nowNode)))
	}

	//---------------------update zlList---------------------------
	zl.zLBytesLen += uint32(unsafe.Sizeof(*nowNode))
	zl.zLElementSize += 1

	//---------------------update nextNode---------------------------
	//preLen:=[]byte{0}
	//if headerLen<255{
	//	preRawLenSize=1
	//
	//}else{
	//	preRawLenSize=5
	//	buffer:=bytes.NewBuffer([]byte{})
	//	if err:=binary.Write(buffer,binary.BigEndian,headerLen);err!=nil{
	//		fmt.Println(err)
	//	}
	//	copy(preLen[:],buffer.Bytes()[:])
	//}
	//
	//aim:=byte(zl.zLTail)-byte(&zl.headEntryX)
	//fmt.Println()

	return nowNode
}

/*
	ZipList struct {
		zLBytesLen uint32
		zLTailShift uint32
		zLElementSize uint16
		zLTail *ZipListEntry
		headEntryX ZipListEntry
	}
*/

func main() {
	test := "1234"
	test1 := "7889"
	zl := NewZipList()
	node1 := zl.InsertZipListNode(zl.zLTail, test, uint(len(test)))
	fmt.Println(zl)
	zl.InsertZipListNode(node1, test1, uint(len(test)))
	fmt.Println(zl)
	fmt.Println(zl.zLTail)
}
