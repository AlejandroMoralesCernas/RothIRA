import React, { useState, useMemo, useRef, useEffect } from "react"; // useState lets a functional component keep track of state during re-renders, useMemo for memoizing regex, avoiding recreation
import { useNavigate } from "react-router-dom";
import "./authentication.css";


// (where we talk to our backend) read base URL from env, fallback to local backend port 8081
const API_BASE = "http://localhost:8080";
// Creating a box (object) that holds address of API endpoints
const API = {
  signup: `${API_BASE}/api/auth/create-user`,
  login:  `${API_BASE}/api/auth/login-user`,
  verify: `${API_BASE}/api/auth/verify-session`, // ✅ new endpoint for verifying cached session
};

// helper function to make POST requests with JSON body and parse JSON response
async function postJSON(url, body, timeoutMs = 3000) { // sends data to server (url), sends data (body) in JSON format, timeout after 3 seconds (3000 ms)
  const ctrl = new AbortController(); // to abort fetch request if it takes too long
  const id = setTimeout(() => ctrl.abort(), timeoutMs); // hey javascript. here's a function. run it later, after 3 seconds

  console.log("posting to " + url)
  try {
    const res = await fetch(url, { // send a network request to the url using these extra options
      method: "POST", // sending data to server
      headers: { "Content-Type": "application/json" }, // we are sending JSON data
      body: JSON.stringify(body), // convert JS object to JSON string
      signal: ctrl.signal, // link abort controller to this fetch request
      credentials: "include", // include cookies in request
    });

    console.log("response received", res)

    // (small robustness tweak) only try JSON if content-type says JSON; otherwise fall back to text
    const ctype = res.headers.get("content-type") || "";
    const data = ctype.includes("application/json")
      ? await res.json().catch(() => ({})) // try to parse JSON response, if fails return empty object
      : { message: await res.text().catch(() => "") };

    return { ok: res.ok, status: res.status, data }; // return whether response was ok (status 200-299), status code, and parsed data
  } finally { // runs whether success or failure
    clearTimeout(id); // clear the timeout to avoid memory leaks
  }
}

