package ginutils

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	//"addr.itskoshkin.ru/internal/utils/useragent"
	"link-availability-checker/internal/utils/useragent"
)

const PathWidth = 42

func TwoLinedAccessLog(param gin.LogFormatterParams) string {
	ipAddr, referer := ExtractIPAndReferer(param)

	prefix := fmt.Sprintf("[GIN] %s | %7s ", param.TimeStamp.Format("2006/01/02 - 15:04:05"), param.Method)
	indentLen := len(prefix) + PathWidth + 1
	indent := strings.Repeat(" ", indentLen)

	firstLine := prefix + param.Path + "\n"
	secondLine := fmt.Sprintf("%s| %3d | %10v | %-15s | %s%s\n",
		indent,
		param.StatusCode,
		param.Latency,
		ipAddr,
		useragent.ShortenUserAgent(param.Request.UserAgent()),
		referer,
	)
	return firstLine + secondLine
}
