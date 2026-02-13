"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { fetchJson } from "@/lib/api";
import type { Device } from "@/lib/types";

export default function AgentRegisterPage() {
  const [token, setToken] = useState("");
  const [hostname, setHostname] = useState("");
  const [ipAddress, setIpAddress] = useState("");
  const [os, setOs] = useState("");
  const [result, setResult] = useState<Device | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setLoading(true);

    try {
      const device = await fetchJson<Device>("/agent/register", {
        method: "POST",
        body: { token, hostname, ipAddress, os },
      });
      setResult(device);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Registration failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="grid gap-6 lg:grid-cols-[1.2fr_0.8fr]">
      <Card>
        <CardHeader>
          <CardTitle>Register a device agent</CardTitle>
          <p className="text-sm text-muted-foreground">
            Use the company API token to register a machine.
          </p>
        </CardHeader>
        <CardContent>
          <form className="space-y-4" onSubmit={handleSubmit}>
            <div className="space-y-2">
              <label className="text-sm font-medium" htmlFor="token">
                API token
              </label>
              <Input
                id="token"
                value={token}
                onChange={(event) => setToken(event.target.value)}
                placeholder="Paste company API token"
                required
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium" htmlFor="hostname">
                Hostname
              </label>
              <Input
                id="hostname"
                value={hostname}
                onChange={(event) => setHostname(event.target.value)}
                placeholder="agent-01"
                required
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium" htmlFor="ip">
                IP address
              </label>
              <Input
                id="ip"
                value={ipAddress}
                onChange={(event) => setIpAddress(event.target.value)}
                placeholder="10.0.10.14"
                required
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium" htmlFor="os">
                Operating system
              </label>
              <Input
                id="os"
                value={os}
                onChange={(event) => setOs(event.target.value)}
                placeholder="Ubuntu 22.04"
                required
              />
            </div>
            {error ? (
              <p className="rounded-md border border-rose-200 bg-rose-50 px-3 py-2 text-sm text-rose-700">
                {error}
              </p>
            ) : null}
            <Button className="w-full" disabled={loading} type="submit">
              {loading ? "Registering..." : "Register device"}
            </Button>
          </form>
        </CardContent>
      </Card>

      <Card className="border-dashed">
        <CardHeader>
          <CardTitle>Latest registration</CardTitle>
        </CardHeader>
        <CardContent>
          {result ? (
            <div className="space-y-2 text-sm">
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Hostname</span>
                <span>{result.hostname}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">IP</span>
                <span>{result.ipAddress}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">OS</span>
                <span>{result.os}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Status</span>
                <span>{result.status}</span>
              </div>
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">
              No device registered yet.
            </p>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
