package model

const SuccessMsgType = 1
const ErrorMsgType = 0
const InfoMsgType = 2

// Msg websocket消息
type Msg struct {
	Type  int    `json:"type"`  // 消息类型
	Title string `json:"title"` // 消息标题
	Msg   string `json:"msg"`   // 消息内容
}

// NewMsg 普通消息
func NewMsg(title, msg string) *Msg {
	return &Msg{Type: InfoMsgType, Title: title, Msg: msg}
}

// SuccessMsg 成功消息
func SuccessMsg(title, msg string) *Msg {
	return &Msg{Type: SuccessMsgType, Title: title, Msg: msg}
}

// ErrMsg 错误消息
func ErrMsg(title, msg string) *Msg {
	return &Msg{Type: ErrorMsgType, Title: title, Msg: msg}
}
