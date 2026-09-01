/**
 * Pusat kota Bandung.
 *
 * Dipakai sebagai titik acuan saat lokasi pengguna belum atau tidak tersedia
 * (§4.2 — "tolak → fallback: pusat kota Bandung"), termasuk saat halaman
 * dirender di server sebelum browser sempat menanyakan izin.
 *
 * Dipisahkan dari lib/location.ts karena file itu hanya boleh jalan di browser
 * ("use client"), sementara nilai ini juga dibutuhkan komponen server.
 */
export const BANDUNG = { lat: -6.9175, lng: 107.6191 } as const;
