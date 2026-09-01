import type { Metadata } from "next";
import { notFound } from "next/navigation";

import { BelumAdaNobar } from "@/components/EmptyState";
import { MatchCard } from "@/components/MatchCard";
import { NobarList } from "@/components/NobarList";
import { ApiError, getFacilities, getNobarForMatch, getSeasonMatches } from "@/lib/api";
import { BANDUNG } from "@/lib/geo";

export async function generateMetadata(props: PageProps<"/match/[id]">): Promise<Metadata> {
  const { id } = await props.params;
  const match = await cariMatch(Number(id));
  if (!match) return { title: "Laga tidak ditemukan" };
  return {
    title: `Nobar ${match.home_team.name} vs ${match.away_team.name}`,
    description: `Daftar tempat nonton bareng ${match.home_team.name} vs ${match.away_team.name} di Bandung, terurut jarak.`,
  };
}

/** Daftar nobar untuk satu laga yang dipilih dari halaman /jadwal. */
export default async function HalamanMatch(props: PageProps<"/match/[id]">) {
  const { id } = await props.params;
  const matchId = Number(id);
  if (!Number.isInteger(matchId) || matchId <= 0) notFound();

  let nobar;
  try {
    nobar = await getNobarForMatch(matchId, {
      lat: BANDUNG.lat,
      lng: BANDUNG.lng,
      sort: "recommended",
      per_page: 20,
    });
  } catch (e) {
    if (e instanceof ApiError && e.status === 404) notFound();
    throw e;
  }

  const [match, facilities] = await Promise.all([cariMatch(matchId), getFacilities()]);

  return (
    <div className="space-y-4 sm:space-y-5">
      {match ? (
        <MatchCard match={match} />
      ) : (
        <section className="rounded-2xl border border-border bg-surface p-4">
          <h1 className="font-bold">{nobar.meta.match.label}</h1>
        </section>
      )}

      {nobar.meta.total === 0 ? (
        <BelumAdaNobar />
      ) : (
        <NobarList matchId={matchId} awal={nobar} facilities={facilities} />
      )}
    </div>
  );
}

/**
 * Mengambil detail laga dari daftar musim.
 *
 * API belum punya `GET /v1/matches/{id}` tersendiri — endpoint itu belum
 * dibutuhkan siapa pun sampai sekarang. Daftar musim satu laga per klub relatif
 * pendek (34 laga liga per musim menurut §1), jadi menyaringnya di sini jauh
 * lebih murah daripada menambah endpoint yang cuma dipakai satu halaman.
 */
async function cariMatch(id: number) {
  try {
    const musim = await getSeasonMatches();
    return musim.find((m) => m.id === id) ?? null;
  } catch {
    return null;
  }
}
