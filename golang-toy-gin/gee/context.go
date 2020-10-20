/*=================================
@Author :tix_hjq
@Date   :2020/10/7 下午6:47
@File   :context.go
@email  :hjq1922451756@gmail.com or 1922451756@qq.com
@version:1.15.2
=================================*/

package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ParamInfo map[string]interface{}

/*
	struct(bound=package):
		private:smaller
		public:bigger
*/

type Context struct {
	Write       http.ResponseWriter
	Request     *http.Request
	Path        string
	Method      string
	StatusCode  int
	Params      map[string]string
	HandleFuncS []HandleFunc
	HandleIdx   int
	Engine      *Engine
}

func NewContext(w http.ResponseWriter, r *http.Request) (context *Context) {
	return &Context{
		Write:     w,
		Request:   r,
		Path:      r.URL.Path,
		Method:    r.Method,
		Params:    make(map[string]string),
		HandleIdx: -1,
	}
}

func (context *Context) Query(key string) string {
	/*
		URL.Query?==>
			e.g google search google
				query==[?q=google&oq=google+&aqs=chrome..69i57j0l3j69i60l4.2008j0j4&client=ubuntu&sourceid=chrome&ie=UTF-8]
			similar to param?but URL default have parma,different???I don't know...
		guess:
			param==need to change param or inputStream param
			query==default or Transparent param for user(e.g client:ubuntu,auto get)

		p.s Query source code return type==(dict,error)==>[Get(key)] or [dict,_=Query() dict[key]]
	*/
	return context.Request.URL.Query().Get(key)
}

func (context *Context) PostForm(key string) string {
	return context.Request.FormValue(key)
}

func (context *Context) Status(code int) {
	context.StatusCode = code
	context.Write.WriteHeader(code)
}

func (context *Context) SetHeader(key string, value string) {
	context.Write.Header().Set(key, value)
}

func (context *Context) String(code int, format string, values ...interface{}) {
	context.SetHeader("Content-Type", "text/plain")
	context.Status(code)
	//func Sprintf(format string, a ...interface{}) string {}
	_, err := context.Write.Write([]byte(fmt.Sprintf(format, values...)))
	if err != nil {
		fmt.Println("values Write Error")
	}
}

func (context *Context) Data(code int, data []byte) {
	context.Status(code)
	_, err := context.Write.Write(data)
	if err != nil {
		fmt.Println("Data Write Error")
	}
}

//html data to write
func (context *Context) HTML(code int, name string, data interface{}) {
	//1.set header before write
	context.SetHeader("Context-Type", "text/html")
	context.Status(code)
	//2.write html
	err := context.Engine.htmlTemplates.ExecuteTemplate(context.Write, name, data)
	if err != nil {
		fmt.Println("Html Write Error...")
	}
}

//JSON data to write
func (context *Context) JSON(code int, obj interface{}) {
	//1.set header before write
	context.SetHeader("Content-Type", "application/json")
	context.Status(code)
	//2.json format and write
	encoder := json.NewEncoder(context.Write)
	if err := encoder.Encode(obj); err != nil {
		http.Error(context.Write, err.Error(), 500)
	}
}

func (context *Context) Param(dParams string) string {
	/*
		Method URL ==> to get Param
	*/
	return context.Params[dParams]
}

func (context *Context) Next() {
	/*
		Middle Wares Sequence Control
			Skip now middle wares
		Format:==>to -- ==> format ++
	*/
	context.HandleIdx++
	for ; context.HandleIdx < len(context.HandleFuncS); context.HandleIdx++ {
		context.HandleFuncS[context.HandleIdx](context)
	}
}
