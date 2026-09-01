"use client";

import { useEffect, useState } from "react";

import { adminFetch } from "@/lib/admin";
import { apiBase } from "@/lib/api";
import { formatKickoff } from "@/lib/format";
import type { Match } from "@/lib/types";

/**
 * Kelola jadwal (§14.1).
 *
 * Input manual, sesuai §3.1 prinsip 3 — sinkronisasi otomatis jadwal ditunda ke
 * Fase 5. Persib main sekitar 34 laga liga per musim (§1); mengetiknya sekali
 * per putaran jauh lebih murah daripada membangun dan memelihara scraper.
 */
export default function HalamanJadwalAdmin() {
  const [matches, setMatches] = useState<Match[]>([]);
  const [pesan, setPesan] = useState<string | null>(null);
  const [galat, setGalat] = useState<string | null>(null);
  const [mengirim, setMengirim] = useState(false);

  const muat = () =>
    fetch(`${apiBase()}/matches`)
      .then((r) => r.json())
      .then((d: { data: Match[] | null }) => setMatches(d.data ?? []))
      .catch(() => setGalat("Gagal memuat jadwal"));

  useEffect(() => {
    void muat();
  }, []);

  async function simpan(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setMengirim(true);
    setPesan(null);
    setGalat(null);

    const form = e.currentTarget;
    const f = new FormData(form);

    // Input datetime-local tidak membawa zona waktu. Seluruh jadwal Liga 1
    // ditulis dalam WIB, jadi offset +07:00 ditempelkan eksplisit — tanpa itu,
    // server akan menafsirkannya sebagai UTC dan laga bergeser tujuh jam.
    const kickoff = `${String(f.get("kickoff_at"))}:00+07:00`;

    try {
      await adminFetch("/admin/matches", {
        method: "POST",
        body: JSON.stringify({
          competition_id: 1,
          home_team_slug: f.get("home_team_slug"),
          away_team_slug: f.get("away_team_slug"),
          kickoff_at: kickoff,
          venue_name: f.get("venue_name"),
          broadcast_tv: f.get("broadcast_tv"),
        }),
      });
      setPesan("Laga tersimpan.");
      form.reset();
      await muat();
    } catch (err) {
      setGalat(err instanceof Error ? err.message : "Gagal menyimpan");
    } finally {
      setMengirim(false);
    }
  }

  return (
    <div className="space-y-6">
      <section>
        <h2 className="judul-bagian mb-2 font-semibold">Tambah laga</h2>
        <form onSubmit={simpan} className="space-y-3">
          <div className="grid grid-cols-2 gap-3">
            <Isian nama="home_team_slug" label="Tuan rumah (slug)" petunjuk="persija-jakarta" wajib />
            <Isian nama="away_team_slug" label="Tamu (slug)" petunjuk="persib-bandung" wajib />
          </div>
          <Isian nama="kickoff_at" label="Kickoff (WIB)" tipe="datetime-local" wajib />
          <div className="grid grid-cols-2 gap-3">
            <Isian nama="venue_name" label="Stadion" />
            <Isian nama="broadcast_tv" label="Siaran" petunjuk="Indosiar" />
          </div>

          {pesan && (
            <p className="rounded-xl bg-confirm-soft px-3 py-2 text-sm text-confirm">{pesan}</p>
          )}
          {galat && (
            <p className="rounded-xl bg-warn-soft px-3 py-2 text-sm text-warn" role="alert">
              {galat}
            </p>
          )}

          <button
            type="submit"
            disabled={mengirim}
            className="min-h-12 w-full rounded-xl bg-brand font-semibold text-on-brand shadow-brand transition-colors hover:bg-brand-lift disabled:opacity-60"
          >
            {mengirim ? "Menyimpan…" : "Simpan laga"}
          </button>
        </form>
      </section>

      <section>
        <h2 className="judul-bagian mb-2 font-semibold">Jadwal tersimpan ({matches.length})</h2>
        {matches.length === 0 ? (
          <p className="rounded-xl border border-dashed border-border px-4 py-8 text-center text-sm text-text-muted">
            Belum ada laga.
          </p>
        ) : (
          <ul className="divide-y divide-border overflow-hidden rounded-xl border border-border bg-surface text-sm">
            {matches.map((m) => (
              <li key={m.id} className="px-3 py-2.5">
                <p className="font-medium">
                  {m.home_team.name} vs {m.away_team.name}
                </p>
                <p className="text-xs text-text-faint">
                  {formatKickoff(m.kickoff_at)} · {m.nobar_count} venue tayang
                </p>
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  );
}

function Isian({
  nama,
  label,
  tipe = "text",
  wajib,
  petunjuk,
}: {
  nama: string;
  label: string;
  tipe?: string;
  wajib?: boolean;
  petunjuk?: string;
}) {
  return (
    <label className="block">
      <span className="text-sm font-medium">
        {label}
        {wajib && <span className="text-warn"> *</span>}
      </span>
      <input
        name={nama}
        type={tipe}
        required={wajib}
        className="mt-1 min-h-11 w-full rounded-xl border border-border bg-surface px-3"
      />
      {petunjuk && <span className="mt-0.5 block text-xs text-text-faint">contoh: {petunjuk}</span>}
    </label>
  );
}
