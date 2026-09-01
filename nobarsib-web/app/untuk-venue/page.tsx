import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Untuk venue",
  description:
    "Cafe, resto, atau warkop yang menayangkan nobar Persib bisa tampil gratis di NOBARSIB.",
};

/**
 * Halaman akuisisi venue (§13.1).
 *
 * Bukan halaman pelengkap: inilah pintu masuk sisi penawaran, dan marketplace
 * mati kalau salah satu sisinya sepi (§3.1 prinsip 4). Portal venue sendiri
 * baru dibangun di Fase 5, jadi untuk sekarang pendaftaran lewat WhatsApp —
 * cara yang memang paling dipakai pemilik cafe di Indonesia (§12.1).
 */
export default function HalamanUntukVenue() {
  const nomorAdmin = process.env.NEXT_PUBLIC_ADMIN_WA ?? "";
  const pesan = encodeURIComponent(
    "Halo NOBARSIB, saya pemilik/pengelola tempat dan mau mendaftarkan venue untuk nobar Persib.",
  );

  return (
    <div className="space-y-8 sm:space-y-12">
      {/* Judul dan tombol daftar berdampingan mulai md: keduanya adalah satu
          ajakan yang sama, dan menaruh tombol selebar layar di bawah paragraf
          hanya mendorong sisa halaman turun tanpa menambah apa pun. */}
      <header className="md:flex md:items-end md:justify-between md:gap-10">
        <div className="max-w-2xl">
          <h1 className="text-2xl font-extrabold leading-tight sm:text-3xl">
            Nayangin nobar Persib? Tampil di sini, gratis.
          </h1>
          <p className="mt-2 text-text-muted">
            Bobotoh yang sedang mencari tempat nonton membuka NOBARSIB dan melihat
            venue terdekat lebih dulu. Mendaftar tidak dipungut biaya.
          </p>
        </div>

        {nomorAdmin ? (
          <a
            href={`https://wa.me/${nomorAdmin.replace(/\D/g, "")}?text=${pesan}`}
            target="_blank"
            rel="noopener noreferrer"
            className="mt-5 flex min-h-12 w-full items-center justify-center rounded-xl bg-brand px-4 text-center font-semibold text-on-brand shadow-brand transition-colors hover:bg-brand-lift md:mt-0 md:w-auto md:shrink-0 md:px-6"
          >
            Daftarkan venue lewat WhatsApp
          </a>
        ) : (
          <p className="mt-5 rounded-xl border border-dashed border-border px-4 py-5 text-center text-sm text-text-muted md:mt-0 md:max-w-xs md:shrink-0">
            Nomor pendaftaran sedang disiapkan. Sementara ini kirim info venue lewat
            direct message Instagram kami.
          </p>
        )}
      </header>

      <section>
        <h2 className="mb-3 font-semibold">Yang kamu dapat</h2>
        {/* Tiga poin sejajar, bukan bertumpuk: ketiganya setara, dan menumpuknya
            di layar lebar membuat halaman ini terasa jauh lebih panjang daripada
            isinya. */}
        <ul className="grid gap-3 md:grid-cols-3">
          <Manfaat judul="Muncul saat orang mencari">
            Venue kamu tampil di daftar untuk setiap laga yang kamu nayangin,
            terurut jarak dari calon penonton.
          </Manfaat>
          <Manfaat judul="Perkiraan jumlah pengunjung">
            Kamu bisa melihat berapa orang membuka pengumumanmu dan berapa yang
            menekan &ldquo;Buka di Maps&rdquo; — indikator paling dekat dengan
            niat datang, berguna untuk menghitung stok dan staf.
          </Manfaat>
          <Manfaat judul="Penilaian yang spesifik">
            Penonton menilai hal yang benar-benar menentukan: layar terlihat atau
            tidak, suara terdengar atau tidak, suasananya kondusif atau tidak.
          </Manfaat>
        </ul>
      </section>

      <section>
        <h2 className="mb-3 font-semibold">Yang kami minta</h2>
        <ol className="grid gap-4 md:grid-cols-3">
          <Langkah nomor={1} judul="Data tempat sekali saja">
            Nama, alamat, titik lokasi, kontak WhatsApp, fasilitas, dan jam buka.
            Sekitar sepuluh menit, dan tidak perlu diulang.
          </Langkah>
          <Langkah nomor={2} judul="Kabari kalau nayangin">
            Setiap kali ada laga yang kamu nayangin, beri tahu kami: jam buka
            pintu, biaya masuk, dan catatan tambahan. Tiga isian.
          </Langkah>
          <Langkah nomor={3} judul="Balas konfirmasi H-1">
            Sehari sebelum pertandingan kami kirim satu pesan WhatsApp. Cukup
            tekan tautannya untuk memastikan nobarnya jadi.
          </Langkah>
        </ol>
        <p className="mt-4 rounded-xl border border-border bg-surface-alt px-4 py-3 text-sm text-text-muted">
          Langkah ketiga yang paling menentukan. Venue yang mengonfirmasi mendapat
          badge <span className="font-semibold text-confirm">Dikonfirmasi</span> dan
          tampil lebih atas, karena itulah yang membuat pengguna percaya pada
          daftar ini.
        </p>
      </section>

      <section>
        <h2 className="judul-bagian mb-2 font-semibold">Pertanyaan yang sering muncul</h2>
        {/* gap-px di atas warna garis: satu cara mendapat garis pemisah yang
            benar baik saat bertumpuk maupun saat bersebelahan, tanpa mengganti
            divide-y jadi divide-x di titik henti tertentu. */}
        <dl className="grid gap-px overflow-hidden rounded-xl border border-border bg-border text-sm md:grid-cols-3">
          <Tanya
            tanya="Berapa biayanya?"
            jawab="Tidak ada. Kalau nanti ada layanan berbayar untuk tampil lebih menonjol, tampil biasa tetap gratis."
          />
          <Tanya
            tanya="Apakah harus punya izin siar?"
            jawab="Menayangkan siaran secara komersial idealnya punya izin dari pemegang hak siar. Itu tanggung jawab venue; kami tidak memverifikasinya."
          />
          <Tanya
            tanya="Kalau nobar batal?"
            jawab="Kabari kami, pengumumannya kami turunkan. Lebih baik batal tercatat daripada penonton datang ke tempat yang tidak menayangkan."
          />
        </dl>
      </section>
    </div>
  );
}

function Manfaat({ judul, children }: { judul: string; children: React.ReactNode }) {
  return (
    <li className="rounded-xl border border-border bg-surface p-4">
      <p className="font-semibold">{judul}</p>
      <p className="mt-1 text-sm text-text-muted">{children}</p>
    </li>
  );
}

function Langkah({
  nomor,
  judul,
  children,
}: {
  nomor: number;
  judul: string;
  children: React.ReactNode;
}) {
  return (
    <li className="flex gap-3">
      <span className="grid size-7 shrink-0 place-items-center rounded-full bg-brand-soft text-sm font-bold text-brand-accent">
        {nomor}
      </span>
      <div>
        <p className="font-semibold">{judul}</p>
        <p className="mt-0.5 text-sm text-text-muted">{children}</p>
      </div>
    </li>
  );
}

function Tanya({ tanya, jawab }: { tanya: string; jawab: string }) {
  return (
    <div className="bg-surface px-4 py-3">
      <dt className="font-medium">{tanya}</dt>
      <dd className="mt-0.5 text-text-muted">{jawab}</dd>
    </div>
  );
}
