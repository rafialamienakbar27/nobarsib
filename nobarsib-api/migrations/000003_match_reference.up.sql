-- Referensi pertandingan (blueprint §7.2)
--
-- Prinsip §3.1 no. 2: jangan hardcode Persib. Tim dan kompetisi disimpan sebagai
-- data, bukan logika program. Menambah Timnas nanti cukup satu INSERT.

CREATE TABLE competition (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,      -- 'Liga 1', 'Piala Presiden'
    slug        VARCHAR(60) UNIQUE NOT NULL,
    country     VARCHAR(50) NOT NULL DEFAULT 'Indonesia',
    is_active   BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE team (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,      -- 'Persib Bandung'
    short_name  VARCHAR(30),                -- 'Persib'
    slug        VARCHAR(60) UNIQUE NOT NULL,
    logo_url    TEXT,
    is_featured BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE match (
    id              BIGSERIAL PRIMARY KEY,
    competition_id  INT REFERENCES competition(id),
    home_team_id    INT NOT NULL REFERENCES team(id),
    away_team_id    INT NOT NULL REFERENCES team(id),
    kickoff_at      TIMESTAMPTZ NOT NULL,
    venue_name      VARCHAR(150),           -- stadion
    broadcast_tv    VARCHAR(100),           -- 'Indosiar', 'Vidio'
    status          VARCHAR(20) NOT NULL DEFAULT 'scheduled',
    score_home      SMALLINT,
    score_away      SMALLINT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT match_status_valid CHECK (
        status IN ('scheduled', 'live', 'finished', 'postponed')
    ),
    CONSTRAINT match_teams_differ CHECK (home_team_id <> away_team_id)
);

CREATE INDEX idx_match_kickoff ON match(kickoff_at);
CREATE INDEX idx_match_teams   ON match(home_team_id, away_team_id);

CREATE TRIGGER trg_match_updated_at
    BEFORE UPDATE ON match
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
