package orchestrator

import (
	"errors"
	"strings"
	"unicode"
)

func InfixToPostfix(expression string) (string, error) {
	var output strings.Builder
	var stack []rune
	rawExpression := strings.ReplaceAll(expression, " ", "")
	if !isValidExpression(rawExpression) {
		return "", errors.New("invalid expression")
	}

	for _, char := range rawExpression {
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

func Contains(arr []rune, element rune) bool {
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
			if Contains([]rune{'+', '-', '*', '/', '(', ' '}, rune(expression[i-1])) {
				return false
			}
		case '-', '+':
			if i == 0 || i == 1 || expression[i-1] == '(' {
				continue
			}
			if Contains([]rune{'+', '-', '*', '/', ' '}, rune(expression[i-1])) &&
				Contains([]rune{'+', '-', '*', '/', '(', ' '}, rune(expression[i-2])) {
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

// func addBrackets(expression string) string {
// 	return expression
// }

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
