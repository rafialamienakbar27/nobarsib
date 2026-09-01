-- Ekstensi yang dipakai seluruh skema (blueprint §7.2)
CREATE EXTENSION IF NOT EXISTS postgis;      -- query geospasial §9.1
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";  -- uuid_generate_v4()
CREATE EXTENSION IF NOT EXISTS pg_trgm;      -- pencarian nama venue §8.2, matching §10.4

-- Trigger pembantu: menjaga updated_at benar-benar ikut berubah.
-- Blueprint mendeklarasikan kolom updated_at dengan DEFAULT now() tapi tidak
-- pernah memperbaruinya, sehingga nilainya akan selamanya sama dengan created_at.
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
