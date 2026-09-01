import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Sengaja TIDAK memakai output: "standalone".
  //
  // Server minimal yang dihasilkannya tidak menerapkan proxy.ts maupun
  // headers() — keduanya diam-diam tidak jalan, tanpa error. Penjelasan
  // lengkap beserta hasil verifikasinya ada di Dockerfile dan proxy.ts.

  // Header keamanan TIDAK dipasang di sini, melainkan di proxy.ts.
  // Server standalone yang dipakai Dockerfile tidak menerapkan headers() dari
  // konfigurasi ini — penjelasan lengkapnya ada di proxy.ts.
  // Nginx di depan (§6.3) tetap yang menambahkan TLS dan HSTS.
};

export default nextConfig;
