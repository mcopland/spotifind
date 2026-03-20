import { useQuery } from "@tanstack/react-query";
import { getPlaylists } from "../api/playlists";
import DataTable, { type ColumnDef } from "../components/table/DataTable";
import type { Playlist } from "../types";

const columns: ColumnDef<Playlist>[] = [
  {
    id: "name",
    header: "Playlist",
    accessorKey: "name",
    cell: ({ row }) => (
      <div className="flex items-center gap-3">
        {row.original.image_url && (
          <img src={row.original.image_url} alt="" className="w-9 h-9 rounded" />
        )}
        <div>
          <div className="font-medium text-white">{row.original.name}</div>
          {row.original.description && (
            <div className="text-xs text-gray-600 truncate max-w-xs">
              {row.original.description}
            </div>
          )}
        </div>
      </div>
    ),
  },
  {
    id: "track_count",
    header: "Tracks",
    accessorKey: "track_count",
    cell: ({ getValue }) => <span className="text-gray-500">{getValue() as number}</span>,
  },
  {
    id: "is_public",
    header: "Public",
    accessorKey: "is_public",
    cell: ({ getValue }) => (
      <span className={`text-xs ${getValue() ? "text-green-500" : "text-gray-600"}`}>
        {getValue() ? "Yes" : "No"}
      </span>
    ),
  },
  {
    id: "collaborative",
    header: "Collaborative",
    accessorKey: "collaborative",
    cell: ({ getValue }) => (
      <span className={`text-xs ${getValue() ? "text-blue-400" : "text-gray-600"}`}>
        {getValue() ? "Yes" : "No"}
      </span>
    ),
  },
];

export default function PlaylistsPage() {
  const { data: playlists = [], isLoading, isError, refetch } = useQuery({
    queryKey: ["playlists"],
    queryFn: getPlaylists,
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
    <div className="h-full overflow-auto">
      <DataTable data={playlists} columns={columns} isLoading={isLoading} />
    </div>
  );
}
