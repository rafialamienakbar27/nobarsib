"use client";

import Link from "next/link";
import { useEffect, useState } from "react";

import { getSummary, type AdminSummary } from "@/lib/admin";
import { formatKickoff } from "@/lib/format";

/** Dashboard ringkasan (§14.1). */
export default function HalamanRingkasan() {
  const [data, setData] = useState<AdminSummary | null>(null);
  const [galat, setGalat] = useState<string | null>(null);

  useEffect(() => {
    getSummary()
      .then(setData)
      .catch((e) => setGalat(e instanceof Error ? e.message : "Gagal memuat"));
  }, []);

  if (galat) {
    return <p className="rounded-xl bg-warn-soft px-3 py-2 text-sm text-warn">{galat}</p>;
  }
  if (!data) {
    return <p className="py-8 text-center text-sm text-text-faint">Memuat…</p>;
  }

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-3">
        <Link
          href="/admin/antrian"
          className="rounded-2xl border border-border bg-surface p-4 transition-colors hover:border-brand-line"
        >
          <p className="text-3xl font-extrabold">{data.pending_events}</p>
          <p className="mt-1 text-sm text-text-muted">Menunggu tinjauan</p>
        </Link>
        <div className="rounded-2xl border border-border bg-surface p-4">
          <p className="text-3xl font-extrabold">{data.upcoming_matches}</p>
          <p className="mt-1 text-sm text-text-muted">Laga dalam sebulan</p>
        </div>
      </div>

      {data.next_match ? (
        <section className="rounded-2xl border border-border bg-surface p-4">
          <p className="text-xs font-semibold uppercase tracking-wider text-brand-accent">
            Laga terdekat
          </p>
          <p className="mt-1 font-bold">
            {data.next_match.home_team.name} vs {data.next_match.away_team.name}
          </p>
          <p className="mt-0.5 text-sm text-text-muted">
            {formatKickoff(data.next_match.kickoff_at)}
          </p>
          <p className="mt-1 text-sm">
            <span className="font-semibold">{data.next_match.nobar_count}</span>{" "}
            <span className="text-text-muted">venue tayang</span>
          </p>
        </section>
      ) : (
        <p className="rounded-2xl border border-dashed border-border px-4 py-8 text-center text-sm text-text-muted">
          Belum ada laga terjadwal. Tambahkan di menu Jadwal.
        </p>
      )}
    </div>
  );
}
