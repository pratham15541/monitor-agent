"use client";

import { usePathname, useRouter } from "next/navigation";
import { useEffect, useMemo, useState } from "react";
import { AppShell } from "@/components/app/AppShell";
import { getToken } from "@/lib/auth";

const titleMap: Record<string, string> = {
  "/dashboard": "Device Overview",
  "/agents": "Agent Register",
  "/company": "Company Profile",
};

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const router = useRouter();
  const pathname = usePathname();
  const [ready, setReady] = useState(false);

  useEffect(() => {
    if (!getToken()) {
      router.replace("/login");
      return;
    }
    setReady(true);
  }, [router]);

  const title = useMemo(() => {
    if (pathname?.startsWith("/devices/")) return "Device Detail";
    return titleMap[pathname ?? ""] ?? "Control Room";
  }, [pathname]);

  if (!ready) return null;

  return <AppShell title={title}>{children}</AppShell>;
}
