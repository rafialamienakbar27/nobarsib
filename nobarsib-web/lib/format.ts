import type { EntryType } from "./types";

const WIB = "Asia/Jakarta";

/** "Sabtu, 5 Sep · 19:00 WIB" seperti kartu laga di §13.2. */
export function formatKickoff(iso: string): string {
  const d = new Date(iso);
  const tanggal = new Intl.DateTimeFormat("id-ID", {
    weekday: "long",
    day: "numeric",
    month: "short",
    timeZone: WIB,
  }).format(d);
  return `${tanggal} · ${formatTime(iso)} WIB`;
}

export function formatTime(iso: string): string {
  return new Intl.DateTimeFormat("id-ID", {
    hour: "2-digit",
    minute: "2-digit",
    timeZone: WIB,
  }).format(new Date(iso));
}

export function formatDate(iso: string): string {
  return new Intl.DateTimeFormat("id-ID", {
    day: "numeric",
    month: "long",
    year: "numeric",
    timeZone: WIB,
  }).format(new Date(iso));
}

/** "2,4 km" — koma sebagai desimal, sesuai kebiasaan Indonesia (§13.2). */
export function formatDistance(km: number): string {
  if (km < 1) return `${Math.round(km * 1000)} m`;
  return `${km.toFixed(1).replace(".", ",")} km`;
}

/** "25rb" — ringkas supaya kartu tetap muat 5 baris (§13.3). */
export function formatRupiahShort(amount: number): string {
  if (amount >= 1_000_000) return `${(amount / 1_000_000).toFixed(1).replace(".0", "")}jt`;
  if (amount >= 1000) return `${Math.round(amount / 1000)}rb`;
  return String(amount);
}

export function formatRupiah(amount: number): string {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    maximumFractionDigits: 0,
  }).format(amount);
}

/** Label biaya masuk untuk kartu venue. */
export function formatEntry(type: EntryType, amount: number): string {
  switch (type) {
    case "free":
      return "Gratis";
    case "min_order":
      return `Min order ${formatRupiahShort(amount)}`;
    case "ticket":
      return `Tiket ${formatRupiahShort(amount)}`;
    case "donation":
      return amount > 0 ? `Donasi ${formatRupiahShort(amount)}` : "Seikhlasnya";
  }
}

/** "4,6" — rating dengan koma desimal. */
export function formatRating(v: number): string {
  return v.toFixed(1).replace(".", ",");
}

/**
 * Label skor kondusif. §11.4 melarang menampilkan angka sebelum ada 3 review,
 * dan API sudah mengirim null pada kondisi itu — di sini tinggal dibaca jujur.
 */
export function kondusifLabel(score: number | null): string | null {
  if (score === null) return null;
  if (score >= 4.2) return "Kondusif";
  if (score >= 3.2) return "Cukup kondusif";
  return "Kurang kondusif";
}

/** Waktu relatif singkat untuk badge konfirmasi dan tanggal review. */
export function timeAgo(iso: string): string {
  const diffMs = Date.now() - new Date(iso).getTime();
  const menit = Math.round(diffMs / 60000);
  if (menit < 1) return "baru saja";
  if (menit < 60) return `${menit} menit lalu`;
  const jam = Math.round(menit / 60);
  if (jam < 24) return `${jam} jam lalu`;
  const hari = Math.round(jam / 24);
  if (hari < 30) return `${hari} hari lalu`;
  return formatDate(iso);
}

/** Nomor WA Indonesia -> tautan wa.me (§13.3 tombol "Chat WA"). */
export function waLink(phone: string, text?: string): string {
  const digits = phone.replace(/\D/g, "").replace(/^0/, "62");
  const q = text ? `?text=${encodeURIComponent(text)}` : "";
  return `https://wa.me/${digits}${q}`;
}

/**
 * Tautan ke Google Maps (§13.3 tombol utama "Buka di Maps").
 *
 * Memakai koordinat, bukan nama, supaya tidak pernah salah tempat — banyak
 * cafe di Bandung punya nama yang mirip.
 */
export function mapsLink(lat: number, lng: number, label?: string): string {
  const q = label ? `${encodeURIComponent(label)}@${lat},${lng}` : `${lat},${lng}`;
  return `https://www.google.com/maps/search/?api=1&query=${q}`;
}

export function instagramLink(handle: string): string {
  return `https://instagram.com/${handle.replace(/^@/, "")}`;
}

const HARI = ["Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"];

export function namaHari(dow: string | number): string {
  return HARI[Number(dow)] ?? "";
}
