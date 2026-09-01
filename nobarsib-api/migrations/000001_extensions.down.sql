DROP FUNCTION IF EXISTS set_updated_at();

-- Ekstensi sengaja TIDAK di-drop.
--
-- Image postgis/postgis sudah memasang postgis_tiger_geocoder dan
-- postgis_topology yang bergantung pada postgis, sehingga
-- `DROP EXTENSION postgis` gagal dan menandai database sebagai dirty:
--
--   pq: cannot drop extension postgis because other objects depend on it
--   error: Dirty database version -1. Fix and force version.
--
-- Memakai CASCADE akan ikut menghapus ekstensi milik image tersebut — merusak
-- lebih banyak daripada yang dipulihkan. Ekstensi juga infrastruktur bersama
-- setingkat database, bukan milik satu migrasi: membiarkannya tetap terpasang
-- tidak menghalangi `migrate up` berikutnya karena semuanya memakai
-- CREATE EXTENSION IF NOT EXISTS.
--
-- Kalau benar-benar ingin database kosong, buang volume-nya: `make reset`.
