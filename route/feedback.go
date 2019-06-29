package route

import (
	"kidshelloworld.com/bindb/bdata"
	"kidshelloworld.com/bindb/mod"
	"github.com/gin-gonic/gin"
	logger "github.com/sirupsen/logrus"
	"net/http"
)

//bindata feeback approximate, not sure
func feedback(ctx *gin.Context) {
	var bindata mod.BinData
	if err := ctx.BindJSON(&bindata); err != nil {
		logger.Error(err)
		ctx.JSON(http.StatusOK, mod.ResponseValue{Code: mod.ResponseCodeFailure, Msg: "无法解析request body"})
	}
	//判断入参是否合法
	if !verifyBinData(bindata) {
		ctx.JSON(http.StatusOK, mod.ResponseValue{Code: mod.ResponseCodeInvalidParams, Msg: "非法参数"})
		return
	}
	if err := bdata.CreateBinData(ctx.Param("bin"), bindata, true); err != nil {
		logger.Error(err)
		ctx.JSON(http.StatusOK, mod.ResponseValue{Code: mod.ResponseCodeFailure, Msg: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, mod.ResponseValue{Code: mod.ResponseCodeSuccess, Msg: "成功"})
}

//bindata feeback, sure
func feedback_t(ctx *gin.Context) {
	var bindata mod.BinData
	if err := ctx.BindJSON(&bindata); err != nil {
		logger.Error(err)
		ctx.JSON(http.StatusOK, mod.ResponseValue{Code: mod.ResponseCodeFailure, Msg: "无法解析request body"})
	}
	//判断入参是否合法
	if !verifyBinData(bindata) {
		ctx.JSON(http.StatusOK, mod.ResponseValue{Code: mod.ResponseCodeInvalidParams, Msg: "非法参数"})
		return
	}
	if err := bdata.CreateBinData(ctx.Param("bin"), bindata, false); err != nil {
		logger.Error(err)
		ctx.JSON(http.StatusOK, mod.ResponseValue{Code: mod.ResponseCodeFailure, Msg: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, mod.ResponseValue{Code: mod.ResponseCodeSuccess, Msg: "成功"})
}

func verifyBinData(binData mod.BinData) bool {
	if binData.BankName == "" || binData.CardType == "" {
		return false
	}
	return true
}

//新增银行中文名称对应关系
func addBankNameCn(ctx *gin.Context) {
	if err := bdata.CreateBankNameMapping(ctx.Param("key"), ctx.Param("name")); err != nil {
		ctx.JSON(http.StatusOK, mod.ResponseValue{Code: mod.ResponseCodeFailure, Msg: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, mod.ResponseValue{Code: mod.ResponseCodeSuccess, Msg: "成功"})
}

//新增国家中文名称对应关系
func addCountryCn(ctx *gin.Context) {
	if err := bdata.CreateCountryCnNameMapping(ctx.Param("key"), ctx.Param("name")); err != nil {
		ctx.JSON(http.StatusOK, mod.ResponseValue{Code: mod.ResponseCodeFailure, Msg: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, mod.ResponseValue{Code: mod.ResponseCodeSuccess, Msg: "成功"})
}
