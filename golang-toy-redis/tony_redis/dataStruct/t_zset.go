/*=================================
@Author :tix_hjq
@Date   :2020/10/21 上午8:46
@File   :t_zset.go
@email  :hjq1922451756@gmail.com or 1922451756@qq.com
@version:1.15.2
=================================*/

package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"time"
)

//var LevelRandomState = time.Now().UnixNano()

const (
	ZSKPLIST_MAXLEVEL = 64
	ZSKPLIST_P        = 0.25
)

type (
	ZSKipListLevel struct {
		forward *ZSkipListNode
		span    uint
	}

	ZSkipListNode struct {
		element  *[]byte
		score    float32
		backward *ZSkipListNode
		level    []ZSKipListLevel
	}

	ZSkipList struct {
		head   *ZSkipListNode
		tail   *ZSkipListNode
		length uint
		level  int
	}
)

func ZSLRandomLevel() int {
	/*
		ZSKPLIST_P[default=0.25]
		while random<ZSKPLIST_P
			level++
		level=1:(1-p)
		level=2:p(1-p)
		level=n:p^n-1(1-p)

		EX=(1-p)+2p(1-p)+...np^n-1(1-p)==>1/(1-p)
			==>1/(0.75)~=1.3
	*/
	level := 1
	rand.Seed(time.Now().UnixNano())
	p_ := rand.Float64()
	for p_ < ZSKPLIST_P {
		//rand.Seed(time.Now().UnixNano())
		p_ = rand.Float64()
		level++
	}
	if level < ZSKPLIST_MAXLEVEL {
		return level
	} else {
		return ZSKPLIST_MAXLEVEL
	}
}

func NewZSKipListNode(level int, score float32, element *[]byte) *ZSkipListNode {
	return &ZSkipListNode{level: make([]ZSKipListLevel, level), score: score, element: element}
}

func NewZSkipList() *ZSkipList {
	/*
		init zSkipList head.
	*/
	zsl := ZSkipList{
		length: 0, level: 1,
		head: NewZSKipListNode(ZSKPLIST_MAXLEVEL, 0, nil),
		tail: nil}
	for idx := 0; idx < ZSKPLIST_MAXLEVEL; idx++ {
		zsl.head.level[idx].forward = nil
		zsl.head.level[idx].span = 0
	}
	zsl.head.backward = nil

	return &zsl
}

func (zsk *ZSkipList) FindEachLevelInfo(score float32, element *[]byte) (*[]uint, *[]*ZSkipListNode) {
	/*
		function:find need to insert position
		init params:
			preNode: need insert preNode
			disHeadStack[levelIdx]: aimNode's preNode head to now dis
			preNodeStack[levelIdx]: aimNode's preNode
	*/
	preNode := zsk.head
	disHeadStack := make([]uint, ZSKPLIST_MAXLEVEL)
	preNodeStack := make([]*ZSkipListNode, ZSKPLIST_MAXLEVEL)

	for levelIdx := zsk.level - 1; levelIdx >= 0; levelIdx-- {
		//init dis[
		//if initLevel ==> 0,because at ori point
		//	else ==> preLevel]
		if levelIdx == zsk.level-1 {
			disHeadStack[levelIdx] = 0
		} else {

			disHeadStack[levelIdx] = disHeadStack[levelIdx+1]
		}
		//judge forward==None,forward.score>score,if score==forward preNode>aimNode
		for (preNode.level[levelIdx].forward != nil) && (preNode.level[levelIdx].forward.score < score || ((preNode.level[levelIdx].forward.score < score) &&
			(bytes.Compare(*preNode.level[levelIdx].forward.element, *element) < 0))) {
			disHeadStack[levelIdx] += preNode.level[levelIdx].span
			preNode = preNode.level[levelIdx].forward
		}
		preNodeStack[levelIdx] = preNode
	}

	return &disHeadStack, &preNodeStack
}

func (zsk *ZSkipList) FormatNewLevel(disHeadStack *[]uint, preNodeStack *[]*ZSkipListNode) (int, *[]uint, *[]*ZSkipListNode) {
	aimLevel := ZSLRandomLevel()
	//build new level[new node level>maxLevel]
	if aimLevel > zsk.level {
		for levelIdx := zsk.level - 1; levelIdx < aimLevel; levelIdx++ {
			(*disHeadStack)[levelIdx] = 0
			(*preNodeStack)[levelIdx] = zsk.head
			//forward==>None length == zsk.length
			(*preNodeStack)[levelIdx].level[levelIdx].span = zsk.length
		}
		zsk.level = aimLevel
	}

	return aimLevel, disHeadStack, preNodeStack
}

