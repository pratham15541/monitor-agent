"use client";

import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet";
import { Menu } from "lucide-react";
import { Sidebar } from "@/components/app/Sidebar";
import { clearCompanyProfile, clearToken, getCompanyProfile } from "@/lib/auth";

export function TopBar({ title }: { title: string }) {
  const router = useRouter();
  const profile = getCompanyProfile();

  function handleLogout() {
    clearToken();
    clearCompanyProfile();
    router.replace("/login");
  }

  return (
    <header className="flex items-center justify-between gap-4 border-b bg-white/70 px-6 py-4 backdrop-blur">
      <div className="flex items-center gap-3">
        <Sheet>
          <SheetTrigger asChild>
            <Button size="icon" variant="ghost" className="xl:hidden">
              <Menu className="size-5" />
            </Button>
          </SheetTrigger>
          <SheetContent side="left" className="p-0">
            <Sidebar className="flex" />
          </SheetContent>
        </Sheet>
        <div>
          <p className="text-xs uppercase tracking-[0.3em] text-muted-foreground">
            Control Room
          </p>
          <h1 className="text-xl font-semibold text-foreground">{title}</h1>
        </div>
      </div>
      <div className="flex items-center gap-4">
        <div className="hidden text-right sm:block">
          <p className="text-sm font-medium text-foreground">
            {profile?.name ?? "Company"}
          </p>
          <p className="text-xs text-muted-foreground">
            {profile?.email ?? "Signed in"}
          </p>
        </div>
        <Button variant="outline" onClick={handleLogout}>
          Log out
        </Button>
      </div>
    </header>
  );
}
