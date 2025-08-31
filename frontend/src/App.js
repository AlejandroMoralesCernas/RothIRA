import React from "react";
import RothIRACalculator from "./components/graph";


function App() {
return (
<div style={{ padding: 40, fontFamily: "sans-serif" }}>
<h1 className="text-2xl font-bold mb-6">Roth IRA Calculator</h1>
<RothIRACalculator />
</div>
);
}


export default App;