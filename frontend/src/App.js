import React, { useState } from "react";

const API_BASE =
  process.env.REACT_APP_API_BASE ||
  (window.location.hostname === "localhost" || window.location.hostname === "127.0.0.1"
    ? "http://localhost:8080" // Local dev: backend on your machine
    : "http://app:8080");     // Docker: backend service name from docker-compose

function App() {
  const [randomNumber, setRandomNumber] = useState(null);
  const [error, setError] = useState(null);

  const fetchRandomNumber = async () => {
    try {
      const response = await fetch(`${API_BASE}/random-number`);
      if (!response.ok) throw new Error("Network response was not ok");
      const number = await response.text();
      setRandomNumber(number);
      setError(null);
    } catch (err) {
      setError("Failed to fetch random number.");
      setRandomNumber(null);
    }
  };

  return (
    <div style={{ padding: 40, fontFamily: "sans-serif" }}>
      <h1>Random Number Generator</h1>
      <button onClick={fetchRandomNumber}>Get Random Number from Backend</button>
      {randomNumber && (
        <div style={{ marginTop: 20, fontSize: 24 }}>
          <strong>Random Number: {randomNumber}</strong>
        </div>
      )}
      {error && (
        <div style={{ marginTop: 20, fontSize: 18, color: "red" }}>
          {error}
        </div>
      )}
      <div style={{ marginTop: 20, fontSize: 14, color: "gray" }}>
        Using API_BASE: {API_BASE}
      </div>
    </div>
  );
}

export default App;
