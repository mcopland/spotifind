import type { ReactNode } from "react";
import Spark from "./Spark";

interface StatCardProps {
  label: ReactNode;
  value: string | number;
  delta?: string;
  deltaUp?: boolean;
  sparkData?: number[];
}

export default function StatCard({ label, value, delta, deltaUp, sparkData }: StatCardProps) {
  return (
    <div
      style={{
        border: "1px solid var(--hair)",
        background: "var(--bg-1)",
        borderRadius: "var(--radius)",
        padding: 12,
        display: "flex",
        flexDirection: "column",
        gap: 4,
      }}
    >
      <div
        style={{
          fontSize: 11,
          color: "var(--fg-2)",
          display: "flex",
          alignItems: "center",
          gap: 5,
        }}
      >
        {label}
      </div>
      <div
        style={{
          fontSize: 22,
          fontWeight: 600,
          letterSpacing: "-0.02em",
          fontVariantNumeric: "tabular-nums",
          fontFamily: "var(--font-mono)",
        }}
      >
        {value}
      </div>
      {delta && (
        <div
          style={{
            fontSize: 11,
            fontVariantNumeric: "tabular-nums",
            color:
              deltaUp === true
                ? "var(--ok)"
                : deltaUp === false
                  ? "var(--err)"
                  : "var(--fg-2)",
          }}
        >
          {delta}
        </div>
      )}
      {sparkData && sparkData.length > 1 && (
        <div style={{ marginTop: 2 }}>
          <Spark data={sparkData} width={160} height={22} />
        </div>
      )}
    </div>
  );
}
