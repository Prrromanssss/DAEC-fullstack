export enum ROUTES {
  EXPRESSIONS = "Expressions",
  OPERATIONS = "Operations",
  AGENTS = "Agents",
  LOGIN = "Login",
}

export enum EXPRESSION_STATUS {
  READY_FOR_COMPUTATION = "ready_for_computation",
  COMPUTING = "computing",
  RESULT = "result",
  TERMINATED = "terminated",
}

export enum AGENT_STATUS {
  RUNNING = "running",
  WAITING = "waiting",
  SLEEPING = "sleeping",
  TERMINATED = "terminated",
}