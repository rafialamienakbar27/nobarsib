// Bentuk data dari API (§8.2). Ditulis manual, bukan digenerate, supaya
// perubahan kontrak di backend memunculkan error TypeScript di sini.

export type SortMode = "recommended" | "nearest" | "rating";
export type EntryType = "free" | "min_order" | "ticket" | "donation";
export type CrowdLevel = "longgar" | "ramai" | "penuh";

export interface Team {
  name: string;
  slug: string;
  logo_url?: string;
}

export interface Match {
  id: number;
  competition?: string;
  home_team: Team;
  away_team: Team;
  kickoff_at: string;
  venue_name?: string;
  broadcast_tv?: string;
  status: string;
  nobar_count: number;
}

export interface VenueCard {
  id: string;
  name: string;
  slug: string;
  district?: string;
  distance_km: number;
  primary_photo?: string;
  google_rating: number | null;
  nobar_rating: number | null;
  nobar_rating_count: number;
  /** null selama review < 3 (§11.4) — tampilkan "Belum ada penilaian". */
  kondusif_score: number | null;
  kid_friendly_score: number | null;
  facilities: string[];
}

export interface NobarEvent {
  event_id: string;
  venue: VenueCard;
  doors_open_at?: string;
  entry_type: EntryType;
  entry_amount: number;
  crowd_level?: CrowdLevel;
  is_confirmed: boolean;
  confirmed_at?: string;
  is_promoted: boolean;
  notes?: string;
}

export interface NobarListMeta {
  total: number;
  page: number;
  per_page: number;
  match: { id: number; kickoff_at: string; label: string };
}

export interface NobarListResponse {
  data: NobarEvent[];
  meta: NobarListMeta;
}

export interface Facility {
  code: string;
  label: string;
  category?: string;
}

export interface VenuePhoto {
  url: string;
  caption?: string;
  is_primary: boolean;
}

export interface VenueReview {
  rating_overall: number;
  rating_kondusif: number | null;
  is_kid_friendly: boolean | null;
  crowd_actual?: string;
  comment?: string;
  created_at: string;
}

export interface NobarHistory {
  match_label: string;
  kickoff_at: string;
  entry_type: EntryType;
  entry_amount: number;
  is_confirmed: boolean;
}

export interface VenueDetail {
  id: string;
  name: string;
  slug: string;
  address: string;
  district?: string;
  city: string;
  lat: number;
  lng: number;
  phone?: string;
  whatsapp?: string;
  instagram_handle?: string;
  website?: string;
  google_rating: number | null;
  nobar_rating: number | null;
  nobar_rating_count: number;
  kondusif_score: number | null;
  kid_friendly_score: number | null;
  data_completeness: number;
  status: string;
  /**
   * Asal data profil dan kapan terakhir dipastikan manusia.
   *
   * Opsional, dan itu bukan kelalaian: venue yang masuk sebelum kolom ini ada
   * memang tidak diketahui asalnya, dan API sengaja menghilangkan field-nya
   * (`omitempty`) daripada mengirim string kosong yang gampang terbaca sebagai
   * "sudah diverifikasi".
   */
  data_source?: "google-places" | "venue" | "manual";
  last_verified_at?: string; // "2026-09-02"
  facilities: string[];
  photos: VenuePhoto[];
  opening_hours: Record<string, { open: string; close: string }>;
  recent_reviews: VenueReview[];
  nobar_history: NobarHistory[];
}

/** Parameter GET /v1/matches/{id}/nobar (§8.2). */
export interface NobarQuery {
  lat?: number;
  lng?: number;
  sort?: SortMode;
  radius_km?: number;
  facilities?: string[];
  entry_type?: EntryType | "";
  open_until_end?: boolean;
  page?: number;
  per_page?: number;
}
