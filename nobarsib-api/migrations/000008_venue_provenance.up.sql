-- Asal-usul data venue, dan satu kolom website.
--
-- Sampai sekarang tidak ada cara membedakan venue yang datanya disalin dari
-- Google Places, dilaporkan pemiliknya, atau diketik tangan setelah survei.
-- Perbedaan itu menentukan seberapa jauh datanya boleh dipercaya: profil dari
-- Google bisa basi berbulan-bulan tanpa ada yang tahu, sementara yang datang
-- dari pemilik venue biasanya benar pada hari ia dikirim.
--
-- §21 menyebut data basi sebagai risiko fatal yang membunuh FANZO, dan §13.3
-- membangun seluruh kepercayaan aplikasi ini di atas badge "Dikonfirmasi".
-- Dua kolom di bawah adalah versi yang sama untuk data PROFIL, bukan untuk
-- pengumuman nobar: dari mana asalnya, dan kapan terakhir ada manusia yang
-- memastikannya.

ALTER TABLE venue
    ADD COLUMN website          TEXT,
    ADD COLUMN data_source      VARCHAR(20),
    ADD COLUMN last_verified_at DATE;

-- Nilainya dibatasi di tingkat basis data, bukan hanya di Go.
--
-- Kolom asal-usul yang boleh diisi teks bebas akan segera berisi "google",
-- "Google Places", "gplaces", dan "manual " berspasi — dan pertanyaan "berapa
-- banyak venue yang datanya belum pernah diverifikasi manusia" jadi mustahil
-- dijawab. NULL tetap sah: venue lama memang tidak diketahui asalnya, dan
-- mengarang nilai untuk mereka lebih buruk daripada mengakui tidak tahu.
ALTER TABLE venue
    ADD CONSTRAINT venue_data_source_valid CHECK (
        data_source IS NULL OR data_source IN ('google-places', 'venue', 'manual')
    );

COMMENT ON COLUMN venue.website IS
    'URL situs resmi venue. Opsional — sebagian besar venue di Bandung hanya punya Instagram (§12.1).';
COMMENT ON COLUMN venue.data_source IS
    'Asal data profil: google-places | venue | manual. NULL = tidak diketahui.';
COMMENT ON COLUMN venue.last_verified_at IS
    'Tanggal terakhir data profil dipastikan manusia. DATE, bukan timestamp: verifikasi adalah peristiwa harian, dan menyimpan jam-menit di sini hanya memberi kesan presisi yang tidak pernah ada.';

-- Indeks parsial untuk pertanyaan operasional yang akan sering diulang saat
-- Fase 4 berjalan: "venue mana yang paling lama tidak diverifikasi". Parsial
-- karena baris ber-NULL tidak pernah menjadi jawabannya.
CREATE INDEX idx_venue_last_verified
    ON venue(last_verified_at)
    WHERE last_verified_at IS NOT NULL;
