"use client";

import { useCallback, useEffect, useRef, useState } from "react";

import { getNobarForMatch } from "@/lib/api";
import {
  BANDUNG,
  lupakanPenolakan,
  mintaLokasi,
  pantauIzin,
  pernahMenolak,
  statusIzin,
  type IzinLokasi,
} from "@/lib/location";
import type { Facility, NobarEvent, NobarListResponse, SortMode } from "@/lib/types";

import {
  AjakanLokasi,
  FilterChips,
  LocationBar,
  SortTabs,
  type FilterState,
} from "./Controls";
import { BelumAdaNobar, EmptyState } from "./EmptyState";
import { VenueCard, VenueCardSkeleton } from "./VenueCard";

const RADIUS_AWAL = 15;
const RADIUS_LUAS = 50;

/**
 * Grid daftar venue.
 *
 * Satu kolom di HP — itu bentuk aslinya dan tidak berubah. Yang berubah adalah
 * apa yang terjadi pada ruang berlebih di layar besar: ia dipakai untuk
 * menambah KOLOM, bukan melebarkan kartu. Kartu selebar 1000px berisi lima
 * baris teks pendek meninggalkan bidang kosong di kanan setiap barisnya, dan
 * daftar dua puluh venue jadi gulungan yang panjangnya tidak masuk akal.
 *
 * Ambang dipilih dari lebar kartu yang masih nyaman (~330–360px), bukan dari
 * angka bulat: dua kolom baru muncul di md (768px), tiga di xl (1280px).
 */
const GRID_KARTU = "grid grid-cols-1 gap-2.5 md:grid-cols-2 xl:grid-cols-3";

/**
 * Daftar venue untuk satu laga — layar utama aplikasi.
 *
 * Hasil pertama dirender di server memakai pusat kota Bandung, sehingga daftar
 * sudah terbaca sebelum JavaScript selesai dimuat (§13.6: LCP < 2,5 detik di 4G)
 * dan tetap berguna kalau izin lokasi ditolak. Komponen ini kemudian meminta
 * lokasi asli dan menyegarkan urutannya.
 */
