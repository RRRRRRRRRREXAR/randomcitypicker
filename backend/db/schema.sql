CREATE TABLE IF NOT EXISTS cities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    country_code TEXT NOT NULL,
    population INTEGER,
    latitude REAL,
    longitude REAL
);

CREATE TABLE IF NOT EXISTS city_picks (
    city_id INTEGER PRIMARY KEY REFERENCES cities(id) ON DELETE CASCADE,
    pick_count INTEGER NOT NULL DEFAULT 0,
    first_picked_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_picked_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_cities_country_pop ON cities(country_code, population);
CREATE INDEX IF NOT EXISTS idx_cities_name ON cities(name);
