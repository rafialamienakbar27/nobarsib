import type {
  Facility,
  Match,
  NobarListResponse,
  NobarQuery,
  VenueDetail,
} from "./types";

// Dua alamat berbeda, dan bedanya penting.
//
// Di dalam Docker, server Next memanggil API lewat nama service (`api:8080`),
// sementara browser pengguna hanya bisa memanggil alamat yang terlihat dari
// luar. Kalau keduanya disamakan, salah satu sisi pasti gagal.
const SERVER_BASE = process.env.API_URL ?? "http://localhost:8080/v1";
const BROWSER_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/v1";

export function apiBase(): string {
  return typeof window === "undefined" ? SERVER_BASE : BROWSER_BASE;
}

/** Dilempar saat API menjawab dengan status error. */
export class ApiError extends Error {
  constructor(
    readonly status: number,
    readonly code: string,
    message: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

interface ErrorBody {
  error?: { code?: string; message?: string };
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${apiBase()}${path}`, {
    ...init,
    headers: { Accept: "application/json", ...init?.headers },
    // Data nobar berubah sampai menit terakhir: venue mengonfirmasi lewat link
    // H-1 (§15.3) dan badge "Dikonfirmasi" harus segera terlihat. Beban
    // permintaan sudah ditahan cache Redis di sisi API (§13.6), jadi tidak
    // perlu ditahan lagi di sini.
    cache: "no-store",
  });

  if (!res.ok) {
    let body: ErrorBody = {};
    try {
      body = (await res.json()) as ErrorBody;
    } catch {
      // Response bukan JSON — pakai pesan bawaan di bawah.
    }
    throw new ApiError(
      res.status,
      body.error?.code ?? "UNKNOWN",
      body.error?.message ?? `Permintaan gagal (${res.status})`,
    );
  }
  return (await res.json()) as T;
}

function qs(params: Record<string, string | number | boolean | undefined>): string {
  const sp = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) {
    if (v !== undefined && v !== "") sp.set(k, String(v));
  }
  const s = sp.toString();
  return s ? `?${s}` : "";
}

export async function getUpcomingMatches(limit = 5): Promise<Match[]> {
  const res = await request<{ data: Match[] | null }>(`/matches/upcoming${qs({ limit })}`);
  return res.data ?? [];
}

export async function getSeasonMatches(): Promise<Match[]> {
  const res = await request<{ data: Match[] | null }>("/matches");
  return res.data ?? [];
}

export async function getNobarForMatch(
  matchId: number,
  q: NobarQuery = {},
): Promise<NobarListResponse> {
  const res = await request<NobarListResponse>(
    `/matches/${matchId}/nobar${qs({
      lat: q.lat,
      lng: q.lng,
      sort: q.sort,
      radius_km: q.radius_km,
      facilities: q.facilities?.length ? q.facilities.join(",") : undefined,
      entry_type: q.entry_type,
      open_until_end: q.open_until_end,
      page: q.page,
      per_page: q.per_page,
    })}`,
  );
  // API mengembalikan `data: null` kalau tidak ada hasil; komponen di bawah
  // selalu ingin array supaya tidak perlu menjaga dua bentuk.
  return { ...res, data: res.data ?? [] };
}

export async function getVenue(slug: string): Promise<VenueDetail> {
  const res = await request<{ data: VenueDetail }>(`/venues/${encodeURIComponent(slug)}`);
  return res.data;
}

export async function getFacilities(): Promise<Facility[]> {
  const res = await request<{ data: Facility[] | null }>("/facilities");
  return res.data ?? [];
}

/**
 * Mencatat interaksi pengguna (§8.2).
 *
 * Sengaja tidak pernah melempar error: statistik tidak boleh merusak
 * pengalaman orang yang sedang memilih tempat nonton.
 */
export function track(
  eventId: string,
  action: "view_card" | "open_detail" | "open_maps" | "click_wa",
  deviceHash: string,
): void {
  const body = JSON.stringify({ action, device_hash: deviceHash });
  const url = `${BROWSER_BASE}/events/${eventId}/track`;

  // sendBeacon bertahan meski halaman langsung ditutup — persis yang terjadi
  // saat pengguna menekan "Buka di Maps" dan berpindah aplikasi.
  if (typeof navigator !== "undefined" && navigator.sendBeacon) {
    navigator.sendBeacon(url, new Blob([body], { type: "application/json" }));
    return;
  }
  void fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body,
    keepalive: true,
  }).catch(() => {});
}
