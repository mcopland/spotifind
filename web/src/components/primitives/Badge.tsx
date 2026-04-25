import type { CSSProperties, ReactNode } from "react";

type BadgeVariant = "default" | "ok" | "warn" | "err" | "info" | "acc";

interface BadgeProps {
  children: ReactNode;
  variant?: BadgeVariant;
  style?: CSSProperties;
  onClick?: () => void;
}

const VARIANT_STYLES: Record<BadgeVariant, CSSProperties> = {
  default: {
    background: "var(--bg-2)",
    color: "var(--fg-1)",
    borderColor: "var(--hair)",
  },
  ok: {
    background: "color-mix(in oklch, var(--ok) 12%, var(--bg-1))",
    color: "color-mix(in oklch, var(--ok) 60%, var(--fg))",
    borderColor: "color-mix(in oklch, var(--ok) 20%, var(--hair))",
  },
  warn: {
    background: "color-mix(in oklch, var(--warn) 15%, var(--bg-1))",
    color: "color-mix(in oklch, var(--warn) 55%, var(--fg))",
    borderColor: "color-mix(in oklch, var(--warn) 25%, var(--hair))",
  },
  err: {
    background: "color-mix(in oklch, var(--err) 12%, var(--bg-1))",
    color: "color-mix(in oklch, var(--err) 55%, var(--fg))",
    borderColor: "color-mix(in oklch, var(--err) 25%, var(--hair))",
  },
  info: {
    background: "color-mix(in oklch, var(--info) 12%, var(--bg-1))",
    color: "color-mix(in oklch, var(--info) 55%, var(--fg))",
    borderColor: "color-mix(in oklch, var(--info) 25%, var(--hair))",
  },
  acc: {
    background: "var(--acc-soft)",
    color: "var(--acc-ink)",
    borderColor: "color-mix(in oklch, var(--acc) 25%, var(--hair))",
  },
};

export default function Badge({ children, variant = "default", style, onClick }: BadgeProps) {
  return (
    <span
      onClick={onClick}
      style={{
        display: "inline-flex",
        alignItems: "center",
        gap: 4,
        padding: "1px 6px",
        height: 18,
        borderRadius: 3,
        border: "1px solid",
        fontFamily: "var(--font-mono)",
        fontSize: 11,
        fontVariantNumeric: "tabular-nums",
        cursor: onClick ? "pointer" : undefined,
        ...VARIANT_STYLES[variant],
        ...style,
      }}
    >
      {children}
    </span>
  );
}
