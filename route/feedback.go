package route

import (
	"git.thinkinpower.net/bindb/data"
	"git.thinkinpower.net/bindb/mod"
	"github.com/gin-gonic/gin"
	"net/http"
)

//bindata feeback
func feedback(ctx *gin.Context) {
	var bindata mod.BinData
	if err := ctx.BindJSON(&bindata); err != nil {
		ctx.JSON(http.StatusOK, mod.ResponseValue{Code: mod.ResponseCodeFailure, Msg: "无法解析request body"})
	}
	data.CreateBinData(ctx.Param("bin"), bindata)
	ctx.JSON(http.StatusOK, mod.ResponseValue{Code: mod.ResponseCodeSuccess, Msg: "成功"})
}

//新增银行中文名称对应关系
func addBankNameCn(ctx *gin.Context) {
	data.CreateBankNameMapping(ctx.Param("key"), ctx.Param("name"))
	ctx.JSON(http.StatusOK, mod.ResponseValue{Code: mod.ResponseCodeSuccess, Msg: "成功"})
}

//新增国家中文名称对应关系
func addCountryCn(ctx *gin.Context) {
	data.CreateCountryCnNameMapping(ctx.Param("key"), ctx.Param("name"))
	ctx.JSON(http.StatusOK, mod.ResponseValue{Code: mod.ResponseCodeSuccess, Msg: "成功"})
}
