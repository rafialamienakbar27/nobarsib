import Link from "next/link";

import { EmptyState, TidakAdaLaga } from "@/components/EmptyState";
import { MatchCard } from "@/components/MatchCard";
import { NobarList } from "@/components/NobarList";
import { getFacilities, getNobarForMatch, getUpcomingMatches } from "@/lib/api";
import { BANDUNG } from "@/lib/geo";

/**
 * Beranda (§13.2).
 *
 * Dirender di server dengan lokasi jatuh ke pusat kota Bandung, supaya daftar
 * sudah terbaca sebelum JavaScript dimuat dan sebelum pengguna menjawab dialog
 * izin lokasi. NobalList kemudian menyegarkan urutannya dengan lokasi asli.
 */
export default async function Beranda() {
  let matches;
  try {
    matches = await getUpcomingMatches(1);
  } catch {
    return <GagalMemuat />;
  }

  const match = matches[0];
  if (!match) {
    return <TidakAdaLaga />;
  }

  const [nobar, facilities] = await Promise.all([
    getNobarForMatch(match.id, {
      lat: BANDUNG.lat,
      lng: BANDUNG.lng,
      sort: "recommended",
      per_page: 20,
    }),
    getFacilities(),
  ]);

  return (
    <div className="space-y-4 sm:space-y-5">
      <MatchCard match={match} />
      <NobarList matchId={match.id} awal={nobar} facilities={facilities} />

      {/* Ajakan memasang diletakkan SETELAH daftar, bukan sebelumnya: orang yang
          sudah menggulir sampai sini sudah tahu aplikasinya berguna. Spanduk
          pasang di layar pertama cuma menghalangi jawaban yang mereka cari. */}
      <section className="mt-2 flex flex-col gap-3 rounded-2xl border border-brand-line bg-brand-soft p-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <p className="font-semibold">Pasang NOBARSIB di HP</p>
          <p className="mt-0.5 text-sm text-text-muted">
            Ikonnya di layar utama, tidak perlu mengetik alamatnya lagi tiap laga.
          </p>
        </div>
        <Link
          href="/unduh"
          className="flex min-h-11 shrink-0 items-center justify-center rounded-xl bg-brand px-5 font-semibold text-on-brand shadow-brand transition-colors hover:bg-brand-lift"
        >
          Unduh aplikasi
        </Link>
      </section>

      <p className="text-center text-sm text-text-muted">
        Menayangkan laga ini di tempatmu?{" "}
        <Link href="/untuk-venue" className="font-semibold text-brand-accent underline underline-offset-2">
          Daftarkan venue
        </Link>
      </p>
    </div>
  );
}

function GagalMemuat() {
  return (
    <EmptyState
      judul="Tidak bisa memuat data"
      pesan="Server sedang tidak bisa dihubungi. Coba muat ulang halaman ini sebentar lagi."
    />
  );
}
