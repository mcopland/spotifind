import {
  type ReactNode,
  type RefObject,
  useEffect,
  useLayoutEffect,
  useRef,
} from "react";
import { createPortal } from "react-dom";

interface PopoverProps {
  anchor: RefObject<HTMLElement | null>;
  open: boolean;
  onClose: () => void;
  children: ReactNode;
  align?: "start" | "end";
}

export default function Popover({
  anchor,
  open,
  onClose,
  children,
  align = "start",
}: PopoverProps) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    const onDoc = (e: MouseEvent) => {
      if (!ref.current) return;
      if (ref.current.contains(e.target as Node)) return;
      if (anchor.current?.contains(e.target as Node)) return;
      onClose();
    };
    const onEsc = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    document.addEventListener("mousedown", onDoc);
    document.addEventListener("keydown", onEsc);
    return () => {
      document.removeEventListener("mousedown", onDoc);
      document.removeEventListener("keydown", onEsc);
    };
  }, [open, anchor, onClose]);

  // Position the popover directly via DOM after paint to avoid cascading renders
  useLayoutEffect(() => {
    if (!open || !anchor.current || !ref.current) return;
    const r = anchor.current.getBoundingClientRect();
    const left = align === "end" ? r.right - 220 : r.left;
    ref.current.style.top = `${String(Math.round(r.bottom + 4))}px`;
    ref.current.style.left = `${String(Math.round(left))}px`;
    ref.current.style.visibility = "visible";
  }, [open, anchor, align]);

  if (!open) return null;

  return createPortal(
    <div
      ref={ref}
      style={{
        position: "fixed",
        top: 0,
        left: 0,
        visibility: "hidden",
        zIndex: 30,
        minWidth: 200,
        background: "var(--bg)",
        border: "1px solid var(--hair-strong)",
        borderRadius: "var(--radius)",
        boxShadow: "var(--shadow-pop)",
        padding: 6,
      }}
    >
      {children}
    </div>,
    document.body,
  );
}

/** Popover section heading */
export function PopoverGroup({ children }: { children: ReactNode }) {
  return (
    <div
      style={{
        padding: "4px 8px",
        fontSize: 10,
        color: "var(--fg-3)",
        textTransform: "uppercase",
        letterSpacing: "0.08em",
      }}
    >
      {children}
    </div>
  );
}

/** Popover option row */
export function PopoverOption({
  children,
  onClick,
  kbd,
}: {
  children: ReactNode;
  onClick: () => void;
  kbd?: string;
}) {
  return (
    <button
      onClick={onClick}
      style={{
        display: "flex",
        alignItems: "center",
        gap: 8,
        padding: "5px 8px",
        borderRadius: 3,
        fontSize: 12,
        color: "var(--fg-1)",
        cursor: "pointer",
        width: "100%",
        textAlign: "left",
        background: "none",
        border: "none",
      }}
      onMouseEnter={(e) => {
        (e.currentTarget as HTMLElement).style.background = "var(--bg-2)";
        (e.currentTarget as HTMLElement).style.color = "var(--fg)";
      }}
      onMouseLeave={(e) => {
        (e.currentTarget as HTMLElement).style.background = "none";
        (e.currentTarget as HTMLElement).style.color = "var(--fg-1)";
      }}
    >
      <span style={{ flex: 1 }}>{children}</span>
      {kbd && (
        <span
          style={{
            color: "var(--fg-3)",
            fontFamily: "var(--font-mono)",
            fontSize: 10,
          }}
        >
          {kbd}
        </span>
      )}
    </button>
  );
}
