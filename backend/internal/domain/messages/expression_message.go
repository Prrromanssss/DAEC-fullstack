package messages

type ExpressionMessage struct {
	ExpressionID int32  `json:"expression_id"`
	Token        string `json:"token"`
	Expression   string `json:"expression"`
	Result       int    `json:"result"`
	IsPing       bool   `json:"is_ping"`
	AgentID      int32  `json:"agent_id"`
	UserID       int32  `json:"user_id"`
}

type ResultAndTokenMessage struct {
	Result string `json:"result"`
	Token  string `json:"token"`
}
