-- Venue, fasilitas, dan foto (blueprint §7.2)

CREATE TABLE venue (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name                VARCHAR(150) NOT NULL,
    slug                VARCHAR(180) UNIQUE NOT NULL,
    address             TEXT NOT NULL,
    district            VARCHAR(80),        -- kecamatan
    city                VARCHAR(80) NOT NULL DEFAULT 'Kota Bandung',
    location            GEOGRAPHY(POINT, 4326) NOT NULL,
    phone               VARCHAR(30),
    whatsapp            VARCHAR(30),
    instagram_handle    VARCHAR(60),
    google_place_id     VARCHAR(120) UNIQUE,
    google_rating       NUMERIC(2,1),
    google_rating_count INT,

    opening_hours       JSONB,

    -- Skor internal, dihitung ulang tiap malam oleh worker (§11.3)
    nobar_rating        NUMERIC(3,2),
    nobar_rating_count  INT NOT NULL DEFAULT 0,
    kondusif_score      NUMERIC(3,2),
    kid_friendly_score  NUMERIC(3,2),
    data_completeness   NUMERIC(3,2) NOT NULL DEFAULT 0,  -- 0..1, §9.3

    status              VARCHAR(20) NOT NULL DEFAULT 'unclaimed',
    owner_user_id       UUID REFERENCES app_user(id) ON DELETE SET NULL,
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,

    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT venue_status_valid CHECK (
        status IN ('unclaimed', 'claimed', 'verified', 'suspended')
    ),
    CONSTRAINT venue_completeness_range CHECK (data_completeness BETWEEN 0 AND 1)
);

-- Bentuk JSONB opening_hours yang disepakati, sudah menampung PERBAIKAN #2.
--
-- Blueprint §9.4 memfilter jam tutup dengan `close::time > (kickoff + 2 jam)::time`.
-- Untuk venue yang tutup lewat tengah malam perbandingan itu jadi `02:00 > 21:00`
-- = false, sehingga venue justru terbuang — padahal Jabarano (04:00),
-- Rooftop Coffee (03:00), Ludo (03:00), Barrack (02:00) dan Grow (02:00) adalah
-- kandidat terbaik di §16.2. Ditambah '24:00'::time melempar error di Postgres.
--
-- Karena itu jam tutup disimpan sebagai menit sejak tengah malam hari buka,
-- dan boleh melebihi 1440 kalau tutupnya keesokan hari:
--   tutup 23:00 -> 1380 | tutup 24:00 -> 1440 | tutup 02:00 -> 1560
-- Perbandingan di query §9.4 dilakukan dalam menit, bukan tipe `time`.
--
-- Format per hari (0 = Minggu ... 6 = Sabtu):
--   {"0": {"open": "08:00", "close": "23:00", "close_minutes": 1380}, ...}
-- Hari yang tutup penuh: {"1": null}
COMMENT ON COLUMN venue.opening_hours IS
    'Jam operasional per hari, kunci 0=Minggu..6=Sabtu. close_minutes = menit sejak tengah malam hari buka, >1440 bila tutup keesokan hari. Lihat migrasi 000004.';

CREATE INDEX idx_venue_location  ON venue USING GIST(location);
CREATE INDEX idx_venue_district  ON venue(district);
CREATE INDEX idx_venue_name_trgm ON venue USING GIN(name gin_trgm_ops);

CREATE TRIGGER trg_venue_updated_at
    BEFORE UPDATE ON venue
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Fasilitas dibuat tabel terpisah supaya bisa nambah tanpa migrasi (§7.3)
CREATE TABLE facility (
    id          SERIAL PRIMARY KEY,
    code        VARCHAR(40) UNIQUE NOT NULL,
    label       VARCHAR(80) NOT NULL,
    icon        VARCHAR(40),
    category    VARCHAR(40)      -- 'nonton' | 'kenyamanan' | 'akses'
);

-- 13 baris ini ikut migrasi, bukan file seed: kodenya dirujuk langsung oleh
-- filter API (§8.2 param `facilities`) sehingga bagian dari kontrak skema.
INSERT INTO facility (code, label, category) VALUES
('layar_besar',    'Layar besar / proyektor', 'nonton'),
('multi_layar',    'Lebih dari 1 layar',      'nonton'),
('sound_system',   'Sound system memadai',    'nonton'),
('indoor',         'Area indoor',             'kenyamanan'),
('outdoor',        'Area outdoor',            'kenyamanan'),
('ac',             'Ber-AC',                  'kenyamanan'),
('area_anak',      'Ramah anak',              'kenyamanan'),
('musala',         'Musala',                  'kenyamanan'),
('toilet_bersih',  'Toilet bersih',           'kenyamanan'),
('parkir_mobil',   'Parkir mobil',            'akses'),
('parkir_motor',   'Parkir motor',            'akses'),
('non_smoking',    'Ada area non-smoking',    'kenyamanan'),
('wifi',           'WiFi',                    'kenyamanan');

CREATE TABLE venue_facility (
    venue_id    UUID REFERENCES venue(id) ON DELETE CASCADE,
    facility_id INT  REFERENCES facility(id),
    note        VARCHAR(200),
    PRIMARY KEY (venue_id, facility_id)
);

CREATE TABLE venue_photo (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    venue_id    UUID NOT NULL REFERENCES venue(id) ON DELETE CASCADE,
    url         TEXT NOT NULL,
    caption     VARCHAR(200),
    is_primary  BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order  SMALLINT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_venue_photo_venue ON venue_photo(venue_id, sort_order);

-- Hanya boleh ada satu foto utama per venue; kartu daftar (§13.3) memakainya
-- sebagai thumbnail dan tidak boleh ambigu.
CREATE UNIQUE INDEX idx_venue_photo_one_primary
    ON venue_photo(venue_id) WHERE is_primary;
