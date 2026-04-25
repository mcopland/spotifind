import { ChevronLeft, ChevronRight } from "lucide-react";

interface PaginationProps {
  page: number;
  pageSize: number;
  total: number;
  onPageChange: (page: number) => void;
}

export default function Pagination({ page, pageSize, total, onPageChange }: PaginationProps) {
  const totalPages = Math.ceil(total / pageSize);
  const from = total > 0 ? (page - 1) * pageSize + 1 : 0;
  const to = Math.min(page * pageSize, total);

  const btnStyle = (disabled: boolean) => ({
    display: "inline-flex",
    alignItems: "center",
    gap: 6,
    padding: "4px 9px",
    height: 26,
    border: "1px solid var(--hair)",
    borderRadius: "var(--radius-sm)",
    background: "var(--bg)",
    color: "var(--fg-1)",
    fontSize: 12,
    cursor: disabled ? "not-allowed" : "pointer",
    opacity: disabled ? 0.3 : 1,
  });

  return (
    <div
      style={{
        padding: "12px 20px",
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        borderTop: "1px solid var(--hair)",
      }}
    >
      <span
        style={{
          fontSize: 11,
          color: "var(--fg-2)",
          fontFamily: "var(--font-mono)",
        }}
      >
        {total > 0
          ? `Showing ${String(from)}–${String(to)} of ${total.toLocaleString()}`
          : "0 results"}
      </span>
      <div style={{ display: "flex", alignItems: "center", gap: 4 }}>
        <button
          onClick={() => { onPageChange(page - 1); }}
          disabled={page <= 1}
          style={btnStyle(page <= 1)}
        >
          <ChevronLeft size={13} /> Prev
        </button>
        <button
          onClick={() => { onPageChange(page + 1); }}
          disabled={page >= totalPages}
          style={btnStyle(page >= totalPages)}
        >
          Next <ChevronRight size={13} />
        </button>
      </div>
    </div>
  );
}
