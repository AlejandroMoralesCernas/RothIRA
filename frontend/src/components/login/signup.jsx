import React, { useState } from "react";
import "./signup.css"; // <-- import the stylesheet

const SignUpForm = () => {
  const [mode, setMode] = useState("signup"); // "signup" or "login"
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState("");

  const handleSubmit = (e) => {
    e.preventDefault();

    if (mode === "signup") {
      if (password !== confirmPassword) {
        setError("Passwords do not match.");
        return;
      }
      if (password.length < 6) {
        setError("Password must be at least 6 characters long.");
        return;
      }
      console.log("Signing up with:", { email, password });
      alert("Sign up successful (simulated)!");
    } else {
      console.log("Logging in with:", { email, password });
      alert("Login successful (simulated)!");
    }

    setError("");
    setEmail("");
    setPassword("");
    setConfirmPassword("");
  };

  return (
    <div className="auth">
      <div className="auth__card">
        <h2 className="auth__title">{mode === "signup" ? "Sign Up" : "Log In"}</h2>

        {error && <p className="auth__error">{error}</p>}

        <form onSubmit={handleSubmit} className="auth__form">
          <div className="auth__field">
            <label htmlFor="email" className="auth__label">Email</label>
            <input
              type="email"
              id="email"
              className="auth__input"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              placeholder="you@example.com"
            />
          </div>

          <div className="auth__field">
            <label htmlFor="password" className="auth__label">Password</label>
            <input
              type="password"
              id="password"
              className="auth__input"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              placeholder="••••••••"
            />
          </div>

          {mode === "signup" && (
            <div className="auth__field">
              <label htmlFor="confirmPassword" className="auth__label">Confirm Password</label>
              <input
                type="password"
                id="confirmPassword"
                className="auth__input"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                required
                placeholder="••••••••"
              />
            </div>
          )}

          <button type="submit" className="auth__submit">
            {mode === "signup" ? "Create Account" : "Log In"}
          </button>
        </form>

        <p className="auth__switch">
          {mode === "signup" ? (
            <>
              Already have an account?{" "}
              <button
                type="button"
                className="auth__linkbutton"
                onClick={() => setMode("login")}
              >
                Log In
              </button>
            </>
          ) : (
            <>
              Don’t have an account?{" "}
              <button
                type="button"
                className="auth__linkbutton"
                onClick={() => setMode("signup")}
              >
                Create one
              </button>
            </>
          )}
        </p>
      </div>
    </div>
  );
};

export default SignUpForm;
