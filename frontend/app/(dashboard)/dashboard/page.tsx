"use client";

import { useEffect, useMemo, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { DeviceCard } from "@/components/app/DeviceCard";
import { EmptyState } from "@/components/app/EmptyState";
import { StatTile } from "@/components/app/StatTile";
import { fetchJson } from "@/lib/api";
import type { Device } from "@/lib/types";
import { RefreshCcw } from "lucide-react";

export default function DashboardPage() {
  const [devices, setDevices] = useState<Device[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [query, setQuery] = useState("");
  const [autoRefresh, setAutoRefresh] = useState(true);

  async function loadDevices() {
    setError(null);
    try {
      const data = await fetchJson<Device[]>("/devices");
      setDevices(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load devices");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void loadDevices();
  }, []);

  useEffect(() => {
    if (!autoRefresh) return;
    const interval = setInterval(() => {
      void loadDevices();
    }, 15000);
    return () => clearInterval(interval);
  }, [autoRefresh]);

  const filtered = useMemo(() => {
    const needle = query.trim().toLowerCase();
    if (!needle) return devices;
    return devices.filter((device) => {
      return (
        device.hostname?.toLowerCase().includes(needle) ||
        device.ipAddress?.toLowerCase().includes(needle) ||
        device.os?.toLowerCase().includes(needle)
      );
    });
  }, [devices, query]);

  const online = devices.filter((device) => device.status === "ONLINE").length;
  const offline = devices.filter(
    (device) => device.status === "OFFLINE",
  ).length;

  return (
    <div className="space-y-6">
      <section className="grid gap-4 md:grid-cols-3">
        <StatTile label="Total" value={String(devices.length)} />
        <StatTile label="Online" value={String(online)} helper="Active" />
        <StatTile
          label="Offline"
          value={String(offline)}
          helper="Needs check"
        />
      </section>

      <section className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex w-full max-w-md items-center gap-2">
          <Input
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Search by hostname, IP, OS"
          />
        </div>
        <div className="flex items-center gap-3">
          <Button
            variant={autoRefresh ? "default" : "outline"}
            onClick={() => setAutoRefresh((prev) => !prev)}
          >
            {autoRefresh ? "Auto refresh: On" : "Auto refresh: Off"}
          </Button>
          <Button variant="outline" onClick={loadDevices}>
            <RefreshCcw className="size-4" />
            Refresh
          </Button>
        </div>
      </section>

      {error ? (
        <div className="rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">
          {error}
        </div>
      ) : null}

      {loading ? (
        <div className="rounded-xl border bg-white/60 p-10 text-center text-sm text-muted-foreground">
          Loading devices...
        </div>
      ) : null}

      {!loading && !filtered.length ? (
        <EmptyState
          title="No devices yet"
          description="Register an agent to start streaming metrics."
        />
      ) : null}

      <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        {filtered.map((device) => (
          <DeviceCard key={device.id} device={device} />
        ))}
      </section>
    </div>
  );
}
