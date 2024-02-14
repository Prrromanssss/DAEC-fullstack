package orchestrator

import (
	"fmt"
	"strings"
)

func handleParseExpression(parseExpression string) {
	ind := 0
	tokens := strings.Split(parseExpression, " ")
	for ind+2 < len(tokens) {
		if isNumber(string(tokens[ind])) &&
			isNumber(string(tokens[ind+1])) &&
			!isNumber(string(tokens[ind+2])) {
			fmt.Println("Send to RabbitMQ", string(tokens[ind]), string(tokens[ind+1]), string(tokens[ind+2]))
			ind += 2
		}
		ind++
	}
}
