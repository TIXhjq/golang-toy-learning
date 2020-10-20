package gee

import (
	"fmt"
	"strings"
)

type HandleFunc func(ctx *Context)

//func(http.ResponseWriter,*http.Request)

type Router struct {
	/*
		routerMaps==dict{} ==>format(key-values)
			key(user use index)=method:[e.g get,post]+func_saveDir:[/hello]
			value(user to use index==>function)
		routerMaps to save all {dir&func},when user to use index==>display aimFunc to user
	*/
	Router map[string]HandleFunc //key-val save handle
	Tree   map[string]*Trie      //prefix tree for match url
}

func NewRouter() *Router {
	//__init__
	return &Router{Router: make(map[string]HandleFunc), Tree: make(map[string]*Trie)}
}

func (router *Router) GetRouterMap() {
	//output router map
	for k, v := range router.Router {
		fmt.Printf("method:%s,path:%s", k, v)
	}
}

func (router *Router) AddRoute(method string, pattern string, handleFunc HandleFunc) {
	/*
		format key-value & add to router
						 & add to SearchMap

	*/

	key := method + "-" + pattern
	router.Router[key] = handleFunc
	if router.Tree[method] == nil {
		router.Tree[method] = NewTrie()
	}
	formatPattern := router.FormatRouterPath(pattern)
	router.Tree[method].Insert(formatPattern)
}

func (router *Router) ParseURL(context *Context) map[string]string {
	/*

		get param from tree
			format:{":param":"dfd"},{"*filepath":"dfd/Dfd"}

	*/
	rURL, method := context.Request.URL.Path, context.Method
	patternInput := router.FormatRouterPath(rURL)

	return router.Tree[method].SearchMatch(patternInput)
}

func (router *Router) FormatRouterPath(routerPath string) []string {
	/*
		format pattern <string==>list>
			same to string.split("/")
	*/
	return strings.Split(routerPath, "/")
}

func (router *Router) Handler(context *Context, patternPath string) {
	/*
		handle router handles &middle prepare
	*/

	if handler, ok := router.Router[patternPath]; ok {
		context.HandleFuncS = append(context.HandleFuncS, handler)
	} else {
		_, err := fmt.Println("404 not found")
		if err != nil {
			fmt.Println("Output Page Error")
		}
	}
	context.Next()
}

func (router *Router) RouterHandleWrapper(context *Context) {
	/*

		Handle Wrapper:Handle data format input
		key==[method(post,..)+path]
		params:Format Match Params to list from params

	*/
	parseRes := router.ParseURL(context)
	//i:=0
	if len(parseRes) == 0 {
		patternPath := context.Method + "-" + context.Path
		router.Handler(context, patternPath)

	} else {
		for k, v := range parseRes {
			resKey := strings.Split(k, " ")
			context.Params[resKey[0]] = v
			patternPath := context.Method + "-" + resKey[1]
			router.Handler(context, patternPath)
		}

	}
}
