// importing necessary React hooks and navigation function
import { useState, useMemo, useRef, useEffect } from "react"; // useState lets component hold state w/ re-renders, useMemo caches a value to skip unnecessary recalculations, useRef keeps values between renders, useEffect runs side code after render
// useState when changes should update what the user sees, useRef when you just need to keep or access a value silently in the background, useState triggers re-renders, useRef does not
import { useNavigate } from "react-router-dom"; // allows navigation between routes without a link click (i.e., after login)

// importing CSS for styling
import "./authentication.css";


// (where we talk to our backend) read base URL from env, fallback to local backend port 8081
// API_BASE: variable holding the base URL for the API
const API_BASE = "http://localhost:8080";

// Creating a box (object) that holds address of API endpoints
const API = {
  signup: `${API_BASE}/api/auth/create-user`,
  login:  `${API_BASE}/api/auth/login-user`,
  verify: `${API_BASE}/api/auth/verify-session`, 
};

// helper function to make POST requests with JSON body and parse JSON response
async function postJSON(url, body, timeoutMs = 3000) { // where to send the request (backend url), the data we want to send (body) in JSON format, how long to wait before giving up (3000 ms)
  // to abort fetch request if it takes too long
  const ctrl = new AbortController(); // creates a kill switch for fetch requests
  const id = setTimeout(() => ctrl.abort(), timeoutMs);  // sets a timer to trigger abort after timeoutMs, the timer starts

  // trying to send the request
  console.log("posting to " + url) // log the URL we are posting to (8080)
  // starts a try block, any fail will jump to finally block
  try {
    const res = await fetch(url, { // calling fetch to make network request, await pauses until fetch completes
      method: "POST", // POST, meaning we are sending data to server
      headers: { "Content-Type": "application/json" }, // informing server that the content type is JSON
      body: JSON.stringify(body), // convert JS object to JSON string for request body
      signal: ctrl.signal, // connects request to abort controller, so if timeout triggersed, request is aborted
      credentials: "include", // instructs fetch to include cookies in requests
    });
    console.log("response received", res)

    const ctype = res.headers.get("content-type") || ""; // when response arrives, get content-type header, default to empty string if missing
      let data; // declare variable to hold parsed response data, 'let' allows reassignment
      if (ctype.includes("application/json")) { // check if content-type is JSON

      // res is the response object from fetch, .json() parses the response body as JSON, .catch handles any parsing errors (arrow function returns empty object on error)
        data = await res.json().catch(() => ({}));
      } else {
        data = { message: await res.text().catch(() => "") }; // else get raw text response
      }
    return { ok: res.ok, status: res.status, data }; // return whether response was ok (status 200-299), status code, and parsed data
  } finally { // runs whether success or failure
    clearTimeout(id); // clear the timeout to avoid memory leaks
  }
}

