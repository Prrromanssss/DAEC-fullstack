package orchestrator

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

func ParseExpression(expression string) (string, error) {
	rawExpression := strings.ReplaceAll(expression, " ", "")
	if !isValidExpression(rawExpression) {
		return "", errors.New("invalid expression")
	}
	rawExpression = addBrackets(rawExpression)
	result, err := infixToPostfix(rawExpression)
	if err != nil {
		return "", err
	}
	return result, nil
}

func infixToPostfix(expression string) (string, error) {
	var output strings.Builder
	var stack []rune
	for _, char := range expression {
		switch char {
		case '(':
			stack = append(stack, char)
		case ')':
			err := popUntilOpeningParenthesis(&stack, &output)
			if err != nil {
				return "", err
			}
		case '+', '-', '*', '/':
			popOperatorsWithHigherPrecedence(char, &stack, &output)
			stack = append(stack, char)
			output.WriteRune(' ')
		default:
			output.WriteRune(char)
		}
	}

	for len(stack) > 0 {
		popTopOperator(&stack, &output)
	}

	return strings.ReplaceAll(strings.TrimSpace(output.String()), "  ", " "), nil
}

func contains(arr []rune, element rune) bool {
	for _, elem := range arr {
		if elem == element {
			return true
		}
	}
	return false
}

func isValidExpression(expression string) bool {
	stack := make([]rune, 0)

	for i, char := range expression {
		switch char {
		case '(':
			stack = append(stack, char)
		case ')':
			if len(stack) == 0 || stack[len(stack)-1] != '(' {
				return false
			}
			stack = stack[:len(stack)-1]
		case '*', '/':
			if i == 0 {
				return false
			}
			if contains([]rune{'+', '-', '*', '/', '(', ' '}, rune(expression[i-1])) {
				return false
			}
		case '-', '+':
			if i == 0 || i == 1 || expression[i-1] == '(' {
				continue
			}
			if contains([]rune{'+', '-', '*', '/', ' '}, rune(expression[i-1])) &&
				contains([]rune{'+', '-', '*', '/', '(', ' '}, rune(expression[i-2])) {
				return false
			}

		default:
			if !unicode.IsDigit(char) {
				return false
			}
		}
	}

	return len(stack) == 0
}

func replaceUnaryPlusAndMinus(expression string) string {
	var result strings.Builder

	length := len(expression)
	ind := 0
	for ind < length {
		if ind+1 < length && contains([]rune{'+', '-', '*', '/'}, rune(expression[ind])) && expression[ind+1] == '+' {
			result.WriteRune(rune(expression[ind]))
			result.WriteRune('&')
			ind++
		} else if ind == 0 && expression[ind] == '+' {
			result.WriteRune('&')
		} else if ind+1 < length && contains([]rune{'+', '-', '*', '/'}, rune(expression[ind])) && expression[ind+1] == '-' {
			result.WriteRune(rune(expression[ind]))
			result.WriteRune('$')
			ind++
		} else if ind == 0 && expression[ind] == '-' {
			result.WriteRune('$')
		} else {
			result.WriteRune(rune(expression[ind]))
		}
		ind++
	}
	return result.String()
}

func orderPlusMinus(expression string) []rune {
	res := make([]rune, 0)
	for _, char := range expression {
		if char == '-' || char == '+' {
			res = append(res, char)
		}
	}
	return res
}

func isNumber(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil || s[0] == '$' || s[0] == '&'
}

func addBrackets(expression string) string {
	var result string

	parts := strings.FieldsFunc(replaceUnaryPlusAndMinus(expression), func(r rune) bool {
		return r == '+' || r == '-'
	})
	fmt.Println(parts)
	length := len(parts)
	sliceOfOrdersPlusMinus := orderPlusMinus(replaceUnaryPlusAndMinus(expression))
	var ind, indForOrdersPlusMinus int
	if len(parts) < 2 {
		return expression
	}
	for ind < length {
		if ind == 0 && isNumber(parts[ind]) && isNumber(parts[ind+1]) {
			result += "(" + parts[ind] + string(sliceOfOrdersPlusMinus[indForOrdersPlusMinus]) + parts[ind+1] + ")"
			indForOrdersPlusMinus++
			ind++
		} else if ind == 0 && ((isNumber(parts[ind]) && !isNumber(parts[ind+1])) || !isNumber(parts[ind])) {
			result += parts[ind]
		} else if ind+1 < length && isNumber(parts[ind]) && isNumber(parts[ind+1]) {
			result += string(sliceOfOrdersPlusMinus[indForOrdersPlusMinus]) + "(" + parts[ind]
			indForOrdersPlusMinus++
			result += string(sliceOfOrdersPlusMinus[indForOrdersPlusMinus]) + parts[ind+1] + ")"
			indForOrdersPlusMinus++
			ind++
		} else {
			result += string(sliceOfOrdersPlusMinus[indForOrdersPlusMinus]) + parts[ind]
			indForOrdersPlusMinus++
		}
		ind++
	}
	result = strings.ReplaceAll(result, "&", "+")
	result = strings.ReplaceAll(result, "$", "-")
	return result
}

func popUntilOpeningParenthesis(stack *[]rune, output *strings.Builder) error {
	for len(*stack) > 0 && (*stack)[len(*stack)-1] != '(' {
		popTopOperator(stack, output)
	}
	if len(*stack) == 0 {
		return errors.New("invalid expression")
	}
	*stack = (*stack)[:len(*stack)-1]
	return nil
}

func popOperatorsWithHigherPrecedence(operator rune, stack *[]rune, output *strings.Builder) {
	for len(*stack) > 0 && precedence((*stack)[len(*stack)-1]) >= precedence(operator) {
		popTopOperator(stack, output)
	}
}

func popTopOperator(stack *[]rune, output *strings.Builder) {
	output.WriteRune(' ')
	output.WriteRune((*stack)[len(*stack)-1])
	*stack = (*stack)[:len(*stack)-1]
}

func precedence(operator rune) int {
	switch operator {
	case '+', '-':
		return 1
	case '*', '/':
		return 2
	default:
		return 0
	}
}
