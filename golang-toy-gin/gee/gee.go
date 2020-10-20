/*=================================
@Author :tix_hjq
@Date   :2020/10/7 下午6:47
@File   :gee.go
@email  :hjq1922451756@gmail.com or 1922451756@qq.com
@version:1.15.2
=================================*/
package gee

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
)

type (
	Engine struct {
		*RouterGroup  //father RouterGroup
		Router        *Router
		Groups        []*RouterGroup
		htmlTemplates *template.Template
		funcMap       template.FuncMap
	}

	RouterGroup struct {
		Prefix      string
		MiddleWares []HandleFunc
		Parent      *RouterGroup
		Engine      *Engine
	}
)

func NewEngine() *Engine {
	router := NewRouter()
	engine := Engine{Router: router}
	engine.RouterGroup = &RouterGroup{Engine: &engine}
	engine.Groups = []*RouterGroup{engine.RouterGroup}
	return &engine
}

func (group *RouterGroup) Group(prefix string) *RouterGroup {
	/*
		function:{
			control routerGroup[e.g prefix control,add router to right group]
		}
	*/
	engine := group.Engine
	newGroup := RouterGroup{
		Engine: engine,
		Prefix: group.Prefix + prefix,
		Parent: group,
	}
	engine.Groups = append(engine.Groups, &newGroup)

	return &newGroup
}

func (group *RouterGroup) AddRoute(method string, withinGroupPath string, handleFunc HandleFunc) {
	pattern := group.Prefix + withinGroupPath
	log.Printf("Add Routers:%s\n", pattern)
	group.Engine.Router.AddRoute(method, pattern, handleFunc)
}

func (group *RouterGroup) GET(pattern string, handleFunc HandleFunc) {
	//get func format{get+dir:aim_func}
	group.AddRoute("GET", pattern, handleFunc)
}

func (group *RouterGroup) POST(pattern string, handleFunc HandleFunc) {
	//post func format{post+dir:aim_func}
	group.AddRoute("POST", pattern, handleFunc)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	/*

		why this function exist?
			to listen:==>http.ListenAndServe(string,Handler)
				Handler?[type Handler interface {ServeHTTP(ResponseWriter, *Request)}]
			achieve ServeHTTP to change struct[router]==>interface[Handler]
		this function is used to judge:
			runFunction if (userUseIndex exist router map) else Error

		function:
			input : wrapper data
			return: create handle conn

	*/
	context := NewContext(w, r)
	context.Engine = engine
	context.HandleFuncS = engine.MiddleWrapper(r)
	engine.Router.RouterHandleWrapper(context)
}

func (group *RouterGroup) AddMiddleWares(handleFunc ...HandleFunc) {
	group.MiddleWares = append(group.MiddleWares, handleFunc...)
}

func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

func (engine *Engine) LoadHTMLGlob(pattern string) {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}

func (engine *Engine) MiddleWrapper(r *http.Request) []HandleFunc {
	/*
		format group router to choose middleWares input
		final==>MiddleHandles to for Context.HandlesFuncS
	*/
	middleWares := []HandleFunc{}
	for group_idx := range engine.Groups {
		if strings.HasPrefix(r.URL.Path, engine.Groups[group_idx].Prefix) {
			middleWares = append(middleWares, engine.Groups[group_idx].MiddleWares...)
		}
	}

	return middleWares
}

func (group *RouterGroup) CreateStaticHandler(relativePath string, fs http.FileSystem) HandleFunc {
	absolutePath := path.Join(group.Prefix, relativePath)
	/*
		http.StripPrefix remove URL path prefix & use root path to replace it
	*/
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))

	return func(c *Context) {
		file := c.Param("filepath")
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Write, c.Request)
	}
}

func (group *RouterGroup) Static(relativePath string, root string) {
	//http.Dir==[path Open interface wrapper]
	handler := group.CreateStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	group.GET(urlPattern, handler)
}

func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}
