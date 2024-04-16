package messages

type MessageFromOrchestrator struct {
	ExpressionID int32  `json:"expression_id"`
	Expression   string `json:"expression"`
}
