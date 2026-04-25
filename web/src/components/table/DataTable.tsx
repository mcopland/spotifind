import { ChevronDown, ChevronUp, ChevronsUpDown } from "lucide-react";
import { type ReactNode } from "react";

export type SortingState = { id: string; desc: boolean }[];

export interface ColumnDef<T> {
  id: string;
  header: string;
  accessorKey?: string;
  accessorFn?: (row: T) => unknown;
  enableSorting?: boolean;
  numeric?: boolean;
  cell?: (ctx: { row: { original: T }; getValue: () => unknown }) => ReactNode;
}

interface DataTableProps<T> {
  data: T[];
  columns: ColumnDef<T>[];
  sorting?: SortingState;
  onSortingChange?: React.Dispatch<React.SetStateAction<SortingState>>;
  isLoading?: boolean;
}

export default function DataTable<T>({
  data,
  columns,
  sorting,
  onSortingChange,
  isLoading,
}: DataTableProps<T>) {
  function handleSort(colId: string) {
    if (!onSortingChange) return;
    const current = sorting?.[0];
    if (current?.id === colId) {
      onSortingChange(current.desc ? [] : [{ id: colId, desc: true }]);
    } else {
      onSortingChange([{ id: colId, desc: false }]);
    }
  }

  if (isLoading) {
    return (
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          height: 240,
        }}
      >
        <div
          style={{
            width: 22,
            height: 22,
            border: "2px solid var(--acc)",
            borderTopColor: "transparent",
            borderRadius: "50%",
            animation: "spin 0.7s linear infinite",
          }}
        />
        <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
      </div>
    );
  }

  return (
    <div style={{ overflowX: "auto" }}>
      <table
        style={{
          width: "100%",
          borderCollapse: "separate",
          borderSpacing: 0,
          fontSize: 12,
        }}
      >
        <thead>
          <tr>
            {columns.map((col) => (
              <th
                key={col.id}
                onClick={col.enableSorting ? () => { handleSort(col.id); } : undefined}
                style={{
                  position: "sticky",
                  top: 0,
                  background: "var(--bg-1)",
                  fontWeight: 500,
                  color: "var(--fg-2)",
                  textAlign: col.numeric ? "right" : "left",
                  padding: `var(--cell-py) var(--cell-px)`,
                  height: 28,
                  borderBottom: "1px solid var(--hair)",
                  fontSize: 11,
                  whiteSpace: "nowrap",
                  userSelect: "none",
                  cursor: col.enableSorting ? "pointer" : undefined,
                }}
              >
                <span
                  style={{
                    display: "inline-flex",
                    alignItems: "center",
                    gap: 4,
                  }}
                >
                  {col.header}
                  {col.enableSorting && (
                    <span style={{ color: "var(--fg-3)", display: "inline-flex" }}>
                      {sorting?.[0]?.id === col.id ? (
                        sorting[0].desc ? (
                          <ChevronDown size={10} />
                        ) : (
                          <ChevronUp size={10} />
                        )
                      ) : (
                        <ChevronsUpDown size={10} />
                      )}
                    </span>
                  )}
                </span>
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {data.length === 0 ? (
            <tr>
              <td
                colSpan={columns.length}
                style={{
                  padding: "48px 16px",
                  textAlign: "center",
                  color: "var(--fg-3)",
                }}
              >
                No results
              </td>
            </tr>
          ) : (
            data.map((row, i) => (
              <tr
                key={i}
                onMouseEnter={(e) => {
                  const cells = (e.currentTarget as HTMLTableRowElement).querySelectorAll("td");
                  cells.forEach((td) => {
                    (td as HTMLElement).style.background = "var(--bg-1)";
                  });
                }}
                onMouseLeave={(e) => {
                  const cells = (e.currentTarget as HTMLTableRowElement).querySelectorAll("td");
                  cells.forEach((td) => {
                    (td as HTMLElement).style.background = "";
                  });
                }}
              >
                {columns.map((col) => {
                  const value = col.accessorFn
                    ? col.accessorFn(row)
                    : col.accessorKey
                      ? (row as Record<string, unknown>)[col.accessorKey]
                      : undefined;
                  return (
                    <td
                      key={col.id}
                      style={{
                        padding: `var(--cell-py) var(--cell-px)`,
                        height: "var(--row-h)",
                        color: "var(--fg)",
                        verticalAlign: "middle",
                        whiteSpace: "nowrap",
                        overflow: "hidden",
                        textOverflow: "ellipsis",
                        maxWidth: 360,
                        textAlign: col.numeric ? "right" : "left",
                      }}
                    >
                      {col.cell
                        ? col.cell({ row: { original: row }, getValue: () => value })
                        : typeof value === "string" ||
                            typeof value === "number" ||
                            typeof value === "boolean"
                          ? String(value)
                          : ""}
                    </td>
                  );
                })}
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  );
}
