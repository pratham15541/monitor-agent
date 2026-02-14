"use client";

import { useMemo, useState } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { ConnectionInfo } from "@/lib/types";

type Props = {
  connections: ConnectionInfo[];
};

function formatAddress(ip: string, port: number) {
  if (!ip || ip === "0.0.0.0" || ip === "::") {
    return `*:${port}`;
  }
  return `${ip}:${port}`;
}

export function ConnectionsView({ connections }: Props) {
  const [search, setSearch] = useState("");

  const filtered = useMemo(() => {
    const needle = search.trim().toLowerCase();
    if (!needle) return connections;

    return connections.filter((conn) => {
      const localAddr = formatAddress(
        conn.local.ip,
        conn.local.port,
      ).toLowerCase();
      const remoteAddr = formatAddress(
        conn.remote.ip,
        conn.remote.port,
      ).toLowerCase();
      return (
        localAddr.includes(needle) ||
        remoteAddr.includes(needle) ||
        conn.status.toLowerCase().includes(needle) ||
        conn.pid.toString().includes(needle)
      );
    });
  }, [connections, search]);

  const grouped = useMemo(() => {
    const listening = filtered.filter((c) => c.status === "LISTEN");
    const established = filtered.filter((c) => c.status === "ESTABLISHED");
    const other = filtered.filter(
      (c) => c.status !== "LISTEN" && c.status !== "ESTABLISHED",
    );

    return { listening, established, other };
  }, [filtered]);

  return (
    <Card>
      <CardHeader>
        <CardTitle>Network Connections ({filtered.length})</CardTitle>
        <Input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search by address, port, status, or PID"
          className="mt-2"
        />
      </CardHeader>
      <CardContent>
        <div className="space-y-6">
          {grouped.listening.length > 0 && (
            <div>
              <h3 className="mb-2 text-sm font-semibold text-muted-foreground">
                Listening ({grouped.listening.length})
              </h3>
              <div className="max-h-[300px] overflow-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>PID</TableHead>
                      <TableHead>Local</TableHead>
                      <TableHead>Type</TableHead>
                      <TableHead>Status</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {grouped.listening.map((conn, idx) => (
                      <TableRow key={idx}>
                        <TableCell className="font-mono text-xs">
                          {conn.pid}
                        </TableCell>
                        <TableCell className="font-mono text-xs">
                          {formatAddress(conn.local.ip, conn.local.port)}
                        </TableCell>
                        <TableCell className="text-xs">
                          {conn.type === 1
                            ? "TCP"
                            : conn.type === 2
                              ? "UDP"
                              : "Other"}
                        </TableCell>
                        <TableCell className="text-xs text-muted-foreground">
                          {conn.status}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </div>
          )}

          {grouped.established.length > 0 && (
            <div>
              <h3 className="mb-2 text-sm font-semibold text-muted-foreground">
                Established ({grouped.established.length})
              </h3>
              <div className="max-h-[300px] overflow-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>PID</TableHead>
                      <TableHead>Local</TableHead>
                      <TableHead>Remote</TableHead>
                      <TableHead>Type</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {grouped.established.map((conn, idx) => (
                      <TableRow key={idx}>
                        <TableCell className="font-mono text-xs">
                          {conn.pid}
                        </TableCell>
                        <TableCell className="font-mono text-xs">
                          {formatAddress(conn.local.ip, conn.local.port)}
                        </TableCell>
                        <TableCell className="font-mono text-xs">
                          {formatAddress(conn.remote.ip, conn.remote.port)}
                        </TableCell>
                        <TableCell className="text-xs">
                          {conn.type === 1
                            ? "TCP"
                            : conn.type === 2
                              ? "UDP"
                              : "Other"}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </div>
          )}

          {grouped.other.length > 0 && (
            <div>
              <h3 className="mb-2 text-sm font-semibold text-muted-foreground">
                Other ({grouped.other.length})
              </h3>
              <div className="max-h-[300px] overflow-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>PID</TableHead>
                      <TableHead>Local</TableHead>
                      <TableHead>Remote</TableHead>
                      <TableHead>Status</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {grouped.other.map((conn, idx) => (
                      <TableRow key={idx}>
                        <TableCell className="font-mono text-xs">
                          {conn.pid}
                        </TableCell>
                        <TableCell className="font-mono text-xs">
                          {formatAddress(conn.local.ip, conn.local.port)}
                        </TableCell>
                        <TableCell className="font-mono text-xs">
                          {formatAddress(conn.remote.ip, conn.remote.port)}
                        </TableCell>
                        <TableCell className="text-xs text-muted-foreground">
                          {conn.status}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </div>
          )}

          {filtered.length === 0 && (
            <div className="py-8 text-center text-sm text-muted-foreground">
              No connections found
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
