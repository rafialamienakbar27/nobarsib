"use client";

import type { Facility, SortMode } from "@/lib/types";

const SORT_LABEL: Record<SortMode, string> = {
  recommended: "Rekomendasi",
  nearest: "Terdekat",
  rating: "Rating",
};

/** Tab urutan (§13.2). */
export function SortTabs({
  value,
  onChange,
}: {
  value: SortMode;
  onChange: (v: SortMode) => void;
}) {
  return (
    <div
      role="tablist"
      aria-label="Urutkan venue"
      // sm:w-fit — tiga tab yang dipaksa membagi lebar 1100px menghasilkan
      // tombol selebar 350px berisi satu kata. Di HP tab tetap memenuhi lebar
      // (target sentuh), di layar besar ia menyusut ke ukuran isinya.
      className="flex gap-1 rounded-xl border border-border bg-surface-alt p-1 sm:w-fit"
    >
      {(Object.keys(SORT_LABEL) as SortMode[]).map((mode) => (
        <button
          key={mode}
          role="tab"
          aria-selected={value === mode}
          onClick={() => onChange(mode)}
          className={
            // Tab aktif diisi penuh biru, bukan putih di atas abu-abu seperti
            // sebelumnya: bedanya kini terbaca sekilas, dan tetap terbaca
            // pada layar HP di bawah sinar matahari.
            "min-h-10 flex-1 rounded-lg px-3 text-sm font-semibold transition-colors sm:flex-none sm:px-6 " +
            (value === mode
              ? "bg-brand text-on-brand shadow-sm"
              : "font-medium text-text-muted hover:text-text")
          }
        >
          {SORT_LABEL[mode]}
        </button>
      ))}
    </div>
  );
}

export interface FilterState {
  facilities: string[];
  gratis: boolean;
  openUntilEnd: boolean;
}

/**
 * Chip filter (§13.2).
 *
 * Hanya fasilitas yang benar-benar menentukan pilihan yang diangkat ke chip —
 * daftar lengkap 13 fasilitas ada di halaman detail. Chip yang terlalu banyak
 * membuat barisnya digeser tanpa henti dan justru tidak dipakai.
 */
const CHIP_UTAMA = ["area_anak", "indoor", "outdoor", "parkir_mobil", "musala", "ac"];

export function FilterChips({
  facilities,
  value,
  onChange,
}: {
  facilities: Facility[];
  value: FilterState;
  onChange: (v: FilterState) => void;
}) {
  const chips = facilities.filter((f) => CHIP_UTAMA.includes(f.code));

  const toggleFacility = (code: string) => {
    const ada = value.facilities.includes(code);
    onChange({
      ...value,
      facilities: ada
        ? value.facilities.filter((c) => c !== code)
        : [...value.facilities, code],
    });
  };

  return (
    // Digeser mendatar hanya selama layarnya memang sempit. Mulai sm barisnya
    // membungkus: di layar lebar chip yang terpotong di tepi kiri (seperti
    // "parkir mobil" yang tinggal "mobil") adalah filter yang tidak akan pernah
    // ditemukan orang, padahal ruang untuk menampung semuanya jelas ada.
    <div className="no-scrollbar -mx-4 flex gap-2 overflow-x-auto px-4 py-0.5 sm:mx-0 sm:flex-wrap sm:overflow-x-visible sm:px-0">
      <Chip
        aktif={value.gratis}
        onClick={() => onChange({ ...value, gratis: !value.gratis })}
      >
        Gratis
      </Chip>

      {chips.map((f) => (
        <Chip
          key={f.code}
          aktif={value.facilities.includes(f.code)}
          onClick={() => toggleFacility(f.code)}
        >
          {f.label}
        </Chip>
      ))}

      {/*
        Filter jam tutup dibuat bisa dimatikan, dan menyala secara default.
        §9.4: tanpa filter ini daftar akan penuh venue yang percuma, karena
        banyak cafe tutup 22.00–23.00 sementara Liga 1 kickoff 19.00 atau lebih malam.
      */}
      <Chip
        aktif={value.openUntilEnd}
        onClick={() => onChange({ ...value, openUntilEnd: !value.openUntilEnd })}
      >
        Buka sampai selesai
      </Chip>
    </div>
  );
}

function Chip({
  aktif,
  onClick,
  children,
}: {
  aktif: boolean;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      type="button"
      aria-pressed={aktif}
      onClick={onClick}
      className={
        "min-h-9 shrink-0 whitespace-nowrap rounded-full border px-3.5 text-sm transition-colors " +
        (aktif
          ? "border-brand bg-brand font-semibold text-on-brand"
          : "border-border bg-surface font-medium text-text-muted hover:border-brand-line hover:text-text")
      }
    >
      {children}
    </button>
  );
}

/** Penanda sumber lokasi + tombol mengaktifkan lokasi (§13.5 baris 4). */
export function LocationBar({
  pakaiLokasiAsli,
  diblokir,
  sedangMinta,
  onAktifkan,
}: {
  pakaiLokasiAsli: boolean;
  diblokir: boolean;
  sedangMinta: boolean;
  onAktifkan: () => void;
}) {
  if (pakaiLokasiAsli) {
    return (
      <p className="flex items-center gap-1.5 text-xs text-text-faint">
        <IkonPin />
        Jarak dihitung dari lokasimu
      </p>
    );
  }

  // Izin sudah diblokir permanen: menawarkan tombol di sini adalah berbohong.
  // Menekannya tidak akan memunculkan dialog apa pun, karena keputusannya sudah
  // tersimpan di peramban dan hanya bisa dicabut dari setelan situs.
  if (diblokir) {
    return (
      <p className="flex items-center gap-1.5 text-xs text-text-faint">
        <IkonPin />
        Jarak dari pusat kota Bandung — izin lokasi diblokir
      </p>
    );
  }

  return (
    <div className="flex flex-wrap items-center gap-x-2 gap-y-1 text-xs text-text-faint">
      <span>Jarak dihitung dari pusat kota Bandung.</span>
      <button
        type="button"
        onClick={onAktifkan}
        disabled={sedangMinta}
        className="font-semibold text-brand-accent underline underline-offset-2 disabled:opacity-60"
      >
        {sedangMinta ? "Meminta izin…" : "Aktifkan lokasi"}
      </button>
    </div>
  );
}

