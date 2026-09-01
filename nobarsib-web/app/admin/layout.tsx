import type { Metadata } from "next";

import { AdminShell } from "@/components/admin/AdminShell";

export const metadata: Metadata = {
  title: "Panel Admin",
  // Panel admin tidak boleh masuk indeks mesin pencari.
  robots: { index: false, follow: false },
};

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  return <AdminShell>{children}</AdminShell>;
}
