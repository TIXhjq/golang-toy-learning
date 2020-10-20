/*=================================
@Author :tix_hjq
@Date   :2020/10/11 下午5:22
@File   :dynamicMethod.go
@email  :hjq1922451756@gmail.com or 1922451756@qq.com
@version:1.15.2
=================================*/
package gee

import (
	"strings"
)

type Info struct {
	mod     int
	oriNode string
}

type Trie struct {
	/*
			router search to ==> format router path ,to reduce dynamic search time
			p.s
				1.baseMethod:[Trie]
		       	2.only support single pattern to search
					format:
						[e.g {/p/*${name}/doc}==True {/p/*${name}/*{name}==False]
	*/

	tree     map[string]*Trie
	nodeInfo map[string]*Info
}

/** Initialize your data structure here. */
func NewTrie() *Trie {
	return &Trie{make(map[string]*Trie), make(map[string]*Info)}
}

func NewInfo(mod int, formatNode string) *Info {
	return &Info{mod, formatNode}
}

func (this *Trie) FormatRouter(router string, temp *Trie) (string, int) {
	mod := -1
	if router != "" {
		mod = this.CheckPattern(router)
		if mod != -1 {
			router = router[1:]
		}
	}
	return router, mod
}

/** Inserts a word into the trie. */
func (this *Trie) Insert(word []string) {
	temp := this
	keyword := ""
	mod := -1

	for i := range word {
		if keyword, mod = this.FormatRouter(word[i], temp); keyword == "" {
			continue
		}
		if keyword == "" {
			continue
		}
		if temp.tree[keyword] == nil {
			temp.tree[keyword] = &Trie{make(map[string]*Trie), make(map[string]*Info)}
			temp.nodeInfo[keyword] = NewInfo(mod, word[i])
		}
		temp = temp.tree[keyword]
	}
	temp.tree["NULL"] = &Trie{make(map[string]*Trie), make(map[string]*Info)}
}

func (this *Trie) FindTail(temp map[string]*Trie) string {
	res := []string{}
	for temp["NULL"] == nil {
		for k, v := range temp {
			temp = v.tree
			res = append(res, k)
		}
	}
	return strings.Join(res, "/")
}

func (this *Trie) SearchMatch(routerPattern []string) map[string]string {
	temp := this
	keyword := ""
	res := make(map[string]string)

	for idx_ := range routerPattern {
		keyword = routerPattern[idx_]
		if keyword == "" {
			continue
		}

		//judge node exist
		for node, _ := range temp.tree {
			switch temp.nodeInfo[node].mod {
			case 1:
				if this.AllMatch(temp.tree[node].tree, routerPattern[idx_+1:]) == true {
					patternKey := strings.Join(routerPattern[:idx_], "/") + "/:" + node + "/" + strings.Join(routerPattern[idx_+1:], " ")
					res[node+" "+patternKey] = keyword
				}
			case 2:
				patternKey := strings.Join(routerPattern[:idx_], "/") + "/*" + node + "/" + this.FindTail(temp.tree[node].tree)
				res[node+" "+patternKey] = strings.Join(routerPattern[idx_:], "/")
			}
		}
		if temp.tree[keyword] == nil {
			break
		}
		temp = temp.tree[keyword]
	}

	return res
}

func (this *Trie) AllMatch(routerList map[string]*Trie, pattern []string) bool {
	/*
		format:mod[":"],a/:b/c,a/d/c,a/m
				==>>{b:d} not have {b:m}
	*/
	temp := routerList
	for idx_ := range pattern {
		if temp[pattern[idx_]] == nil {
			return false
			//if temp["NULL"]==nil {
			//	return false
			//}
			//break
		}
		temp = temp[pattern[idx_]].tree
	}
	if len(pattern) == 0 {
		return false
	}
	return true
}

func (this *Trie) CheckPattern(pattern string) int {
	/*
		func:judge pattern mod
		return: mod int
			mod:
				case 1: : [a,pattern,b]==[a,:,b] if b!=b NG
				case 2: * [a,pattern,b]==[a,*] pattern
	*/
	switch pattern[0] {
	case ':':
		return 1
	case '*':
		return 2
	default:
		return -1
	}
}

func (this *Trie) GetSonRouter(routerNode map[string]*Trie, patternRouter string) []string {
	res := []string{}
	for k, _ := range routerNode {
		if k != patternRouter {
			res = append(res, k)
		}
	}

	return res
}
