import Link from "next/link";

/**
 * Tampilan keadaan kosong (§13.5).
 *
 * Blueprint menyebut bagian ini "sering dilupakan tapi penting untuk aplikasi
 * baru yang datanya masih tipis" — dan di awal, inilah yang paling sering
 * dilihat pengguna pertama. Karena itu setiap keadaan punya pesan DAN langkah
 * lanjutannya sendiri, bukan satu pesan generik.
 */
export function EmptyState({
  judul,
  pesan,
  aksi,
}: {
  judul: string;
  pesan: string;
  aksi?: { label: string; href?: string; onClick?: () => void };
}) {
  return (
    <div className="mx-auto max-w-2xl rounded-2xl border border-dashed border-border bg-surface px-5 py-10 text-center">
      <h3 className="font-semibold">{judul}</h3>
      <p className="mx-auto mt-1.5 max-w-sm text-sm leading-relaxed text-text-muted">{pesan}</p>
      {aksi &&
        (aksi.href ? (
          <Link
            href={aksi.href}
            className="mt-4 inline-flex min-h-11 items-center rounded-xl bg-brand px-4 py-2.5 text-sm font-semibold text-on-brand shadow-brand transition-colors hover:bg-brand-lift"
          >
            {aksi.label}
          </Link>
        ) : (
          <button
            type="button"
            onClick={aksi.onClick}
            className="mt-4 inline-flex min-h-11 items-center rounded-xl bg-brand px-4 py-2.5 text-sm font-semibold text-on-brand shadow-brand transition-colors hover:bg-brand-lift"
          >
            {aksi.label}
          </button>
        ))}
    </div>
  );
}

/** §13.5 baris 1 — belum ada laga mendatang. */
export function TidakAdaLaga() {
  return (
    <EmptyState
      judul="Belum ada jadwal laga"
      pesan="Persib belum punya jadwal terdekat yang bisa ditampilkan. Cek jadwal lengkap musim ini."
      aksi={{ label: "Lihat jadwal lengkap", href: "/jadwal" }}
    />
  );
}

/** §13.5 baris 2 — ada laga, belum ada venue. */
export function BelumAdaNobar() {
  return (
    <EmptyState
      judul="Belum ada info nobar untuk laga ini"
      pesan="Belum ada venue yang mengumumkan nobar. Punya info tempat yang menayangkan? Kirim ke kami, nanti kami verifikasi."
      aksi={{ label: "Kirim info nobar", href: "/untuk-venue" }}
    />
  );
}
