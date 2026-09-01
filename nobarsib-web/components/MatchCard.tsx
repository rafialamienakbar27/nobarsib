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

import { formatKickoff } from "@/lib/format";
import type { Match } from "@/lib/types";

/**
 * Kartu laga (§13.2).
 *
 * Diletakkan paling atas karena §3.1 prinsip 1: aplikasi ini berpusat pada
 * PERTANDINGAN, bukan pada venue. Pengguna datang dengan pertanyaan "Sabtu
 * nanti nonton di mana", jadi laganya harus terjawab lebih dulu.
 */
export function MatchCard({ match, sebagaiTautan = false }: { match: Match; sebagaiTautan?: boolean }) {
  // Dua tampilan yang sengaja berbeda.
  //
  // Sebagai KEPALA beranda, kartu ini satu-satunya di halaman dan boleh
  // mengambil bidang biru penuh — ia menjawab "laga apa" sebelum mata turun ke
  // daftar venue.
  //
  // Sebagai BARIS di /jadwal ia berdampingan dengan belasan kartu lain; bidang
  // biru berulang akan berubah jadi dinding warna yang justru tidak terbaca.
  // Di sana warnanya turun jadi aksen tipis saja.
  return sebagaiTautan ? <BarisJadwal match={match} /> : <KepalaLaga match={match} />;
}

function KepalaLaga({ match }: { match: Match }) {
  return (
    // Di HP kartu ini menumpuk ke bawah; mulai md ia jadi satu baris — laga di
    // kiri, jumlah tempat nobar di kanan. Bidang biru selebar layar dengan
    // tulisan yang menggerombol di kiri atas justru terbaca sebagai kartu yang
    // "kurang isi", padahal isinya persis sama.
    <section className="bidang-brand overflow-hidden rounded-2xl p-4 shadow-brand sm:p-6 md:flex md:items-center md:justify-between md:gap-8">
      <div className="min-w-0">
        <p className="inline-flex rounded-full bg-white/15 px-2.5 py-1 text-[11px] font-bold uppercase tracking-[0.12em]">
          Laga berikutnya
        </p>

        {/* text-lg di HP, bukan text-xl: "Persija Jakarta vs Persib Bandung"
            masih muat satu baris di layar 360px pada ukuran ini, dan tidak
            terpotong. Baru dinaikkan setelah lebarnya memang ada. */}
        <div className="mt-3 flex items-center gap-2.5 text-lg leading-tight sm:gap-3 sm:text-xl lg:text-2xl">
          <TeamName team={match.home_team} />
          <span className="shrink-0 text-sm font-semibold text-white/55">vs</span>
          <TeamName team={match.away_team} />
        </div>

        <p className="mt-2.5 text-sm font-medium text-white/85">
          {formatKickoff(match.kickoff_at)}
        </p>
      </div>

      {/* Dipisah garis: jumlah tempat nobar adalah yang menghubungkan kartu ini
          ke daftar di bawahnya, bukan sekadar keterangan tambahan laga.
          Garisnya pindah dari atas ke samping saat kartu jadi satu baris. */}
      <div className="mt-3 flex flex-wrap items-center gap-x-2 gap-y-1 border-t border-white/15 pt-3 text-sm md:mt-0 md:shrink-0 md:flex-col md:items-end md:gap-y-0.5 md:border-l md:border-t-0 md:pl-8 md:pt-0 md:text-right">
        <span className="font-semibold md:text-base">
          {match.nobar_count > 0
            ? `${match.nobar_count} tempat nobar`
            : "Belum ada info nobar"}
        </span>
        {match.broadcast_tv && (
          <span className="text-white/70">
            <span className="md:hidden">· </span>Siaran {match.broadcast_tv}
          </span>
        )}
      </div>
    </section>
  );
}

function BarisJadwal({ match }: { match: Match }) {
  return (
    <Link
      href={`/match/${match.id}`}
      // h-full: di /jadwal kartu ini berdiri dalam grid, dan baris yang
      // kartunya tidak sama tinggi terbaca sebagai daftar yang berantakan.
      className="block h-full rounded-2xl border border-border bg-surface p-4 shadow-card transition-colors hover:border-brand-line hover:bg-surface-alt"
    >
      <p className="text-xs font-semibold uppercase tracking-wider text-brand-accent">
        {match.competition}
      </p>

      <div className="mt-2 flex items-center gap-3">
        <TeamName team={match.home_team} />
        <span className="text-sm font-medium text-text-faint">vs</span>
        <TeamName team={match.away_team} />
      </div>

      <p className="mt-2 text-sm text-text-muted">{formatKickoff(match.kickoff_at)}</p>

      <p className="mt-1 text-sm text-text-muted">
        {match.broadcast_tv && <>{match.broadcast_tv} · </>}
        <span className="font-medium text-text">
          {match.nobar_count > 0
            ? `${match.nobar_count} tempat nobar`
            : "Belum ada info nobar"}
        </span>
      </p>
    </Link>
  );
}

/**
 * Nama tim. Persib ditebalkan supaya laga yang dicari langsung terbaca, tapi
 * penandanya datang dari data (`is_featured` di tabel team), bukan dari
 * perbandingan nama di kode — §3.1 prinsip 2 melarang hardcode Persib.
 */
function TeamName({ team }: { team: Match["home_team"] }) {
  return (
    <span className="flex min-w-0 items-center gap-1.5">
      {team.logo_url && (
        <img src={team.logo_url} alt="" width={20} height={20} className="size-5 object-contain" />
      )}
      <span className="truncate font-bold">{team.name}</span>
    </span>
  );
}
