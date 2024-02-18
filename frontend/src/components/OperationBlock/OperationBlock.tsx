import styles from "./OperationBlock.module.css";
import { OperationBlockProps } from "src/ts/interfaces"
import { useState } from "react";
import { Button } from "../Button/Button";
import { Input } from "../Input/Input";

export const OperationBlock = ({ operation, saveChanges }: OperationBlockProps) => {
  const [operationName, setOperationName] = useState(Number(operation.execution_time));
  const isChanged = operationName !== operation.execution_time;

  return (
    <div>
      <p className={styles.title}>Тип операции (сек): {operation.operation_type}</p>
      <div className={styles.block}>
        <Input
          type="number"
          value={operationName}
          onChange={(e) => setOperationName(Number(e))}
        />
        <Button
          title="Save"
          disabled={!isChanged}
          onClick={() => saveChanges(operationName)}
        />
      </div>
    </div>
  )
}