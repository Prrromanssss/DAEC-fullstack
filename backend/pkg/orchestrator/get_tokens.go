package orchestrator

import (
	"errors"
	"fmt"
	"strings"
)

func GetTokens(parseExpression string) []string {
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

func InsertResultToToken(parseExpression, token string, result int) (string, string, error) {
	ind := 0
	tokens := strings.Split(parseExpression, " ")
	res := ""
	sourceTokens := strings.Split(token, " ")
	newToken := ""
	isTokenFind := false
	if len(tokens) == 3 {
		return fmt.Sprint(result), "", nil
	}
	for ind+2 < len(tokens) {
		if string(tokens[ind]) == sourceTokens[0] &&
			string(tokens[ind+1]) == sourceTokens[1] &&
			string(tokens[ind+2]) == sourceTokens[2] {
			res += fmt.Sprint(result)
			isTokenFind = true
			if ind > 0 && ind+3 >= len(tokens) {
				return "", "", errors.New("invalidate expression")
			}
			if ind > 0 && isNumber(string(tokens[ind-1])) && !isNumber(string(tokens[ind+3])) {
				newToken = string(tokens[ind-1]) + fmt.Sprint(result) + string(tokens[ind+3])
			}
			ind += 2
		} else {
			res += tokens[ind]
		}
		ind++
	}
	if !isTokenFind {
		return "", "", errors.New("can't find token")
	}
	return res, newToken, nil
}
