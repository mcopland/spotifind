import { X } from "lucide-react";
import { type ReactNode, useRef, useState } from "react";
import Popover from "./Popover";

interface FilterChipProps {
  label: string;
  value?: string;
  applied?: boolean;
  onRemove?: () => void;
  children?: ReactNode;
}

export default function FilterChip({
  label,
  value,
  applied = false,
  onRemove,
  children,
}: FilterChipProps) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLButtonElement>(null);

  if (!children) {
    // Display-only applied chip (no popover)
    return (
      <span
        style={{
          display: "inline-flex",
          alignItems: "center",
          gap: 5,
          height: 24,
          padding: "0 8px",
          border: `1px solid ${applied ? "color-mix(in oklch, var(--acc) 40%, var(--hair))" : "var(--hair-strong)"}`,
          borderStyle: applied ? "solid" : "dashed",
          borderRadius: 999,
          fontSize: 11.5,
          color: applied ? "var(--acc-ink)" : "var(--fg-1)",
          background: applied ? "var(--acc-soft)" : "var(--bg)",
          cursor: "default",
        }}
      >
        <span style={{ color: applied ? "var(--acc-ink)" : "var(--fg-2)", opacity: applied ? 0.7 : 1 }}>
          {label}
        </span>
        {value !== undefined && (
          <span style={{ fontWeight: 500 }}>{value}</span>
        )}
        {applied && onRemove && (
          <button
            onClick={(e) => {
              e.stopPropagation();
              onRemove();
            }}
            style={{
              marginLeft: 2,
              opacity: 0.5,
              display: "flex",
              alignItems: "center",
              cursor: "pointer",
            }}
            onMouseEnter={(e) => {
              (e.currentTarget as HTMLElement).style.opacity = "1";
            }}
            onMouseLeave={(e) => {
              (e.currentTarget as HTMLElement).style.opacity = "0.5";
            }}
          >
            <X size={10} />
          </button>
        )}
      </span>
    );
  }

  return (
    <>
      <button
        ref={ref}
        onClick={() => { setOpen((o) => !o); }}
        style={{
          display: "inline-flex",
          alignItems: "center",
          gap: 5,
          height: 24,
          padding: "0 8px",
          border: `1px solid ${applied ? "color-mix(in oklch, var(--acc) 40%, var(--hair))" : "var(--hair-strong)"}`,
          borderStyle: applied ? "solid" : "dashed",
          borderRadius: 999,
          fontSize: 11.5,
          color: applied ? "var(--acc-ink)" : "var(--fg-1)",
          background: applied ? "var(--acc-soft)" : "var(--bg)",
          cursor: "pointer",
        }}
      >
        <span style={{ color: applied ? "var(--acc-ink)" : "var(--fg-2)", opacity: applied ? 0.7 : 1 }}>
          {label}
        </span>
        {value !== undefined && (
          <span style={{ fontWeight: 500 }}>{value}</span>
        )}
        {applied && onRemove && (
          <button
            onClick={(e) => {
              e.stopPropagation();
              onRemove();
              setOpen(false);
            }}
            style={{
              marginLeft: 2,
              opacity: 0.5,
              display: "flex",
              alignItems: "center",
              cursor: "pointer",
            }}
            onMouseEnter={(e) => {
              (e.currentTarget as HTMLElement).style.opacity = "1";
            }}
            onMouseLeave={(e) => {
              (e.currentTarget as HTMLElement).style.opacity = "0.5";
            }}
          >
            <X size={10} />
          </button>
        )}
      </button>
      <Popover anchor={ref} open={open} onClose={() => { setOpen(false); }}>
        {children}
      </Popover>
    </>
  );
}
