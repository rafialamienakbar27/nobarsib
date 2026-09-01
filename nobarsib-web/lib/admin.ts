"use client";

import { apiBase } from "./api";

/**
 * Klien API panel admin.
 *
 * Access token disimpan di localStorage, bukan cookie HttpOnly. Itu keputusan
 * sadar, bukan kelalaian: panel admin hanya dipakai oleh satu-dua orang di
 * perangkat sendiri, dan cookie HttpOnly menuntut alur CSRF tersendiri yang
 * tidak sepadan untuk itu. Yang menahan dampaknya adalah umur access token
 * yang cuma 15 menit, dan refresh token yang dicabut begitu dipakai (rotasi).
 *
 * Kalau nanti portal venue dibuka untuk publik di Fase 5, keputusan ini harus
 * ditinjau ulang.
 */
const KUNCI_AKSES = "nobarsib:admin:access";
const KUNCI_REFRESH = "nobarsib:admin:refresh";
const KUNCI_USER = "nobarsib:admin:user";

export interface AdminUser {
  id: string;
  email: string;
  full_name: string;
  role: string;
}

export function simpanSesi(access: string, refresh: string, user: AdminUser): void {
  try {
    localStorage.setItem(KUNCI_AKSES, access);
    localStorage.setItem(KUNCI_REFRESH, refresh);
    localStorage.setItem(KUNCI_USER, JSON.stringify(user));
  } catch {
    // Penyimpanan diblokir: sesi hanya bertahan selama halaman terbuka.
  }
}

export function hapusSesi(): void {
  for (const k of [KUNCI_AKSES, KUNCI_REFRESH, KUNCI_USER]) {
    try {
      localStorage.removeItem(k);
    } catch {
      /* diabaikan */
    }
  }
}

export function penggunaSaatIni(): AdminUser | null {
  try {
    const raw = localStorage.getItem(KUNCI_USER);
    return raw ? (JSON.parse(raw) as AdminUser) : null;
  } catch {
    return null;
  }
}

/*
  Tiga fungsi di bawah dipakai useSyncExternalStore agar React membaca
  localStorage dengan benar.

  Membacanya di dalam useEffect lalu memanggil setState memicu render
  bertingkat, dan React 19 menandainya sebagai kesalahan. localStorage adalah
  penyimpanan di luar React, jadi bacanya lewat jalur yang memang disediakan
  untuk itu.

  snapshotSesi sengaja mengembalikan STRING mentah, bukan objek hasil parse:
  useSyncExternalStore membandingkan hasilnya dengan Object.is, dan objek baru
  pada setiap pemanggilan akan membuat React merender tanpa henti.
*/
export function subscribeSesi(onChange: () => void): () => void {
  // Peristiwa "storage" hanya menyala di tab LAIN. Itu justru yang diinginkan:
  // keluar di satu tab akan mengeluarkan tab lain juga.
  window.addEventListener("storage", onChange);
  return () => window.removeEventListener("storage", onChange);
}

export function snapshotSesi(): string {
  try {
    return localStorage.getItem(KUNCI_USER) ?? "";
  } catch {
    return "";
  }
}

// Saat dirender di server, tidak ada sesi yang bisa dibaca.
export function snapshotSesiServer(): string {
  return "";
}

function ambilToken(kunci: string): string | null {
  try {
    return localStorage.getItem(kunci);
  } catch {
    return null;
  }
}

export class SesiBerakhir extends Error {
  constructor() {
    super("Sesi berakhir, silakan masuk lagi");
    this.name = "SesiBerakhir";
  }
}

/**
 * Memanggil API admin, memperbarui token sekali kalau kedaluwarsa.
 *
 * Access token berumur 15 menit, sementara admin bisa membuka antrian selama
 * setengah jam. Tanpa pembaruan otomatis, tinjauan yang seharusnya 10 detik
 * (§14.2) akan terputus oleh layar login di tengah jalan.
 */
export async function adminFetch<T>(path: string, init?: RequestInit): Promise<T> {
  let res = await panggil(path, init, ambilToken(KUNCI_AKSES));

  if (res.status === 401) {
    const baru = await perbaruiToken();
    if (!baru) {
      hapusSesi();
      throw new SesiBerakhir();
    }
    res = await panggil(path, init, baru);
  }

  if (!res.ok) {
    let pesan = `Permintaan gagal (${res.status})`;
    try {
      const body = (await res.json()) as { error?: { message?: string } };
      if (body.error?.message) pesan = body.error.message;
    } catch {
      /* pakai pesan bawaan */
    }
    throw new Error(pesan);
  }

  if (res.status === 204) return undefined as T;
  return (await res.json()) as T;
}

function panggil(path: string, init: RequestInit | undefined, token: string | null) {
  return fetch(`${apiBase()}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...init?.headers,
    },
    cache: "no-store",
  });
}

async function perbaruiToken(): Promise<string | null> {
  const refresh = ambilToken(KUNCI_REFRESH);
  if (!refresh) return null;

  const res = await fetch(`${apiBase()}/auth/refresh`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ refresh_token: refresh }),
  });
  if (!res.ok) return null;

  const body = (await res.json()) as {
    access_token: string;
    refresh_token: string;
    user: AdminUser;
  };
  simpanSesi(body.access_token, body.refresh_token, body.user);
  return body.access_token;
}

export async function login(email: string, password: string): Promise<AdminUser> {
  const res = await fetch(`${apiBase()}/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });

  if (!res.ok) {
    const body = (await res.json().catch(() => ({}))) as { error?: { message?: string } };
    throw new Error(body.error?.message ?? "Gagal masuk");
  }

  const body = (await res.json()) as {
    access_token: string;
    refresh_token: string;
    user: AdminUser;
  };
  simpanSesi(body.access_token, body.refresh_token, body.user);
  return body.user;
}

export async function logout(): Promise<void> {
  const refresh = ambilToken(KUNCI_REFRESH);
  if (refresh) {
    await fetch(`${apiBase()}/auth/logout`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: refresh }),
    }).catch(() => {});
  }
  hapusSesi();
}

// ---------------------------------------------------------------- tipe data

export interface AdminEvent {
  id: string;
  venue_name: string;
  venue_slug: string;
  district: string;
  match_label: string;
  kickoff_at: string;
  doors_open_at?: string;
  entry_type: string;
  entry_amount: number;
  notes?: string;
  status: string;
  created_at: string;
}

export interface AdminSummary {
  pending_events: number;
  upcoming_matches: number;
  next_match: {
    id: number;
    home_team: { name: string };
    away_team: { name: string };
    kickoff_at: string;
    nobar_count: number;
  } | null;
}

export const getSummary = () =>
  adminFetch<{ data: AdminSummary }>("/admin/summary").then((r) => r.data);

export const getAntrian = () =>
  adminFetch<{ data: AdminEvent[] | null }>("/admin/events").then((r) => r.data ?? []);

export const tinjauEvent = (id: string, aksi: "approve" | "reject" | "confirm" | "cancel") =>
  adminFetch<{ data: { id: string; status: string } }>(`/admin/events/${id}/${aksi}`, {
    method: "POST",
  });
