package database

import (
	"time"

	"github.com/google/uuid"
)

type ExpressionTransformed struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Data      string    `json:"data"`
	Status    string    `json:"status"`
}

func DatabaseExpressionToExpression(dbExpr Expression) ExpressionTransformed {
	return ExpressionTransformed{
		ID:        dbExpr.ID,
		CreatedAt: dbExpr.CreatedAt,
		UpdatedAt: dbExpr.UpdatedAt,
		Data:      dbExpr.Data,
		Status:    dbExpr.Status,
	}
}

func DatabaseExpressionsToExpressions(dbExprs []Expression) []ExpressionTransformed {
	exprs := []ExpressionTransformed{}
	for _, dbExpr := range dbExprs {
		exprs = append(exprs, DatabaseExpressionToExpression(dbExpr))
	}
	return exprs
}
