import styles from "./OperationPage.module.css";
import { useEffect, useState } from "react";
import { Operation } from "../../ts/interfaces";
import { OperationBlock } from "../../components/OperationBlock/OperationBlock";
import { getOperations, updateOperation } from "../../services/api";

export const OperationsPage = () => {
  const [operations, setOperations] = useState<Operation[]>([]);

  const saveChanges = (newValue: number, operation: Operation) => {
    updateOperation({ ...operation, execution_time: newValue }).then(() => {
      getOperations().then(data => setOperations(data));
    })
  };

  useEffect(() => {
    getOperations().then(data => setOperations(data));
  }, []);

  return (
    <div className={styles.container}>
      {operations.map(operation => (
        <OperationBlock
          key={operation.id}
          operation={operation}
          saveChanges={(newValue) => saveChanges(newValue, operation)}
        />
      ))}
    </div>
  )
}