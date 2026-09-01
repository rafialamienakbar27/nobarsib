import type { Metadata } from "next";
import Link from "next/link";

import { APK, apkTersedia, apkUrl } from "@/lib/site";

export const metadata: Metadata = {
  title: "Unduh aplikasi",
  description:
    "Pasang NOBARSIB di HP Android lewat berkas APK, atau tambahkan ke layar utama di iPhone.",
};

/**
 * Halaman unduh aplikasi.
 *
 * Aplikasi Android-nya adalah pembungkus tipis (Trusted Web Activity) di atas
 * situs yang sama — bukan aplikasi terpisah. Konsekuensinya jujur ditulis di
 * halaman ini: isinya selalu ikut situs, dan tetap butuh koneksi.
 *
 * Dua hal yang sengaja TIDAK disembunyikan, karena menyembunyikannya justru
 * membuat orang berhenti di tengah pemasangan dan mengira aplikasinya rusak:
 *
 *   1. Android akan menampilkan peringatan "sumber tidak dikenal". Itu selalu
 *      terjadi di luar Play Store, dan halaman ini menjelaskannya lebih dulu.
 *   2. iPhone tidak bisa memasang APK sama sekali. Daripada tombol yang tidak
 *      berfungsi, pengguna iOS diberi jalur "Tambah ke Layar Utama" yang
 *      hasilnya sama-sama satu ikon di layar utama.
 */
export default function HalamanUnduh() {
  return (
    <div className="mx-auto max-w-3xl space-y-8 sm:space-y-10">
      <header className="max-w-2xl">
        <h1 className="text-2xl font-extrabold leading-tight sm:text-3xl">
          Pasang NOBARSIB di HP
        </h1>
        <p className="mt-2 text-text-muted">
          Supaya tidak perlu mengetik alamatnya lagi tiap kali Persib main. Ikonnya
          ada di layar utama, dibuka layar penuh tanpa bilah alamat peramban.
        </p>
      </header>

      <Android />
      <IPhone />
      <Catatan />
    </div>
  );
}

function Android() {
  return (
    <section>
      <h2 className="judul-bagian mb-3 font-semibold">Android</h2>

      {apkTersedia ? (
        <>
          {/*
            download + tanpa target="_blank": berkas .apk harus diunduh, bukan
            dibuka di tab baru — sebagian peramban Android menampilkan halaman
            kosong kalau dibuka sebagai navigasi.
          */}
          <a
            href={apkUrl}
            download
            // w-fit, bukan w-auto: elemen flex tetap setingkat blok dan akan
            // memenuhi lebar induknya kalau hanya diberi w-auto.
            className="flex min-h-14 w-full items-center justify-center gap-2.5 rounded-xl bg-brand px-6 text-base font-semibold text-on-brand shadow-brand transition-colors hover:bg-brand-lift sm:w-fit sm:min-w-72"
          >
            <svg viewBox="0 0 16 16" className="size-5" fill="currentColor" aria-hidden>
              <path d="M8 1.5a.75.75 0 0 1 .75.75v6.19l2.22-2.22a.75.75 0 1 1 1.06 1.06l-3.5 3.5a.75.75 0 0 1-1.06 0l-3.5-3.5a.75.75 0 0 1 1.06-1.06l2.22 2.22V2.25A.75.75 0 0 1 8 1.5M2.75 12a.75.75 0 0 1 .75.75v.75h9v-.75a.75.75 0 0 1 1.5 0v1.5a.75.75 0 0 1-.75.75h-10.5a.75.75 0 0 1-.75-.75v-1.5a.75.75 0 0 1 .75-.75" />
            </svg>
            Unduh aplikasi ({APK.ukuran})
          </a>

          <p className="mt-2 text-sm text-text-faint">
            Versi {APK.versi} · {APK.androidMinimal} ke atas · berkas .apk
          </p>

          <h3 className="mt-6 font-semibold">Cara memasang</h3>
          <ol className="mt-2 space-y-3">
            <Langkah nomor={1}>
              Tekan tombol di atas, lalu tunggu unduhannya selesai.
            </Langkah>
            <Langkah nomor={2}>
              Buka berkas <code className="rounded bg-surface-alt px-1">{APK.nama}</code>{" "}
              dari notifikasi unduhan, atau dari aplikasi Berkas di folder Download.
            </Langkah>
            <Langkah nomor={3}>
              Android akan menolak sekali dan menawarkan membuka Setelan. Nyalakan{" "}
              <span className="font-medium text-text">
                &ldquo;Izinkan dari sumber ini&rdquo;
              </span>
              , lalu tekan kembali.
            </Langkah>
            <Langkah nomor={4}>
              Tekan <span className="font-medium text-text">Pasang</span>, lalu{" "}
              <span className="font-medium text-text">Buka</span>. Ikon NOBARSIB
              muncul di layar utama.
            </Langkah>
          </ol>

          <p className="mt-5 rounded-xl border border-border bg-surface-alt px-4 py-3 text-sm text-text-muted">
            <span className="font-semibold text-text">
              Kenapa ada peringatan keamanan?
            </span>{" "}
            Karena aplikasinya diunduh langsung dari situs ini, bukan dari Play
            Store. Android menampilkan peringatan yang sama untuk setiap aplikasi
            di luar Play Store — itu bukan tanda ada yang salah dengan berkasnya.
          </p>
        </>
      ) : (
        <p className="rounded-xl border border-dashed border-border px-4 py-6 text-sm text-text-muted">
          Berkas aplikasinya sedang disiapkan dan belum bisa diunduh. Sementara
          ini, buka situs ini di Chrome lalu pilih menu ⋮ →{" "}
          <span className="font-medium text-text">&ldquo;Tambahkan ke layar utama&rdquo;</span>{" "}
          — hasilnya sama: satu ikon di layar utama yang membuka layar penuh.
        </p>
      )}
    </section>
  );
}

