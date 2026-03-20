import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { getTracks } from "../api/tracks";
import FilterSidebar from "../components/filters/FilterSidebar";
import DataTable, { type ColumnDef, type SortingState } from "../components/table/DataTable";
import Pagination from "../components/table/Pagination";
import { useFilterStore } from "../stores/filterStore";
import type { Track } from "../types";

function formatDuration(ms: number): string {
  const s = Math.floor(ms / 1000);
  return `${String(Math.floor(s / 60))}:${String(s % 60).padStart(2, "0")}`;
}

const columns: ColumnDef<Track>[] = [
  {
    id: "name",
    header: "Title",
    accessorKey: "name",
    enableSorting: true,
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
    enableSorting: true,
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
    id: "year",
    header: "Year",
    accessorFn: r => r.album?.release_year ?? "",
    enableSorting: true,
    cell: ({ getValue }) => <span className="text-gray-500">{getValue() as string}</span>,
  },
  {
    id: "duration",
    header: "Duration",
    accessorKey: "duration_ms",
    enableSorting: true,
    cell: ({ row }) => (
      <span className="text-gray-500">{formatDuration(row.original.duration_ms)}</span>
    ),
  },
  {
    id: "popularity",
    header: "Pop.",
    accessorKey: "popularity",
    enableSorting: true,
    cell: ({ row }) => (
      <div className="flex items-center gap-1.5">
        <div className="w-12 h-1 bg-gray-800 rounded-full overflow-hidden">
          <div className="h-full bg-[#1DB954]" style={{ width: `${String(row.original.popularity)}%` }} />
        </div>
        <span className="text-gray-600 text-xs">{row.original.popularity}</span>
      </div>
    ),
  },
];

export default function TracksPage() {
  const [sorting, setSorting] = useState<SortingState>([]);
  const {
    search,
    genres,
    yearMin,
    yearMax,
    popularityMin,
    popularityMax,
    explicit,
    page,
    pageSize,
    setPage,
  } = useFilterStore();

  const sortBy = sorting[0]?.id;
  const sortDir = sorting[0] ? (sorting[0].desc ? "desc" : "asc") : undefined;

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: [
      "tracks",
      {
        search,
        genres,
        yearMin,
        yearMax,
        popularityMin,
        popularityMax,
        explicit,
        page,
        pageSize,
        sortBy,
        sortDir,
      },
    ],
    queryFn: () =>
      getTracks({
        search,
        genres,
        year_min: yearMin,
        year_max: yearMax,
        popularity_min: popularityMin,
        popularity_max: popularityMax,
        explicit,
        page,
        page_size: pageSize,
        sort_by: sortBy,
        sort_dir: sortDir,
      }),
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
    <div className="flex h-full">
      <FilterSidebar />
      <div className="flex-1 flex flex-col overflow-hidden">
        <DataTable
          data={data?.items ?? []}
          columns={columns}
          sorting={sorting}
          onSortingChange={setSorting}
          isLoading={isLoading}
        />
        <Pagination
          page={page}
          pageSize={pageSize}
          total={data?.total ?? 0}
          onPageChange={setPage}
        />
      </div>
    </div>
  );
}
