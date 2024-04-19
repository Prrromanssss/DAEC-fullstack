import './App.css'
import { useEffect, useState } from 'react';
import { Header } from './components/Header/Header'
import { ROUTES } from './ts/enums';
import { AgentsPage } from './pages/Agents/AgentsPage';
import { ExpressionsPage } from './pages/Expressions/ExpressionsPage';
import { OperationsPage } from './pages/Operations/OperationsPage';

function App() {
  const [activePage, setActivePage] = useState("");

  const changePage = (page: ROUTES) => {
    sessionStorage.setItem("page", page);
    setActivePage(page);
  }

  useEffect(() => {
    const pageFromStorage = sessionStorage.getItem("page") || ROUTES.EXPRESSIONS;
    if (pageFromStorage) setActivePage(pageFromStorage);
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
      </div>
    </>
  )
}

export default App
