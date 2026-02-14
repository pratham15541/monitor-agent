"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { ServicesSnapshot } from "@/lib/types";

type Props = {
  services: ServicesSnapshot;
};

export function ServicesView({ services }: Props) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>
          Services & Background Processes
          {services.source && (
            <span className="ml-2 text-sm font-normal text-muted-foreground">
              (via {services.source})
            </span>
          )}
        </CardTitle>
      </CardHeader>
      <CardContent>
        {services.output ? (
          <pre className="max-h-[600px] overflow-auto rounded-lg border bg-muted/30 p-4 text-xs font-mono leading-relaxed">
            {services.output}
          </pre>
        ) : (
          <div className="py-8 text-center text-sm text-muted-foreground">
            No service information available
          </div>
        )}
      </CardContent>
    </Card>
  );
}
