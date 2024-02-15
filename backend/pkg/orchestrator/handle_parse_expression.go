package orchestrator

import (
	"fmt"
	"strings"
)

func HandleParseExpression(parseExpression string) []string {
	res := make([]string, 0)
	ind := 0
	tokens := strings.Split(parseExpression, " ")
	for ind+2 < len(tokens) {
		if isNumber(string(tokens[ind])) &&
			isNumber(string(tokens[ind+1])) &&
			!isNumber(string(tokens[ind+2])) {
			res = append(res, fmt.Sprint(string(tokens[ind]), string(tokens[ind+1]), string(tokens[ind+2])))
			ind += 2
		}
		ind++
	}
	return res
}
