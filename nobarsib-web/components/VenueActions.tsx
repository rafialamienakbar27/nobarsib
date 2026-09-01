"use client";

import { track } from "@/lib/api";
import { deviceHash } from "@/lib/device";
import { instagramLink, mapsLink, waLink } from "@/lib/format";
import type { VenueDetail } from "@/lib/types";

/**
 * Tombol aksi di halaman detail (§13.4 blok 5).
 *
 * §13.3 menyebut "Buka di Maps" dan "Chat WA" sebagai dua tombol yang
 * benar-benar dipakai orang, jadi keduanya dibuat sejajar dan paling menonjol.
 * Instagram menyusul sebagai tautan sekunder.
 *
 * Setiap penekanan dicatat lewat /track: klik "Buka di Maps" adalah indikator
 * niat datang yang paling kuat (§15.4), dan itulah angka yang nanti membuat
 * venue mau berlangganan.
 */
export function VenueActions({
  venue,
  eventId,
}: {
  venue: VenueDetail;
  eventId?: string;
}) {
  const catat = (aksi: "open_maps" | "click_wa") => {
    if (eventId) track(eventId, aksi, deviceHash());
  };

  const pesanWA = `Halo ${venue.name}, mau tanya soal nobar Persib. Masih ada tempat?`;

  return (
    <div className="grid grid-cols-2 gap-2">
      <a
        href={mapsLink(venue.lat, venue.lng, venue.name)}
        target="_blank"
        rel="noopener noreferrer"
        onClick={() => catat("open_maps")}
        className="flex min-h-12 items-center justify-center gap-2 rounded-xl bg-brand px-4 font-semibold text-on-brand shadow-brand transition-colors hover:bg-brand-lift"
      >
        <svg viewBox="0 0 16 16" className="size-4" fill="currentColor" aria-hidden>
          <path d="M8 1a5 5 0 0 0-5 5c0 3.6 5 9 5 9s5-5.4 5-9a5 5 0 0 0-5-5m0 7a2 2 0 1 1 0-4 2 2 0 0 1 0 4" />
        </svg>
        Buka di Maps
      </a>

      {venue.whatsapp ? (
        <a
          href={waLink(venue.whatsapp, pesanWA)}
          target="_blank"
          rel="noopener noreferrer"
          onClick={() => catat("click_wa")}
          className="flex min-h-12 items-center justify-center gap-2 rounded-xl border border-border bg-surface px-4 font-semibold shadow-card transition-colors hover:border-brand-line"
        >
          <svg viewBox="0 0 16 16" className="size-4" fill="currentColor" aria-hidden>
            <path d="M8 1.5A6.4 6.4 0 0 0 2.5 11L1.6 14.4l3.5-.9A6.4 6.4 0 1 0 8 1.5m3.6 9c-.2.5-.9.9-1.3.9-.4 0-.8.2-2.5-.6a8 8 0 0 1-3.2-3.4c-.3-.5-.6-1.2-.6-1.8s.3-1 .5-1.2c.1-.2.3-.2.4-.2h.4c.1 0 .3 0 .4.3l.6 1.4c0 .1 0 .3-.1.4l-.3.3c-.1.1-.2.2-.1.4a5 5 0 0 0 2.4 2.1c.2 0 .3 0 .4-.1l.5-.6c.1-.2.3-.2.4-.1l1.3.7c.2.1.3.2.3.3s0 .4-.1.6" />
          </svg>
          Chat WA
        </a>
      ) : (
        <span className="flex min-h-12 items-center justify-center rounded-xl border border-dashed border-border px-4 text-sm text-text-faint">
          Nomor WA belum ada
        </span>
      )}

      {venue.instagram_handle && (
        <a
          href={instagramLink(venue.instagram_handle)}
          target="_blank"
          rel="noopener noreferrer"
          className="col-span-2 flex min-h-11 items-center justify-center rounded-xl border border-border bg-surface px-4 text-sm font-medium text-text-muted"
        >
          @{venue.instagram_handle.replace(/^@/, "")} di Instagram
        </a>
      )}
    </div>
  );
}
