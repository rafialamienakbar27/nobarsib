"use client";

/* eslint-disable @next/next/no-img-element --
   Foto venue masih memakai <img> biasa, bukan next/image.
   next/image mewajibkan hostname sumber didaftarkan di images.remotePatterns,
   sementara host penyimpanan foto (Cloudflare R2, §6.1) baru disiapkan di
   Fase 4 bersama pengisian data venue sungguhan. Mendaftarkan hostname "**"
   demi menyenangkan linter akan mengubah pengoptimal gambar menjadi proksi
   terbuka untuk domain mana pun — harga yang jelas terlalu mahal.
   Sampai host itu ada, ukuran gambar dikunci lewat atribut width/height dan
   loading="lazy" supaya tata letak tidak bergeser. */

import type { VenuePhoto } from "@/lib/types";

/**
 * Galeri foto (§13.4 blok 1).
 *
 * Digeser mendatar dengan scroll-snap, bukan carousel ber-JavaScript: satu
 * elemen native lebih ringan, bisa dipakai keyboard, dan tetap jalan sebelum
 * hidrasi selesai.
 */
export function Gallery({ photos, nama }: { photos: VenuePhoto[]; nama: string }) {
  if (photos.length === 0) {
    return (
      <div className="grid aspect-[16/9] w-full place-items-center rounded-2xl bg-surface-alt text-text-faint">
        <span className="text-sm">Belum ada foto</span>
      </div>
    );
  }

  if (photos.length === 1) {
    return (
      <img
        src={photos[0].url}
        alt={photos[0].caption || nama}
        className="aspect-[16/9] w-full rounded-2xl bg-surface-alt object-cover"
      />
    );
  }

  /*
    Geser mendatar hanya di layar sempit, di mana itu memang gerakan yang
    paling wajar. Mulai sm foto disusun jadi kisi: satu foto besar di atas,
    sisanya kecil di bawahnya — menggeser foto satu per satu dengan mouse
    adalah pekerjaan, dan di layar lebar semuanya bisa terlihat sekaligus.
  */
  return (
    <div className="no-scrollbar -mx-4 flex snap-x snap-mandatory gap-2 overflow-x-auto px-4 sm:mx-0 sm:grid sm:grid-cols-3 sm:overflow-visible sm:px-0">
      {photos.map((p, i) => (
        <img
          key={p.url}
          src={p.url}
          alt={p.caption || `${nama} — foto ${i + 1}`}
          loading={i === 0 ? "eager" : "lazy"}
          decoding="async"
          className={
            "aspect-[16/9] w-[85%] shrink-0 snap-center rounded-2xl bg-surface-alt object-cover sm:w-full " +
            (i === 0 ? "sm:col-span-3" : "sm:aspect-square")
          }
        />
      ))}
    </div>
  );
}
