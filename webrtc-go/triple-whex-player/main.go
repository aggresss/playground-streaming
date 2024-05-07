package main

import (
	"time"

	"github.com/gin-gonic/gin"
)

var epoch = time.Unix(0, 0).Format(time.RFC1123)

var noCacheHeaders = map[string]string{
	"Expires":         epoch,
	"Cache-Control":   "no-cache, private, max-age=0",
	"Pragma":          "no-cache",
	"X-Accel-Expires": "0",
}

var etagHeaders = []string{
	"ETag",
	"If-Modified-Since",
	"If-Match",
	"If-None-Match",
	"If-Range",
	"If-Unmodified-Since",
}

func noCache() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		for _, v := range etagHeaders {
			if ctx.Request.Header.Get(v) != "" {
				ctx.Request.Header.Del(v)
			}
		}
		for k, v := range noCacheHeaders {
			ctx.Writer.Header().Set(k, v)
		}
	}
}

func main() {
	router := gin.Default()
	router.Use(noCache())
	router.Static("/", "./examples")
	router.Run(":18080")
}
