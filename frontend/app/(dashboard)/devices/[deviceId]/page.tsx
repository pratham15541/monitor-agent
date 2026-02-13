"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import { useParams } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { MetricChart } from "@/components/app/MetricChart";
import { StatTile } from "@/components/app/StatTile";
import { fetchJson } from "@/lib/api";
import { formatDateTime, formatNumber } from "@/lib/format";
import type { Device, DeviceStatus, Metric } from "@/lib/types";
import { RefreshCcw } from "lucide-react";
import { Client } from "@stomp/stompjs";
import SockJS from "sockjs-client";

const DEFAULT_WS_URL = "http://localhost:8080/ws";
const MAX_METRICS = 60;

function normalizeSockJsUrl(url: string) {
  if (url.startsWith("ws://")) {
    return `http://${url.slice(5)}`;
  }
  if (url.startsWith("wss://")) {
    return `https://${url.slice(6)}`;
  }
  return url;
}

type RealtimeState = "connecting" | "connected" | "disconnected";

export default function DeviceDetailPage() {
  const params = useParams<{ deviceId?: string }>();
  const deviceId = params?.deviceId;
  const [device, setDevice] = useState<Device | null>(null);
  const [metrics, setMetrics] = useState<Metric[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [realtimeState, setRealtimeState] =
    useState<RealtimeState>("disconnected");

  async function loadData() {
    setError(null);
    if (!deviceId || deviceId === "undefined") {
      setError("Missing device id in route.");
      setLoading(false);
      return;
    }
    try {
      const [deviceList, deviceMetrics] = await Promise.all([
        fetchJson<Device[]>("/devices"),
        fetchJson<Metric[]>(`/devices/${deviceId}/metrics`),
      ]);
      setDevice(deviceList.find((item) => item.id === deviceId) ?? null);
      setMetrics(deviceMetrics);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load device");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void loadData();
  }, [deviceId]);

  useEffect(() => {
    if (!deviceId || deviceId === "undefined") {
      return;
    }

    const wsUrl = normalizeSockJsUrl(
      process.env.NEXT_PUBLIC_WS_URL ?? DEFAULT_WS_URL,
    );

    const client = new Client({
      reconnectDelay: 5000,
      webSocketFactory: () => new SockJS(wsUrl),
      onConnect: () => {
        setRealtimeState("connected");
        client.subscribe(`/topic/device/${deviceId}`, (message) => {
          if (!message.body) {
            return;
          }
          try {
            const parsed = JSON.parse(message.body) as Metric;
            setMetrics((prev) => [parsed, ...prev].slice(0, MAX_METRICS));
            if (parsed.createdAt) {
              setDevice((current) =>
                current
                  ? {
                      ...current,
                      lastSeenAt: parsed.createdAt,
                    }
                  : current,
              );
            }
          } catch {
            // Ignore malformed payloads to keep the stream alive.
          }
        });

        client.subscribe(`/topic/device-status/${deviceId}`, (message) => {
          const nextStatus = message.body?.trim() as DeviceStatus | undefined;
          if (!nextStatus) {
            return;
          }
          setDevice((current) =>
            current
              ? {
                  ...current,
                  status: nextStatus,
                }
              : current,
          );
        });
      },
      onStompError: () => {
        setRealtimeState("disconnected");
      },
      onWebSocketClose: () => {
        setRealtimeState("disconnected");
      },
      onWebSocketError: () => {
        setRealtimeState("disconnected");
      },
    });

    setRealtimeState("connecting");
    client.activate();

    return () => {
      setRealtimeState("disconnected");
      client.deactivate();
    };
  }, [deviceId]);

  const chartMetrics = useMemo(() => metrics.slice().reverse(), [metrics]);
  const latest = metrics[0];

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <Link href="/dashboard" className="text-sm text-muted-foreground">
            ← Back to dashboard
          </Link>
          <h2 className="mt-2 text-2xl font-semibold text-foreground">
            {device?.hostname ?? "Device detail"}
          </h2>
          <p className="text-sm text-muted-foreground">
            {device?.ipAddress ?? "Unknown IP"} • {device?.os ?? "Unknown OS"}
          </p>
        </div>
        <Button variant="outline" onClick={loadData}>
          <RefreshCcw className="size-4" />
          Refresh metrics
        </Button>
      </div>

      <div className="text-xs text-muted-foreground">
        Realtime stream: {realtimeState}
      </div>

      {error ? (
        <div className="rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">
          {error}
        </div>
      ) : null}

      {loading ? (
        <div className="rounded-xl border bg-white/60 p-10 text-center text-sm text-muted-foreground">
          Loading device metrics...
        </div>
      ) : null}

      {latest ? (
        <section className="grid gap-4 md:grid-cols-3">
          <StatTile
            label="CPU"
            value={`${formatNumber(latest.cpuUsage)}%`}
            helper={"Latest"}
          />
          <StatTile
            label="Memory"
            value={`${formatNumber(latest.memoryUsage)}%`}
            helper={"Latest"}
          />
          <StatTile
            label="Disk"
            value={`${formatNumber(latest.diskUsage)}%`}
            helper={"Latest"}
          />
        </section>
      ) : null}

      <section className="grid gap-4 lg:grid-cols-2">
        <MetricChart
          title="CPU usage"
          unit="%"
          values={chartMetrics.map((item) => item.cpuUsage)}
          strokeClassName="stroke-emerald-500"
        />
        <MetricChart
          title="Memory usage"
          unit="%"
          values={chartMetrics.map((item) => item.memoryUsage)}
          strokeClassName="stroke-sky-500"
        />
        <MetricChart
          title="Disk usage"
          unit="%"
          values={chartMetrics.map((item) => item.diskUsage)}
          strokeClassName="stroke-amber-500"
        />
        <MetricChart
          title="Network in"
          unit="Mb"
          values={chartMetrics.map((item) => item.networkIn)}
          strokeClassName="stroke-purple-500"
        />
        <MetricChart
          title="Network out"
          unit="Mb"
          values={chartMetrics.map((item) => item.networkOut)}
          strokeClassName="stroke-rose-500"
        />
      </section>

      <Card>
        <CardHeader>
          <CardTitle>Recent samples</CardTitle>
        </CardHeader>
        <CardContent>
          {!metrics.length ? (
            <p className="text-sm text-muted-foreground">No metrics yet.</p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Time</TableHead>
                  <TableHead>CPU</TableHead>
                  <TableHead>Memory</TableHead>
                  <TableHead>Disk</TableHead>
                  <TableHead>Net In</TableHead>
                  <TableHead>Net Out</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {metrics.slice(0, 10).map((metric) => (
                  <TableRow key={metric.id}>
                    <TableCell>{formatDateTime(metric.createdAt)}</TableCell>
                    <TableCell>{formatNumber(metric.cpuUsage)}%</TableCell>
                    <TableCell>{formatNumber(metric.memoryUsage)}%</TableCell>
                    <TableCell>{formatNumber(metric.diskUsage)}%</TableCell>
                    <TableCell>{formatNumber(metric.networkIn)}</TableCell>
                    <TableCell>{formatNumber(metric.networkOut)}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
