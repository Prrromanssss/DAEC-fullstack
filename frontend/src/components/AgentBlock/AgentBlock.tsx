import { AGENT_STATUS } from "../../ts/enums";
import { AgentBlockProps } from "../../ts/interfaces";
import styles from "./AgentBlock.module.css";

const ICONS = {
  [AGENT_STATUS.RUNNING]: "https://cdn-icons-png.flaticon.com/512/7351/7351882.png",
  [AGENT_STATUS.SLEEPING]: "https://thumbs.dreamstime.com/b/%D0%B7%D0%B5%D0%BB%D0%B5%D0%BD%D0%B0%D1%8F-%D0%B3%D0%B0%D0%BB%D0%BE%D1%87%D0%BA%D0%B0-%D0%BF%D0%BE%D0%B4%D1%82%D0%B2%D0%B5%D1%80%D0%B6%D0%B4%D0%B0%D1%8E%D1%82-%D0%B8%D0%BB%D0%B8-%D0%BB%D0%B8%D0%BD%D0%B8%D1%8F-%D0%B7%D0%BD%D0%B0%D1%87%D0%BA%D0%B8-%D0%BA%D0%BE%D0%BD%D1%82%D1%80%D0%BE%D0%BB%D1%8C%D0%BD%D0%BE%D0%B9-%D0%BF%D0%BE%D0%BC%D0%B5%D1%82%D0%BA%D0%B8-185588596.jpg",
  [AGENT_STATUS.WAITING]: "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcSeendsSC1KLb1ljM9hILUWPjJDcu_JPgdvsymh6DZj0YqBILgplXiIpWlELjhynu4R-9M&usqp=CAU",
  [AGENT_STATUS.TERMINATED]: "https://cdn-icons-png.flaticon.com/512/6368/6368418.png",
};

const DESCRIPTION = {
  [AGENT_STATUS.RUNNING]: "the server is calculating expressions and waiting for new ones",
  [AGENT_STATUS.SLEEPING]: "the server is calculating the expressions and is fully occupied",
  [AGENT_STATUS.WAITING]: "the server is waiting for new expressions",
  [AGENT_STATUS.TERMINATED]: "the server is down",
};

export const AgentBlock = ({ agent }: AgentBlockProps) => {
  const createdAt = new Date(agent.created_at).toLocaleString();
  const computingAt = new Date(agent.last_ping).toLocaleString();

  return (
    <div className={styles.expressionBlock}>
      <div className={styles.headerBlock}>
        <img
          className={styles.icon}
          src={ICONS[agent.status]}
        />
        <p className={styles.title}>
          Computing server ({DESCRIPTION[agent.status]})
        </p>
      </div>
      <p className={styles.createdAt}>
        Last ping: {computingAt}
      </p>
      <p className={styles.createdAt}>
      Number of parallel calculations: {agent.number_of_parallel_calculations}
      </p>
      <p className={styles.createdAt}>
        Дата создания: {createdAt}
      </p>
    </div>
  )
}