import './App.css'
import { useEffect, useState } from 'react';
import { Header } from './components/Header/Header'
import { ROUTES } from './ts/enums';
import { AgentsPage } from './pages/Agents/AgentsPage';
import { ExpressionsPage } from './pages/Expressions/ExpressionsPage';
import { OperationsPage } from './pages/Operations/OperationsPage';
import { LoginPage } from './pages/Login/LoginPage';
import axios from 'axios';

function App() {
  const [activePage, setActivePage] = useState("");

  const changePage = (page: ROUTES) => {
    sessionStorage.setItem("page", page);
    setActivePage(page);
  }

  useEffect(() => {
    const pageFromStorage = sessionStorage.getItem("page") || ROUTES.EXPRESSIONS;
    const token = sessionStorage.getItem("token");
    if (pageFromStorage) setActivePage(pageFromStorage);
    if (token) axios.defaults.headers.common = { "Authorization": `Bearer ${token}` };
  }, []);

  return (
    <>
      <Header
        activePage={activePage}
        setActivePage={changePage}
      />
      <div className="page">
        {activePage === ROUTES.AGENTS && <AgentsPage />}
        {activePage === ROUTES.EXPRESSIONS && <ExpressionsPage />}
        {activePage === ROUTES.OPERATIONS && <OperationsPage />}
        {activePage === ROUTES.LOGIN && <LoginPage />}
      </div>
    </>
  )
}

export default App
