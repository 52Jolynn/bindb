package route

import (
	"git.thinkinpower.net/bindb/bdata"
	"git.thinkinpower.net/bindb/mod"
	"github.com/gin-gonic/gin"
	"net/http"
)

func binQuery(ctx *gin.Context) {
	var (
		binData *mod.SimpleBinData
		err     error
	)
	if binData, err = bdata.Query(ctx.Param("bin")); err != nil {
		ctx.JSON(http.StatusOK, mod.ResponseValue{Code: mod.ResponseCodeFailure, Msg: "失败"})
		return
	}
	ctx.JSON(http.StatusOK, mod.ResponseData{ResponseValue: mod.ResponseValue{Code: mod.ResponseCodeSuccess, Msg: "成功 "}, Data: binData})
}
