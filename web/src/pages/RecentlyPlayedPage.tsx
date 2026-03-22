import { useQuery } from "@tanstack/react-query";
import { getRecentlyPlayed } from "../api/recently-played";
import DataTable, { type ColumnDef } from "../components/table/DataTable";
import Pagination from "../components/table/Pagination";
import { useFilterStore } from "../stores/filterStore";
import type { RecentlyPlayedTrack } from "../types";

function formatDuration(ms: number): string {
  const s = Math.floor(ms / 1000);
  return `${String(Math.floor(s / 60))}:${String(s % 60).padStart(2, "0")}`;
}

function formatPlayedAt(iso: string): string {
  return new Date(iso).toLocaleString(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

const columns: ColumnDef<RecentlyPlayedTrack>[] = [
  {
    id: "name",
    header: "Title",
    accessorKey: "name",
    cell: ({ row }) => (
      <div>
        <div className="font-medium text-white truncate max-w-xs">{row.original.name}</div>
        {row.original.explicit && (
          <span className="text-[10px] text-gray-600 border border-gray-700 rounded px-1">E</span>
        )}
      </div>
    ),
  },
  {
    id: "artists",
    header: "Artist",
    accessorFn: r => r.artists.map(a => a.name).join(", "),
    cell: ({ getValue }) => (
      <span className="text-gray-400 truncate max-w-[40] block">{getValue() as string}</span>
    ),
  },
  {
    id: "album",
    header: "Album",
    accessorFn: r => r.album?.name ?? "",
    cell: ({ row }) => (
      <div className="flex items-center gap-2">
        {row.original.album?.image_url && (
          <img src={row.original.album.image_url} alt="" className="w-7 h-7 rounded" />
        )}
        <span className="text-gray-400 truncate max-w-[35]">{row.original.album?.name}</span>
      </div>
    ),
  },
  {
    id: "duration",
    header: "Duration",
    accessorKey: "duration_ms",
    cell: ({ row }) => (
      <span className="text-gray-500">{formatDuration(row.original.duration_ms)}</span>
    ),
  },
  {
    id: "played_at",
    header: "Played At",
    accessorKey: "played_at",
    cell: ({ row }) => (
      <span className="text-gray-500 text-xs">{formatPlayedAt(row.original.played_at)}</span>
    ),
  },
];

export default function RecentlyPlayedPage() {
  const { page, pageSize, setPage } = useFilterStore();

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ["recently-played", { page, pageSize }],
    queryFn: () => getRecentlyPlayed({ page, page_size: pageSize }),
  });

  if (isError) {
    return (
      <div className="flex flex-col items-center justify-center h-64 gap-4">
        <p className="text-red-400">Failed to load data.</p>
        <button
          onClick={() => void refetch()}
          className="px-4 py-2 bg-[#1DB954] text-black text-sm font-medium rounded hover:bg-[#1ed760] transition-colors"
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      <DataTable
        data={data?.items ?? []}
        columns={columns}
        sorting={[]}
        onSortingChange={() => undefined}
        isLoading={isLoading}
      />
      <Pagination
        page={page}
        pageSize={pageSize}
        total={data?.total ?? 0}
        onPageChange={setPage}
      />
    </div>
  );
}
