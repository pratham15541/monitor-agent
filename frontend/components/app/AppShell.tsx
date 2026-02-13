"use client";

import type { ReactNode } from "react";
import { Sidebar } from "@/components/app/Sidebar";
import { TopBar } from "@/components/app/TopBar";

export function AppShell({
  title,
  children,
}: {
  title: string;
  children: ReactNode;
}) {
  return (
    <div className="min-h-screen bg-transparent">
      <div className="grid min-h-screen grid-cols-1 xl:grid-cols-[16rem_1fr]">
        <Sidebar className="hidden xl:flex" />
        <div className="flex min-h-screen flex-col">
          <TopBar title={title} />
          <main className="flex-1 px-6 py-8">
            <div className="mx-auto flex w-full max-w-6xl flex-col gap-6">
              {children}
            </div>
          </main>
        </div>
      </div>
    </div>
  );
}
