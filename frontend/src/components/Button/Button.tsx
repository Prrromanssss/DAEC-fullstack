import styles from "./Button.module.css";
import { ButtonProps } from "src/ts/interfaces";

export const Button = ({ title, onClick, disabled }: ButtonProps) => {
  return (
    <button
      className={styles.btn}
      onClick={onClick}
      disabled={disabled}
    >
      {title}
    </button>
  )
}