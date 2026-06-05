import { useEffect, useRef, useState } from "react";
import { pickRandomCity, getCountries, confirmPick, type RandomCityResponse } from "../api";

interface CityPickerProps {
  onPickConfirmed?: () => void;
}

export default function CityPicker({ onPickConfirmed }: CityPickerProps) {
  const [countries, setCountries] = useState<string[]>([]);
  const [country, setCountry] = useState("");
  const [minPop, setMinPop] = useState("200");
  const [maxPop, setMaxPop] = useState("50000");
  const [result, setResult] = useState<RandomCityResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [confirming, setConfirming] = useState(false);
  const [error, setError] = useState("");
  const abortControllerRef = useRef<AbortController | null>(null);

  useEffect(() => {
    getCountries()
      .then((codes) => {
        setCountries(codes);
        if (codes.includes("FR")) {
          setCountry("FR");
        } else if (codes.length > 0) {
          setCountry(codes[0]);
        }
      })
      .catch(() => {
        setCountries([]);
      });
  }, []);

  const handlePick = async () => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }
    const controller = new AbortController();
    abortControllerRef.current = controller;

    setLoading(true);
    setError("");
    try {
      const data = await pickRandomCity(
        country || "FR",
        minPop === "" ? 0 : Number(minPop),
        maxPop === "" ? 9999999 : Number(maxPop),
        controller.signal
      );
      setResult(data);
    } catch (err: any) {
      if (err.name === "AbortError") {
        return;
      }
      setError(err.message || "Failed to pick city");
    } finally {
      setLoading(false);
    }
  };

  const handleConfirm = async () => {
    if (!result) return;
    setConfirming(true);
    setError("");
    try {
      const data = await confirmPick(result.city.id);
      setResult({ ...result, pick_count: data.pick_count });
      onPickConfirmed?.();
    } catch (err: any) {
      setError(err.message || "Failed to confirm pick");
    } finally {
      setConfirming(false);
    }
  };

  return (
    <div style={{ marginBottom: "2rem" }}>
      <h2>Pick a Random City</h2>
      <div style={{ display: "flex", gap: "1rem", flexWrap: "wrap", marginBottom: "1rem" }}>
        <div>
          <label>Country</label>
          <select
            value={country}
            onChange={(e) => setCountry(e.target.value)}
            disabled={countries.length === 0}
          >
            {countries.length === 0 && (
              <option value="">No countries loaded</option>
            )}
            {countries.map((c) => (
              <option key={c} value={c}>
                {c}
              </option>
            ))}
          </select>
        </div>
        <div>
          <label>Min Population</label>
          <input
            type="number"
            value={minPop}
            onChange={(e) => setMinPop(e.target.value)}
            min={0}
          />
        </div>
        <div>
          <label>Max Population</label>
          <input
            type="number"
            value={maxPop}
            onChange={(e) => setMaxPop(e.target.value)}
            min={0}
          />
        </div>
        <div style={{ display: "flex", alignItems: "flex-end" }}>
          <button onClick={handlePick} disabled={loading || countries.length === 0}>
            {loading ? "Picking..." : "Pick Random City"}
          </button>
        </div>
      </div>

      {error && <p style={{ color: "red" }}>{error}</p>}

      {result && (
        <div
          style={{
            border: "1px solid #ccc",
            borderRadius: "8px",
            padding: "1rem",
            background: "#f9f9f9",
          }}
        >
          <h3>{result.city.name}</h3>
          {result.image_url && (
            <img
              src={result.image_url}
              alt={result.city.name}
              style={{
                maxWidth: "100%",
                maxHeight: "240px",
                borderRadius: "6px",
                marginBottom: "0.75rem",
                display: "block",
              }}
            />
          )}
          {result.summary && (
            <p style={{ lineHeight: 1.5, marginBottom: "0.75rem" }}>{result.summary}</p>
          )}
          <p>Country: {result.city.country_code}</p>
          <p>Population: {result.city.population.toLocaleString()}</p>
          <p>
            Coordinates: {result.city.latitude.toFixed(4)}, {result.city.longitude.toFixed(4)}
          </p>
          <p>
            <strong>
              Picked {result.pick_count} time{result.pick_count > 1 ? "s" : ""} total
            </strong>
          </p>
          <button onClick={handleConfirm} disabled={confirming}>
            {confirming ? "Confirming..." : "Confirm Pick"}
          </button>
        </div>
      )}
    </div>
  );
}
