import type { Metadata, Viewport } from "next";
import Link from "next/link";

import "./globals.css";
import { RegisterSW } from "@/components/RegisterSW";

export const metadata: Metadata = {
  title: {
    default: "NOBARSIB — Cari tempat nobar Persib di Bandung",
    template: "%s · NOBARSIB",
  },
  description:
    "Cari tempat nonton bareng Persib di Bandung: terurut jarak, lengkap dengan " +
    "fasilitas, biaya masuk, dan seberapa kondusif suasananya.",
  applicationName: "NOBARSIB",
  appleWebApp: { capable: true, title: "NOBARSIB", statusBarStyle: "default" },
  formatDetection: { telephone: false },
};

export const viewport: Viewport = {
  // Satu warna untuk kedua mode: bilah alamat peramban menyambung ke header,
  // dan headernya biru di terang maupun gelap.
  themeColor: "#171b87",
  width: "device-width",
  initialScale: 1,
  // Tidak dikunci: memperbesar teks adalah kebutuhan aksesibilitas, bukan bug.
  maximumScale: 5,
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="id">
      {/* Kolom flex setinggi layar: halaman pendek — /jadwal saat baru satu laga,
          atau layar galat — tidak lagi menyisakan bidang kosong di bawah footer
          yang mengambang di tengah layar. */}
      <body className="flex min-h-dvh flex-col">
        <a
          href="#konten"
          className="sr-only focus:not-sr-only focus:absolute focus:left-3 focus:top-3 focus:z-50 focus:rounded-lg focus:bg-surface focus:px-4 focus:py-2 focus:shadow-lg"
        >
          Lompat ke konten
        </a>

        {/*
          Header adalah bidang biru terbesar di aplikasi, dan itu disengaja:
          satu bilah warna di puncak setiap halaman memberi identitas tanpa
          mewarnai konten di bawahnya. Daftar venue tetap kartu putih di atas
          latar terang, karena itu yang harus mudah dipindai (§13.3), bukan
          yang harus tampak menarik.
        */}
        <header className="bidang-brand sticky top-0 z-30 shadow-brand">
          <div className="wadah flex h-14 items-center justify-between sm:h-16">
            <Link
              href="/"
              className="flex items-center text-lg font-extrabold tracking-tight sm:text-xl"
            >
              <span>NOBAR</span>
              <span className="ml-1 rounded-md bg-white px-1.5 py-0.5 leading-none text-brand-ink">
                SIB
              </span>
            </Link>
            {/*
              "Untuk venue" disembunyikan di bawah 640px, bukan "Unduh".
              Di layar sempit hanya muat dua tautan, dan halaman venue adalah
              sisi penawaran yang datang lewat tautan langsung dari kami —
              sementara tombol unduh justru paling dibutuhkan persis di HP.
              Tautannya tetap ada di footer, jadi tidak ada yang hilang.
            */}
            <nav className="flex items-center gap-0.5 text-sm sm:gap-1">
              <Link
                href="/jadwal"
                className="rounded-lg px-3 py-2 font-medium text-white/75 transition-colors hover:bg-white/12 hover:text-white"
              >
                Jadwal
              </Link>
              <Link
                href="/untuk-venue"
                className="hidden rounded-lg px-3 py-2 font-medium text-white/75 transition-colors hover:bg-white/12 hover:text-white sm:block"
              >
                Untuk venue
              </Link>
              <Link
                href="/unduh"
                className="ml-1 rounded-lg bg-white/15 px-3 py-2 font-semibold text-white transition-colors hover:bg-white/25"
              >
                Unduh
              </Link>
            </nav>
          </div>
        </header>

        <main id="konten" className="wadah flex-1 pb-16 pt-4 sm:pt-6">
          {children}
        </main>

        {/* Footer melebar bersama halaman; di layar lebar tautan dan penyangkalan
            berdiri berdampingan, bukan bertumpuk dengan sisa ruang kosong di
            kanannya. */}
        <footer className="border-t-2 border-brand bg-surface">
          <div className="wadah py-8 text-sm text-text-muted md:flex md:items-start md:justify-between md:gap-10">
            <div className="flex flex-wrap gap-x-5 gap-y-2 md:shrink-0">
              <Link href="/unduh" className="hover:text-text">
                Unduh aplikasi
              </Link>
              <Link href="/tentang" className="hover:text-text">
                Tentang
              </Link>
              <Link href="/untuk-venue" className="hover:text-text">
                Daftarkan venue
              </Link>
              <Link href="/jadwal" className="hover:text-text">
                Jadwal
              </Link>
            </div>
            <p className="mt-4 max-w-2xl text-xs leading-relaxed text-text-faint md:mt-0 md:text-right">
              NOBARSIB adalah direktori independen dan tidak berafiliasi dengan klub
              mana pun. Informasi nobar berasal dari venue dan pengguna; kami tidak
              memverifikasi izin siar masing-masing tempat.
            </p>
          </div>
        </footer>

        <RegisterSW />
      </body>
    </html>
  );
}
