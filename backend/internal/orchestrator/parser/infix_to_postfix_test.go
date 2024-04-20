package parser_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Prrromanssss/DAEE-fullstack/internal/orchestrator/parser"
)

func TestInfixToPostfix(t *testing.T) {
	testCases := []struct {
		name             string
		expression       string
		wantedExpression string
		err              error
	}{
		{
			name:             "Basic arithmetic expressions - '+'",
			expression:       "3+4",
			wantedExpression: "3 4 +",
			err:              nil,
		},
		{
			name:             "Basic arithmetic expressions - '-'",
			expression:       "3-4",
			wantedExpression: "3 4 -",
			err:              nil,
		},
		{
			name:             "Basic arithmetic expressions - '*'",
			expression:       "3*4",
			wantedExpression: "3 4 *",
			err:              nil,
		},
		{
			name:             "Basic arithmetic expressions - '/'",
			expression:       "3/4",
			wantedExpression: "3 4 /",
			err:              nil,
		},
		{
			name:             "Expression with parentheses at the beginning with plus operator",
			expression:       "(3+4)*5",
			wantedExpression: "3 4 + 5 *",
			err:              nil,
		},
		{
			name:             "Expression with parentheses at the beginning with product operator",
			expression:       "(3*4)+5",
			wantedExpression: "3 4 * 5 +",
			err:              nil,
		},
		{
			name:             "Expression with parentheses at the end with plus operator",
			expression:       "3*(4+5)",
			wantedExpression: "3 4 5 + *",
			err:              nil,
		},
		{
			name:             "Expression with 2 parentheses at the end and at the beginning",
			expression:       "(3+4)*(5-6)",
			wantedExpression: "3 4 + 5 6 - *",
			err:              nil,
		},
		{
			name:             "Complex expressions 1",
			expression:       "3+4*5",
			wantedExpression: "3 4 5 * +",
			err:              nil,
		},
		{
			name:             "Complex expressions 2",
			expression:       "3*4+5",
			wantedExpression: "3 4 * 5 +",
			err:              nil,
		},
		{
			name:             "Complex expressions 3",
			expression:       "(3+4)*5-6/2",
			wantedExpression: "3 4 + 5 * 6 2 / -",
			err:              nil,
		},
		{
			name:             "Complex expressions 4",
			expression:       "3*(4+5)-6/(2+1)",
			wantedExpression: "3 4 5 + * 6 2 1 + / -",
			err:              nil,
		},
		{
			name:             "Expression with multiple operators of the same precedence 1",
			expression:       "3+4-5",
			wantedExpression: "3 4 + 5 -",
			err:              nil,
		},
		{
			name:             "Expression with multiple operators of the same precedence 2",
			expression:       "3*4/5",
			wantedExpression: "3 4 * 5 /",
			err:              nil,
		},
		{
			name:             "Expression with multiple operators of the same precedence 3",
			expression:       "(3+4)-(5+6)",
			wantedExpression: "3 4 + 5 6 + -",
			err:              nil,
		},
		{
			name:             "Expression with multiple operators of the same precedence 4",
			expression:       "(3*4)/(5*6)",
			wantedExpression: "3 4 * 5 6 * /",
			err:              nil,
		},
		{
			name:             "Invalid expression",
			expression:       "5+)+3",
			wantedExpression: "",
			err:              errors.New("invalid expression"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := parser.InfixToPostfix(tc.expression)
			if got != tc.wantedExpression {
				t.Errorf(
					"InfixToPostfix(%v) = %v, %v; want %v, but got %v",
					tc.expression, got, err,
					tc.wantedExpression, got,
				)
			}
			if tc.err != nil && (err == nil || !strings.Contains(err.Error(), tc.err.Error())) {
				t.Errorf(
					"InfixToPostfix(%v) = %v, %v; expected error containing '%v', but got %v",
					tc.expression, got, err,
					tc.err, err,
				)
			} else if tc.err == nil && err != nil {
				t.Errorf(
					"InfixToPostfix(%v) = %v, %v; expected no error, but got %v",
					tc.expression, got, err,
					err,
				)
			}
		})
	}
}