function IkonPin() {
  return (
    <svg viewBox="0 0 16 16" className="size-3.5 shrink-0" fill="currentColor" aria-hidden>
      <path d="M8 1a5 5 0 0 0-5 5c0 3.6 5 9 5 9s5-5.4 5-9a5 5 0 0 0-5-5m0 7a2 2 0 1 1 0-4 2 2 0 0 1 0 4" />
    </svg>
  );
}

/**
 * Ajakan menyalakan lokasi, hanya saat tab "Terdekat" dipilih tanpa lokasi asli.
 *
 * Sengaja TIDAK muncul saat halaman dibuka. Peramban sudah punya dialog izinnya
 * sendiri; menaruh ajakan kita di depan dialog itu berarti pengguna harus
 * menyetujui dua kali, dan lapisan pertama justru menurunkan jumlah yang
 * sampai ke lapisan kedua.
 *
 * Di sinilah tempatnya, karena di sinilah ia baru benar-benar berarti: meminta
 * "Terdekat" sambil mengukur jarak dari pusat kota adalah menjawab pertanyaan
 * yang tidak diajukan siapa pun.
 */
export function AjakanLokasi({
  keadaan,
  sedangMinta,
  onAktifkan,
}: {
  keadaan: "belum" | "ditolak" | "diblokir" | "gagal";
  sedangMinta: boolean;
  onAktifkan: () => void;
}) {
  if (keadaan === "diblokir") {
    return (
      <Bingkai>
        <p className="font-semibold">Izin lokasi diblokir untuk situs ini</p>
        <p className="mt-1 text-text-muted">
          Urutannya masih dihitung dari pusat kota Bandung. Untuk membukanya:
          tekan ikon gembok atau tombol setelan di sebelah alamat situs di bilah
          atas peramban, pilih{" "}
          <span className="font-medium text-text">Izin</span> →{" "}
          <span className="font-medium text-text">Lokasi</span>, lalu ubah ke{" "}
          <span className="font-medium text-text">Izinkan</span>. Daftarnya
          langsung menyesuaikan, tidak perlu memuat ulang halaman.
        </p>
      </Bingkai>
    );
  }

  // Ditolak, tapi belum tentu permanen — dialog yang ditutup tanpa dipilih
  // berakhir di sini. Tombolnya masih berguna: menekannya akan memunculkan
  // dialog itu lagi. Kalimat terakhir disiapkan untuk kemungkinan sebaliknya,
  // termasuk di peramban yang tidak mau memberi tahu status izinnya (Safari).
  if (keadaan === "ditolak") {
    return (
      <Bingkai>
        <p className="font-semibold">Lokasi belum diizinkan</p>
        <p className="mt-1 text-text-muted">
          Urutannya masih dihitung dari pusat kota Bandung.
        </p>
        <TombolAktifkan sedangMinta={sedangMinta} onAktifkan={onAktifkan} label="Coba lagi" />
        <p className="mt-2 text-xs text-text-faint">
          Kalau menekan tombol ini tidak memunculkan dialog izin, berarti lokasi
          sudah diblokir untuk situs ini dan harus diubah dari setelan situs di
          peramban.
        </p>
      </Bingkai>
    );
  }

  if (keadaan === "gagal") {
    return (
      <Bingkai>
        <p className="font-semibold">Lokasi tidak bisa didapat</p>
        <p className="mt-1 text-text-muted">
          Sinyalnya mungkin sedang lemah, atau layanan lokasi di perangkat sedang
          mati. Urutannya sementara dihitung dari pusat kota Bandung.
        </p>
        <TombolAktifkan sedangMinta={sedangMinta} onAktifkan={onAktifkan} label="Coba lagi" />
      </Bingkai>
    );
  }

  return (
    <Bingkai>
      <p className="font-semibold">Urutan ini belum benar-benar dari tempatmu</p>
      <p className="mt-1 text-text-muted">
        Jaraknya masih dihitung dari pusat kota Bandung. Nyalakan lokasi supaya
        yang paling atas benar-benar yang paling dekat.
      </p>
      <TombolAktifkan sedangMinta={sedangMinta} onAktifkan={onAktifkan} label="Nyalakan lokasi" />
    </Bingkai>
  );
}

function Bingkai({ children }: { children: React.ReactNode }) {
  return (
    <div
      role="status"
      className="rounded-xl border border-brand-line bg-brand-soft px-4 py-3 text-sm"
    >
      {children}
    </div>
  );
}

function TombolAktifkan({
  sedangMinta,
  onAktifkan,
  label,
}: {
  sedangMinta: boolean;
  onAktifkan: () => void;
  label: string;
}) {
  return (
    <button
      type="button"
      onClick={onAktifkan}
      disabled={sedangMinta}
      className="mt-3 inline-flex min-h-10 items-center rounded-lg bg-brand px-4 text-sm font-semibold text-on-brand transition-colors hover:bg-brand-lift disabled:opacity-60"
    >
      {sedangMinta ? "Meminta izin…" : label}
    </button>
  );
}
