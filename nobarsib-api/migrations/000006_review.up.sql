-- Review dan analitik ringan (blueprint §7.2)

CREATE TABLE review (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nobar_event_id  UUID REFERENCES nobar_event(id) ON DELETE SET NULL,
    venue_id        UUID NOT NULL REFERENCES venue(id) ON DELETE CASCADE,

    -- 3 pertanyaan inti (§11.2). Tiga adalah batas atas, bukan target awal.
    rating_overall  SMALLINT NOT NULL CHECK (rating_overall BETWEEN 1 AND 5),
    rating_kondusif SMALLINT CHECK (rating_kondusif BETWEEN 1 AND 5),
    is_kid_friendly BOOLEAN,

    crowd_actual    VARCHAR(20),
    comment         VARCHAR(500),

    -- Anti-spam tanpa login (§11.5): SHA256(fingerprint browser + salt server)
    device_hash     VARCHAR(64) NOT NULL,
    ip_hash         VARCHAR(64),

    is_hidden       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (nobar_event_id, device_hash),

    CONSTRAINT review_crowd_actual_valid CHECK (
        crowd_actual IS NULL OR crowd_actual IN ('sepi', 'pas', 'penuh', 'overload')
    )
);

CREATE INDEX idx_review_venue ON review(venue_id, created_at DESC);

-- Dipakai worker recalculate-scores (§11.3) yang memfilter is_hidden = FALSE
-- dan created_at dalam 18 bulan terakhir.
CREATE INDEX idx_review_scoring ON review(venue_id, created_at)
    WHERE is_hidden = FALSE;

-- Analitik ringan. Sumber data statistik venue (§15.4) dan estimasi keramaian
-- (§9.5), jadi dipasang sejak awal meski belum ada yang membacanya.
CREATE TABLE event_view (
    id              BIGSERIAL PRIMARY KEY,
    nobar_event_id  UUID REFERENCES nobar_event(id) ON DELETE CASCADE,
    device_hash     VARCHAR(64),
    action          VARCHAR(30),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT event_view_action_valid CHECK (
        action IN ('view_card', 'open_detail', 'open_maps', 'click_wa')
    )
);

CREATE INDEX idx_view_event ON event_view(nobar_event_id, action, created_at);

-- §11.5 mensyaratkan review hanya bisa dikirim kalau device tersebut pernah
-- membuka detail event itu. Index ini melayani pengecekan tersebut.
CREATE INDEX idx_view_device_detail ON event_view(device_hash, nobar_event_id)
    WHERE action = 'open_detail';
