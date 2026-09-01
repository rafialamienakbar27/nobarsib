"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";

import { login } from "@/lib/admin";

export default function HalamanMasuk() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [sandi, setSandi] = useState("");
  const [galat, setGalat] = useState<string | null>(null);
  const [mengirim, setMengirim] = useState(false);

  async function kirim(e: React.FormEvent) {
    e.preventDefault();
    setMengirim(true);
    setGalat(null);
    try {
      await login(email, sandi);
      router.replace("/admin");
    } catch (err) {
      setGalat(err instanceof Error ? err.message : "Gagal masuk");
      setMengirim(false);
    }
  }

  return (
    <form onSubmit={kirim} className="space-y-4">
      <div>
        <h1 className="text-xl font-bold">Masuk panel admin</h1>
        <p className="mt-1 text-sm text-text-muted">
          Akun dibuat lewat <code className="text-xs">make admin-create</code>.
        </p>
      </div>

      <label className="block">
        <span className="text-sm font-medium">Email</span>
        <input
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          required
          autoComplete="username"
          className="mt-1 min-h-11 w-full rounded-xl border border-border bg-surface px-3"
        />
      </label>

      <label className="block">
        <span className="text-sm font-medium">Kata sandi</span>
        <input
          type="password"
          value={sandi}
          onChange={(e) => setSandi(e.target.value)}
          required
          autoComplete="current-password"
          className="mt-1 min-h-11 w-full rounded-xl border border-border bg-surface px-3"
        />
      </label>

      {galat && (
        <p className="rounded-xl bg-warn-soft px-3 py-2 text-sm text-warn" role="alert">
          {galat}
        </p>
      )}

      <button
        type="submit"
        disabled={mengirim}
        className="min-h-12 w-full rounded-xl bg-brand font-semibold text-on-brand shadow-brand transition-colors hover:bg-brand-lift disabled:opacity-60"
      >
        {mengirim ? "Memeriksa…" : "Masuk"}
      </button>
    </form>
  );
}
