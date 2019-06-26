package route

import (
	"git.thinkinpower.net/bindb/data"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func Register(r *gin.Engine) {
	g := r.Group("/bindb")
	{
		g.GET("/index", func(context *gin.Context) {
			context.String(http.StatusOK, "Hello bindb, date: %s", time.Now().Format(data.DateTimePattern))
		})

		g.GET("/query/:bin", binQuery)
		g.POST("/feedback/:bin", feedback)
		g.POST("/feedback/bank_name/:key/:name", addBankNameCn)
		g.POST("/feedback/country/:key/:name", addCountryCn)
	}
}
