import { LoginForm } from "src/components/LoginForm/LoginForm";
import styles from "./LoginPage.module.css";
import { useState } from "react";
import { FormVariant } from "src/ts/types";
import { login, registration } from "src/services/api";
import { toast } from "react-toastify";

export const LoginPage = () => {
  const [variant, setVariant] = useState<FormVariant>("login");
  const [data, setData] = useState({
    email: "",
    password: "",
  });

  const handler = () => {
    if (variant === "login") {
      login(data).then(() => {
        setData({ email: "", password: "" })
        toast.success("Успешно!");
      });
    } else {
      registration(data).then(() => {
        toast.success("Успешно!");
        setVariant("login");
      })
      .catch((err) => {
        toast.error(err.response.data.error);
      });
    }
  };

  return (
    <div className={styles.container}>
      <div className={styles.containerBtn}>
        <div
          onClick={() => {
            setData({ email: "", password: "" });
            setVariant("login");
          }}
          className={styles.activeElement}
          style={{
            background: variant === "login" ? "blue" : "white",
            color: variant === "login" ? "white" : "black",
          }}
        >
          Login
        </div>
        <div>/</div>
        <div
          onClick={() => {
            setData({ email: "", password: "" });
            setVariant("reg");
          }}
          className={styles.activeElement}
          style={{
            background: variant === "reg" ? "blue" : "white",
            color: variant === "reg" ? "white" : "black",
          }}
        >
          Registration
        </div>
      </div>
      <LoginForm
        variant={variant}
        handler={handler}
        data={data}
        setData={setData}
      />
    </div>
  )
}