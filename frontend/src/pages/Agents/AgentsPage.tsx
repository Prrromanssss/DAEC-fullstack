import { useEffect, useState } from "react";
import { Agent } from "../../ts/interfaces";
import { getAgents } from "../../services/api";
import { AgentBlock } from "../../components/AgentBlock/AgentBlock";
import styles from "./AgentsPage.module.css";

export const AgentsPage = () => {
  const [agents, setAgents] = useState<Agent[]>([]);

  useEffect(() => {
    getAgents().then(data => setAgents(data));
  }, []);

  return (
    <div className={styles.container}>
      {agents.map(agent => <AgentBlock agent={agent} />)}
    </div>
  )
}