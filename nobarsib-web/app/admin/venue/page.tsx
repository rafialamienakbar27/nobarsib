"use client";

import { useState } from "react";

import { adminFetch } from "@/lib/admin";
import { apiBase } from "@/lib/api";
import type { Facility, VenueCard } from "@/lib/types";
import { useEffect } from "react";

/**
 * Kelola venue (§14.1).
 *
 * Formulir ini untuk menambah satu-dua venue atau memperbaiki data yang keliru.
 * Untuk mengetik 20–30 venue hasil survei sekaligus, pakai importer:
 *
 *     make import FILE=nobarsib-api/testdata/venues.json
 *
 * Berkas JSON bisa diperiksa ulang, disimpan, dan dijalankan lagi setelah
 * dikoreksi — tiga hal yang tidak bisa dilakukan formulir.
 */
export default function HalamanVenue() {
  const [q, setQ] = useState("");
  const [hasil, setHasil] = useState<VenueCard[]>([]);
  const [fasilitas, setFasilitas] = useState<Facility[]>([]);
  const [pesan, setPesan] = useState<string | null>(null);
  const [galat, setGalat] = useState<string | null>(null);
  const [mengirim, setMengirim] = useState(false);

  useEffect(() => {
    fetch(`${apiBase()}/facilities`)
      .then((r) => r.json())
      .then((d: { data: Facility[] | null }) => setFasilitas(d.data ?? []))
      .catch(() => {});
  }, []);

  async function cari(e: React.FormEvent) {
    e.preventDefault();
    if (q.trim().length < 2) return;
    const r = await fetch(`${apiBase()}/venues/search?q=${encodeURIComponent(q)}`);
    const d = (await r.json()) as { data: VenueCard[] | null };
    setHasil(d.data ?? []);
  }

  async function simpan(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setMengirim(true);
    setPesan(null);
    setGalat(null);

    const f = new FormData(e.currentTarget);
    const pilihan = f.getAll("facilities").map(String);

    try {
      const res = await adminFetch<{ data: { slug: string; data_completeness: number } }>(
        "/admin/venues",
        {
          method: "POST",
          body: JSON.stringify({
            name: f.get("name"),
            address: f.get("address"),
            district: f.get("district"),
            lat: Number(f.get("lat")),
            lng: Number(f.get("lng")),
            whatsapp: f.get("whatsapp"),
            instagram_handle: f.get("instagram_handle"),
            facilities: pilihan,
            status: "unclaimed",
          }),
        },
      );
      setPesan(
        `Venue "${res.data.slug}" tersimpan. Kelengkapan data ${res.data.data_completeness.toFixed(2)} ` +
          `— jam buka dan foto belum terisi, lengkapi lewat importer supaya naik.`,
      );
      e.currentTarget.reset();
    } catch (err) {
      setGalat(err instanceof Error ? err.message : "Gagal menyimpan");
    } finally {
      setMengirim(false);
    }
  }

  return (
    <div className="space-y-6">
      <section>
        <h2 className="judul-bagian mb-2 font-semibold">Cari venue</h2>
        <form onSubmit={cari} className="flex gap-2">
          <input
            value={q}
            onChange={(ev) => setQ(ev.target.value)}
            placeholder="nama venue…"
            className="min-h-11 flex-1 rounded-xl border border-border bg-surface px-3"
          />
          <button className="min-h-11 rounded-xl border border-border px-4 text-sm font-medium">
            Cari
          </button>
        </form>
        {hasil.length > 0 && (
          <ul className="mt-2 divide-y divide-border overflow-hidden rounded-xl border border-border bg-surface text-sm">
            {hasil.map((v) => (
              <li key={v.id} className="flex items-center justify-between gap-3 px-3 py-2">
                <span className="truncate">
                  {v.name}
                  <span className="text-text-faint"> · {v.district}</span>
                </span>
                <a
                  href={`/venue/${v.slug}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="shrink-0 text-brand-accent underline underline-offset-2"
                >
                  lihat
                </a>
              </li>
            ))}
          </ul>
        )}
      </section>

      <section>
        <h2 className="mb-1 font-semibold">Tambah venue</h2>
        <p className="mb-3 text-sm text-text-muted">
          Untuk memasukkan banyak venue sekaligus, pakai{" "}
          <code className="rounded bg-surface-alt px-1 text-xs">make import</code> — lebih
          cepat dan hasilnya bisa dikoreksi lalu dijalankan ulang.
        </p>

        <form onSubmit={simpan} className="space-y-3">
          <Isian nama="name" label="Nama venue" wajib />
          <Isian nama="address" label="Alamat" wajib />
          <Isian nama="district" label="Kecamatan" />

          <div className="grid grid-cols-2 gap-3">
            <Isian nama="lat" label="Lintang (lat)" tipe="number" step="any" wajib
                   petunjuk="klik kanan di Google Maps" />
            <Isian nama="lng" label="Bujur (lng)" tipe="number" step="any" wajib />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <Isian nama="whatsapp" label="WhatsApp" petunjuk="628xxxxxxxxxx" />
            <Isian nama="instagram_handle" label="Instagram" petunjuk="tanpa @" />
          </div>

          <fieldset>
            <legend className="text-sm font-medium">Fasilitas</legend>
            <div className="mt-1.5 grid grid-cols-2 gap-1.5">
              {fasilitas.map((f) => (
                <label key={f.code} className="flex items-center gap-2 text-sm">
                  <input type="checkbox" name="facilities" value={f.code} />
                  {f.label}
                </label>
              ))}
            </div>
          </fieldset>

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
            {mengirim ? "Menyimpan…" : "Simpan venue"}
          </button>
        </form>
      </section>
    </div>
  );
}

function Isian({
  nama,
  label,
  tipe = "text",
  step,
  wajib,
  petunjuk,
}: {
  nama: string;
  label: string;
  tipe?: string;
  step?: string;
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
        step={step}
        required={wajib}
        className="mt-1 min-h-11 w-full rounded-xl border border-border bg-surface px-3"
      />
      {petunjuk && <span className="mt-0.5 block text-xs text-text-faint">{petunjuk}</span>}
    </label>
  );
}
