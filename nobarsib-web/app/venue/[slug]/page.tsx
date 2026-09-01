import type { Metadata } from "next";
import { notFound } from "next/navigation";

import { Gallery } from "@/components/Gallery";
import { StatusBadge } from "@/components/VenueCard";
import { VenueActions } from "@/components/VenueActions";
import { ApiError, getFacilities, getVenue } from "@/lib/api";
import {
  formatDate,
  formatEntry,
  formatRating,
  kondusifLabel,
  namaHari,
  timeAgo,
} from "@/lib/format";
import type { VenueDetail } from "@/lib/types";

export async function generateMetadata(props: PageProps<"/venue/[slug]">): Promise<Metadata> {
  const { slug } = await props.params;
  try {
    const v = await getVenue(slug);
    return {
      title: v.name,
      description: `Nobar Persib di ${v.name}${v.district ? `, ${v.district}` : ""}. ${v.address}`,
    };
  } catch {
    return { title: "Venue tidak ditemukan" };
  }
}

/**
 * Halaman detail venue.
 *
 * Urutan blok mengikuti §13.4 persis, dari atas ke bawah:
 *   1 galeri · 2 nama/kecamatan · 3 skor · 4 info nobar · 5 tombol aksi
 *   6 fasilitas · 7 jam buka · 8 review terbaru · 9 riwayat nobar
 */
export default async function HalamanVenue(props: PageProps<"/venue/[slug]">) {
  const { slug } = await props.params;

  let venue: VenueDetail;
  try {
    venue = await getVenue(slug);
  } catch (e) {
    if (e instanceof ApiError && e.status === 404) notFound();
    throw e;
  }
  const facilities = await getFacilities();
  const labelFasilitas = new Map(facilities.map((f) => [f.code, f.label]));

  /*
    Dua baris grid, bukan satu kolom panjang.

    Urutan blok §13.4 dipertahankan PERSIS di DOM — 1…9 tetap berurutan dalam
    sumbernya, jadi pembaca layar dan layar sempit membacanya sama seperti
    sebelumnya. Yang berubah hanya penempatannya di layar lebar: blok 1–5
    (galeri dan keputusan: siapa, seberapa bagus, tombol pergi ke sana)
    berdampingan di baris pertama, blok 6–9 (rincian yang dibaca kalau masih
    ragu) di dua kolom baris kedua. Satu kolom sempit di tengah layar 27 inci
    membuat halaman ini butuh empat kali gulir untuk isi yang muat satu layar.
  */
  return (
    <article className="space-y-6 sm:space-y-8">
      <div className="grid gap-5 lg:grid-cols-2 lg:items-start lg:gap-8">
        {/* 1 */}
        <Gallery photos={venue.photos} nama={venue.name} />

        <div className="space-y-5">
          {/* 2 */}
          <header>
            <h1 className="text-2xl font-extrabold leading-tight sm:text-3xl">
              {venue.name}
            </h1>
            <p className="mt-1 text-text-muted">
              {venue.district && <>{venue.district} · </>}
              {venue.city}
            </p>
            <p className="mt-1 text-sm text-text-muted">{venue.address}</p>
          </header>

          {/* 3 */}
          <Skor venue={venue} />

          {/* 4 — info nobar malam ini diisi di Fase 5 bersama portal venue;
                 sekarang halaman ini dipakai untuk profil dan riwayat. */}

          {/* 5 */}
          <VenueActions venue={venue} />
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-2 lg:items-start lg:gap-8">
        <div className="space-y-6">
          {/* 6 */}
          {venue.facilities.length > 0 && (
            <section>
              <h2 className="judul-bagian mb-2 font-semibold">Fasilitas</h2>
              <ul className="grid grid-cols-2 gap-x-3 gap-y-1.5 text-sm text-text-muted sm:grid-cols-3 lg:grid-cols-2">
                {venue.facilities.map((code) => (
                  <li key={code} className="flex items-center gap-1.5">
                    <span className="text-confirm" aria-hidden>
                      ✓
                    </span>
                    {labelFasilitas.get(code) ?? code}
                  </li>
                ))}
              </ul>
            </section>
          )}

          {/* 7 */}
          <JamBuka hours={venue.opening_hours} />
        </div>

        <div className="space-y-6">
          {/* 8 */}
          <Review venue={venue} />

          {/* 9 */}
          <Riwayat venue={venue} />
        </div>
      </div>
    </article>
  );
}

function Skor({ venue }: { venue: VenueDetail }) {
  const kondusif = kondusifLabel(venue.kondusif_score);
  const cukupReview = venue.nobar_rating !== null;

  return (
    <section className="grid grid-cols-3 gap-2 rounded-2xl border border-border bg-surface p-4 text-center shadow-card">
      <Angka
        label="Rating nobar"
        nilai={cukupReview ? formatRating(venue.nobar_rating!) : null}
        keterangan={cukupReview ? `${venue.nobar_rating_count} penilaian` : "Belum ada"}
      />
      {/* §11.4: sebelum 3 review, "Belum ada penilaian" lebih jujur daripada
          angka dari satu orang yang kebetulan sedang kesal. */}
      <Angka
        label="Kondusif"
        nilai={venue.kondusif_score !== null ? formatRating(venue.kondusif_score) : null}
        keterangan={kondusif ?? "Belum ada penilaian"}
        bergaris
      />
      <Angka
        label="Ramah anak"
        nilai={
          venue.kid_friendly_score !== null ? formatRating(venue.kid_friendly_score) : null
        }
        keterangan={venue.kid_friendly_score !== null ? "menurut penonton" : "Belum ada"}
      />
    </section>
  );
}

