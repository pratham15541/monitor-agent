import { Card, CardContent } from "@/components/ui/card";

type StatTileProps = {
  label: string;
  value: string;
  helper?: string;
};

export function StatTile({ label, value, helper }: StatTileProps) {
  return (
    <Card className="border border-muted/40 bg-card/70">
      <CardContent className="px-5 py-4">
        <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">
          {label}
        </p>
        <p className="mt-2 text-2xl font-semibold text-foreground">{value}</p>
        {helper ? (
          <p className="mt-1 text-xs text-muted-foreground">{helper}</p>
        ) : null}
      </CardContent>
    </Card>
  );
}
