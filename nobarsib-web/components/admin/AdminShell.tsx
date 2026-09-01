"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, useMemo, useSyncExternalStore } from "react";

import {
  logout,
  snapshotSesi,
  snapshotSesiServer,
  subscribeSesi,
  type AdminUser,
} from "@/lib/admin";

const MENU = [
  { href: "/admin", label: "Ringkasan" },
  { href: "/admin/antrian", label: "Antrian" },
  { href: "/admin/venue", label: "Venue" },
  { href: "/admin/jadwal", label: "Jadwal" },
];

/**
 * Kerangka panel admin: memeriksa sesi, lalu menampilkan navigasi.
 *
 * Penjaga di sini hanya untuk kenyamanan — yang benar-benar melindungi data
 * adalah middleware di API, yang menolak setiap permintaan tanpa token sah.
 * Menyembunyikan tombol bukan keamanan.
 */
export function AdminShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();

  const raw = useSyncExternalStore(subscribeSesi, snapshotSesi, snapshotSesiServer);
  const user = useMemo<AdminUser | null>(() => {
    if (!raw) return null;
    try {
      return JSON.parse(raw) as AdminUser;
    } catch {
      return null;
    }
  }, [raw]);

  const halamanLogin = pathname === "/admin/masuk";

  useEffect(() => {
    if (!user && !halamanLogin) router.replace("/admin/masuk");
  }, [user, halamanLogin, router]);

  if (halamanLogin) {
    return <div className="mx-auto max-w-sm py-12">{children}</div>;
  }

  if (!user) {
    return (
      <p className="py-12 text-center text-sm text-text-faint">
        Mengalihkan ke halaman masuk…
      </p>
    );
  }

  // Panel admin dibatasi lebih sempit daripada halaman publik: isinya formulir
  // dan daftar tinjauan, dan baris formulir selebar layar penuh justru lebih
  // sulit dipakai.
  return (
    <div className="mx-auto max-w-4xl space-y-5">
      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-border pb-3">
        <div>
          <h1 className="text-lg font-bold">Panel Admin</h1>
          <p className="text-xs text-text-faint">{user.email}</p>
        </div>
        <button
          onClick={async () => {
            await logout();
            router.replace("/admin/masuk");
          }}
          className="min-h-9 rounded-lg border border-border px-3 text-sm text-text-muted hover:text-text"
        >
          Keluar
        </button>
      </div>

      <nav className="flex gap-1 overflow-x-auto">
        {MENU.map((m) => {
          const aktif = pathname === m.href;
          return (
            <Link
              key={m.href}
              href={m.href}
              className={
                "min-h-10 shrink-0 rounded-lg px-3 py-2 text-sm font-medium " +
                (aktif ? "bg-brand-soft text-brand-accent" : "text-text-muted hover:bg-surface-alt")
              }
            >
              {m.label}
            </Link>
          );
        })}
      </nav>

      {children}
    </div>
  );
}
