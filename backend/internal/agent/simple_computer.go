package agent

import (
	"time"

	"github.com/Prrromanssss/DAEC-fullstack/internal/domain/messages"
)

// simpleComputer calculates a simple expression consisting of 2 operands.
func simpleComputer(
	exprMsg *messages.ExpressionMessage,
	digit1, digit2 int,
	oper string,
	timer *time.Timer,
	res chan<- *messages.ExpressionMessage,
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
