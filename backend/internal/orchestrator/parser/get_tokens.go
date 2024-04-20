package parser

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/Prrromanssss/DAEE-fullstack/internal/domain/messages"
)

// GetTokens gets tokens "<operand> <operand> <operator>" from parseExpression.
func GetTokens(log *slog.Logger, parseExpression string) []string {
	const fn = "parser.GetTokens"

	res := make([]string, 0)
	ind := 0
	tokens := strings.Split(parseExpression, " ")

	log.Info("finding tokens", slog.String("fn", fn), slog.Any("tokens", tokens))

	for ind+2 < len(tokens) {
		if IsNumber(string(tokens[ind])) &&
			IsNumber(string(tokens[ind+1])) &&
			!IsNumber(string(tokens[ind+2])) {
			res = append(res, fmt.Sprint(string(tokens[ind]), " ", string(tokens[ind+1]), " ", string(tokens[ind+2])))
			ind += 2
		}
		ind++
	}

	return res
}

// InsertResultToToken inserts result into parseExpression to the place of the token.
// Returns new token if the insertion was at the beginning of the parseExpression.
func InsertResultToToken(parseExpression, token string, result int) (messages.ResultAndTokenMessage, error) {
	ind := 0
	tokens := strings.Split(parseExpression, " ")
	res := make([]string, 0)
	sourceTokens := strings.Split(token, " ")
	newToken := ""
	isTokenFind := false

	if len(tokens) == 3 {
		return messages.ResultAndTokenMessage{Result: fmt.Sprint(result)}, nil
	}

	for ind+2 < len(tokens) {
		if string(tokens[ind]) == sourceTokens[0] &&
			string(tokens[ind+1]) == sourceTokens[1] &&
			string(tokens[ind+2]) == sourceTokens[2] {

			res = append(res, strconv.Itoa(result))

			isTokenFind = true
			if ind > 0 && ind+3 >= len(tokens) {
				return messages.ResultAndTokenMessage{}, errors.New("invalid expression")
			}
			if ind > 0 && IsNumber(string(tokens[ind-1])) && !IsNumber(string(tokens[ind+3])) {
				newToken = fmt.Sprint(tokens[ind-1], " ", result, " ", tokens[ind+3])
			} else if ind > 0 && ind+4 < len(tokens) && !IsNumber(string(tokens[ind-1])) &&
				IsNumber(string(tokens[ind+3])) &&
				!IsNumber(string(tokens[ind+4])) {
				newToken = fmt.Sprint(result, " ", tokens[ind+3], " ", tokens[ind+4])
			} else if ind == 0 && ind+4 < len(tokens) && IsNumber(string(tokens[ind+3])) &&
				!IsNumber(string(tokens[ind+4])) {
				newToken = fmt.Sprint(result, " ", tokens[ind+3], " ", tokens[ind+4])
			}

			ind += 3
			break
		} else {
			res = append(res, tokens[ind])
		}
		ind++
	}

	for ind < len(tokens) {
		res = append(res, tokens[ind])
		ind++
	}

	if !isTokenFind {
		return messages.ResultAndTokenMessage{}, errors.New("can't find token")
	}

	return messages.ResultAndTokenMessage{
		Result: strings.Join(res, " "),
		Token:  newToken,
	}, nil
}