func (zsk *ZSkipList) MergeNode(aimLevel int, aimNode *ZSkipListNode, preNodeStack *[]*ZSkipListNode, disHeadStack *[]uint) {
	/*
		merge node,pre&next&tail
		pre: (back==head)?==>nil
	*/

	for levelIdx := 0; levelIdx < aimLevel; levelIdx++ {
		//aim==>preNode.forward
		aimNode.level[levelIdx].forward = (*preNodeStack)[levelIdx].level[levelIdx].forward
		//preNode.forward==>aim
		(*preNodeStack)[levelIdx].level[levelIdx].forward = aimNode
		/*
			level 1  nil  preNode nil      aim  nil {disHead[level1]=dis[head,preNode]}
			level 0  node node    preNode  aim  nil {disHead[level0]=dis[head,preNode]}
				dis[preNode,aim]=disHead[0]-disHead[1]
				dis[aim,preNodeForward]=preNode.forward.span-dis[preNode,aim]
				preNode.span=dis[inFact(at 0 level dis)+1(insertNode)]
		*/
		aimNode.level[levelIdx].span = (*preNodeStack)[levelIdx].level[levelIdx].span - ((*disHeadStack)[0] - (*disHeadStack)[levelIdx])
		(*preNodeStack)[levelIdx].level[levelIdx].span = (*disHeadStack)[0] - (*disHeadStack)[levelIdx] + 1
	}

	//without insert high node,span++
	for levelIdx := aimLevel; levelIdx < zsk.level; levelIdx++ {
		(*preNodeStack)[levelIdx].level[levelIdx].span++
	}

	//pre
	if (*preNodeStack)[0] != zsk.head {
		aimNode.backward = (*preNodeStack)[0]
	} else {
		aimNode.backward = nil
	}

	//tail
	if aimNode.level[0].forward != nil {
		aimNode.level[0].forward.backward = aimNode
	} else {
		zsk.tail = aimNode
	}

	zsk.length++
}

func (zsk *ZSkipList) InsertZSkipList(score float32, element *[]byte) *ZSkipListNode {
	disHeadStack, preNodeStack := zsk.FindEachLevelInfo(score, element)
	aimLevel, disHeadStack, preNodeStack := zsk.FormatNewLevel(disHeadStack, preNodeStack)
	aimNode := NewZSKipListNode(aimLevel, score, element)
	zsk.MergeNode(aimLevel, aimNode, preNodeStack, disHeadStack)

	return aimNode
}

func (zsk *ZSkipList) FindZSKipNode(score float32, element *[]byte) (*ZSkipListNode, *[]*ZSkipListNode, error) {
	_, preNodeStack := zsk.FindEachLevelInfo(score, element)

	aimNode := (*preNodeStack)[0].level[0].forward
	if aimNode != nil && aimNode.score == score && bytes.Compare(*(aimNode.element), *element) == 0 {
		return aimNode, preNodeStack, nil
	} else {
		return nil, nil, fmt.Errorf("Not exist node in SkipList")
	}
}

func (zsk *ZSkipList) DeleteZSkipNode(aimNode *ZSkipListNode, preNodeStack *[]*ZSkipListNode, err error) error {
	if err != nil {
		return err
	}

	// preNode-->aimNode-->forward,preNode-->forward
	for levelIdx := 0; levelIdx < zsk.level; levelIdx++ {
		if (*preNodeStack)[levelIdx].level[levelIdx].forward == aimNode {
			(*preNodeStack)[levelIdx].level[levelIdx].forward = aimNode.level[levelIdx].forward
			(*preNodeStack)[levelIdx].level[levelIdx].span += aimNode.level[levelIdx].span
		} else {
			(*preNodeStack)[levelIdx].level[levelIdx].span--
		}
	}

	//tail
	if (*preNodeStack)[0].level[0].forward == nil {
		(*preNodeStack)[0].level[0].forward.backward = (*preNodeStack)[0]
	} else {
		zsk.tail = (*preNodeStack)[0]
	}

	//check level,zsk.level>1,not exist levelNode,zsk.level--
	for zsk.level > 1 && zsk.head.level[zsk.level-1].forward == nil {
		zsk.level--
	}
	zsk.length--

	return nil
}

func (zsk *ZSkipList) FreeZSkNode(aimNode *ZSkipListNode) {
	aimNode.backward = nil
	aimNode.element = nil

	for levelIdx := 0; levelIdx < len(aimNode.level); levelIdx++ {
		aimNode.level[levelIdx].forward = nil
	}
	aimNode.level = nil
}

func DeleteZSkList(zsk **ZSkipList) {
	/*
		destroy ZSkipList
	*/
	zSkNode := (*zsk).head
	for length := uint(0); length < (*zsk).length; length++ {
		nextNode := zSkNode.level[0].forward
		(*zsk).FreeZSkNode(zSkNode)
		zSkNode = nextNode
	}
	(*zsk).head = nil
	(*zsk).tail = nil
	(*zsk) = nil

	//return zsk
}

func main() {
	zsk := NewZSkipList()
	a := []byte{1, 3, 4}
	zsk.InsertZSkipList(20, &a)
	zsk.InsertZSkipList(13, &a)
	zsk.InsertZSkipList(12, &a)
	zsk.InsertZSkipList(9, &a)
	if err := zsk.DeleteZSkipNode(zsk.FindZSKipNode(12, &a)); err != nil {
		fmt.Println(err)
	}
	if err := zsk.DeleteZSkipNode(zsk.FindZSKipNode(12, &a)); err != nil {
		fmt.Println(err)
	}
	DeleteZSkList(&zsk)
	fmt.Println(zsk)
}
