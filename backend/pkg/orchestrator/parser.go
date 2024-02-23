package orchestrator

import (
	"errors"
	"log"
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
			if len(stack) == 0 || stack[len(stack)-1] != '(' || (i > 0 && expression[i-1] == '(') {
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
			if i+1 < len(expression) && expression[i+1] == '0' {
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

func addZeroToUnaryPlusAndMinus(expression string) string {
	var result strings.Builder
	log.Println(expression)
	length := len(expression)
	ind := 0
	for ind < length {
		if ind+1 < length && contains([]rune{'+', '-', '*', '/'}, rune(expression[ind])) && expression[ind+1] == '+' {
			result.WriteRune(rune(expression[ind]))
			result.WriteRune('0')
			result.WriteRune('+')
			ind++
		} else if ind == 0 && expression[ind] == '+' {
			result.WriteRune('0')
			result.WriteRune('+')
		} else if ind+1 < length && contains([]rune{'+', '-', '*', '/'}, rune(expression[ind])) && expression[ind+1] == '-' {
			result.WriteRune(rune(expression[ind]))
			result.WriteRune('0')
			result.WriteRune('-')
			ind++
		} else if ind == 0 && expression[ind] == '-' {
			result.WriteRune('0')
			result.WriteRune('-')
		} else {
			result.WriteRune(rune(expression[ind]))
		}
		ind++
		log.Println(result.String())
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

func IsNumber(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func addBrackets(expression string) string {
	var result string

	parts := strings.FieldsFunc(addZeroToUnaryPlusAndMinus(expression), func(r rune) bool {
		return r == '+' || r == '-'
	})
	length := len(parts)
	sliceOfOrdersPlusMinus := orderPlusMinus(addZeroToUnaryPlusAndMinus(expression))
	var ind, indForOrdersPlusMinus int
	if len(parts) <= 2 {
		return expression
	}
	for ind < length {
		currentOperator := string(sliceOfOrdersPlusMinus[indForOrdersPlusMinus])
		currentSymbol := parts[ind]

		if ind == 0 &&
			IsNumber(currentSymbol) &&
			IsNumber(parts[ind+1]) &&
			sliceOfOrdersPlusMinus[indForOrdersPlusMinus+1] == '+' {

			result += "(" + currentSymbol + currentOperator + parts[ind+1] + ")"
			indForOrdersPlusMinus++
			ind++
		} else if ind == 0 &&
			((IsNumber(currentSymbol) && !IsNumber(parts[ind+1])) ||
				!IsNumber(currentSymbol)) {

			result += currentSymbol
		} else if ind == 0 {
			result += currentSymbol
		} else if ind+1 < length &&
			IsNumber(currentSymbol) &&
			IsNumber(parts[ind+1]) &&
			currentOperator == "+" &&
			(indForOrdersPlusMinus+2 >= len(sliceOfOrdersPlusMinus) ||
				sliceOfOrdersPlusMinus[indForOrdersPlusMinus+2] == '+') {

			result += currentOperator + "(" + currentSymbol
			indForOrdersPlusMinus++
			result += string(sliceOfOrdersPlusMinus[indForOrdersPlusMinus]) + parts[ind+1] + ")"
			indForOrdersPlusMinus++
			ind++
		} else {
			result += currentOperator + currentSymbol
			indForOrdersPlusMinus++
		}
		ind++
	}
	result = strings.ReplaceAll(result, "&", "+")
	result = strings.ReplaceAll(result, "$", "-")
	return result
}
