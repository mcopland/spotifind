import { ChevronDown, ChevronsUpDown, ChevronUp } from "lucide-react";
import { type ReactNode } from "react";

export type SortingState = { id: string; desc: boolean }[];

export interface ColumnDef<T> {
  id: string;
  header: string;
  accessorKey?: string;
  accessorFn?: (row: T) => unknown;
  enableSorting?: boolean;
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
      <div className="flex items-center justify-center h-64">
        <div className="w-6 h-6 border-2 border-[#1DB954] border-t-transparent rounded-full animate-spin" />
      </div>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-gray-800">
            {columns.map(col => (
              <th
                key={col.id}
                className={`px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wide ${
                  col.enableSorting ? "cursor-pointer select-none hover:text-gray-300" : ""
                }`}
                onClick={col.enableSorting ? () => { handleSort(col.id); } : undefined}
              >
                <div className="flex items-center gap-1">
                  {col.header}
                  {col.enableSorting && (
                    <span className="text-gray-600">
                      {sorting?.[0]?.id === col.id ? (
                        sorting[0].desc ? (
                          <ChevronDown className="w-3 h-3" />
                        ) : (
                          <ChevronUp className="w-3 h-3" />
                        )
                      ) : (
                        <ChevronsUpDown className="w-3 h-3" />
                      )}
                    </span>
                  )}
                </div>
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {data.length === 0 ? (
            <tr>
              <td colSpan={columns.length} className="px-4 py-16 text-center text-gray-600">
                No results
              </td>
            </tr>
          ) : (
            data.map((row, i) => (
              <tr
                key={i}
                className="border-b border-gray-900 hover:bg-gray-900/50 transition-colors"
              >
                {columns.map(col => {
                  const value = col.accessorFn
                    ? col.accessorFn(row)
                    : col.accessorKey
                      ? (row as Record<string, unknown>)[col.accessorKey]
                      : undefined;
                  return (
                    <td key={col.id} className="px-4 py-2.5 text-gray-300">
                      {col.cell
                        ? col.cell({ row: { original: row }, getValue: () => value })
                        : (typeof value === "string" || typeof value === "number" || typeof value === "boolean" ? String(value) : "")}
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