// Authentication component handles login and signup UI and logic
export default function Authentication() {
  const navigate = useNavigate();
  // react returns a pair: current value and function to update it
  // react stores an actual value internally in its state memory, not in these variables directly
  // const is so that we don't accidentally reassign the variable itself =

  const [tab, setTab] = useState("login"); // track which tab is active: "login" or "signup"
  const [cachedUser, setCachedUser] = useState(null);   // store cached username if JWT still valid
  const [checkingCache, setCheckingCache] = useState(true); // wait flag to avoid UI flashing

  // login form state
  const [identifier, setIdentifier] = useState("");     // email or username input
  const [loginPassword, setLoginPassword] = useState("");
  const [loginBusy, setLoginBusy] = useState(false);    // disables button while logging in
  const [loginMsg, setLoginMsg] = useState(null);       // holds feedback message (success/error)

  // sigup form state
  const [email, setEmail] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [signupBusy, setSignupBusy] = useState(false);
  const [signupMsg, setSignupMsg] = useState(null);

  // creates a regex for email validation, memoized to avoid recreation on each render
  const emailRe = useMemo(() => /^[^\s@]+@[^\s@]+\.[^\s@]+$/, []);

  // DOM Element: a live object that represents something drawn on the screen.

  // Refs to the first input fields of each form (used to auto-focus when switching tabs)
  const loginFirstFieldRef = useRef(null);
  const signupFirstFieldRef = useRef(null);

  // Automatically focus the correct input field whenever the active tab changes
  // After every render where the value of tab changes, run this function
  useEffect(() => {
  // If Login tab is active, focus the login input field
  if (tab === "login" && loginFirstFieldRef.current)
    loginFirstFieldRef.current.focus();

  // If Signup tab is active, focus the signup input field
  if (tab === "signup" && signupFirstFieldRef.current)
    signupFirstFieldRef.current.focus();

  }, [tab]); // Re-run this effect whenever 'tab' changes

  // HANDLE CACHED JWT SESSION ON LOAD
  // This useEffect runs once on component load to check if a saved JWT session exists and verify it with the backend
  useEffect(() => {
    // load token and username from local storage, localStorage is a browser API for persistent key-value storage
    const token = localStorage.getItem("token"); // get saved JWT from local storage
    const username = localStorage.getItem("username"); // get saved username
    // if no token in cache, user is not logged in, so skip the fetch, stop here
    if (!token) {
      setCheckingCache(false);
      return;
    }
    // if there is a token, verify token with backend
    // send GET request to /api/auth/verify-session to check if token is valid
    fetch(API.verify, {
      method: "GET",
      headers: { Authorization: `Bearer ${token}` },
    })

    // if backend says token is valid, set cachedUser to username to show welcome screen
      .then((res) => res.json())
      .then((data) => {
        if (data.ok) {
          setCachedUser(username); // valid token, show welcome screen
        } else {
          // invalid or expired token, clear cache
          localStorage.removeItem("token");
          localStorage.removeItem("username");
        }
      })
      .catch(() => {
        // any network error, clear cache
        localStorage.removeItem("token");
        localStorage.removeItem("username");
      })
      .finally(() => setCheckingCache(false)); // always stop loading state
  }, []);

  // HANDLE LOGIN SUBMIT
  async function handleLogin(e) {
    e.preventDefault(); // prevent full-page reload
    if (loginBusy) return; // prevent double click
    setLoginMsg(null); // clear old messages, like "Invalid password" for example

    // basic input validation
    if (!identifier.trim() || !loginPassword.trim()) {
      setLoginMsg({ type: "error", text: "Please enter identifier and password." });
      return;
    }

    setLoginBusy(true); // disable button while logging in, prevents spamming
    // try-finally to ensure we always reset button state
    try {
      // send login request
      // await pauses execution until promise resolves, wait here until postJSON finishes
      const { ok, status, data } = await postJSON(
        API.login,
        { identifier: identifier.trim(), password: loginPassword }
      ).catch(err => ({
        ok: false,
        status: 0,
        data: { error: err.name === "AbortError" ? "Request timed out" : "Network error" }
      }));

      if (ok) {
        setLoginMsg({ type: "success", text: "Logged in successfully." });

        // store JWT + username for cached login next time
        localStorage.setItem("token", data.token);
        localStorage.setItem("username", identifier.trim());
        navigate("/app"); // redirect after success, specifically to CalculatorPage in this case 
      } else { // the login failed
        const serverMsg = data?.error || data?.message || null; // extract server error message if available, it reads from the response data from the backend to determine the error message
        // update message state to show error to user
        setLoginMsg({
          type: "error",
          text: serverMsg || (status === 0 ? "Network error" : "Login failed"),
        });
      }
    } finally {
      setLoginBusy(false); // always reset button state
    }
  }

  // HANDLE SIGNUP SUBMIT
  async function handleSignup(e) {
    e.preventDefault();
    if (signupBusy) return;
    setSignupMsg(null);

    // validation checks
    if (!emailRe.test(email.trim())) {
      setSignupMsg({ type: "error", text: "Enter a valid email." });
      return;
    }
    if (!/^[a-zA-Z0-9_]{3,32}$/.test(username.trim())) {
      setSignupMsg({ type: "error", text: "Username must be 3–32 chars of letters, numbers, underscore." });
      return;
    }
    if (password.length < 8) {
      setSignupMsg({ type: "error", text: "Password must be at least 8 characters." });
      return;
    }
    setSignupBusy(true);
    try {
      const { ok, status, data } = await postJSON(API.signup, {
        email: email.trim(),
        username: username.trim(),
        password,
        firstName: firstName.trim(),
        lastName: lastName.trim(),
      }).catch(err => ({
        ok: false,
        status: 0,
        data: { error: err.name === "AbortError" ? "Request timed out" : "Network error" }
      }));

      if (ok) {
        setSignupMsg({ type: "success", text: "Account created! You can log in now." });
      } else {
        const serverMsg =
          typeof data?.error === "string" ? data.error :
          typeof data?.message === "string" ? data.message : null;

        setSignupMsg({
          type: "error",
          text: serverMsg || (status === 0 ? "Network error" : "Sign-up failed"),
        });
      }
    } finally {
      setSignupBusy(false);
    }
  }

  // Conditional rendering for cached JWT session or login form
  // if still checking cache, show nothing to avoid flicker
  if (checkingCache) return null;
  // if cachedUser exists, show "Welcome back" overlay
  if (cachedUser) {
    return (
      <div className="auth-root">
        <div className="auth-card cached-card">
          <h2>Welcome back, {cachedUser}!</h2>
          <p>You’re still signed in.</p>
          <div className="cached-actions">
            <button className="auth-btn" onClick={() => navigate("/app")}>
              Continue
            </button>
            <button
              className="auth-btn logout-btn"
              onClick={() => {
                localStorage.removeItem("token");
                localStorage.removeItem("username");
                setCachedUser(null);
                navigate("/"); // return to login screen
              }}
            >
              Not You?
            </button>
          </div>
        </div>
      </div>
    );
  }

  // MAIN AUTH UI (LOGIN + SIGNUP FORMS)
  return (
    <div className="auth-root">
      <div className="auth-card">
        {/* ===== Tabs: Login / Signup ===== */}
        <div className="auth-tabs" role="tablist" aria-label="Authentication tabs">
          {/* --- Login Tab Button --- */}
          <button
            id="login-tab"
            className={`auth-tab ${tab === "login" ? "active" : ""}`} // active class if this tab is selected
            onClick={() => setTab("login")}
            type="button"
            role="tab" // tells the computer’s accessibility tools what the element is supposed to be
            aria-selected={tab === "login"} // tells screen readers if this tab is selected
            aria-controls="login-panel" // tells screen readers which panel this tab controls
            tabIndex={tab === "login" ? 0 : -1} // makes tab focusable only if selected
          >
            Login
          </button>

          {/* --- Signup Tab Button --- */}
          <button
            id="signup-tab"
            className={`auth-tab ${tab === "signup" ? "active" : ""}`} // active class if this tab is selected
            onClick={() => setTab("signup")}
            type="button"
            role="tab" 
            aria-selected={tab === "signup"}
            aria-controls="signup-panel"
            tabIndex={tab === "signup" ? 0 : -1}
          >
            Sign Up
          </button>
        </div>

        {/* ===== Conditional Panels ===== */}
        {tab === "login" ? (
          // LOGIN FORM PANEL
          <form
            className="auth-form"
            onSubmit={handleLogin}
            id="login-panel"
            role="tabpanel"
            aria-labelledby="login-tab"
          >
            <label className="auth-label" htmlFor="login-identifier">Email or Username</label>
            <input
              id="login-identifier"
              className="auth-input"
              type="text"
              value={identifier}
              onChange={(e) => setIdentifier(e.target.value)}
              placeholder="email@example.com or username"
              autoComplete="username"
              required
              ref={loginFirstFieldRef}
            />

            <label className="auth-label" htmlFor="login-password">Password</label>
            <input
              id="login-password"
              className="auth-input"
              type="password"
              value={loginPassword}
              onChange={(e) => setLoginPassword(e.target.value)}
              placeholder="••••••••"
              autoComplete="current-password"
              required
              minLength={1}
            />

            {loginMsg && (
              <div className={`auth-msg ${loginMsg.type}`} role="alert" aria-live="polite">
                {loginMsg.text}
              </div>
            )}

            <button className="auth-btn" type="submit" disabled={loginBusy}>
              {loginBusy ? "Logging in..." : "Log In"}
            </button>
          </form>
        ) : (
          
          // SIGNUP FORM PANEL
          <form
            className="auth-form"
            onSubmit={handleSignup}
            id="signup-panel"
            role="tabpanel"
            aria-labelledby="signup-tab"
          >
            <label className="auth-label" htmlFor="signup-email">Email</label>
            <input
              id="signup-email"
              className="auth-input"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="email@example.com"
              autoComplete="email"
              required
            />

            <label className="auth-label" htmlFor="signup-username">Username</label>
            <input
              id="signup-username"
              className="auth-input"
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="your_username"
              autoComplete="username"
              required
              minLength={3}
              maxLength={32}
              pattern="^[a-zA-Z0-9_]{3,32}$"
            />

            <label className="auth-label" htmlFor="signup-password">Password</label>
            <input
              id="signup-password"
              className="auth-input"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="at least 8 characters"
              autoComplete="new-password"
              required
              minLength={8}
            />

            {/* First/Last name fields */}
            <div className="auth-name-grid">
              <div>
                <label className="auth-label" htmlFor="signup-first">First name</label>
                <input
                  id="signup-first"
                  className="auth-input"
                  type="text"
                  value={firstName}
                  onChange={(e) => setFirstName(e.target.value)}
                  placeholder="Alex"
                  autoComplete="given-name"
                  ref={signupFirstFieldRef}
                />
              </div>
              <div>
                <label className="auth-label" htmlFor="signup-last">Last name</label>
                <input
                  id="signup-last"
                  className="auth-input"
                  type="text"
                  value={lastName}
                  onChange={(e) => setLastName(e.target.value)}
                  placeholder="Osorio"
                  autoComplete="family-name"
                />
              </div>
            </div>

            {signupMsg && (
              <div className={`auth-msg ${signupMsg.type}`} role="alert" aria-live="polite">
                {signupMsg.text}
              </div>
            )}

            <button className="auth-btn" type="submit" disabled={signupBusy}>
              {signupBusy ? "Creating..." : "Create Account"}
            </button>
          </form>
        )}
      </div>
    </div>
  );
}