package routers

import (
	"guoshi/routers/base"
	"guoshi/routers/v1"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

type PortalRouter struct {
	base.BaseRouter
}

var (
	portal *PortalRouter
)

func init() {
	portal = &PortalRouter{}
	portal.Register("", NewPingRouter())
	portal.Register("/guoshi/api/v1", v1.NewV1Router())
}

func Load(engine *gin.Engine) {
	engine.Use(loggerMiddleware)

	engine.Use(CORSMiddleware())
	engine.Use(parseForm)
	portal.Load(engine.Group("/"))
}

func loggerMiddleware(c *gin.Context) {
	// Start timer
	start := time.Now()
	url := c.Request.URL

	// Process request
	c.Next()

	// Stop timer
	end := time.Now()
	latency := end.Sub(start)

	clientIP := c.ClientIP()
	method := c.Request.Method
	statusCode := c.Writer.Status()

	log.Printf("[GIN] %v | %3d | %13v | %15s | %s %s",
		end.Format("2006/01/02 - 15:04:05"),
		statusCode,
		latency,
		clientIP,
		method,
		url,
	)
}

func CORSMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("origin")
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, PATCH, OPTIONS, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, XMLHttpRequest, "+
			"Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.String(200, "ok")
			return
		}
		c.Next()
	}
}

func parseForm(c *gin.Context) {
	c.Request.ParseForm()
	c.Next()
}
