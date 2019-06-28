package route

import (
	"git.thinkinpower.net/bindb/bdata"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func Register(r *gin.Engine) {
	g := r.Group("/bindb")
	{
		g.GET("/index", func(context *gin.Context) {
			context.String(http.StatusOK, "Hello bindb, date: %s", time.Now().Format(bdata.DateTimePattern))
		})

		v1 := g.Group("/v1")

		v1.GET("/bin/query/:bin", binQuery)
		v1.POST("/bin/feedback/:bin", feedback)
		v1.POST("/bin_t/feedback/:bin", feedback_t)
		v1.POST("/bank/feedback/:key/:name", addBankNameCn)
		v1.POST("/country/feedback/:key/:name", addCountryCn)
	}
}
