import { ROUTES, EXPRESSION_STATUS, AGENT_STATUS } from "./enums";
import { FormVariant } from "./types";

export interface Expression {
  expression_id: number,
  created_at: string,
  updated_at: string,
  data: string,
  status: EXPRESSION_STATUS,
  is_ready: boolean,
  result: number,
  parse_data: string,
  user_id: number,
  agent_id: number,
}

export interface Operation {
  operation_id: number,
  operation_type: string,
  execution_time: number,
  user_id: number,
}

export interface Agent {
  agent_id: number,
  number_of_parallel_calculations: number,
  last_ping: string,
  status: AGENT_STATUS,
  created_at: string,
  number_of_active_calculations: number,
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

export interface FormData {
  email: string,
  password: string,
}

export interface FormProps {
  variant: FormVariant,
  handler: () => void;
  data: FormData;
  setData: (data: FormData) => void;
}

