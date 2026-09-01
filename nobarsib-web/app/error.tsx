"use client";

export default function Galat({ reset }: { error: Error; reset: () => void }) {
  return (
    <div className="mx-auto max-w-2xl rounded-2xl border border-dashed border-border bg-surface px-5 py-12 text-center">
      <h1 className="text-lg font-bold">Ada yang salah</h1>
      <p className="mx-auto mt-2 max-w-sm text-sm text-text-muted">
        Halaman ini gagal dimuat. Biasanya karena koneksi terputus atau server
        sedang sibuk.
      </p>
      <button
        onClick={reset}
        className="mt-5 inline-flex min-h-11 items-center rounded-xl bg-brand px-4 py-2.5 text-sm font-semibold text-on-brand shadow-brand transition-colors hover:bg-brand-lift"
      >
        Coba lagi
      </button>
    </div>
  );
}
