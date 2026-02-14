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
import type { ProcessInfo } from "@/lib/types";

type Props = {
  processes: ProcessInfo[];
};

function formatBytes(bytes?: number) {
  if (!bytes) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`;
}

function formatPercent(value?: number) {
  if (value === undefined || value === null) return "0%";
  return `${value.toFixed(1)}%`;
}

export function ProcessListView({ processes }: Props) {
  const [search, setSearch] = useState("");

  const filtered = useMemo(() => {
    const needle = search.trim().toLowerCase();
    if (!needle) return processes;

    return processes.filter((proc) => {
      return (
        proc.name?.toLowerCase().includes(needle) ||
        proc.exe?.toLowerCase().includes(needle) ||
        proc.cmdline?.toLowerCase().includes(needle) ||
        proc.username?.toLowerCase().includes(needle) ||
        proc.pid.toString().includes(needle)
      );
    });
  }, [processes, search]);

  const sorted = useMemo(() => {
    return [...filtered].sort((a, b) => {
      const cpuA = a.cpuPercent ?? 0;
      const cpuB = b.cpuPercent ?? 0;
      return cpuB - cpuA;
    });
  }, [filtered]);

  return (
    <Card>
      <CardHeader>
        <CardTitle>Processes ({sorted.length})</CardTitle>
        <Input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search by name, exe, cmdline, username, or PID"
          className="mt-2"
        />
      </CardHeader>
      <CardContent>
        <div className="max-h-[600px] overflow-auto">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>PID</TableHead>
                <TableHead>Name</TableHead>
                <TableHead>User</TableHead>
                <TableHead>CPU</TableHead>
                <TableHead>Memory</TableHead>
                <TableHead>RSS</TableHead>
                <TableHead>Threads</TableHead>
                <TableHead>Status</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {sorted.length === 0 ? (
                <TableRow>
                  <TableCell
                    colSpan={8}
                    className="text-center text-muted-foreground"
                  >
                    No processes found
                  </TableCell>
                </TableRow>
              ) : (
                sorted.map((proc) => (
                  <TableRow key={proc.pid}>
                    <TableCell className="font-mono text-xs">
                      {proc.pid}
                    </TableCell>
                    <TableCell
                      className="max-w-[200px] truncate font-medium"
                      title={proc.name || proc.exe}
                    >
                      {proc.name || proc.exe || "—"}
                    </TableCell>
                    <TableCell className="text-xs text-muted-foreground">
                      {proc.username || "—"}
                    </TableCell>
                    <TableCell className="font-mono text-xs">
                      {formatPercent(proc.cpuPercent)}
                    </TableCell>
                    <TableCell className="font-mono text-xs">
                      {formatPercent(proc.memoryPercent)}
                    </TableCell>
                    <TableCell className="font-mono text-xs">
                      {formatBytes(proc.memoryRssBytes)}
                    </TableCell>
                    <TableCell className="text-xs">
                      {proc.threads ?? "—"}
                    </TableCell>
                    <TableCell className="text-xs text-muted-foreground">
                      {proc.status?.[0] ?? "—"}
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>
  );
}