export function NobarList({
  matchId,
  awal,
  facilities,
}: {
  matchId: number;
  awal: NobarListResponse;
  facilities: Facility[];
}) {
  const [events, setEvents] = useState<NobarEvent[]>(awal.data);
  const [total, setTotal] = useState(awal.meta.total);
  const [sort, setSort] = useState<SortMode>("recommended");
  const [filter, setFilter] = useState<FilterState>({
    facilities: [],
    gratis: false,
    openUntilEnd: true,
  });
  const [radius, setRadius] = useState(RADIUS_AWAL);
  const [koordinat, setKoordinat] = useState<{ lat: number; lng: number } | null>(null);
  const [memuat, setMemuat] = useState(false);
  const [mintaIzin, setMintaIzin] = useState(false);
  const [galat, setGalat] = useState<string | null>(null);
  const [izin, setIzin] = useState<IzinLokasi>("unknown");
  // Dibedakan dari `izin`: percobaan yang gagal karena sinyal bukan penolakan,
  // dan pesannya pun harus berbeda.
  const [lokasiGagal, setLokasiGagal] = useState(false);
  // Pernah ditolak di sesi ini, TAPI peramban tidak memastikannya sebagai
  // blokir permanen. Keadaan ini nyata dan sering: menutup dialog izin tanpa
  // memilih apa pun menghasilkan PERMISSION_DENIED, sementara status izinnya
  // tetap "prompt" — bertanya lagi masih akan memunculkan dialog. Menyamakannya
  // dengan blokir permanen berarti menyuruh orang mengubah setelan peramban
  // untuk sesuatu yang cukup diselesaikan dengan satu ketukan tombol.
  const [ditolak, setDitolak] = useState(false);

  // Melewati pengambilan ulang pada render pertama: datanya sudah datang dari
  // server, dan memuat ulang seketika hanya membuat daftar berkedip.
  const pertama = useRef(true);

  const ambil = useCallback(async () => {
    setMemuat(true);
    setGalat(null);
    try {
      const res = await getNobarForMatch(matchId, {
        lat: koordinat?.lat ?? BANDUNG.lat,
        lng: koordinat?.lng ?? BANDUNG.lng,
        sort,
        radius_km: radius,
        facilities: filter.facilities,
        entry_type: filter.gratis ? "free" : "",
        open_until_end: filter.openUntilEnd,
        per_page: 20,
      });
      setEvents(res.data);
      setTotal(res.meta.total);
    } catch {
      setGalat("Gagal memuat daftar. Periksa koneksi lalu coba lagi.");
    } finally {
      setMemuat(false);
    }
  }, [matchId, koordinat, sort, radius, filter]);

  useEffect(() => {
    if (pertama.current) {
      pertama.current = false;
      return;
    }
    void ambil();
  }, [ambil]);

  // Izin lokasi diminta sekali saat halaman dibuka (§4.2), kecuali pengguna
  // pernah menolaknya — menanyakan berulang kali terasa seperti aplikasi yang
  // tidak ingat apa-apa.
  useEffect(() => {
    let batal = false;

    void (async () => {
      const keadaan = await statusIzin();
      if (batal) return;
      setIzin(keadaan);

      // Izin sudah diberikan mengalahkan catatan penolakan lama.
      //
      // Penanda "pernah menolak" hanya ada untuk menghindari dialog izin yang
      // muncul berulang — dan kalau izinnya sudah diberikan, tidak ada dialog
      // yang perlu dihindari. Tanpa baris ini, orang yang dulu menolak lalu
      // belakangan mengizinkan lewat setelan peramban akan tetap terkunci di
      // jarak dari pusat kota, karena penandanya tidak pernah kedaluwarsa.
      if (keadaan === "granted") {
        lupakanPenolakan();
      } else if (keadaan === "denied" || pernahMenolak()) {
        // Sudah diblokir: memanggil getCurrentPosition hanya menghasilkan
        // penolakan seketika tanpa dialog. Tidak ada gunanya dicoba.
        return;
      }

      const s = await mintaLokasi();
      if (batal) return;
      if (s.status === "granted") {
        setKoordinat({ lat: s.lat, lng: s.lng });
        setIzin("granted");
      } else if (s.status === "denied") {
        // Tanya ulang statusnya: hanya peramban yang tahu apakah ini blokir
        // permanen atau dialog yang ditutup begitu saja.
        const sesudah = await statusIzin();
        if (batal) return;
        setIzin(sesudah);
        setDitolak(true);
      } else {
        setLokasiGagal(true);
      }
    })();

    return () => {
      batal = true;
    };
  }, []);

  // Pengguna yang membuka setelan peramban lalu mengizinkan lokasi tidak perlu
  // memuat ulang halaman — persis janji yang ditulis di petunjuk AjakanLokasi.
  useEffect(() => {
    return pantauIzin((keadaan) => {
      setIzin(keadaan);
      if (keadaan !== "granted") return;
      setLokasiGagal(false);
      void mintaLokasi().then((s) => {
        if (s.status === "granted") setKoordinat({ lat: s.lat, lng: s.lng });
      });
    });
  }, []);

  const aktifkanLokasi = async () => {
    setMintaIzin(true);
    setLokasiGagal(false);
    lupakanPenolakan();
    const s = await mintaLokasi();
    setMintaIzin(false);

    if (s.status === "granted") {
      setKoordinat({ lat: s.lat, lng: s.lng });
      setIzin("granted");
      setDitolak(false);
      return;
    }
    // Tanpa dua cabang di bawah, penolakan berakhir tanpa jejak apa pun di
    // layar — tombolnya seolah rusak.
    if (s.status === "denied") {
      setIzin(await statusIzin());
      setDitolak(true);
      return;
    }
    setLokasiGagal(true);
  };

  const adaFilterAktif =
    filter.facilities.length > 0 || filter.gratis || !filter.openUntilEnd;

  return (
    <section className="space-y-3">
      <SortTabs value={sort} onChange={setSort} />
      <FilterChips facilities={facilities} value={filter} onChange={setFilter} />

      {/* Hanya di tab "Terdekat", dan hanya kalau lokasinya memang belum ada. */}
      {sort === "nearest" && koordinat === null && (
        <AjakanLokasi
          keadaan={
            izin === "denied"
              ? "diblokir"
              : lokasiGagal
                ? "gagal"
                : ditolak
                  ? "ditolak"
                  : "belum"
          }
          sedangMinta={mintaIzin}
          onAktifkan={() => void aktifkanLokasi()}
        />
      )}

      <div className="flex flex-wrap items-center justify-between gap-x-3 gap-y-1">
        <LocationBar
          pakaiLokasiAsli={koordinat !== null}
          diblokir={izin === "denied"}
          sedangMinta={mintaIzin}
          onAktifkan={() => void aktifkanLokasi()}
        />
        {!memuat && events.length > 0 && (
          <p className="shrink-0 text-xs text-text-faint">{total} tempat</p>
        )}
      </div>

      {galat && (
        <div className="rounded-xl border border-border bg-warn-soft px-4 py-3 text-sm text-warn">
          {galat}{" "}
          <button onClick={() => void ambil()} className="font-semibold underline">
            Coba lagi
          </button>
        </div>
      )}

      {memuat ? (
        <div className={GRID_KARTU} aria-busy>
          <VenueCardSkeleton />
          <VenueCardSkeleton />
          <VenueCardSkeleton />
          <VenueCardSkeleton />
          <VenueCardSkeleton />
          <VenueCardSkeleton />
        </div>
      ) : events.length > 0 ? (
        <ul className={GRID_KARTU}>
          {events.map((e) => (
            <li key={e.event_id}>
              <VenueCard event={e} />
            </li>
          ))}
        </ul>
      ) : adaFilterAktif ? (
        // Kosong karena filter, bukan karena datanya tidak ada — bedanya
        // penting supaya pengguna tahu apa yang harus diubah.
        <EmptyState
          judul="Tidak ada yang cocok dengan filter"
          pesan="Coba longgarkan filternya, atau matikan salah satu untuk melihat lebih banyak tempat."
          aksi={{
            label: "Hapus semua filter",
            onClick: () =>
              setFilter({ facilities: [], gratis: false, openUntilEnd: true }),
          }}
        />
      ) : radius < RADIUS_LUAS ? (
        // §13.5 baris 3 — radius kosong.
        <EmptyState
          judul={`Tidak ada nobar dalam ${radius} km`}
          pesan="Belum ada venue yang menayangkan laga ini di sekitarmu. Coba perluas radius pencarian."
          aksi={{ label: `Perluas ke ${RADIUS_LUAS} km`, onClick: () => setRadius(RADIUS_LUAS) }}
        />
      ) : (
        // §13.5 baris 2 — ada laga, belum ada venue sama sekali.
        <BelumAdaNobar />
      )}
    </section>
  );
}
