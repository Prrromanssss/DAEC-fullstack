package agent

import "time"

func simpleComputer(
	exprMsg *ExpressionMessage,
	digit1, digit2 int,
	oper string,
	timer *time.Timer,
	res chan<- *ExpressionMessage,
) {
	switch {
	case oper == "+":
		<-timer.C
		exprMsg.Result = digit1 + digit2
		res <- exprMsg
	case oper == "-":
		<-timer.C
		exprMsg.Result = digit1 - digit2
		res <- exprMsg
	case oper == "/":
		<-timer.C
		exprMsg.Result = digit1 / digit2
		res <- exprMsg
	case oper == "*":
		<-timer.C
		exprMsg.Result = digit1 * digit2
		res <- exprMsg
	}
}
