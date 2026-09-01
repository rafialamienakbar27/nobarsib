-- Fungsi pembantu untuk query pencarian (Fase 2).

-- =====================================================================
-- PERBAIKAN #2 — filter jam tutup yang melewati tengah malam
-- =====================================================================
--
-- Blueprint §9.4 memfilter dengan `close::time > (kickoff + 2 jam)::time`.
-- Untuk venue yang tutup lewat tengah malam perbandingan itu menjadi
-- `02:00 > 21:00` = false, sehingga yang terbuang justru kandidat terbaik di
-- §16.2: Jabarano (04:00), Rooftop (03:00), Ludo (03:00), Barrack (02:00),
-- Grow (02:00). Selain itu `'24:00'::time` melempar error di Postgres dan
-- cabang OR di blueprint tidak menyelamatkannya karena cast tetap dievaluasi.
--
-- Fungsi ini memakai close_minutes yang sudah dinormalisasi (menit sejak tengah
-- malam hari buka, >1440 bila tutup keesokan hari — lihat migrasi 000004 dan
-- paket Go internal/pkg/openhours).
--
-- Tiga keadaan dibedakan, dan pembedaan itu penting:
--   hari tidak tercatat / close_minutes kosong -> TIDAK DIKETAHUI -> lolos
--   hari tercatat bernilai null                -> TUTUP          -> gugur
--   close_minutes ada                          -> dibandingkan
--
-- Venue baru yang jam bukanya belum terisi tidak boleh disembunyikan oleh
-- filter: peringkatnya sudah dihukum lewat data_completeness (§9.3), dan §13.5
-- menegaskan data tipis adalah kondisi normal di awal.
CREATE OR REPLACE FUNCTION venue_open_until(
    hours            JSONB,
    dow              INT,      -- 0 = Minggu ... 6 = Sabtu
    required_minutes INT       -- menit sejak tengah malam hari yang sama
) RETURNS BOOLEAN AS $$
DECLARE
    day_key  TEXT := dow::text;
    day_data JSONB;
    close_m  INT;
BEGIN
    IF hours IS NULL OR NOT (hours ? day_key) THEN
        RETURN TRUE;                      -- tidak diketahui
    END IF;

    day_data := hours -> day_key;
    IF jsonb_typeof(day_data) = 'null' THEN
        RETURN FALSE;                     -- tercatat tutup di hari itu
    END IF;

    close_m := NULLIF(day_data ->> 'close_minutes', '')::int;
    IF close_m IS NULL THEN
        RETURN TRUE;                      -- tercatat tapi belum dinormalisasi
    END IF;

    -- Tegas (>), sama seperti §9.4 dan internal/pkg/openhours.OpenUntil.
    RETURN close_m > required_minutes;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

COMMENT ON FUNCTION venue_open_until(JSONB, INT, INT) IS
    'Apakah venue masih buka melewati required_minutes pada hari dow. Cermin SQL dari internal/pkg/openhours.Week.OpenUntil.';

-- =====================================================================
-- data_completeness (§9.3)
-- =====================================================================
--
-- kelengkapan = jumlah_field_terisi / total_field_penting
--
-- Enam field penting menurut §9.3: koordinat, telepon/WA, minimal 1 foto,
-- minimal 3 fasilitas, jam buka lengkap, IG handle.
--
-- Skor ini mendorong venue melengkapi profilnya tanpa perlu dipaksa, karena
-- berpengaruh langsung ke urutan tampil.
CREATE OR REPLACE FUNCTION venue_data_completeness(v_id UUID)
RETURNS NUMERIC AS $$
DECLARE
    v          venue%ROWTYPE;
    filled     INT := 0;
    total      CONSTANT INT := 6;
    hours_days INT;
BEGIN
    SELECT * INTO v FROM venue WHERE id = v_id;
    IF NOT FOUND THEN
        RETURN 0;
    END IF;

    -- 1. Koordinat. Kolomnya NOT NULL, tapi titik nol berarti belum disurvei.
    IF v.location IS NOT NULL
       AND ST_X(v.location::geometry) <> 0
       AND ST_Y(v.location::geometry) <> 0 THEN
        filled := filled + 1;
    END IF;

    -- 2. Telepon atau WhatsApp — salah satu cukup.
    IF COALESCE(NULLIF(v.whatsapp, ''), NULLIF(v.phone, '')) IS NOT NULL THEN
        filled := filled + 1;
    END IF;

    -- 3. Minimal satu foto.
    IF EXISTS (SELECT 1 FROM venue_photo WHERE venue_id = v_id) THEN
        filled := filled + 1;
    END IF;

    -- 4. Minimal tiga fasilitas.
    IF (SELECT count(*) FROM venue_facility WHERE venue_id = v_id) >= 3 THEN
        filled := filled + 1;
    END IF;

    -- 5. Jam buka lengkap: ketujuh hari tercatat (nilai null pun sah, artinya
    --    "tutup" — yang tidak boleh adalah harinya tidak disebut sama sekali).
    IF v.opening_hours IS NOT NULL THEN
        SELECT count(*) INTO hours_days
        FROM jsonb_object_keys(v.opening_hours) AS k
        WHERE k IN ('0','1','2','3','4','5','6');
        IF hours_days = 7 THEN
            filled := filled + 1;
        END IF;
    END IF;

    -- 6. Instagram handle.
    IF NULLIF(v.instagram_handle, '') IS NOT NULL THEN
        filled := filled + 1;
    END IF;

    RETURN round(filled::numeric / total, 2);
END;
$$ LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION venue_data_completeness(UUID) IS
    'Kelengkapan profil venue 0..1 menurut blueprint §9.3, dipakai sebagai komponen skor rekomendasi.';
