import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { getAlbums } from "../api/albums";
import FilterSidebar from "../components/filters/FilterSidebar";
import DataTable, { type ColumnDef, type SortingState } from "../components/table/DataTable";
import Pagination from "../components/table/Pagination";
import { useFilterStore } from "../stores/filterStore";
import type { Album } from "../types";

const columns: ColumnDef<Album>[] = [
  {
    id: "name",
    header: "Album",
    accessorKey: "name",
    enableSorting: true,
    cell: ({ row }) => (
      <div className="flex items-center gap-3">
        {row.original.image_url && (
          <img src={row.original.image_url} alt="" className="w-9 h-9 rounded" />
        )}
        <span className="font-medium text-white truncate max-w-xs">{row.original.name}</span>
      </div>
    ),
  },
  {
    id: "artists",
    header: "Artist",
    accessorFn: r => r.artists?.map(a => a.name).join(", ") ?? "",
    cell: ({ getValue }) => <span className="text-gray-400">{getValue() as string}</span>,
  },
  {
    id: "release_year",
    header: "Year",
    accessorKey: "release_year",
    enableSorting: true,
    cell: ({ getValue }) => <span className="text-gray-500">{getValue() as number}</span>,
  },
  {
    id: "album_type",
    header: "Type",
    accessorKey: "album_type",
    cell: ({ getValue }) => (
      <span className="text-xs text-gray-600 uppercase">{getValue() as string}</span>
    ),
  },
  {
    id: "total_tracks",
    header: "Tracks",
    accessorKey: "total_tracks",
    enableSorting: true,
    cell: ({ getValue }) => <span className="text-gray-500">{getValue() as number}</span>,
  },
];

export default function AlbumsPage() {
  const [sorting, setSorting] = useState<SortingState>([]);
  const { search, genres, yearMin, yearMax, page, pageSize, setPage } = useFilterStore();

  const sortBy = sorting[0]?.id;
  const sortDir = sorting[0] ? (sorting[0].desc ? "desc" : "asc") : undefined;

  const { data, isLoading } = useQuery({
    queryKey: ["albums", { search, genres, yearMin, yearMax, page, pageSize, sortBy, sortDir }],
    queryFn: () =>
      getAlbums({
        search,
        genres,
        year_min: yearMin,
        year_max: yearMax,
        page,
        page_size: pageSize,
        sort_by: sortBy,
        sort_dir: sortDir,
      }),
  });

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
