import styles from "./ExpressionsPage.module.css";
import { useEffect, useState } from "react";
import { Expression } from "../../ts/interfaces";
import { Button } from "../../components/Button/Button";
import { Input } from "../../components/Input/Input";
import { createExpression, getExpressions } from "../../services/api";
import { ExpressionBlock } from "../../components/ExpressionBlock/ExpressionBlock";

export const ExpressionsPage = () => {
  const [expressions, setExpressions] = useState<Expression[]>([]);
  const [newExpression, setNewExpression] = useState<string>("");

  const createHandler = () => {
    createExpression(newExpression).then(() => {
      getExpressions().then(data => setExpressions(data));
    }).finally(() => {
      setNewExpression("");
    });
  };

  useEffect(() => {
    getExpressions().then(data => setExpressions(data));
  }, []);

  return (
    <div>
      <div className={styles.actionsBlock}>
        <Input
          placeholder="Enter expression to calculate"
          value={newExpression}
          onChange={(e) => setNewExpression(e)}
        />
        <Button
          onClick={createHandler}
          disabled={!newExpression.length}
          title="Create"
        />
      </div>
      <div className={styles.items}>
        {expressions.map(expression => <ExpressionBlock key={expression.id} expression={expression} />)}
      </div>
    </div>
  )
}