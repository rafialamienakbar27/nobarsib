"use client";

import { useEffect } from "react";

/**
 * Mendaftarkan service worker.
 *
 * Hanya di produksi: di mode pengembangan, service worker menyimpan aset lama
 * dan membuat perubahan kode seolah-olah tidak tersimpan.
 */
export function RegisterSW() {
  useEffect(() => {
    if (process.env.NODE_ENV !== "production") return;
    if (!("serviceWorker" in navigator)) return;

    // Didaftarkan setelah halaman selesai dimuat supaya tidak ikut berebut
    // bandwidth dengan tampilan pertama — LCP < 2,5 detik di 4G (§13.6).
    const daftar = () => {
      void navigator.serviceWorker.register("/sw.js").catch(() => {
        // Pendaftaran gagal (mode penyamaran, konteks tidak aman). Aplikasi
        // tetap berjalan normal tanpa dukungan luring.
      });
    };

    if (document.readyState === "complete") {
      daftar();
      return;
    }
    window.addEventListener("load", daftar);
    return () => window.removeEventListener("load", daftar);
  }, []);

  return null;
}
