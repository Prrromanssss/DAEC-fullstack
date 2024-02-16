import { ButtonProps } from "../../ts/interfaces";
import styles from "./Button.module.css";

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