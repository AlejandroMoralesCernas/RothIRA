import React, { useState } from "react";
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import SignUpForm from './components/login/signup';

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<SignUpForm />} />
        {/* <Route path="/random-number" element={<RandomNumber />} /> */}
      </Routes>
    </Router>
  );
}

export default App;