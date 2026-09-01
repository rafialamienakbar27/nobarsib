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

import Link from "next/link";

import { track } from "@/lib/api";
import { deviceHash } from "@/lib/device";
import {
  formatDistance,
  formatEntry,
  formatRating,
  formatTime,
  kondusifLabel,
} from "@/lib/format";
import type { NobarEvent } from "@/lib/types";

/**
 * Kartu venue di daftar hasil.
 *
 * §13.3 membatasi kartu maksimal LIMA baris informasi, karena lebih dari itu
 * tidak terbaca di layar HP. Kelima baris itu:
 *
 *   1. nama venue
 *   2. kecamatan · jarak            <- jarak selalu tampil, alasan utama memilih
 *   3. rating · kondusif
 *   4. biaya masuk · jam buka pintu
 *   5. badge konfirmasi
 *
 * Catatan venue dan daftar fasilitas sengaja TIDAK ikut di sini; keduanya ada
 * di halaman detail. Setiap tambahan di kartu memotong keterbacaan seluruh daftar.
 */
export function VenueCard({ event }: { event: NobarEvent }) {
  const v = event.venue;
  const kondusif = kondusifLabel(v.kondusif_score);
  const rating = v.nobar_rating ?? v.google_rating;
  const ratingDariNobar = v.nobar_rating !== null;

  return (
    <Link
      href={`/venue/${v.slug}`}
      onClick={() => track(event.event_id, "open_detail", deviceHash())}
      // h-full: kartu berdiri dalam grid, dan tanpa ini kartu tanpa rating jadi
      // lebih pendek daripada tetangganya di baris yang sama.
      className="flex h-full gap-3 rounded-xl border border-border bg-surface p-3 shadow-card transition-colors hover:border-brand-line hover:bg-surface-alt"
    >
      <Thumbnail url={v.primary_photo} name={v.name} />

      <div className="flex min-w-0 flex-1 flex-col">
        {/* Baris 1 */}
        <div className="flex items-start justify-between gap-2">
          <h3 className="truncate font-semibold leading-tight">{v.name}</h3>
          {event.is_promoted && event.is_confirmed && (
            <span className="shrink-0 rounded border border-brand-line bg-brand-soft px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-brand-accent">
              Promo
            </span>
          )}
        </div>

        {/* Baris 2 — jarak selalu tampil (§13.3).
            Diberi warna biru: §13.3 menyebut jarak sebagai alasan utama orang
            memilih, jadi ia satu-satunya angka di kartu yang boleh menarik mata
            lebih dulu daripada nama venue. */}
        <p className="mt-0.5 truncate text-sm text-text-muted">
          {v.district && <>{v.district} · </>}
          <span className="font-semibold text-brand-accent">
            {formatDistance(v.distance_km)}
          </span>
        </p>

        {/* Baris 3 */}
        <p className="mt-0.5 flex flex-wrap items-center gap-x-2 text-sm text-text-muted">
          {rating !== null ? (
            <span>
              <span aria-hidden>★</span> {formatRating(rating)}
              <span className="text-text-faint">
                {" "}
                {ratingDariNobar ? `(${v.nobar_rating_count} nobar)` : "(Google)"}
              </span>
            </span>
          ) : (
            <span className="text-text-faint">Belum ada rating</span>
          )}
          {/* §11.4: sebelum 3 review, tidak ada angka kondusif yang boleh muncul */}
          {kondusif && <span className="text-confirm">· {kondusif}</span>}
        </p>

        {/* Baris 4 */}
        <p className="mt-0.5 text-sm text-text-muted">
          {formatEntry(event.entry_type, event.entry_amount)}
          {event.doors_open_at && <> · Buka {formatTime(event.doors_open_at)}</>}
        </p>

        {/* Baris 5. mt-auto mendorong badge ke dasar kartu supaya di dalam grid
            semua badge sebaris — status adalah hal yang dipindai paling cepat
            (§4.3), dan memindainya jadi sulit kalau letaknya naik-turun. */}
        <div className="mt-auto pt-1.5">
          <StatusBadge isConfirmed={event.is_confirmed} />
        </div>
      </div>
    </Link>
  );
}

/**
 * Thumbnail, bukan foto besar (§13.3). Sengaja memakai <img> biasa, bukan
 * next/image: sumbernya adalah URL foto venue dari domain yang belum tentu
 * terdaftar di konfigurasi, dan gambar sekecil ini tidak cukup berat untuk
 * membenarkan pipeline optimasi.
 */
function Thumbnail({ url, name }: { url?: string; name: string }) {
  if (!url) {
    return (
      <div
        className="grid size-20 shrink-0 place-items-center rounded-lg bg-brand-soft text-xl font-bold text-brand-accent"
        aria-hidden
      >
        {name.charAt(0).toUpperCase()}
      </div>
    );
  }
  return (
    <img
      src={url}
      alt=""
      loading="lazy"
      decoding="async"
      width={80}
      height={80}
      className="size-20 shrink-0 rounded-lg bg-surface-alt object-cover"
    />
  );
}

/**
 * Badge status.
 *
 * Ini elemen terpenting di seluruh kartu. §4.3 menyebut konfirmasi H-1 sebagai
 * "mekanisme paling penting di seluruh sistem", dan §21 menyebut data basi
 * sebagai risiko fatal yang membunuh FANZO. Karena itu venue yang belum
 * konfirmasi tidak disembunyikan — informasinya tetap berguna — tapi statusnya
 * dinyatakan apa adanya.
 */
export function StatusBadge({ isConfirmed }: { isConfirmed: boolean }) {
  if (isConfirmed) {
    return (
      <span className="inline-flex items-center gap-1 rounded-full border border-confirm-line bg-confirm-soft px-2 py-0.5 text-xs font-semibold text-confirm">
        <svg viewBox="0 0 16 16" className="size-3.5" fill="currentColor" aria-hidden>
          <path d="M6.5 11.8 3.2 8.5l1.1-1.1 2.2 2.2 5-5 1.1 1.1z" />
        </svg>
        Dikonfirmasi
      </span>
    );
  }
  return (
    <span className="inline-flex items-center gap-1 rounded-full bg-warn-soft px-2 py-0.5 text-xs font-medium text-warn">
      Belum dikonfirmasi
    </span>
  );
}

export function VenueCardSkeleton() {
  return (
    <div className="denyut flex h-full gap-3 rounded-xl border border-border bg-surface p-3">
      <div className="size-20 shrink-0 rounded-lg bg-surface-alt" />
      <div className="flex-1 space-y-2 py-1">
        <div className="h-4 w-2/3 rounded bg-surface-alt" />
        <div className="h-3 w-1/3 rounded bg-surface-alt" />
        <div className="h-3 w-1/2 rounded bg-surface-alt" />
        <div className="h-5 w-28 rounded-full bg-surface-alt" />
      </div>
    </div>
  );
}
