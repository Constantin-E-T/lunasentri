"use client";

import { Area, AreaChart, CartesianGrid, XAxis } from "recharts";
import { ChartConfig, ChartContainer } from "@/components/ui/chart";
import { useRef, useState } from "react";
import { useSpring, useMotionValueEvent } from "motion/react";
import type { MetricSample } from "@/lib/useMetrics";

interface MetricSparklineProps {
  data: MetricSample[];
  metric: 'cpu_pct' | 'mem_used_pct';
  className?: string;
}

const chartConfig = {
  cpu_pct: {
    label: "CPU",
    color: "hsl(var(--chart-1))",
  },
  mem_used_pct: {
    label: "Memory", 
    color: "hsl(var(--chart-2))",
  },
} satisfies ChartConfig;

export function MetricSparkline({ data, metric, className }: MetricSparklineProps) {
  const chartRef = useRef<HTMLDivElement>(null);
  const [axis, setAxis] = useState(0);

  // motion values for smooth tracking
  const springX = useSpring(0, {
    damping: 30,
    stiffness: 100,
  });
  const springY = useSpring(0, {
    damping: 30,
    stiffness: 100,
  });

  useMotionValueEvent(springX, "change", (latest) => {
    setAxis(latest);
  });

  // Transform data for the chart
  const chartData = data.map((sample, index) => ({
    index,
    value: sample[metric],
    timestamp: sample.timestamp,
  }));

  // Loading skeleton if no data
  if (data.length === 0) {
    return (
      <div className={`h-20 bg-muted/30 rounded-lg animate-pulse ${className}`}>
        <div className="flex items-end justify-center h-full p-2">
          <div className="flex items-end space-x-1 h-full">
            {Array.from({ length: 20 }, (_, i) => (
              <div
                key={i}
                className="bg-muted/60 rounded-sm w-1"
                style={{ height: `${Math.random() * 60 + 20}%` }}
              />
            ))}
          </div>
        </div>
      </div>
    );
  }

  const currentValue = chartData[chartData.length - 1]?.value || 0;
  const metricColor = chartConfig[metric].color;

  return (
    <div className={className}>
      <ChartContainer
        ref={chartRef}
        className="h-20 w-full"
        config={chartConfig}
      >
        <AreaChart
          className="overflow-visible"
          accessibilityLayer
          data={chartData}
          onMouseMove={(state) => {
            const x = state.activeCoordinate?.x;
            const dataValue = state.activePayload?.[0]?.value;
            if (x && dataValue !== undefined) {
              springX.set(x);
              springY.set(dataValue);
            }
          }}
          onMouseLeave={() => {
            const chartWidth = chartRef.current?.getBoundingClientRect().width || 0;
            springX.set(chartWidth);
            springY.jump(currentValue);
          }}
          margin={{
            top: 5,
            right: 5,
            left: 5,
            bottom: 5,
          }}
        >
          <CartesianGrid
            vertical={false}
            horizontal={false}
            strokeDasharray="3 3"
            strokeOpacity={0.1}
          />
          <XAxis
            dataKey="index"
            hide
            domain={['dataMin', 'dataMax']}
          />
          
          {/* Main filled area with clipping effect */}
          <Area
            dataKey="value"
            type="monotone"
            fill={`url(#gradient-${metric})`}
            fillOpacity={0.3}
            stroke={metricColor}
            strokeWidth={2}
            clipPath={`inset(0 ${
              Number(chartRef.current?.getBoundingClientRect().width) - axis
            } 0 0)`}
            dot={false}
            activeDot={false}
          />
          
          {/* Animated vertical tracking line */}
          <line
            x1={axis}
            y1={0}
            x2={axis}
            y2="100%"
            stroke={metricColor}
            strokeDasharray="2 2"
            strokeOpacity={0.4}
            strokeWidth={1}
          />
          
          {/* Ghost line behind the filled area */}
          <Area
            dataKey="value"
            type="monotone"
            fill="none"
            stroke={metricColor}
            strokeOpacity={0.15}
            strokeWidth={1}
            dot={false}
            activeDot={false}
          />
          
          <defs>
            <linearGradient
              id={`gradient-${metric}`}
              x1="0"
              y1="0"
              x2="0"
              y2="1"
            >
              <stop
                offset="5%"
                stopColor={metricColor}
                stopOpacity={0.4}
              />
              <stop
                offset="95%"
                stopColor={metricColor}
                stopOpacity={0}
              />
            </linearGradient>
          </defs>
        </AreaChart>
      </ChartContainer>
    </div>
  );
}