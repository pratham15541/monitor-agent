"use client";

import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import type { LogsSnapshot } from "@/lib/types";

type Props = {
  logs: LogsSnapshot;
};

export function LogsView({ logs }: Props) {
  const [activeTab, setActiveTab] = useState<"agent" | "system">("agent");

  const activeLog = activeTab === "agent" ? logs.agent : logs.system;
  const lines = activeLog ? activeLog.split("\n").length : 0;

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle>Logs</CardTitle>
          <div className="flex gap-2">
            <Button
              size="sm"
              variant={activeTab === "agent" ? "default" : "outline"}
              onClick={() => setActiveTab("agent")}
            >
              Agent
            </Button>
            <Button
              size="sm"
              variant={activeTab === "system" ? "default" : "outline"}
              onClick={() => setActiveTab("system")}
            >
              System
            </Button>
          </div>
        </div>
        {activeLog && (
          <p className="text-xs text-muted-foreground">
            {lines} line{lines !== 1 ? "s" : ""}
          </p>
        )}
      </CardHeader>
      <CardContent>
        {activeLog ? (
          <pre className="max-h-[600px] overflow-auto rounded-lg border bg-black/95 p-4 text-xs font-mono leading-relaxed text-green-400">
            {activeLog}
          </pre>
        ) : (
          <div className="py-8 text-center text-sm text-muted-foreground">
            No {activeTab} logs available
          </div>
        )}
      </CardContent>
    </Card>
  );
}
