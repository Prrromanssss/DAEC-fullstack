package parser_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Prrromanssss/DAEE-fullstack/internal/orchestrator/parser"
)

func TestParseExpression(t *testing.T) {
	testCases := []struct {
		name             string
		expression       string
		wantedExpression string
		err              error
	}{
		{
			name:             "Valid expression with parentheses",
			expression:       "(3+4)*5",
			wantedExpression: "3 4 + 5 *",
			err:              nil,
		},
		{
			name:             "Valid expression with unary minus",
			expression:       "-3+4*5",
			wantedExpression: "0 3 - 4 5 * +",
			err:              nil,
		},
		{
			name:             "Invalid expression with mismatched parentheses",
			expression:       "(3+4*5",
			wantedExpression: "",
			err:              errors.New("invalid expression"),
		},
		{
			name:             "Invalid expression with invalid characters",
			expression:       "3+x+4*5",
			wantedExpression: "",
			err:              errors.New("invalid expression"),
		},
		{
			name:             "Expression with leading spaces",
			expression:       "   3+4*5",
			wantedExpression: "3 4 5 * +",
			err:              nil,
		},
		{
			name:             "Expression with trailing spaces",
			expression:       "3+4*5   ",
			wantedExpression: "3 4 5 * +",
			err:              nil,
		},
		{
			name:             "Expression with spaces in between",
			expression:       "3 + 4 * 5",
			wantedExpression: "3 4 5 * +",
			err:              nil,
		},
		{
			name:             "Expression with unary minus",
			expression:       "-3+-4*5",
			wantedExpression: "0 3 - 0 + 4 5 * -",
			err:              nil,
		},
		{
			name:             "Expression with six unary minus",
			expression:       "------3",
			wantedExpression: "",
			err:              errors.New("invalid expression"),
		},
		{
			name:             "Valid expression with unary plus",
			expression:       "+3+4*5",
			wantedExpression: "0 3 + 4 5 * +",
			err:              nil,
		},
		{
			name:             "Valid expression with consecutive operators",
			expression:       "3++4*5",
			wantedExpression: "3 0 + 4 5 * +",
			err:              nil,
		},
		{
			name:             "Expression with division by zero",
			expression:       "3/0",
			wantedExpression: "",
			err:              errors.New("invalid expression"),
		},
		{
			name:             "Expression with multiple operators",
			expression:       "3+4*5-6/2",
			wantedExpression: "3 4 5 * + 6 2 / -",
			err:              nil,
		},
		{
			name:             "Expression with excessive parentheses",
			expression:       "(((3+4)*5)-6)/2",
			wantedExpression: "3 4 + 5 * 6 - 2 /",
			err:              nil,
		},
		{
			name:             "Expression with starting unary minus and excessive parentheses",
			expression:       "-(((3+4)*5)-6)/2",
			wantedExpression: "0 3 4 + 5 * 6 - 2 / -",
			err:              nil,
		},
		{
			name:             "Expression with empty input",
			expression:       "",
			wantedExpression: "",
			err:              errors.New("invalid expression"),
		},
		{
			name:             "Expression with single number",
			expression:       "42",
			wantedExpression: "42",
			err:              nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := parser.ParseExpression(tc.expression)
			if got != tc.wantedExpression {
				t.Errorf(
					"ParseExpression(%v) = %v, %v; want %v, but got %v",
					tc.expression,
					got, err,
					tc.wantedExpression, got,
				)
			}
			if tc.err != nil && (err == nil || !strings.Contains(err.Error(), tc.err.Error())) {
				t.Errorf(
					"ParseExpression(%v) = %v, %v; expected error containing '%v', but got %v",
					tc.expression,
					got, err,
					tc.err, err,
				)
			} else if tc.err == nil && err != nil {
				t.Errorf(
					"ParseExpression(%v) = %v, %v; expected no error, but got %v",
					tc.expression,
					got, err,
					err,
				)
			}
		})
	}

}

func TestIsValidExpression(t *testing.T) {
	testCases := []struct {
		name       string
		expression string
		want       bool
	}{
		{
			name:       "Empty expression",
			expression: "",
			want:       false,
		},
		{
			name:       "Simple valid expression",
			expression: "(3+4)*5",
			want:       true,
		},
		{
			name:       "Valid expression with nested parentheses",
			expression: "((3+4)*5)",
			want:       true,
		},
		{
			name:       "Valid expression with multiple operators",
			expression: "3+4*5/2",
			want:       true,
		},
		{
			name:       "Valid expression: unary operator at the beginning",
			expression: "3+-4*5",
			want:       true,
		},
		{
			name:       "Valid expression: unary plus",
			expression: "3+4*5++2",
			want:       true,
		},
		{
			name:       "Invalid expression: incomplete expression",
			expression: "3+4*",
			want:       false,
		},
		{
			name:       "Invalid expression: division by zero",
			expression: "3+4/0",
			want:       false,
		},
		{
			name:       "Invalid expression: unbalanced parentheses",
			expression: "3+(4*5",
			want:       false,
		},
		{
			name:       "Valid expression with multiple parentheses",
			expression: "3+(4*(5-6)*2)/2",
			want:       true,
		},
		{
			name:       "Two unary operators",
			expression: "--3+--4*5",
			want:       false,
		},
		{
			name:       "Valid expression with leading zero in a number",
			expression: "03+4*5",
			want:       false,
		},
		{
			name:       "Invalid expression: operator at the end",
			expression: "3+4*5+",
			want:       false,
		},
		{
			name:       "Valid expression with negative number",
			expression: "3+(-4)*5",
			want:       true,
		},
		{
			name:       "Valid expression: division by negative number",
			expression: "3+4/(-2)",
			want:       true,
		},
		{
			name:       "Valid expression with negative number in parentheses",
			expression: "3+(-(4 * 5))",
			want:       false,
		},
		{
			name:       "Invalid expression: incomplete negative number",
			expression: "3+(-)",
			want:       false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := parser.IsValidExpression(tc.expression)
			if got != tc.want {
				t.Errorf("IsValidExpression(%v) = %v; want %v", tc.expression, got, tc.want)
			}
		})
	}
}

func TestAddBrackets(t *testing.T) {
	testCases := []struct {
		name             string
		expression       string
		wantedExpression string
	}{
		{
			name:             "Unary plus at the beginning",
			expression:       "+3+4*5",
			wantedExpression: "(0+3)+4*5",
		},
		{
			name:             "Unary minus at the beginning",
			expression:       "-3+4*5",
			wantedExpression: "(0-3)+4*5",
		},
		{
			name:             "Unary minus before number",
			expression:       "3+-4*5",
			wantedExpression: "3+0-4*5",
		},
		{
			name:             "Unary plus and minus combined",
			expression:       "-3+-4*5",
			wantedExpression: "(0-3)+0-4*5",
		},
		{
			name:             "Expression with brackets",
			expression:       "3+(4*5)+6",
			wantedExpression: "3+(4*5)+6",
		},
		{
			name:             "Expression with many same operators",
			expression:       "1+1+2+2+3+3",
			wantedExpression: "(1+1)+(2+2)+(3+3)",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := parser.AddBrackets(tc.expression)
			if got != tc.wantedExpression {
				t.Errorf("AddBrackets(%v) = %v; want %v", tc.expression, got, tc.wantedExpression)
			}
		})
	}
}
