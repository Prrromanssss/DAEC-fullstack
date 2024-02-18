import styles from "./ExpressionsPage.module.css";
import { useEffect, useState } from "react";
import { Expression } from "src/ts/interfaces";
import { Button } from "src/components/Button/Button";
import { Input } from "src/components/Input/Input";
import { createExpression, getExpressions } from "src/services/api";
import { ExpressionBlock } from "src/components/ExpressionBlock/ExpressionBlock";

export const ExpressionsPage = () => {
  const [expressions, setExpressions] = useState<Expression[]>([]);
  const [newExpression, setNewExpression] = useState<string>("");

  const createHandler = () => {
    createExpression(newExpression)
      .then(() => {
        getExpressions()
          .then(data => setExpressions(data));
      })
      .finally(() => {
        setNewExpression("");
      });
  };

  useEffect(() => {
    getExpressions()
      .then(data => setExpressions(data));
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
        {expressions.map(expression => (
          <ExpressionBlock
            key={expression.id}
            expression={expression}
          />
        ))}
      </div>
    </div>
  )
}