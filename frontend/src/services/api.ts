import axios from "axios"
import { Agent, Expression, FormData, Operation } from "src/ts/interfaces";

axios.defaults.baseURL = "http://localhost:3000/v1";

export const getExpressions = async (): Promise<Expression[]> => {
  const { data } = await axios.get("/expressions");
  return data;
}

export const createExpression = async (name: string): Promise<Expression> => {
  const { data } = await axios.post("/expressions", { data: name });
  return data;
}

export const getOperations = async (): Promise<Operation[]> => {
  const { data } = await axios.get("/operations");
  return data;
}

export const updateOperation = async (operation: Operation): Promise<Operation> => {
  const { operation_type, execution_time } = operation;
  const { data } = await axios.patch("/operations", { operation_type, execution_time });
  return data;
}

export const getAgents = async (): Promise<Agent[]> => {
  const { data } = await axios.get("/agents");
  return data;
}

export const login = async (value: FormData): Promise<{ token: string }> => {
  const { data } = await axios.post("/login", value);
  sessionStorage.setItem("token", data.token);
  axios.defaults.headers.common = { "Authorization": `Bearer ${data.token}` }
  return data;
}

export const registration = async (value: FormData): Promise<{ user_id: number }> => {
  const { data } = await axios.post("/register", value);
  return data;
}