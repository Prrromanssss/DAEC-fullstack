import { ROUTES, EXPRESSION_STATUS, AGENT_STATUS } from "./enums";

export interface Expression {
  id: string,
  created_at: string,
  updated_at: string,
  data: string,
  status: EXPRESSION_STATUS,
  is_ready: boolean,
  result: number,
  parse_date: string,
}

export interface Operation {
  id: string,
  operation_type: string,
  execution_time: number,
}

export interface Agent {
  id: string,
  number_of_parallel_calculations: number,
  last_ping: string,
  status: AGENT_STATUS,
  created_at: string,
}

export interface HeaderProps {
  activePage: ROUTES | string,
  setActivePage: (value: ROUTES) => void,
}

export interface OperationBlockProps {
  operation: Operation,
  saveChanges: (value: number) => void;
}

export interface ButtonProps {
  title: string,
  onClick: () => void,
  disabled?: boolean,
}

export interface InputProps {
  value: string | number,
  onChange: (value: string) => void,
  type?: string;
  placeholder?: string;
}

export interface ExpressionBlockProps {
  expression: Expression,
}

export interface AgentBlockProps {
  agent: Agent,
}

