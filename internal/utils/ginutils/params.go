package ginutils

import (
	"github.com/gin-gonic/gin"
)

func ExtractIPAndReferer(param gin.LogFormatterParams) (ipAddr string, referer string) {
	realIP := param.Request.Header.Get("X-Real-IP")
	clientIP := param.ClientIP
	if realIP == "" {
		realIP = clientIP
		clientIP = ""
	}
	ipAddr = realIP
	if clientIP != "" && clientIP != realIP {
		ipAddr += " (" + clientIP + ")"
	}
	referer = param.Request.Referer()
	if referer != "" {
		referer = " | Referer \"" + referer + "\""
	}
	return
}
