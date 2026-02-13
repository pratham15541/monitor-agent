"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import { Home, Cpu, Users, BadgeCheck } from "lucide-react";

const navItems = [
  { href: "/dashboard", label: "Dashboard", icon: Home },
  { href: "/agents", label: "Agent Register", icon: Cpu },
  { href: "/company", label: "Company", icon: Users },
];

export function Sidebar({ className }: { className?: string }) {
  const pathname = usePathname();

  return (
    <aside
      className={cn(
        "h-full w-64 flex-col gap-6 border-r bg-white/70 px-5 py-6 backdrop-blur",
        className,
      )}
    >
      <div>
        <div className="flex items-center gap-3">
          <div className="flex size-10 items-center justify-center rounded-xl bg-primary text-primary-foreground">
            <BadgeCheck className="size-5" />
          </div>
          <div>
            <p className="text-sm uppercase tracking-[0.2em] text-muted-foreground">
              Monitor
            </p>
            <p className="text-lg font-semibold">Control Room</p>
          </div>
        </div>
      </div>
      <nav className="space-y-1">
        {navItems.map((item) => {
          const active = pathname === item.href;
          const Icon = item.icon;
          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition",
                active
                  ? "bg-primary/10 text-primary"
                  : "text-muted-foreground hover:bg-muted/50 hover:text-foreground",
              )}
            >
              <Icon className="size-4" />
              {item.label}
            </Link>
          );
        })}
      </nav>
      <div className="mt-auto rounded-xl border bg-card/70 p-4 text-xs text-muted-foreground">
        Keep the agent connected to stream live metrics. Use the API token in
        the Agent Register screen.
      </div>
    </aside>
  );
}
