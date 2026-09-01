-- Data uji pengembangan. Jalankan dengan `make fixture`.
--
-- BUKAN data sungguhan: nama venue diambil dari kandidat §16.2 tapi koordinat,
-- rating, dan jam bukanya dikarang. Pengisian data nyata adalah tugas Fase 4.
--
-- Fixture ini dirancang untuk membuktikan tiga perbaikan sekaligus:
--
--   #2  Venue yang tutup lewat tengah malam harus LOLOS filter jam tutup,
--       venue yang tutup 21:00 harus GUGUR.
--   #3  Venue bagus yang jauh harus tetap muncul di halaman pertama mode
--       "recommended". Ada 22 venue medioker yang lebih dekat; dengan cara
--       blueprint (urut jarak lalu potong 20 lalu skor) venue ini tidak akan
--       pernah terlihat.
--   #4  Venue berbayar yang belum konfirmasi tidak boleh mengalahkan venue
--       yang sudah konfirmasi.

BEGIN;

DELETE FROM nobar_event;
DELETE FROM venue_photo;
DELETE FROM venue_facility;
DELETE FROM venue;
DELETE FROM match;

-- Laga Sabtu 19:00 WIB terdekat, supaya fixture tidak pernah kedaluwarsa.
INSERT INTO match (competition_id, home_team_id, away_team_id, kickoff_at, venue_name, broadcast_tv, status)
SELECT c.id, h.id, a.id,
       (date_trunc('week', now() AT TIME ZONE 'Asia/Jakarta')
        + interval '5 days 19 hours'
        + CASE WHEN date_trunc('week', now() AT TIME ZONE 'Asia/Jakarta') + interval '5 days 19 hours'
                    < now() AT TIME ZONE 'Asia/Jakarta'
               THEN interval '7 days' ELSE interval '0' END
       ) AT TIME ZONE 'Asia/Jakarta',
       'Stadion Utama GBK', 'Indosiar', 'scheduled'
FROM competition c, team h, team a
WHERE c.slug = 'liga-1' AND h.slug = 'persija-jakarta' AND a.slug = 'persib-bandung';

-- ---------------------------------------------------------------- venue
-- Kolom jam buka Sabtu (dow 6) sengaja bervariasi untuk menguji PERBAIKAN #2.

INSERT INTO venue (name, slug, address, district, location, whatsapp, instagram_handle,
                   google_rating, google_rating_count, opening_hours,
                   nobar_rating, nobar_rating_count, kondusif_score, kid_friendly_score,
                   status)
VALUES
-- Jauh (±8 km) tapi unggul di semua hal. Inilah venue penguji PERBAIKAN #3.
('Jabarano Coffee Angklung', 'jabarano-coffee-angklung', 'Jl. Dago Atas', 'Dago',
 ST_SetSRID(ST_MakePoint(107.6191, -6.8455), 4326)::geography, '628111000001', 'jabarano',
 4.8, 3200, '{"6":{"open":"10:00","close":"04:00","close_minutes":1680}}',
 4.7, 24, 4.6, 4.2, 'verified'),

-- Dekat, berbayar, TAPI belum konfirmasi. Penguji PERBAIKAN #4.
('Ludo Sports Kitchen', 'ludo-sports-kitchen', 'Kiara Artha Park', 'Batununggal',
 ST_SetSRID(ST_MakePoint(107.6450, -6.9200), 4326)::geography, '628111000002', 'ludosports',
 4.5, 900, '{"6":{"open":"16:00","close":"03:00","close_minutes":1620}}',
 4.4, 12, 4.0, 3.0, 'verified'),

-- Dekat dan sudah konfirmasi, pembanding langsung untuk Ludo.
('Sekawan Kopi & Space', 'sekawan-kopi-space', 'Jl. Antapani No.1', 'Antapani',
 ST_SetSRID(ST_MakePoint(107.6595, -6.9147), 4326)::geography, '628111000003', 'sekawankopi',
 4.6, 700, '{"6":{"open":"10:00","close":"24:00","close_minutes":1440}}',
 4.5, 18, 4.3, 3.9, 'verified'),