function IPhone() {
  return (
    <section>
      <h2 className="judul-bagian mb-3 font-semibold">iPhone &amp; iPad</h2>
      <p className="text-sm text-text-muted">
        iOS tidak mengizinkan pemasangan aplikasi dari luar App Store, jadi tidak
        ada berkas yang perlu diunduh. Hasil yang sama didapat lewat tiga langkah:
      </p>
      <ol className="mt-3 space-y-3">
        <Langkah nomor={1}>
          Buka situs ini di <span className="font-medium text-text">Safari</span>{" "}
          (harus Safari — Chrome di iOS tidak punya menu ini).
        </Langkah>
        <Langkah nomor={2}>
          Tekan tombol <span className="font-medium text-text">Bagikan</span> di
          bilah bawah, ikon kotak dengan panah ke atas.
        </Langkah>
        <Langkah nomor={3}>
          Pilih{" "}
          <span className="font-medium text-text">
            &ldquo;Tambah ke Layar Utama&rdquo;
          </span>
          , lalu Tambah.
        </Langkah>
      </ol>
    </section>
  );
}

function Catatan() {
  return (
    <section>
      <h2 className="judul-bagian mb-3 font-semibold">Yang perlu kamu tahu</h2>
      <ul className="space-y-2 text-sm text-text-muted">
        <li className="flex gap-2">
          <span aria-hidden className="text-brand-accent">
            ·
          </span>
          Aplikasinya membungkus situs yang sama, jadi isinya selalu ikut terbarui.
          Tidak ada pembaruan yang perlu kamu pasang sendiri setiap kali daftar
          venue berubah.
        </li>
        <li className="flex gap-2">
          <span aria-hidden className="text-brand-accent">
            ·
          </span>
          Tetap butuh koneksi internet. Status &ldquo;Dikonfirmasi&rdquo; berubah
          sampai menit terakhir menjelang kickoff, dan menampilkan data lama lebih
          berbahaya daripada tidak menampilkan apa-apa.
        </li>
        <li className="flex gap-2">
          <span aria-hidden className="text-brand-accent">
            ·
          </span>
          Tidak perlu mendaftar, dan aplikasinya tidak meminta akses apa pun selain
          lokasi — itu pun hanya untuk mengurutkan venue berdasarkan jarak, dan
          boleh ditolak.
        </li>
      </ul>

      <p className="mt-6 text-sm text-text-muted">
        Ada kendala saat memasang?{" "}
        <Link
          href="/untuk-venue"
          className="font-semibold text-brand-accent underline underline-offset-2"
        >
          Hubungi kami
        </Link>
        .
      </p>
    </section>
  );
}

function Langkah({ nomor, children }: { nomor: number; children: React.ReactNode }) {
  return (
    <li className="flex gap-3">
      <span className="grid size-7 shrink-0 place-items-center rounded-full bg-brand-soft text-sm font-bold text-brand-accent">
        {nomor}
      </span>
      <p className="pt-0.5 text-sm text-text-muted">{children}</p>
    </li>
  );
}
