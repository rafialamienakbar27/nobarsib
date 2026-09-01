"use client";

const KEY = "nobarsib:device";

/**
 * Identitas perangkat untuk statistik dan (nanti di Fase 5) anti-spam review.
 *
 * Sengaja acak dan disimpan lokal, bukan sidik jari browser: UU PDP berlaku
 * (§22) dan blueprint meminta menyimpan sedata mungkin — cukup untuk membedakan
 * perangkat, tidak cukup untuk mengenali orang.
 *
 * Backend tetap melakukan hashing sendiri dengan salt server (§11.5), jadi
 * nilai ini tidak pernah tersimpan apa adanya di database.
 */
export function deviceHash(): string {
  try {
    const ada = localStorage.getItem(KEY);
    if (ada) return ada;
    const baru = crypto.randomUUID();
    localStorage.setItem(KEY, baru);
    return baru;
  } catch {
    // Penyimpanan diblokir: kirim nilai sekali pakai supaya pencatatan tetap
    // jalan, meski perangkatnya tidak bisa dikenali lagi nanti.
    return crypto.randomUUID();
  }
}
