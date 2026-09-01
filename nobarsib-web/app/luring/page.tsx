import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Sedang luring",
  robots: { index: false },
};

/** Halaman cadangan saat perangkat benar-benar tanpa koneksi. */
export default function HalamanLuring() {
  return (
    <div className="mx-auto max-w-2xl rounded-2xl border border-dashed border-border bg-surface px-5 py-12 text-center">
      <h1 className="text-lg font-bold">Tidak ada koneksi</h1>
      <p className="mx-auto mt-2 max-w-sm text-sm leading-relaxed text-text-muted">
        Daftar tempat nobar tidak bisa dimuat tanpa internet. Informasi ini sengaja
        tidak disimpan untuk dibaca luring — status venue berubah sampai menit
        terakhir, dan menampilkan data lama lebih berbahaya daripada tidak
        menampilkan apa-apa.
      </p>
    </div>
  );
}
