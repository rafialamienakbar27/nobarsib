-- Nobar event — inti aplikasi (blueprint §7.2)

CREATE TABLE nobar_event (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    venue_id            UUID NOT NULL REFERENCES venue(id) ON DELETE CASCADE,
    match_id            BIGINT NOT NULL REFERENCES match(id) ON DELETE CASCADE,

    doors_open_at       TIMESTAMPTZ,
    entry_type          VARCHAR(30) NOT NULL DEFAULT 'free',
    entry_amount        INT NOT NULL DEFAULT 0,     -- rupiah
    capacity_estimate   INT,
    crowd_level         VARCHAR(20),
    notes               TEXT,

    status              VARCHAR(20) NOT NULL DEFAULT 'draft',
    confirmed_at        TIMESTAMPTZ,                -- diisi saat konfirmasi H-1 §15.3
    is_promoted         BOOLEAN NOT NULL DEFAULT FALSE,

    created_by          UUID REFERENCES app_user(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- Satu venue hanya boleh punya satu event per pertandingan (§7.3).
    -- Mencegah duplikasi saat impor dari IG dan submit venue bertabrakan.
    UNIQUE (venue_id, match_id),

    -- Status ditulis huruf kecil, konsisten dengan DEFAULT di §7.2.
    -- Diagram §4.5 menulisnya kapital — itu hanya gaya penulisan diagram.
    -- Validasi transisi antar status ada di layer service (Fase 2), CHECK ini
    -- hanya menjaga nilainya tetap dalam himpunan yang sah.
    CONSTRAINT nobar_event_status_valid CHECK (
        status IN ('draft', 'pending_review', 'published',
                   'confirmed', 'rejected', 'cancelled', 'finished')
    ),
    CONSTRAINT nobar_event_entry_type_valid CHECK (
        entry_type IN ('free', 'min_order', 'ticket', 'donation')
    ),
    CONSTRAINT nobar_event_crowd_level_valid CHECK (
        crowd_level IS NULL OR crowd_level IN ('longgar', 'ramai', 'penuh')
    ),
    CONSTRAINT nobar_event_entry_amount_nonneg CHECK (entry_amount >= 0),
    -- Gratis berarti nol. Mencegah kartu venue menampilkan "Gratis" dan
    -- "Rp 25.000" sekaligus (§3.1 no. 5: jangan janji yang tidak bisa ditepati).
    CONSTRAINT nobar_event_free_is_zero CHECK (
        entry_type <> 'free' OR entry_amount = 0
    ),
    -- Status 'confirmed' tidak boleh ada tanpa jejak kapan dikonfirmasi.
    -- Badge "Dikonfirmasi" (§13.3) adalah janji kepercayaan aplikasi; ia harus
    -- selalu bisa ditelusuri ke satu timestamp.
    CONSTRAINT nobar_event_confirmed_has_timestamp CHECK (
        status <> 'confirmed' OR confirmed_at IS NOT NULL
    )
);

CREATE INDEX idx_nobar_match  ON nobar_event(match_id, status);
CREATE INDEX idx_nobar_venue  ON nobar_event(venue_id);
CREATE INDEX idx_nobar_status ON nobar_event(status, confirmed_at);

CREATE TRIGGER trg_nobar_event_updated_at
    BEFORE UPDATE ON nobar_event
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Jejak asal data: dari IG, input manual, atau submit venue (§10.4 mensyaratkan
-- source_url dan raw_caption disimpan untuk penelusuran kalau ada komplain)
CREATE TABLE event_source (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nobar_event_id   UUID REFERENCES nobar_event(id) ON DELETE CASCADE,
    source_type      VARCHAR(30) NOT NULL,
    source_url       TEXT,
    raw_caption      TEXT,
    poster_image_url TEXT,
    extracted_json   JSONB,
    extracted_at     TIMESTAMPTZ,
    reviewed_by      UUID REFERENCES app_user(id) ON DELETE SET NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT event_source_type_valid CHECK (
        source_type IN ('instagram', 'manual', 'venue_portal')
    )
);

CREATE INDEX idx_event_source_event ON event_source(nobar_event_id);
