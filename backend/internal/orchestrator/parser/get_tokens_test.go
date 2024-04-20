package parser_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Prrromanssss/DAEE-fullstack/internal/domain/messages"
	"github.com/Prrromanssss/DAEE-fullstack/internal/lib/logger/handlers/slogdiscard"
	"github.com/Prrromanssss/DAEE-fullstack/internal/orchestrator/parser"
)

func TestGetTokens(t *testing.T) {
	cases := []struct {
		name         string
		expression   string
		wantedTokens []string
	}{
		{
			name:         "one token, only plus operator",
			expression:   "1 1 +",
			wantedTokens: []string{"1 1 +"},
		},
		{
			name:         "one token, only minus operator",
			expression:   "1 1 -",
			wantedTokens: []string{"1 1 -"},
		},
		{
			name:         "one token, only product operator",
			expression:   "1 1 *",
			wantedTokens: []string{"1 1 *"},
		},
		{
			name:         "one token, only division operator",
			expression:   "1 1 /",
			wantedTokens: []string{"1 1 /"},
		},
		{
			name:         "three tokens, only plus operator",
			expression:   "1 1 + 2 2 + + 3 3 + +",
			wantedTokens: []string{"1 1 +", "2 2 +", "3 3 +"},
		},
		{
			name:         "three tokens, big numbers, different operators",
			expression:   "1345 1123 + 9 223 - + 9 3 * +",
			wantedTokens: []string{"1345 1123 +", "9 223 -", "9 3 *"},
		},
	}

	log := slogdiscard.NewDiscardLogger()

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := parser.GetTokens(log, tc.expression)
			if !slicesEqual(got, tc.wantedTokens) {
				t.Errorf("GetTokens(log, %v) = %v; want %v", tc.expression, got, tc.wantedTokens)
			}
		})
	}
}

func TestInsertResultToToken(t *testing.T) {
	cases := []struct {
		name             string
		expression       string
		token            string
		result           int
		err              error
		wantedExpression messages.ResultAndTokenMessage
	}{
		{
			name:       "Correct case with token at the beginning of the expression",
			expression: "3 3 + 4 + 122 +",
			token:      "3 3 +",
			result:     6,
			err:        nil,
			wantedExpression: messages.ResultAndTokenMessage{
				Result: "6 4 + 122 +",
				Token:  "6 4 +",
			},
		},
		{
			name:       "Correct case with token at the end of the expression",
			expression: "3 3 + 4 + 55 67 + +",
			token:      "55 67 +",
			result:     122,
			err:        nil,
			wantedExpression: messages.ResultAndTokenMessage{
				Result: "3 3 + 4 + 122 +",
				Token:  "",
			},
		},
		{
			name:       "Correct case with token in the middle of the expression",
			expression: "1 1 + 2 2 + + 3 3 + +",
			token:      "2 2 +",
			result:     4,
			err:        nil,
			wantedExpression: messages.ResultAndTokenMessage{
				Result: "1 1 + 4 + 3 3 + +",
				Token:  "",
			},
		},
		{
			name:       "Correct case with a token length of a three",
			expression: "45 23 +",
			token:      "45 23 +",
			result:     68,
			err:        nil,
			wantedExpression: messages.ResultAndTokenMessage{
				Result: "68",
				Token:  "",
			},
		},
		{
			name:       "Incorrect case with missing token",
			expression: "3 3 + 4 + 55 67 + +",
			token:      "155 67 +",
			result:     222,
			err:        errors.New("can't find token"),
			wantedExpression: messages.ResultAndTokenMessage{
				Result: "",
				Token:  "",
			},
		},
		{
			name:       "Incorrect case with invalid expression",
			expression: "3 3 + 4 + 55 67 +",
			token:      "55 67 +",
			result:     122,
			err:        errors.New("invalid expression"),
			wantedExpression: messages.ResultAndTokenMessage{
				Result: "",
				Token:  "",
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := parser.InsertResultToToken(tc.expression, tc.token, tc.result)
			if got.Result != tc.wantedExpression.Result {
				t.Errorf(
					"InsertResultToToken(%v, %v, %v) = %v, %v; want %v, but got %v",
					tc.expression, tc.token, tc.result,
					got, err,
					tc.wantedExpression.Result, got.Result,
				)
			}
			if got.Token != tc.wantedExpression.Token {
				t.Errorf(
					"InsertResultToToken(%v, %v, %v) = %v, %v; want %v, but got %v",
					tc.expression, tc.token, tc.result,
					got, err,
					tc.wantedExpression.Token, got.Token,
				)
			}
			if tc.err != nil && (err == nil || !strings.Contains(err.Error(), tc.err.Error())) {
				t.Errorf(
					"InsertResultToToken(%v, %v, %v) = %v, %v; expected error containing '%v', but got %v",
					tc.expression, tc.token, tc.result,
					got, err,
					tc.err, err,
				)
			} else if tc.err == nil && err != nil {
				t.Errorf(
					"InsertResultToToken(%v, %v, %v) = %v, %v; expected no error, but got %v",
					tc.expression, tc.token, tc.result,
					got, err,
					err,
				)
			}
		})
	}
}

func slicesEqual(slice1, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	for i := 0; i < len(slice1); i++ {
		if slice1[i] != slice2[i] {
			return false
		}
	}

	return true
}
