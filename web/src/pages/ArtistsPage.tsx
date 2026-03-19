import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { getArtists } from "../api/artists";
import FilterSidebar from "../components/filters/FilterSidebar";
import DataTable, { type ColumnDef, type SortingState } from "../components/table/DataTable";
import Pagination from "../components/table/Pagination";
import { useFilterStore } from "../stores/filterStore";
import type { Artist } from "../types";

const columns: ColumnDef<Artist>[] = [
  {
    id: "name",
    header: "Artist",
    accessorKey: "name",
    enableSorting: true,
    cell: ({ row }) => (
      <div className="flex items-center gap-3">
        {row.original.image_url && (
          <img src={row.original.image_url} alt="" className="w-9 h-9 rounded-full object-cover" />
        )}
        <span className="font-medium text-white">{row.original.name}</span>
      </div>
    ),
  },
  {
    id: "genres",
    header: "Genres",
    accessorFn: r => r.genres?.slice(0, 3).join(", ") ?? "",
    cell: ({ getValue }) => <span className="text-gray-400 text-xs">{getValue() as string}</span>,
  },
  {
    id: "popularity",
    header: "Popularity",
    accessorKey: "popularity",
    enableSorting: true,
    cell: ({ row }) => (
      <div className="flex items-center gap-1.5">
        <div className="w-16 h-1 bg-gray-800 rounded-full overflow-hidden">
          <div className="h-full bg-[#1DB954]" style={{ width: `${String(row.original.popularity ?? 0)}%` }} />
        </div>
        <span className="text-gray-600 text-xs">{row.original.popularity}</span>
      </div>
    ),
  },
  {
    id: "followers",
    header: "Followers",
    accessorKey: "followers",
    enableSorting: true,
    cell: ({ row }) => (
      <span className="text-gray-500">{row.original.followers?.toLocaleString()}</span>
    ),
  },
];

export default function ArtistsPage() {
  const [sorting, setSorting] = useState<SortingState>([]);
  const { search, genres, page, pageSize, setPage } = useFilterStore();

  const sortBy = sorting[0]?.id;
  const sortDir = sorting[0] ? (sorting[0].desc ? "desc" : "asc") : undefined;

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ["artists", { search, genres, page, pageSize, sortBy, sortDir }],
    queryFn: () =>
      getArtists({ search, genres, page, page_size: pageSize, sort_by: sortBy, sort_dir: sortDir }),
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
