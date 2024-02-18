import styles from "./AgentsPage.module.css";
import { useEffect, useState } from "react";
import { Agent } from "src/ts/interfaces";
import { getAgents } from "src/services/api";
import { AgentBlock } from "src/components/AgentBlock/AgentBlock";

export const AgentsPage = () => {
  const [agents, setAgents] = useState<Agent[]>([]);

  useEffect(() => {
    getAgents()
      .then(data => setAgents(data));
  }, []);

  return (
    <div className={styles.container}>
      {agents.map(agent => <AgentBlock agent={agent} />)}
    </div>
  )
}