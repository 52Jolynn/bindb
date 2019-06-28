package mod

type ResponseValue struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type ResponseData struct {
	ResponseValue
	Data interface{}
}

const (
	//成功
	ResponseCodeSuccess = 1001
	//失败
	ResponseCodeFailure = 1002
	//缺少参数
	ResponseCodeMissingParams = 1003
	//非法参数
	ResponseCodeInvalidParams = 1004
	//数据不存在
	ResponseCodeNotFound = 1010
)
