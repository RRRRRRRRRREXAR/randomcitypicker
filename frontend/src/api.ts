const API_BASE = import.meta.env.VITE_API_URL || "";

export interface City {
  id: number;
  name: string;
  country_code: string;
  population: number;
  latitude: number;
  longitude: number;
}

export interface RandomCityResponse {
  city: City;
  pick_count: number;
  summary?: string;
  image_url?: string;
}

export interface CityPick {
  city_id: number;
  pick_count: number;
  first_picked_at: string;
  last_picked_at: string;
  city: City;
}

export async function getCountries(): Promise<string[]> {
  const res = await fetch(`${API_BASE}/api/countries`);
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export async function pickRandomCity(
  country: string,
  minPop: number,
  maxPop: number,
  signal?: AbortSignal
): Promise<RandomCityResponse> {
  const params = new URLSearchParams({
    country,
    minPop: String(minPop),
    maxPop: String(maxPop),
  });
  const res = await fetch(`${API_BASE}/api/cities/random?${params}`, { signal });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export async function getPickedCities(): Promise<CityPick[]> {
  const res = await fetch(`${API_BASE}/api/cities/picked`);
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export async function confirmPick(
  cityId: number
): Promise<{ city_id: number; pick_count: number }> {
  const res = await fetch(`${API_BASE}/api/cities/confirm`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ city_id: cityId }),
  });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export async function resetPicks(): Promise<void> {
  const res = await fetch(`${API_BASE}/api/cities/reset`, { method: "POST" });
  if (!res.ok) throw new Error(await res.text());
}
