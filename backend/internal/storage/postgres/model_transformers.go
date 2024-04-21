package postgres

import (
	"database/sql"
	"time"
)

type ExpressionTransformed struct {
	ExpressionID int32            `json:"expression_id"`
	UserID       int32            `json:"user_id"`
	AgentID      sql.NullInt32    `json:"agent_id"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
	Data         string           `json:"data"`
	ParseData    string           `json:"parse_data"`
	Status       ExpressionStatus `json:"status"`
	Result       int32            `json:"result"`
	IsReady      bool             `json:"is_ready"`
}

func DatabaseExpressionToExpression(dbExpr Expression) ExpressionTransformed {
	return ExpressionTransformed(dbExpr)
}

func DatabaseExpressionsToExpressions(dbExprs []Expression) []ExpressionTransformed {
	exprs := []ExpressionTransformed{}
	for _, dbExpr := range dbExprs {
		exprs = append(exprs, DatabaseExpressionToExpression(dbExpr))
	}
	return exprs
}

type OperationTransformed struct {
	OperationID   int32  `json:"operation_id"`
	OperationType string `json:"operation_type"`
	ExecutionTime int32  `json:"execution_time"`
	UserID        int32  `json:"user_id"`
}

func DatabaseOperationToOperation(dbOper Operation) OperationTransformed {
	return OperationTransformed(dbOper)
}

func DatabaseOperationsToOperations(dbOpers []Operation) []OperationTransformed {
	opers := []OperationTransformed{}
	for _, dbOper := range dbOpers {
		opers = append(opers, DatabaseOperationToOperation(dbOper))
	}
	return opers
}

type AgentTransformed struct {
	AgentID                      int32       `json:"agent_id"`
	NumberOfParallelCalculations int32       `json:"number_of_parallel_calculations"`
	LastPing                     time.Time   `json:"last_ping"`
	Status                       AgentStatus `json:"status"`
	CreatedAt                    time.Time   `json:"created_at"`
	NumberOfActiveCalculations   int32       `json:"number_of_active_calculations"`
}

func DatabaseAgentToAgent(dbAgent Agent) AgentTransformed {
	return AgentTransformed(dbAgent)
}

func DatabaseAgentsToAgents(dbAgents []Agent) []AgentTransformed {
	agents := []AgentTransformed{}
	for _, dbAgent := range dbAgents {
		agents = append(agents, DatabaseAgentToAgent(dbAgent))
	}
	return agents
}

type UserTransformed struct {
	UserID       int32  `json:"user_id"`
	Email        string `json:"email"`
	PasswordHash []byte `json:"password_hash"`
}

func DatabaseUserToUser(dbUser User) UserTransformed {
	return UserTransformed(dbUser)
}

func DatabaseUsersToUsers(dbUsers []User) []UserTransformed {
	users := []UserTransformed{}
	for _, dbUser := range dbUsers {
		users = append(users, DatabaseUserToUser(dbUser))
	}
	return users
}
