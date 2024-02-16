import { EXPRESSION_STATUS } from "../../ts/enums";
import { ExpressionBlockProps } from "../../ts/interfaces";
import styles from "./ExpressionBlock.module.css";

const ICONS = {
  [EXPRESSION_STATUS.COMPUTING]: "https://cdn-icons-png.flaticon.com/512/7351/7351882.png",
  [EXPRESSION_STATUS.RESULT]: "https://thumbs.dreamstime.com/b/%D0%B7%D0%B5%D0%BB%D0%B5%D0%BD%D0%B0%D1%8F-%D0%B3%D0%B0%D0%BB%D0%BE%D1%87%D0%BA%D0%B0-%D0%BF%D0%BE%D0%B4%D1%82%D0%B2%D0%B5%D1%80%D0%B6%D0%B4%D0%B0%D1%8E%D1%82-%D0%B8%D0%BB%D0%B8-%D0%BB%D0%B8%D0%BD%D0%B8%D1%8F-%D0%B7%D0%BD%D0%B0%D1%87%D0%BA%D0%B8-%D0%BA%D0%BE%D0%BD%D1%82%D1%80%D0%BE%D0%BB%D1%8C%D0%BD%D0%BE%D0%B9-%D0%BF%D0%BE%D0%BC%D0%B5%D1%82%D0%BA%D0%B8-185588596.jpg",
  [EXPRESSION_STATUS.READY_FOR_COMPUTATION]: "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcSeendsSC1KLb1ljM9hILUWPjJDcu_JPgdvsymh6DZj0YqBILgplXiIpWlELjhynu4R-9M&usqp=CAU",
  [EXPRESSION_STATUS.TERMINATED]: "https://cdn-icons-png.flaticon.com/512/6368/6368418.png",
};

const DESCRIPTION = {
  [EXPRESSION_STATUS.READY_FOR_COMPUTATION]: "the expression is accepted, it will be processed soon",
  [EXPRESSION_STATUS.RESULT]: "the expression is ready",
  [EXPRESSION_STATUS.COMPUTING]: "the expression is being processed, it will be calculated soon",
  [EXPRESSION_STATUS.TERMINATED]: "expression parsing error",
};

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
          {expression.data} = {expression.is_ready ? expression.result : "?"} ({DESCRIPTION[expression.status]})
        </p>
      </div>
      <p className={styles.createdAt}>
        Дата создания: {date}
      </p>
    </div>
  )
}