import React from "react";
import { Routes, Route, Navigate } from "react-router-dom";
import Authentication from "./components/authentication/authentication";
import RothIRACalculator from "./components/graph";

function CalculatorPage() {
  const username = localStorage.getItem("username") || "There"; // Fallback to display "There" if username is not found
  return (
    <div style={{ padding: 40, fontFamily: "sans-serif" }}>
      <h1 className="text-2xl font-bold mb-6">Hello {username}!</h1>
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
