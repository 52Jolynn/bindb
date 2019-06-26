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
	ResponseCodeSuccess = 1001
	ResponseCodeFailure = 1002
)
