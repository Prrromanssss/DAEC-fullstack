import styles from "./LoginForm.module.css";
import { FormProps } from "src/ts/interfaces";

export const LoginForm = ({ variant, handler, data, setData }: FormProps) => {
  return (
    <div className={styles.container}>
      <input
        name="email"
        onChange={(e) => setData({ ...data, email: e.target.value })}
        value={data.email}
        className={styles.input}
        placeholder="Email..."
      />
      <input
        name="password"
        type="password"
        onChange={(e) => setData({ ...data, password: e.target.value })}
        value={data.password}
        className={styles.input}
        placeholder="Password..."
      />
      <button
        className={styles.button}
        disabled={!data.email.length || !data.password.length}
        onClick={handler}
      >
        {variant === "login" ? "Войти" : "Регистрация"}
      </button>
    </div>
  )
}