// export default makes this component the default export of this file
// main purpose is to provide the Authentication component to the rest of the app
export default function Authentication() { // declaring a React functional component named Authentication, tab is currently active tab
  const navigate = useNavigate();
  const [tab, setTab] = useState("login"); // state to track which tab is active, default to "login"

  const [cachedUser, setCachedUser] = useState(null); // ✅ new: store cached username if JWT still valid
  const [checkingCache, setCheckingCache] = useState(true); // ✅ new: prevents flashing the UI before checking token

  // login state
  const [identifier, setIdentifier] = useState("");
  const [loginPassword, setLoginPassword] = useState("");
  const [loginBusy, setLoginBusy] = useState(false);
  const [loginMsg, setLoginMsg] = useState(null);

  // signup state
  const [email, setEmail] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [signupBusy, setSignupBusy] = useState(false);
  const [signupMsg, setSignupMsg] = useState(null);

  // memoized regex for email validation
  const emailRe = useMemo(() => /^[^\s@]+@[^\s@]+\.[^\s@]+$/, []);

  // refs to move focus when switching tabs (small a11y/UX improvement)
  const loginFirstFieldRef = useRef(null);
  const signupFirstFieldRef = useRef(null);
  useEffect(() => {
    if (tab === "login" && loginFirstFieldRef.current) loginFirstFieldRef.current.focus();
    if (tab === "signup" && signupFirstFieldRef.current) signupFirstFieldRef.current.focus();
  }, [tab]);

  // ✅ NEW: On component mount, check if a valid JWT exists and is still active
  useEffect(() => {
    const token = localStorage.getItem("token");
    const username = localStorage.getItem("username");

    if (!token) {
      setCheckingCache(false);
      return;
    }

    // Verify token validity via backend
    fetch(API.verify, {
      method: "GET",
      headers: { Authorization: `Bearer ${token}` },
    })
      .then((res) => res.json())
      .then((data) => {
        if (data.ok) {
          setCachedUser(username); // ✅ valid token: show "Welcome back" screen
        } else {
          // invalid or expired → clear cache
          localStorage.removeItem("token");
          localStorage.removeItem("username");
        }
      })
      .catch(() => {
        localStorage.removeItem("token");
        localStorage.removeItem("username");
      })
      .finally(() => setCheckingCache(false));
  }, []);

  // handle login form submission
  async function handleLogin(e) {  // async because we will make network request
    e.preventDefault(); // stop page from reloading on form submit
    if (loginBusy) return; // prevent double submit if already busy
    setLoginMsg(null);  // clear previous messages

    // basic validation
    if (!identifier.trim() || !loginPassword.trim()) {
      setLoginMsg({ type: "error", text: "Please enter identifier and password." });
      return;
    }

    setLoginBusy(true);
    try {
      // make POST request to login endpoint with identifier and password
      const { ok, status, data } = await postJSON( // sends login data to backend using postJSON helper
        API.login,                 // 1st argument: URL
        {                           // 2nd argument:  <<< BODY STARTS here
          identifier: identifier.trim(),
          password: loginPassword,  // last property of the body
        },                          // <<< BODY ENDS here
      ).catch(err => ({
        ok: false,
        status: 0,
        data: {
          error: err.name === "AbortError"
            ? "Request timed out" // if fetch was aborted due to timeout
            : "Network error" // else other network error
        }
      }));

      if (ok) {
        setLoginMsg({ type: "success", text: "Logged in successfully." });

        // ✅ NEW: Save JWT token + username in localStorage for cache login
        localStorage.setItem("token", data.token);
        localStorage.setItem("username", identifier.trim());

        navigate("/app"); // redirect to app after successful login
      } else {
        // prefer server-provided message if present, else fall back to network/failed
        const serverMsg = typeof data?.error === "string" ? data.error :
                          typeof data?.message === "string" ? data.message : null;
        setLoginMsg({ type: "error", text: serverMsg || (status === 0 ? "Network error" : "Login failed") });
      }
    } finally {
      setLoginBusy(false); // ensure busy flag is turned off even if an unexpected error happens
    }
  }

  // handle signup form submission
  async function handleSignup(e) {
    e.preventDefault();
    if (signupBusy) return; // prevent double submit if already busy
    setSignupMsg(null);

    if (!emailRe.test(email.trim())) { // if the trimmed email doesn’t match the regex pattern, show an error and stop
      setSignupMsg({ type: "error", text: "Enter a valid email." });
      return;
    }
    if (!/^[a-zA-Z0-9_]{3,32}$/.test(username.trim())) { // username must be 3-32 chars of letters, numbers, underscore
      setSignupMsg({ type: "error", text: "Username must be 3–32 chars of letters, numbers, underscore." });
      return;
    }
    if (password.length < 8) { // password must be at least 8 characters
      setSignupMsg({ type: "error", text: "Password must be at least 8 characters." });
      return;
    }

    setSignupBusy(true); // indicate signup is in progress
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
        const serverMsg = typeof data?.error === "string" ? data.error :
                          typeof data?.message === "string" ? data.message : null;
        setSignupMsg({ type: "error", text: serverMsg || (status === 0 ? "Network error" : "Sign-up failed") });
      }
    } finally {
      setSignupBusy(false); // ensure busy flag is turned off even if an unexpected error happens
    }
  }

  // ✅ NEW: cached login overlay (display if valid token found)
  if (checkingCache) return null; // wait until we check if token exists/valid

  if (cachedUser) {
    return (
      <div className="auth-root">
        <div className="auth-card cached-card">
          <h2>Welcome back, {cachedUser}!</h2>
          <p>You’re still signed in.</p>
          <div className="cached-actions">
            <button
              className="auth-btn"
              onClick={() => navigate("/app")} // go straight to app
            >
              Continue
            </button>
            <button
              className="auth-btn logout-btn"
              onClick={() => {
                localStorage.removeItem("token");
                localStorage.removeItem("username");
                setCachedUser(null); // clear cached login
              }}
            >
              Not you?
            </button>
          </div>
        </div>
      </div>
    );
  }

  // ✅ Default login/signup UI (unchanged below)
  return (
    // main container for authentication component
    <div className="auth-root">
      <div className="auth-card">
        {/* make the tabs accessible */}
        <div className="auth-tabs" role="tablist" aria-label="Authentication tabs"> 
          <button // the login tab button
            id="login-tab" // link tab to its panel
            className={`auth-tab ${tab === "login" ? "active" : ""}`} // if tab is equal to "login" use "active" otherwise, use empty string
            onClick={() => setTab("login")}
            type="button" // prevent form submission on click
            role="tab" // a11y: this is a tab
            aria-selected={tab === "login"} // a11y: indicate current tab
            aria-controls="login-panel" // a11y: which panel this tab controls
            tabIndex={tab === "login" ? 0 : -1} // a11y: only active tab is tabbable
          >
            Login
          </button>
          <button
            id="signup-tab"
            className={`auth-tab ${tab === "signup" ? "active" : ""}`}
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

        {tab === "login" ? (
          // login form (same as your original)
          <form className="auth-form" onSubmit={handleLogin} id="login-panel" role="tabpanel" aria-labelledby="login-tab">
            <label className="auth-label" htmlFor="login-identifier">Email or Username</label>
            <input id="login-identifier" className="auth-input" type="text" value={identifier}
              onChange={(e) => setIdentifier(e.target.value)} placeholder="email@example.com or username"
              autoComplete="username" required ref={loginFirstFieldRef} />

            <label className="auth-label" htmlFor="login-password">Password</label>
            <input id="login-password" className="auth-input" type="password" value={loginPassword}
              onChange={(e) => setLoginPassword(e.target.value)} placeholder="••••••••"
              autoComplete="current-password" required minLength={1} />

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
          // signup form (same as your original)
          <form className="auth-form" onSubmit={handleSignup} id="signup-panel" role="tabpanel" aria-labelledby="signup-tab">
            <label className="auth-label" htmlFor="signup-email">Email</label>
            <input id="signup-email" className="auth-input" type="email" value={email}
              onChange={(e) => setEmail(e.target.value)} placeholder="email@example.com"
              autoComplete="email" required />

            <label className="auth-label" htmlFor="signup-username">Username</label>
            <input id="signup-username" className="auth-input" type="text" value={username}
              onChange={(e) => setUsername(e.target.value)} placeholder="your_username"
              autoComplete="username" required minLength={3} maxLength={32} pattern="^[a-zA-Z0-9_]{3,32}$" />

            <label className="auth-label" htmlFor="signup-password">Password</label>
            <input id="signup-password" className="auth-input" type="password" value={password}
              onChange={(e) => setPassword(e.target.value)} placeholder="at least 8 characters"
              autoComplete="new-password" required minLength={8} />

            <div className="auth-name-grid">
              <div>
                <label className="auth-label" htmlFor="signup-first">First name</label>
                <input id="signup-first" className="auth-input" type="text" value={firstName}
                  onChange={(e) => setFirstName(e.target.value)} placeholder="Alex"
                  autoComplete="given-name" ref={signupFirstFieldRef} />
              </div>
              <div>
                <label className="auth-label" htmlFor="signup-last">Last name</label>
                <input id="signup-last" className="auth-input" type="text" value={lastName}
                  onChange={(e) => setLastName(e.target.value)} placeholder="Osorio"
                  autoComplete="family-name" />
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
