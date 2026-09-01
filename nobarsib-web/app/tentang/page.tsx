import type { Metadata } from "next";
import Link from "next/link";

export const metadata: Metadata = {
  title: "Tentang",
  description: "Apa itu NOBARSIB, dari mana datanya, dan apa yang tidak kami janjikan.",
};

export default function HalamanTentang() {
  /*
    Halaman teks, jadi lebarnya dibatasi — tapi dibatasi pada BARIS, bukan pada
    halaman. Paragraf sepanjang 1100px tidak terbaca; mata kehilangan barisnya
    saat kembali ke kiri. Ruang sisanya dipakai dengan menaruh blok-blok
    penjelasan berdampingan mulai lg, bukan dengan melebarkan paragraf.
  */
  return (
    <div className="space-y-6 sm:space-y-8">
      <h1 className="text-2xl font-extrabold sm:text-3xl">Tentang NOBARSIB</h1>

      <section className="max-w-2xl space-y-3 text-text-muted">
        <p>
          NOBARSIB adalah direktori tempat nonton bareng Persib di Bandung. Buka
          sebelum pertandingan, lihat venue mana saja yang menayangkan laga malam
          itu, lalu pilih berdasarkan jarak, rating, fasilitas, dan seberapa
          kondusif suasananya.
        </p>
        <p>
          Yang membedakannya dari poster Instagram adalah informasi yang tidak
          pernah ada di poster: apakah suasananya kondusif, aman dibawa anak,
          suaranya terdengar jelas, parkirnya cukup. Data itu hanya bisa datang
          dari penonton yang sudah pernah datang.
        </p>
      </section>

      <div className="grid gap-6 lg:grid-cols-2 lg:gap-x-10 lg:gap-y-8">
        <Blok judul="Dari mana datanya">
          <p>
            Sebagian diisi manual oleh kami setelah mendatangi venue, sebagian
            diumumkan sendiri oleh pemilik tempat. Setiap pengumuman ditinjau
            sebelum tampil.
          </p>
        </Blok>

        <Blok judul="Arti badge Dikonfirmasi">
          <p>
            Sehari sebelum pertandingan, kami menghubungi venue untuk memastikan
            nobarnya benar-benar jadi. Yang menjawab mendapat badge{" "}
            <span className="font-semibold text-confirm">Dikonfirmasi</span>. Yang
            belum menjawab tetap ditampilkan, tapi ditandai apa adanya dan berada
            lebih bawah.
          </p>
          <p>
            Kami memilih menampilkan ketidakpastian, bukan menyembunyikannya. Datang
            ke tempat yang ternyata tidak menayangkan pertandingan adalah kekecewaan
            yang tidak sebanding dengan daftar yang terlihat lebih penuh.
          </p>
        </Blok>

        <Blok judul="Yang tidak kami lakukan">
          <ul className="list-disc space-y-1 pl-5">
            <li>Tidak menyediakan atau menautkan siaran pertandingan.</li>
            <li>Tidak memverifikasi izin siar masing-masing venue.</li>
            <li>Tidak menerima reservasi atau pembayaran.</li>
            <li>
              Tidak menampilkan angka sisa kursi, karena kami tidak punya datanya.
              Indikator keramaian yang kami tampilkan adalah perkiraan berdasarkan
              minat pengguna aplikasi.
            </li>
          </ul>
        </Blok>

        <Blok judul="Data pribadi">
          <p>
            Tidak perlu mendaftar untuk memakai aplikasi ini. Kami tidak menyimpan
            nama, email, atau nomor telepon pengunjung. Untuk membedakan perangkat
            saat menghitung statistik, browser menyimpan satu penanda acak secara
            lokal — dan penanda itu tidak bisa dipakai mengenali siapa pun.
          </p>
        </Blok>

        <Blok judul="Afiliasi">
          <p>
            NOBARSIB tidak berafiliasi dengan Persib Bandung, PT Liga Indonesia
            Baru, maupun pemegang hak siar mana pun. Nama dan lambang klub adalah
            milik pemiliknya masing-masing.
          </p>
        </Blok>
      </div>

      <p className="text-sm text-text-muted">
        Punya masukan atau ingin mendaftarkan venue?{" "}
        <Link href="/untuk-venue" className="font-semibold text-brand-accent underline underline-offset-2">
          Lihat halaman untuk venue
        </Link>
        .
      </p>
    </div>
  );
}

function Blok({ judul, children }: { judul: string; children: React.ReactNode }) {
  return (
    <section>
      <h2 className="judul-bagian mb-2 font-semibold">{judul}</h2>
      <div className="space-y-2 text-text-muted">{children}</div>
    </section>
  );
}
