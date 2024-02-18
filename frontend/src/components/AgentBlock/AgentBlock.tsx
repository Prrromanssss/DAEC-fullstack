import styles from "./AgentBlock.module.css";
import { AGENT_DESCRIPTION, ICONS } from "src/ts/consts";
import { AgentBlockProps } from "src/ts/interfaces";

export const AgentBlock = ({ agent }: AgentBlockProps) => {
  const createdAt = new Date(agent.created_at).toLocaleString();
  const computingAt = new Date(agent.last_ping).toLocaleString();

  return (
    <div className={styles.agentBlock}>
      <div className={styles.headerBlock}>
        <img
          className={styles.icon}
          src={ICONS[agent.status]}
        />
        <p className={styles.title}>
          Computing server ({AGENT_DESCRIPTION[agent.status]})
        </p>
      </div>
      <p className={styles.text}>
        Last ping: {computingAt}
      </p>
      <p className={styles.text}>
        Number of parallel calculations: {agent.number_of_parallel_calculations}
      </p>
      <p className={styles.text}>
        Дата создания: {createdAt}
      </p>
    </div>
  )
}