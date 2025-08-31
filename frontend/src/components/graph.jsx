import { useState } from "react";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  Legend,
  ResponsiveContainer,
  CartesianGrid,
} from "recharts";

export default function RothIRACalculator() {
  const [useSavedBalance, setUseSavedBalance] = useState(false);
  const [useSavedAge, setUseSavedAge] = useState(false);
  const [startingBalance, setStartingBalance] = useState(0);
  const [age, setAge] = useState(29);
  const [contribution, setContribution] = useState(5000);
  const [contributionFreq, setContributionFreq] = useState("yearly");
  const [retireAge, setRetireAge] = useState(65);
  const [rate, setRate] = useState(7);

  // Demo: Assume saved user data
  const savedIncome = 10000;
  const savedAge = 30;

  const getAnnualContribution = () => {
    if (contributionFreq === "daily") return contribution * 365;
    if (contributionFreq === "monthly") return contribution * 12;
    return contribution;
  };

  const calculateData = () => {
    let data = [];
    let balance = useSavedBalance ? savedIncome : Number(startingBalance);
    let yearlyContribution = getAnnualContribution();
    let totalContributed = 0;
    let currentAge = useSavedAge ? savedAge : Number(age);

    for (let year = currentAge; year <= retireAge; year++) {
      totalContributed += yearlyContribution;
      balance = balance * (1 + rate / 100) + yearlyContribution;
      let growth = balance - totalContributed;

      data.push({
        age: year,
        balance: balance,
        contributed: totalContributed,
        growth: growth,
      });
    }
    return data;
  };

  const data = calculateData();
  const finalBalance = data[data.length - 1]?.balance || 0;

  return (
    <div className="p-6 max-w-6xl mx-auto">
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* Left Side - Inputs */}
        <div className="p-4 border rounded-lg space-y-4">
          {/* Starting Balance */}
          <div>
            <label className="font-semibold">Starting Balance</label>
            <div className="flex items-center gap-2">
              <input
                type="number"
                value={useSavedBalance ? savedIncome : startingBalance}
                onChange={(e) => setStartingBalance(e.target.value)}
                disabled={useSavedBalance}
                className="border p-2 rounded w-full"
              />
              <input
                type="checkbox"
                checked={useSavedBalance}
                onChange={(e) => setUseSavedBalance(e.target.checked)}
              />
              <span className="text-sm">Use saved income</span>
            </div>
          </div>

          {/* Age */}
          <div>
            <label className="font-semibold">How old are you?</label>
            <div className="flex items-center gap-2">
              <input
                type="number"
                value={useSavedAge ? savedAge : age}
                onChange={(e) => setAge(e.target.value)}
                disabled={useSavedAge}
                className="border p-2 rounded w-full"
              />
              <input
                type="checkbox"
                checked={useSavedAge}
                onChange={(e) => setUseSavedAge(e.target.checked)}
              />
              <span className="text-sm">Use saved age</span>
            </div>
          </div>

          {/* Contribution */}
          <div>
            <label className="font-semibold">Contribution Amount</label>
            <div className="flex gap-2">
              <input
                type="number"
                value={contribution}
                onChange={(e) => setContribution(e.target.value)}
                className="border p-2 rounded w-full"
              />
              <select
                value={contributionFreq}
                onChange={(e) => setContributionFreq(e.target.value)}
                className="border p-2 rounded"
              >
                <option value="daily">Daily</option>
                <option value="monthly">Monthly</option>
                <option value="yearly">Yearly</option>
              </select>
            </div>
          </div>

          {/* Retirement Age */}
          <div>
            <label className="font-semibold">Retirement Age</label>
            <input
              type="number"
              value={retireAge}
              onChange={(e) => setRetireAge(e.target.value)}
              className="border p-2 rounded w-full"
            />
          </div>

          {/* Rate */}
          <div>
            <label className="font-semibold">Expected Rate of Return (%)</label>
            <input
              type="number"
              value={rate}
              onChange={(e) => setRate(e.target.value)}
              className="border p-2 rounded w-full"
            />
          </div>
        </div>

        {/* Right Side - Results */}
        <div className="md:col-span-2">
          <div className="p-6 border rounded-lg">
            <h2 className="text-xl font-bold mb-4">
              Estimated IRA Balance: ${finalBalance.toLocaleString()}
            </h2>

            <ResponsiveContainer width="100%" height={400}>
              <LineChart data={data}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="age" />
                <YAxis />
                <Tooltip />
                <Legend />
                <Line
                  type="monotone"
                  dataKey="balance"
                  stroke="#2563eb"
                  name="Total Balance"
                />
                <Line
                  type="monotone"
                  dataKey="contributed"
                  stroke="#16a34a"
                  name="Contributed"
                />
                <Line
                  type="monotone"
                  dataKey="growth"
                  stroke="#f59e0b"
                  name="Growth"
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>
    </div>
  );
}
