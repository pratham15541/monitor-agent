import { formatNumber } from "@/lib/format";

type MetricChartProps = {
  title: string;
  unit?: string;
  values: number[];
  strokeClassName?: string;
};

export function MetricChart({
  title,
  unit,
  values,
  strokeClassName = "stroke-primary",
}: MetricChartProps) {
  const width = 240;
  const height = 80;

  if (!values.length) {
    return (
      <div className="rounded-xl border bg-card/60 p-4">
        <p className="text-sm font-medium text-foreground">{title}</p>
        <p className="mt-6 text-sm text-muted-foreground">No data yet</p>
      </div>
    );
  }

  const min = Math.min(...values);
  const max = Math.max(...values);
  const range = max - min || 1;

  const points = values
    .map((value, index) => {
      const x = (index / Math.max(values.length - 1, 1)) * width;
      const y = height - ((value - min) / range) * height;
      return `${x},${y}`;
    })
    .join(" ");

  const latest = values[values.length - 1];

  return (
    <div className="rounded-xl border bg-card/60 p-4">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="text-sm font-medium text-foreground">{title}</p>
          <p className="text-2xl font-semibold text-foreground">
            {formatNumber(latest)}
            {unit ? (
              <span className="text-sm text-muted-foreground"> {unit}</span>
            ) : null}
          </p>
        </div>
        <div className="text-right text-xs text-muted-foreground">
          <div>Min {formatNumber(min)}</div>
          <div>Max {formatNumber(max)}</div>
        </div>
      </div>
      <svg
        viewBox={`0 0 ${width} ${height}`}
        className="mt-4 h-20 w-full"
        preserveAspectRatio="none"
      >
        <polyline
          fill="none"
          strokeWidth="2.5"
          className={strokeClassName}
          points={points}
        />
      </svg>
    </div>
  );
}
