import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import type { Device } from "@/lib/types";
import { formatDateTime } from "@/lib/format";

const statusStyles: Record<string, string> = {
  ONLINE: "bg-emerald-500/10 text-emerald-700 border-emerald-200",
  OFFLINE: "bg-rose-500/10 text-rose-700 border-rose-200",
};

export function DeviceCard({ device }: { device: Device }) {
  const card = (
    <Card className="h-full transition hover:-translate-y-1 hover:shadow-md">
      <CardHeader>
        <div className="flex items-start justify-between gap-3">
          <div>
            <CardTitle className="text-lg">
              {device.hostname || "Unnamed device"}
            </CardTitle>
            <CardDescription>{device.os || "Unknown OS"}</CardDescription>
          </div>
          <Badge
            variant="outline"
            className={statusStyles[device.status] ?? "border-muted"}
          >
            {device.status ?? "UNKNOWN"}
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="space-y-2 text-sm text-muted-foreground">
        <div className="flex items-center justify-between">
          <span>IP</span>
          <span className="text-foreground">{device.ipAddress || "-"}</span>
        </div>
        <div className="flex items-center justify-between">
          <span>Last seen</span>
          <span className="text-foreground">
            {formatDateTime(device.lastSeenAt)}
          </span>
        </div>
        {!device.id ? (
          <p className="text-xs text-rose-600">Missing device id</p>
        ) : null}
      </CardContent>
    </Card>
  );

  if (!device.id) {
    return <div className="h-full opacity-80">{card}</div>;
  }

  return (
    <Link href={`/devices/${device.id}`} className="block h-full">
      {card}
    </Link>
  );
}
