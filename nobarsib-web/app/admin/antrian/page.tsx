"use client";

import { useCallback, useEffect, useState } from "react";

import { getAntrian, tinjauEvent, type AdminEvent } from "@/lib/admin";
import { formatKickoff, formatRupiah, formatTime } from "@/lib/format";

/**
 * Antrian tinjauan (§14.2).
 *
 * Blueprint menetapkan target: satu event selesai ditinjau dalam 10 detik.
 * Yang dilakukan demi itu:
 *
 *   - Semua yang dibutuhkan untuk memutuskan ada di kartu, tanpa perlu membuka
 *     apa pun: venue, laga, jam kickoff, jam buka pintu, biaya, catatan.
 *   - Tiga tombol saja, dan yang paling sering dipakai (Setujui) paling besar.
 *   - Kartu langsung hilang dari daftar begitu ditekan, tanpa memuat ulang
 *     seluruh antrian — menunggu jaringan tiga kali per event akan
 *     menghabiskan anggaran 10 detik itu sendirian.
 *   - Kalau ternyata gagal, kartunya kembali beserta pesan galat.
 */
export default function HalamanAntrian() {
  const [events, setEvents] = useState<AdminEvent[]>([]);
  const [memuat, setMemuat] = useState(true);
  const [galat, setGalat] = useState<string | null>(null);
  const [sibuk, setSibuk] = useState<string | null>(null);

  // muat() sengaja TIDAK memanggil setMemuat(true) di awal: dipanggil dari
  // dalam useEffect, itu berarti setState sinkron di dalam effect, yang memicu
  // render bertingkat dan ditandai sebagai kesalahan oleh React 19.
  // Keadaan awal sudah `memuat = true`, dan tombol "Muat ulang" yang
  // menyalakannya kembali saat dipanggil manual.
  const muat = useCallback(async () => {
    try {
      // Tidak ada setState sebelum await pertama — lihat catatan di atas.
      const data = await getAntrian();
      setEvents(data);
      setGalat(null);
    } catch (e) {
      setGalat(e instanceof Error ? e.message : "Gagal memuat antrian");
    } finally {
      setMemuat(false);
    }
  }, []);

  useEffect(() => {
    /*
      Mengambil data saat komponen dipasang memang harus terjadi di sini:
      antrian dilindungi token yang tersimpan di localStorage, jadi tidak bisa
      diambil komponen server seperti halaman publik. Menarik pustaka data
      (SWR/React Query) hanya demi satu halaman admin tidak sepadan dengan
      dependensinya.
    */
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void muat();
  }, [muat]);

  function muatUlang() {
    setMemuat(true);
    void muat();
  }

  async function tinjau(e: AdminEvent, aksi: "approve" | "reject") {
    setSibuk(e.id);
    setGalat(null);

    const sebelum = events;
    setEvents((d) => d.filter((x) => x.id !== e.id)); // hilang seketika

    try {
      await tinjauEvent(e.id, aksi);
    } catch (err) {
      setEvents(sebelum); // gagal: kembalikan
      setGalat(err instanceof Error ? err.message : "Gagal memproses");
    } finally {
      setSibuk(null);
    }
  }

  if (memuat) {
    return <p className="py-8 text-center text-sm text-text-faint">Memuat antrian…</p>;
  }

  return (
    <section className="space-y-3">
      <div className="flex items-center justify-between">
        <h2 className="font-semibold">
          {events.length > 0
            ? `${events.length} menunggu tinjauan`
            : "Antrian kosong"}
        </h2>
        <button
          onClick={muatUlang}
          className="min-h-9 rounded-lg border border-border px-3 text-sm text-text-muted"
        >
          Muat ulang
        </button>
      </div>

      {galat && (
        <p className="rounded-xl bg-warn-soft px-3 py-2 text-sm text-warn" role="alert">
          {galat}
        </p>
      )}

      {events.length === 0 ? (
        <p className="rounded-2xl border border-dashed border-border px-4 py-10 text-center text-sm text-text-muted">
          Tidak ada pengumuman nobar yang menunggu tinjauan.
        </p>
      ) : (
        <ul className="space-y-2">
          {events.map((e) => (
            <li key={e.id} className="rounded-xl border border-border bg-surface p-4">
              <p className="font-semibold">{e.venue_name}</p>
              <p className="text-sm text-text-muted">
                {e.district && <>{e.district} · </>}
                {e.match_label}
              </p>
              <p className="mt-1 text-sm text-text-muted">{formatKickoff(e.kickoff_at)}</p>
              <p className="mt-0.5 text-sm text-text-muted">
                {e.doors_open_at && <>Buka pintu {formatTime(e.doors_open_at)} · </>}
                {e.entry_type === "free" ? "Gratis" : formatRupiah(e.entry_amount)}
              </p>
              {e.notes && (
                <p className="mt-1 rounded-lg bg-surface-alt px-2.5 py-1.5 text-sm text-text-muted">
                  {e.notes}
                </p>
              )}

              <div className="mt-3 flex gap-2">
                <button
                  onClick={() => void tinjau(e, "approve")}
                  disabled={sibuk === e.id}
                  className="min-h-11 flex-1 rounded-xl bg-brand font-semibold text-on-brand shadow-brand transition-colors hover:bg-brand-lift disabled:opacity-60"
                >
                  Setujui
                </button>
                <a
                  href={`/venue/${e.venue_slug}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex min-h-11 items-center rounded-xl border border-border px-4 text-sm font-medium"
                >
                  Lihat venue
                </a>
                <button
                  onClick={() => void tinjau(e, "reject")}
                  disabled={sibuk === e.id}
                  className="min-h-11 rounded-xl border border-border px-4 text-sm font-medium text-text-muted disabled:opacity-60"
                >
                  Tolak
                </button>
              </div>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
