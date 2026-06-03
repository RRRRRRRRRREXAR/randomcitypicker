import { useEffect, useState } from "react";
import { getPickedCities, resetPicks, type CityPick } from "../api";

interface PickHistoryProps {
  refreshKey?: number;
}

export default function PickHistory({ refreshKey }: PickHistoryProps) {
  const [picks, setPicks] = useState<CityPick[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const fetchPicks = async () => {
    setLoading(true);
    try {
      const data = await getPickedCities();
      setPicks(data ?? []);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchPicks();
  }, []);

  useEffect(() => {
    if (refreshKey !== undefined) {
      fetchPicks();
    }
  }, [refreshKey]);

  const handleReset = async () => {
    if (!confirm("Reset all pick history?")) return;
    try {
      await resetPicks();
      setPicks([]);
    } catch (err: any) {
      setError(err.message);
    }
  };

  return (
    <div>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
        <h2>Pick History</h2>
        <button onClick={handleReset} disabled={loading}>
          Reset History
        </button>
      </div>
      {error && <p style={{ color: "red" }}>{error}</p>}
      {picks?.length === 0 ? (
        <p>No cities picked yet.</p>
      ) : (
        <table style={{ width: "100%", borderCollapse: "collapse" }}>
          <thead>
            <tr>
              <th style={{ textAlign: "left", borderBottom: "1px solid #ccc" }}>City</th>
              <th style={{ textAlign: "left", borderBottom: "1px solid #ccc" }}>Country</th>
              <th style={{ textAlign: "right", borderBottom: "1px solid #ccc" }}>Population</th>
              <th style={{ textAlign: "right", borderBottom: "1px solid #ccc" }}>Times Picked</th>
              <th style={{ textAlign: "left", borderBottom: "1px solid #ccc" }}>Last Picked</th>
            </tr>
          </thead>
          <tbody>
            {picks?.map((p) => (
              <tr key={p.city_id}>
                <td style={{ borderBottom: "1px solid #eee" }}>{p.city.name}</td>
                <td style={{ borderBottom: "1px solid #eee" }}>{p.city.country_code}</td>
                <td style={{ textAlign: "right", borderBottom: "1px solid #eee" }}>
                  {p.city.population.toLocaleString()}
                </td>
                <td style={{ textAlign: "right", borderBottom: "1px solid #eee" }}>{p.pick_count}</td>
                <td style={{ borderBottom: "1px solid #eee" }}>
                  {new Date(p.last_picked_at).toLocaleString()}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
