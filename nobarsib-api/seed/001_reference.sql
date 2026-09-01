-- Seed data referensi. Dijalankan dengan `make seed`, idempoten (aman diulang).
--
-- Dipisah dari migrasi karena isinya berubah tiap musim (roster klub Liga 1),
-- sementara migrasi seharusnya hanya dijalankan sekali.
--
-- Prinsip §3.1 no. 2: Persib disimpan sebagai DATA, bukan logika program.
-- Menambah Timnas Indonesia nanti cukup satu INSERT di file ini.

INSERT INTO competition (name, slug, country, is_active) VALUES
('Liga 1',         'liga-1',         'Indonesia', TRUE),
('Piala Presiden', 'piala-presiden', 'Indonesia', FALSE)
ON CONFLICT (slug) DO NOTHING;

INSERT INTO team (name, short_name, slug, is_featured) VALUES
('Persib Bandung',      'Persib',    'persib-bandung',      TRUE),
('Persija Jakarta',     'Persija',   'persija-jakarta',     FALSE),
('Arema FC',            'Arema',     'arema-fc',            FALSE),
('Persebaya Surabaya',  'Persebaya', 'persebaya-surabaya',  FALSE),
('PSM Makassar',        'PSM',       'psm-makassar',        FALSE),
('Bali United',         'Bali Utd',  'bali-united',         FALSE),
('Borneo FC',           'Borneo',    'borneo-fc',           FALSE),
('Madura United',       'Madura',    'madura-united',       FALSE),
('PSIS Semarang',       'PSIS',      'psis-semarang',       FALSE),
('Persis Solo',         'Persis',    'persis-solo',         FALSE),
('Dewa United',         'Dewa',      'dewa-united',         FALSE),
('Persik Kediri',       'Persik',    'persik-kediri',       FALSE),
('Barito Putera',       'Barito',    'barito-putera',       FALSE),
('PSBS Biak',           'PSBS',      'psbs-biak',           FALSE),
('Semen Padang',        'Semen',     'semen-padang',        FALSE),
('Malut United',        'Malut',     'malut-united',        FALSE),
('Persita Tangerang',   'Persita',   'persita-tangerang',   FALSE),
('PSS Sleman',          'PSS',       'pss-sleman',          FALSE)
ON CONFLICT (slug) DO NOTHING;

-- Catatan: roster di atas perlu dicocokkan dengan peserta Liga 1 musim berjalan
-- sebelum input jadwal di Fase 4. Klub yang degradasi/promosi cukup diubah
-- di file ini, tidak perlu migrasi.
