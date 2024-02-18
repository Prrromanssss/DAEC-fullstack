import styles from "./ExpressionBlock.module.css";
import { EXPRESSION_DESCRIPTION, ICONS } from "src/ts/consts";
import { ExpressionBlockProps } from "src/ts/interfaces";

export const ExpressionBlock = ({ expression }: ExpressionBlockProps) => {
  const date = new Date(expression.created_at).toLocaleString();

  return (
    <div className={styles.expressionBlock}>
      <div className={styles.headerBlock}>
        <img
          className={styles.icon}
          src={ICONS[expression.status]}
        />
        <p className={styles.title}>
          {expression.data} = {expression.is_ready ? expression.result : "?"} ({EXPRESSION_DESCRIPTION[expression.status]})
        </p>
      </div>
      <p className={styles.createdAt}>
        Дата создания: {date}
      </p>
    </div>
  )
}