/**
 * Satu angka skor.
 *
 * Angkanya biru saat ada isinya dan abu-abu saat kosong — supaya sekilas
 * terlihat mana yang sudah punya data dan mana yang belum, tanpa harus
 * membaca keterangan di bawahnya.
 */
function Angka({
  label,
  nilai,
  keterangan,
  bergaris,
}: {
  label: string;
  nilai: string | null;
  keterangan: string;
  bergaris?: boolean;
}) {
  return (
    <div className={bergaris ? "border-x border-border" : undefined}>
      <p className="text-xs text-text-faint">{label}</p>
      <p
        className={
          "mt-0.5 text-xl font-extrabold " +
          (nilai ? "text-brand-accent" : "text-text-faint")
        }
      >
        {nilai ?? "—"}
      </p>
      <p className="text-[11px] leading-tight text-text-faint">{keterangan}</p>
    </div>
  );
}

function JamBuka({ hours }: { hours: VenueDetail["opening_hours"] }) {
  const adaData = Object.keys(hours ?? {}).length > 0;

  return (
    <section>
      <h2 className="judul-bagian mb-2 font-semibold">Jam buka</h2>
      {adaData ? (
        <dl className="divide-y divide-border overflow-hidden rounded-xl border border-border bg-surface text-sm">
          {["0", "1", "2", "3", "4", "5", "6"].map((dow) => {
            // Tiga keadaan, bukan dua.
            //
            // Hari yang TIDAK ADA di data berarti jamnya belum pernah dicatat;
            // hari yang ada tapi bernilai null berarti venue memang tutup.
            // Menyamakan keduanya jadi "Tutup" adalah mengarang: venue yang
            // datanya belum lengkap akan terbaca seolah tutup enam hari
            // seminggu. Pembedaan yang sama sudah dipakai di sisi server
            // (§9.4, openhours.OpenUntil mengembalikan `known`).
            const tercatat = Object.hasOwn(hours, dow);
            const jam = hours[dow];
            return (
              <div key={dow} className="flex justify-between gap-3 px-3 py-2">
                <dt className="text-text-muted">{namaHari(dow)}</dt>
                <dd className={jam ? "font-medium" : "text-text-faint"}>
                  {jam
                    ? `${jam.open} – ${jam.close}`
                    : tercatat
                      ? "Tutup"
                      : "Belum tercatat"}
                </dd>
              </div>
            );
          })}
        </dl>
      ) : (
        <p className="rounded-xl border border-dashed border-border px-3 py-4 text-sm text-text-faint">
          Jam buka belum tercatat.
        </p>
      )}
    </section>
  );
}

function Review({ venue }: { venue: VenueDetail }) {
  return (
    <section>
      <h2 className="judul-bagian mb-2 font-semibold">Kata penonton</h2>
      {venue.recent_reviews.length > 0 ? (
        <ul className="space-y-2">
          {venue.recent_reviews.map((r, i) => (
            <li key={i} className="rounded-xl border border-border bg-surface p-3">
              <div className="flex items-center justify-between text-sm">
                <span className="font-semibold">
                  <span aria-hidden>★</span> {r.rating_overall}/5
                </span>
                <span className="text-xs text-text-faint">{timeAgo(r.created_at)}</span>
              </div>
              {r.comment && <p className="mt-1 text-sm text-text-muted">{r.comment}</p>}
              <div className="mt-1.5 flex flex-wrap gap-1.5 text-[11px]">
                {r.rating_kondusif !== null && (
                  <span className="rounded bg-surface-alt px-1.5 py-0.5 text-text-muted">
                    Kondusif {r.rating_kondusif}/5
                  </span>
                )}
                {r.is_kid_friendly !== null && (
                  <span className="rounded bg-surface-alt px-1.5 py-0.5 text-text-muted">
                    {r.is_kid_friendly ? "Aman bawa anak" : "Kurang cocok untuk anak"}
                  </span>
                )}
                {r.crowd_actual && (
                  <span className="rounded bg-surface-alt px-1.5 py-0.5 text-text-muted">
                    {r.crowd_actual}
                  </span>
                )}
              </div>
            </li>
          ))}
        </ul>
      ) : (
        <p className="rounded-xl border border-dashed border-border px-3 py-4 text-sm text-text-faint">
          Belum ada penilaian. Penilaian bisa diisi setelah nobar berlangsung.
        </p>
      )}
    </section>
  );
}

function Riwayat({ venue }: { venue: VenueDetail }) {
  if (venue.nobar_history.length === 0) return null;

  return (
    <section>
      <h2 className="judul-bagian mb-2 font-semibold">Pernah nobar di sini</h2>
      <ul className="divide-y divide-border overflow-hidden rounded-xl border border-border bg-surface text-sm">
        {venue.nobar_history.map((h, i) => (
          <li key={i} className="flex items-center justify-between gap-3 px-3 py-2.5">
            <div className="min-w-0">
              <p className="truncate font-medium">{h.match_label}</p>
              <p className="text-xs text-text-faint">{formatDate(h.kickoff_at)}</p>
            </div>
            <div className="flex shrink-0 items-center gap-2">
              <span className="text-xs text-text-muted">
                {formatEntry(h.entry_type, h.entry_amount)}
              </span>
              <StatusBadge isConfirmed={h.is_confirmed} />
            </div>
          </li>
        ))}
      </ul>
    </section>
  );
}
