import React from "react";
import { Routes, Route, Navigate } from "react-router-dom";
import Authentication from "./components/authentication/authentication";
import RothIRACalculator from "./components/graph";

function CalculatorPage() {
  const username = localStorage.getItem("username") || "There";

  function handleLogout() {
    localStorage.removeItem("token");
    localStorage.removeItem("username");
    window.location.href = "/";
  } // ✅ this closing brace was missing

  return (
    <div style={{ padding: "40px", fontFamily: "sans-serif" }}>
      <div className="app-header">
        <h1 className="text-2xl font-bold mb-6">Hello {username}!</h1>
        <button className="logout-btn-app" onClick={handleLogout}>
          Log Out
        </button>
      </div>

      <RothIRACalculator />
    </div>
  );
}

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<Authentication />} />
      <Route path="/app" element={<CalculatorPage />} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
