/**
 * Membuat ikon PWA sebagai PNG tanpa dependensi apa pun.
 *
 * Dijalankan sekali dengan `node scripts/generate-icons.mjs`; hasilnya
 * dikomit ke public/ sehingga proses build tidak perlu menjalankannya lagi.
 *
 * Menggambar sendiri, bukan memakai pustaka gambar, karena yang dibutuhkan
 * hanya beberapa bentuk geometris — memasang toolchain grafis demi tiga file
 * statis tidak sepadan.
 *
 * Lambangnya adalah layar dengan segitiga putar: "menonton bersama" — bentuk
 * milik sendiri, bukan lambang klub. §22 melarang memakai logo klub; warna
 * birunya boleh, perisai dan harimaunya tidak.
 */
import { deflateSync } from "node:zlib";
import { writeFileSync, mkdirSync } from "node:fs";

const BIRU = [23, 27, 135]; // #171B87, sama dengan theme_color di manifest
const PUTIH = [255, 255, 255];

function buatIkon(size, { maskable = false } = {}) {
  // Ikon maskable harus menyisakan zona aman: Android memotongnya jadi
  // lingkaran atau bentuk lain sesuai peluncur, dan gambar yang mepet tepi
  // akan terpotong.
  const pad = maskable ? size * 0.22 : size * 0.14;
  const radius = maskable ? 0 : size * 0.22;

  const px = new Uint8Array(size * size * 4);

  const layarKiri = pad;
  const layarKanan = size - pad;
  const layarAtas = pad + (size - pad * 2) * 0.12;
  const layarBawah = size - pad - (size - pad * 2) * 0.22;
  const tebal = Math.max(2, size * 0.055);

  for (let y = 0; y < size; y++) {
    for (let x = 0; x < size; x++) {
      const i = (y * size + x) * 4;

      if (!maskable && diLuarSudutMembulat(x, y, size, radius)) {
        px[i + 3] = 0; // transparan di luar sudut
        continue;
      }

      let warna = BIRU;

      const diBingkaiLayar =
        x >= layarKiri && x <= layarKanan && y >= layarAtas && y <= layarBawah &&
        (x < layarKiri + tebal || x > layarKanan - tebal ||
         y < layarAtas + tebal || y > layarBawah - tebal);

      // Kaki penyangga layar.
      const kakiLebar = (layarKanan - layarKiri) * 0.34;
      const diKaki =
        y > layarBawah && y <= layarBawah + tebal * 1.6 &&
        Math.abs(x - size / 2) < kakiLebar / 2;

      if (diBingkaiLayar || diKaki || diSegitigaPutar(x, y, layarKiri, layarKanan, layarAtas, layarBawah)) {
        warna = PUTIH;
      }

      px[i] = warna[0];
      px[i + 1] = warna[1];
      px[i + 2] = warna[2];
      px[i + 3] = 255;
    }
  }
  return encodePNG(px, size, size);
}

function diLuarSudutMembulat(x, y, size, r) {
  const cx = x < r ? r : x > size - r ? size - r : x;
  const cy = y < r ? r : y > size - r ? size - r : y;
  if (cx === x && cy === y) return false;
  return (x - cx) ** 2 + (y - cy) ** 2 > r * r;
}

function diSegitigaPutar(x, y, kiri, kanan, atas, bawah) {
  const cx = (kiri + kanan) / 2;
  const cy = (atas + bawah) / 2;
  const tinggi = (bawah - atas) * 0.42;
  const lebar = tinggi * 0.9;

  const dy = y - cy;
  if (Math.abs(dy) > tinggi / 2) return false;
  const batasKanan = cx + lebar / 2 - (Math.abs(dy) / (tinggi / 2)) * lebar;
  return x > cx - lebar / 2 && x < batasKanan;
}

/** Encoder PNG minimal: satu IHDR, satu IDAT, satu IEND. */
function encodePNG(pixels, width, height) {
  const baris = Buffer.alloc((width * 4 + 1) * height);
  for (let y = 0; y < height; y++) {
    baris[y * (width * 4 + 1)] = 0; // filter: none
    Buffer.from(pixels.buffer, y * width * 4, width * 4).copy(
      baris, y * (width * 4 + 1) + 1,
    );
  }

  const ihdr = Buffer.alloc(13);
  ihdr.writeUInt32BE(width, 0);
  ihdr.writeUInt32BE(height, 4);
  ihdr[8] = 8;  // bit depth
  ihdr[9] = 6;  // RGBA
  ihdr[10] = 0; // deflate
  ihdr[11] = 0; // filter standar
  ihdr[12] = 0; // tanpa interlace

  return Buffer.concat([
    Buffer.from([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]),
    chunk("IHDR", ihdr),
    chunk("IDAT", deflateSync(baris, { level: 9 })),
    chunk("IEND", Buffer.alloc(0)),
  ]);
}

function chunk(tipe, data) {
  const panjang = Buffer.alloc(4);
  panjang.writeUInt32BE(data.length, 0);
  const isi = Buffer.concat([Buffer.from(tipe, "ascii"), data]);
  const crc = Buffer.alloc(4);
  crc.writeUInt32BE(crc32(isi), 0);
  return Buffer.concat([panjang, isi, crc]);
}

const TABEL_CRC = (() => {
  const t = new Uint32Array(256);
  for (let n = 0; n < 256; n++) {
    let c = n;
    for (let k = 0; k < 8; k++) c = c & 1 ? 0xedb88320 ^ (c >>> 1) : c >>> 1;
    t[n] = c >>> 0;
  }
  return t;
})();

function crc32(buf) {
  let c = 0xffffffff;
  for (const b of buf) c = TABEL_CRC[(c ^ b) & 0xff] ^ (c >>> 8);
  return (c ^ 0xffffffff) >>> 0;
}

mkdirSync("public", { recursive: true });
writeFileSync("public/icon-192.png", buatIkon(192));
writeFileSync("public/icon-512.png", buatIkon(512));
writeFileSync("public/icon-maskable-512.png", buatIkon(512, { maskable: true }));
writeFileSync("public/apple-icon.png", buatIkon(180, { maskable: true }));
console.log("ikon dibuat: 192, 512, maskable-512, apple-icon");
