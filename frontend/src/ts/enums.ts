export enum ROUTES {
  EXPRESSIONS = "Expressions",
  OPERATIONS = "Operations",
  AGENTS = "Agents",
}

export enum EXPRESSION_STATUS {
  READY_FOR_COMPUTATION = "ready for computation",
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