import { useState } from "react";
import CityPicker from "./components/CityPicker";
import PickHistory from "./components/PickHistory";

function App() {
  const [historyKey, setHistoryKey] = useState(0);

  return (
    <div style={{ maxWidth: "900px", margin: "0 auto", padding: "2rem" }}>
      <h1>Random City Picker</h1>
      <CityPicker onPickConfirmed={() => setHistoryKey((k) => k + 1)} />
      <PickHistory refreshKey={historyKey} />
    </div>
  );
}

export default App;
