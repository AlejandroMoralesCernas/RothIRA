import React, { useState } from "react";
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';

function App() {
  const [randomNumber, setRandomNumber] = useState(null);
  const [error, setError] = useState(null);
  const fetchRandomNumber = async () => {
    try {
      // URL assumes backend runs on localhost:8081
      const response = await fetch("http://localhost:8081/random-number");
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
          <strong>Random Number: </strong> {randomNumber}
        </div>
      )}
      {error && (
        <div style={{ marginTop: 20, color: "red" }}>{error}</div>
      )}
    </div>
  );
}

export default App;