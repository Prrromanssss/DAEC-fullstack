import styles from "./Header.module.css";
import { ROUTES } from "../../ts/enums";
import { HeaderProps } from "../../ts/interfaces";

export const Header = ({ activePage, setActivePage }: HeaderProps) => {
  return <div className={styles.container}>
    {Object.values(ROUTES).map(page => {
      const isActive = activePage === page;

      return (
        <p
          key={page}
          onClick={() => setActivePage(page)}
          style={{
            textDecoration: isActive ? "underline" : "none"
          }}
        >
          {page}
        </p>
      )
    })}
  </div>
}