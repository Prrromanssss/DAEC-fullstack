package parser_test

import (
	"testing"

	"github.com/Prrromanssss/DAEE-fullstack/internal/orchestrator/parser"
)

func TestIsValidExpression(t *testing.T) {
	testCases := []struct {
		name       string
		expression string
		want       bool
	}{
		{
			name:       "Empty expression",
			expression: "",
			want:       true,
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
			name:       "Invalid expression: division by negative number",
			expression: "3+4/(-2)",
			want:       false,
		},
		{
			name:       "Valid expression with negative number in parentheses",
			expression: "3+(-(4 * 5))",
			want:       true,
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
