/*=================================
@Author :tix_hjq
@Date   :2020/10/14 下午6:23
@File   :logger.go
@email  :hjq1922451756@gmail.com or 1922451756@qq.com
@version:1.15.2
=================================*/

package gee

import (
	"log"
	"time"
)

func Logger() HandleFunc {
	return func(c *Context) {
		t := time.Now()
		c.Next()
		log.Printf("[%s],%s,%s", time.Since(t), c.StatusCode, c.Request.RequestURI)
	}
}
