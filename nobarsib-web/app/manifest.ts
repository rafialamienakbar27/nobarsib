import type { MetadataRoute } from "next";

/**
 * Manifest PWA.
 *
 * §6.2: PWA dulu, native belakangan. Dengan pola pemakaian 15–20 malam setahun,
 * orang tidak akan memasang aplikasi native — tapi "Add to Home Screen" tetap
 * berguna bagi yang datang lewat tautan Instagram dan ingin membukanya lagi
 * nanti. Checklist rilis §23 mensyaratkan ini bisa dilakukan.
 */
export default function manifest(): MetadataRoute.Manifest {
  return {
    name: "NOBARSIB — Cari tempat nobar Persib",
    short_name: "NOBARSIB",
    description:
      "Cari tempat nonton bareng Persib di Bandung, terurut jarak dan lengkap dengan info fasilitas.",
    lang: "id",
    start_url: "/",
    scope: "/",
    display: "standalone",
    // "any", bukan "portrait": tata letaknya kini responsif sampai lebar
    // tablet, jadi mengunci orientasi hanya membuang kemampuan yang sudah ada —
    // dan memaksa portrait di tablet justru membuat daftar venue kembali jadi
    // satu kolom sempit. APK Android memakai setelan yang sama ("default").
    orientation: "any",
    background_color: "#f4f5fb",
    theme_color: "#171b87",
    categories: ["sports", "lifestyle"],
    icons: [
      { src: "/icon-192.png", sizes: "192x192", type: "image/png", purpose: "any" },
      { src: "/icon-512.png", sizes: "512x512", type: "image/png", purpose: "any" },
      { src: "/icon-maskable-512.png", sizes: "512x512", type: "image/png", purpose: "maskable" },
    ],
    shortcuts: [
      { name: "Jadwal Persib", url: "/jadwal" },
      { name: "Daftarkan venue", url: "/untuk-venue" },
    ],
  };
}