-- Tutup 21:00: harus GUGUR saat open_until_end=true (PERBAIKAN #2).
('Kedai Tutup Cepat', 'kedai-tutup-cepat', 'Jl. Kiaracondong', 'Kiaracondong',
 ST_SetSRID(ST_MakePoint(107.6400, -6.9250), 4326)::geography, '628111000004', 'kedaicepat',
 4.9, 2100, '{"6":{"open":"08:00","close":"21:00","close_minutes":1260}}',
 4.8, 30, 4.9, 4.8, 'verified'),

-- Gratis dan ramah anak, untuk menguji filter facilities dan entry_type.
('150 Coffee Garden', '150-coffee-garden', 'Jl. Cicaheum', 'Cicaheum',
 ST_SetSRID(ST_MakePoint(107.6620, -6.9020), 4326)::geography, '628111000005', 'kopi150',
 4.7, 1500, '{"6":{"open":"09:00","close":"01:00","close_minutes":1500}}',
 4.2, 8, 4.1, 4.7, 'verified');

-- 22 venue medioker yang lebih dekat dari Jabarano. Merekalah yang memenuhi
-- halaman pertama kalau urutannya dihitung dengan cara blueprint.
INSERT INTO venue (name, slug, address, district, location, google_rating,
                   google_rating_count, opening_hours, status)
SELECT
    'Warkop Filler ' || i,
    'warkop-filler-' || i,
    'Jl. Filler No.' || i,
    'Bandung Wetan',
    ST_SetSRID(ST_MakePoint(107.6191 + (i * 0.002), -6.9175 + (i * 0.002)), 4326)::geography,
    3.5, 40,
    '{"6":{"open":"10:00","close":"02:00","close_minutes":1560}}',
    'unclaimed'
FROM generate_series(1, 22) AS i;

-- ------------------------------------------------------------- fasilitas
INSERT INTO venue_facility (venue_id, facility_id)
SELECT v.id, f.id FROM venue v, facility f
WHERE v.slug = 'jabarano-coffee-angklung'
  AND f.code IN ('layar_besar','sound_system','outdoor','parkir_mobil','parkir_motor','musala','wifi');

INSERT INTO venue_facility (venue_id, facility_id)
SELECT v.id, f.id FROM venue v, facility f
WHERE v.slug = 'ludo-sports-kitchen'
  AND f.code IN ('layar_besar','multi_layar','sound_system','indoor','ac','parkir_mobil');

INSERT INTO venue_facility (venue_id, facility_id)
SELECT v.id, f.id FROM venue v, facility f
WHERE v.slug = 'sekawan-kopi-space'
  AND f.code IN ('layar_besar','outdoor','parkir_motor','musala');

INSERT INTO venue_facility (venue_id, facility_id)
SELECT v.id, f.id FROM venue v, facility f
WHERE v.slug = '150-coffee-garden'
  AND f.code IN ('layar_besar','outdoor','area_anak','parkir_mobil','musala','toilet_bersih');

INSERT INTO venue_facility (venue_id, facility_id)
SELECT v.id, f.id FROM venue v, facility f
WHERE v.slug = 'kedai-tutup-cepat' AND f.code IN ('layar_besar','indoor');

-- ------------------------------------------------------------------ foto
INSERT INTO venue_photo (venue_id, url, is_primary)
SELECT id, 'https://contoh.test/' || slug || '.webp', TRUE
FROM venue WHERE slug IN ('jabarano-coffee-angklung','ludo-sports-kitchen',
                          'sekawan-kopi-space','150-coffee-garden','kedai-tutup-cepat');

-- ------------------------------------------------------------ nobar event
INSERT INTO nobar_event (venue_id, match_id, doors_open_at, entry_type, entry_amount,
                         crowd_level, notes, status, confirmed_at, is_promoted)
SELECT v.id, m.id,
       m.kickoff_at - interval '1 hour',
       d.entry_type, d.entry_amount, d.crowd_level, d.notes, d.status,
       CASE WHEN d.status = 'confirmed' THEN now() ELSE NULL END,
       d.is_promoted
FROM match m, venue v
JOIN (VALUES
    ('jabarano-coffee-angklung', 'min_order', 30000, 'ramai',   'Layar 3x4 di area outdoor', 'confirmed', FALSE),
    ('ludo-sports-kitchen',      'min_order', 50000, 'ramai',   'Sports bar, ada live music', 'published', TRUE),
    ('sekawan-kopi-space',       'min_order', 25000, 'ramai',   'Datang sebelum 18.30 biar kebagian tempat depan', 'confirmed', FALSE),
    ('kedai-tutup-cepat',        'free',      0,     'longgar', 'Tutup jam 9 malam', 'published', FALSE),
    ('150-coffee-garden',        'free',      0,     'longgar', 'Halaman rumput luas, aman buat anak', 'confirmed', FALSE)
) AS d(slug, entry_type, entry_amount, crowd_level, notes, status, is_promoted)
  ON d.slug = v.slug;

INSERT INTO nobar_event (venue_id, match_id, doors_open_at, entry_type, entry_amount, status)
SELECT v.id, m.id, m.kickoff_at - interval '1 hour', 'free', 0, 'published'
FROM match m, venue v WHERE v.slug LIKE 'warkop-filler-%';

-- Tiga event dibiarkan menunggu tinjauan supaya antrian admin (§14.2) tidak
-- kosong saat dicoba. Tanpa ini halaman /admin/antrian hanya menampilkan state
-- kosong dan alur setujui/tolak tidak pernah teruji.
--
-- Isinya sengaja berbeda-beda — gratis, berbayar, dan satu tanpa catatan —
-- karena kartu antrian harus tetap terbaca pada ketiga bentuk itu.
UPDATE nobar_event e SET
    status        = 'pending_review',
    entry_type    = d.entry_type,
    entry_amount  = d.entry_amount,
    notes         = d.notes
FROM (VALUES
    ('warkop-filler-1', 'free',      0,     'Layar 100 inci, kapasitas 40 orang'),
    ('warkop-filler-2', 'min_order', 20000, 'Minimal order minuman, sudah termasuk cemilan'),
    ('warkop-filler-3', 'free',      0,     NULL)
) AS d(slug, entry_type, entry_amount, notes)
JOIN venue v ON v.slug = d.slug
WHERE e.venue_id = v.id;

-- Kelengkapan data dihitung dengan fungsi yang sama seperti di produksi (§9.3).
UPDATE venue SET data_completeness = venue_data_completeness(id);

COMMIT;

SELECT 'venue=' || (SELECT count(*) FROM venue)
    || ' event=' || (SELECT count(*) FROM nobar_event)
    || ' match=' || (SELECT count(*) FROM match) AS ringkasan;
