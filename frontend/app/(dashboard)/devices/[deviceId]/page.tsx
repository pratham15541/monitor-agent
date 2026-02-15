"use client";

import Link from "next/link";
import { useEffect, useMemo, useRef, useState } from "react";
import { useParams } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
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
import { ProcessListView } from "@/components/app/ProcessListView";
import { ConnectionsView } from "@/components/app/ConnectionsView";
import { ServicesView } from "@/components/app/ServicesView";
import { LogsView } from "@/components/app/LogsView";
import { fetchJson } from "@/lib/api";
import { getToken } from "@/lib/auth";
import { formatDateTime, formatNumber } from "@/lib/format";
import type {
  Device,
  DeviceStatus,
  Metric,
  MetricDetail,
  DetailedMetricsPayload,
} from "@/lib/types";
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

type CommandResult = {
  deviceId: string;
  commandId: string;
  type: string;
  status: string;
  output?: string;
  error?: string;
  startedAt?: string;
  finishedAt?: string;
};

function formatOutput(text?: string) {
  if (!text) return "";
  const limit = 2000;
  if (text.length <= limit) return text;
  return `${text.slice(0, limit)}...`;
}

export default function DeviceDetailPage() {
  const params = useParams<{ deviceId?: string }>();
  const deviceId = params?.deviceId;
  const [device, setDevice] = useState<Device | null>(null);
  const [metrics, setMetrics] = useState<Metric[]>([]);
  const [detailedMetrics, setDetailedMetrics] = useState<MetricDetail[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [realtimeState, setRealtimeState] =
    useState<RealtimeState>("disconnected");
  const [commandInput, setCommandInput] = useState("");
  const [commandHistory, setCommandHistory] = useState<CommandResult[]>([]);
  const [activeTab, setActiveTab] = useState<"overview" | "detailed">(
    "overview",
  );
  const stompRef = useRef<Client | null>(null);

  async function loadData() {
    setError(null);
    if (!deviceId || deviceId === "undefined") {
      setError("Missing device id in route.");
      setLoading(false);
      return;
    }
    try {
      const [deviceList, deviceMetrics, deviceDetailedMetrics] =
        await Promise.all([
          fetchJson<Device[]>("/devices"),
          fetchJson<Metric[]>(`/devices/${deviceId}/metrics`),
          fetchJson<MetricDetail[]>(`/devices/${deviceId}/metrics-detail`),
        ]);
      setDevice(deviceList.find((item) => item.id === deviceId) ?? null);
      setMetrics(deviceMetrics);
      setDetailedMetrics(deviceDetailedMetrics);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load device");
    } finally {
      setLoading(false);
    }
  }

  async function loadDetailedMetrics() {
    if (!deviceId || deviceId === "undefined") {
      return;
    }

    try {
      const deviceDetailedMetrics = await fetchJson<MetricDetail[]>(
        `/devices/${deviceId}/metrics-detail`,
      );
      setDetailedMetrics(deviceDetailedMetrics);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load details");
    }
  }

  useEffect(() => {
    void loadData();
  }, [deviceId]);

  useEffect(() => {
    if (activeTab !== "detailed") {
      return;
    }

    void loadDetailedMetrics();
    const interval = setInterval(() => {
      void loadDetailedMetrics();
    }, 15000);

    return () => clearInterval(interval);
  }, [activeTab, deviceId]);

  useEffect(() => {
    if (!deviceId || deviceId === "undefined") {
      return;
    }

    const wsUrl = normalizeSockJsUrl(
      process.env.NEXT_PUBLIC_WS_URL ?? DEFAULT_WS_URL,
    );
    const token = getToken();

    const client = new Client({
      reconnectDelay: 5000,
      connectHeaders: token ? { Authorization: `Bearer ${token}` } : {},
      webSocketFactory: () => new SockJS(wsUrl),
      onConnect: () => {
        setRealtimeState("connected");
        stompRef.current = client;
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

        client.subscribe(`/topic/device-detail/${deviceId}`, (message) => {
          if (!message.body) {
            return;
          }
          try {
            const parsed = JSON.parse(message.body) as MetricDetail;
            setDetailedMetrics((prev) => [parsed, ...prev].slice(0, 20));
          } catch {
            // Ignore malformed payloads.
          }
        });

        client.subscribe(`/topic/command-result/${deviceId}`, (message) => {
          if (!message.body) return;
          try {
            const parsed = JSON.parse(message.body) as CommandResult;
            setCommandHistory((prev) => [parsed, ...prev].slice(0, 20));
          } catch {
            // Ignore malformed payloads.
          }
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
      stompRef.current = null;
      client.deactivate();
    };
  }, [deviceId]);

  function sendCommand(
    type: "shell" | "service" | "diagnostics" | "collect-details",
    payload = "",
  ) {
    if (!deviceId || deviceId === "undefined") {
      setError("Missing device id in route.");
      return;
    }

    const client = stompRef.current;
    if (!client || !client.connected) {
      setError("Websocket not connected.");
      return;
    }

    const commandId = crypto.randomUUID();
    client.publish({
      destination: `/app/command/${deviceId}`,
      body: JSON.stringify({
        deviceId,
        commandId,
        type,
        payload,
      }),
    });

    if (type === "shell") {
      setCommandInput("");
    }
  }

  async function handleRefresh() {
    if (activeTab === "detailed") {
      sendCommand("collect-details");
      await new Promise((resolve) => setTimeout(resolve, 1500));
    }

    void loadData();
  }

  const chartMetrics = useMemo(() => metrics.slice().reverse(), [metrics]);
  const latest = metrics[0];

  const latestDetailed = useMemo(() => {
    if (!detailedMetrics.length) return null;
    try {
      const parsed = JSON.parse(
        detailedMetrics[0].detailsJson,
      ) as DetailedMetricsPayload;
      return parsed;
    } catch {
      return null;
    }
  }, [detailedMetrics]);

  const latestDetailedCreatedAt = detailedMetrics[0]?.createdAt ?? null;

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
        <Button variant="outline" onClick={handleRefresh}>
          <RefreshCcw className="size-4" />
          Refresh metrics
        </Button>
      </div>

      <div className="flex items-center justify-between">
        <div className="text-xs text-muted-foreground">
          Realtime stream: {realtimeState}
        </div>
        <div className="flex gap-2">
          <Button
            size="sm"
            variant={activeTab === "overview" ? "default" : "outline"}
            onClick={() => setActiveTab("overview")}
          >
            Overview
          </Button>
          <Button
            size="sm"
            variant={activeTab === "detailed" ? "default" : "outline"}
            onClick={() => setActiveTab("detailed")}
          >
            Detailed Metrics
          </Button>
        </div>
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

      {activeTab === "overview" && (
        <>
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
              <CardTitle>Remote commands</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex flex-wrap items-center gap-2">
                <Input
                  value={commandInput}
                  onChange={(event) => setCommandInput(event.target.value)}
                  placeholder="Run shell command"
                />
                <Button
                  onClick={() => sendCommand("shell", commandInput)}
                  disabled={!commandInput.trim()}
                >
                  Run
                </Button>
              </div>

              <div className="flex flex-wrap gap-2">
                <Button
                  variant="outline"
                  onClick={() => sendCommand("diagnostics", "collect")}
                >
                  Collect diagnostics
                </Button>
                <Button
                  variant="outline"
                  onClick={() => sendCommand("service", "restart")}
                >
                  Restart service
                </Button>
                <Button
                  variant="outline"
                  onClick={() => sendCommand("service", "stop")}
                >
                  Stop service
                </Button>
                <Button
                  variant="outline"
                  onClick={() => sendCommand("service", "start")}
                >
                  Start service
                </Button>
              </div>
              <p className="text-xs text-muted-foreground">
                Background service note: remote commands require the agent to be
                online. Stop/restart will disconnect the agent, and start must
                be run locally if the service is offline.
              </p>

              <div className="rounded-xl border bg-white/70 p-3 text-xs">
                {!commandHistory.length ? (
                  <p className="text-muted-foreground">
                    No command results yet.
                  </p>
                ) : (
                  <div className="space-y-3">
                    {commandHistory.map((item) => (
                      <div key={item.commandId} className="space-y-1">
                        <div className="flex flex-wrap items-center gap-2">
                          <span className="font-medium">{item.type}</span>
                          <span className="text-muted-foreground">
                            {item.finishedAt
                              ? formatDateTime(item.finishedAt)
                              : ""}
                          </span>
                          <span
                            className={
                              item.status === "ok"
                                ? "text-emerald-600"
                                : item.status === "timeout"
                                  ? "text-amber-600"
                                  : "text-rose-600"
                            }
                          >
                            {item.status}
                          </span>
                        </div>
                        {item.error ? (
                          <div className="text-rose-600">{item.error}</div>
                        ) : null}
                        {item.output ? (
                          <pre className="whitespace-pre-wrap text-muted-foreground">
                            {formatOutput(item.output)}
                          </pre>
                        ) : null}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </CardContent>
          </Card>

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
                        <TableCell>
                          {formatDateTime(metric.createdAt)}
                        </TableCell>
                        <TableCell>{formatNumber(metric.cpuUsage)}%</TableCell>
                        <TableCell>
                          {formatNumber(metric.memoryUsage)}%
                        </TableCell>
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
        </>
      )}

      {activeTab === "detailed" && (
        <>
          {!latestDetailed ? (
            <div className="rounded-xl border bg-white/60 p-10 text-center text-sm text-muted-foreground">
              No detailed metrics available yet. Wait for the agent to send
              data.
            </div>
          ) : (
            <div className="space-y-6">
              <div className="text-xs text-muted-foreground">
                Collected at: {latestDetailed.collectedAt} • OS:{" "}
                {latestDetailed.os}
                {latestDetailedCreatedAt
                  ? ` • Received: ${formatDateTime(latestDetailedCreatedAt)}`
                  : ""}
              </div>

              <ProcessListView processes={latestDetailed.processes} />
              <ConnectionsView connections={latestDetailed.connections} />
              <ServicesView services={latestDetailed.services} />
              <LogsView logs={latestDetailed.logs} />
            </div>
          )}
        </>
      )}
    </div>
  );
}
