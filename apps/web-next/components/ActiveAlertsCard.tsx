"use client";

import Link from "next/link";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { AlertTriangle, ChevronRight } from "lucide-react";

export function ActiveAlertsCard() {
  return (
    <Card className="w-full bg-card/70 border border-border/30 backdrop-blur-xl shadow-2xl transition-all duration-300 hover:-translate-y-1 hover:shadow-[0_20px_60px_rgba(20,40,120,0.35)]">
      <CardHeader className="pb-4">
        <CardTitle className="flex items-center gap-2 text-foreground">
          <AlertTriangle className="h-5 w-5" />
          Active Alerts
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="text-center py-8 space-y-3">
          <div className="text-muted-foreground text-sm">
            No active alerts at the moment
          </div>
          <p className="text-xs text-muted-foreground/70">
            Your system is running smoothly
          </p>
        </div>

        <Link
          href="/alerts"
          className="flex items-center justify-center gap-2 text-sm text-primary hover:text-primary/80 transition-colors duration-200 group"
        >
          <span>View all alerts</span>
          <ChevronRight className="h-4 w-4 transition-transform duration-200 group-hover:translate-x-1" />
        </Link>
      </CardContent>
    </Card>
  );
}
