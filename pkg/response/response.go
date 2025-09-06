package response

type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func OK(data interface{}) Result {
	return Result{Code: 0, Msg: "ok", Data: data}
}

func Err(msg string, code int) Result {
	return Result{Code: code, Msg: msg}
}
