package parser

import (
	"errors"
	"strings"
)

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
