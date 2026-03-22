import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { getTopArtists, getTopTracks } from "../api/top";
import DataTable, { type ColumnDef } from "../components/table/DataTable";
import Pagination from "../components/table/Pagination";
import { useFilterStore } from "../stores/filterStore";
import type { TimeRange, TopArtist, TopTrack } from "../types";

const TIME_RANGE_LABELS: Record<TimeRange, string> = {
  short_term: "Last 4 Weeks",
  medium_term: "Last 6 Months",
  long_term: "All Time",
};

const trackColumns: ColumnDef<TopTrack>[] = [
  {
    id: "rank",
    header: "#",
    accessorKey: "rank",
    cell: ({ row }) => <span className="text-gray-400 font-medium">{row.original.rank}</span>,
  },
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
];

const artistColumns: ColumnDef<TopArtist>[] = [
  {
    id: "rank",
    header: "#",
    accessorKey: "rank",
    cell: ({ row }) => <span className="text-gray-400 font-medium">{row.original.rank}</span>,
  },
  {
    id: "name",
    header: "Artist",
    accessorKey: "name",
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
    cell: ({ getValue }) => (
      <span className="text-gray-400 text-xs">{getValue() as string}</span>
    ),
  },
  {
    id: "popularity",
    header: "Popularity",
    accessorKey: "popularity",
    cell: ({ row }) => (
      <div className="flex items-center gap-1.5">
        <div className="w-16 h-1 bg-gray-800 rounded-full overflow-hidden">
          <div className="h-full bg-[#1DB954]" style={{ width: `${String(row.original.popularity ?? 0)}%` }} />
        </div>
        <span className="text-gray-600 text-xs">{row.original.popularity}</span>
      </div>
    ),
  },
];

export default function TopChartsPage() {
  const [activeTab, setActiveTab] = useState<"tracks" | "artists">("tracks");
  const [timeRange, setTimeRange] = useState<TimeRange>("short_term");
  const { page, pageSize, setPage } = useFilterStore();

  const { data: trackData, isLoading: tracksLoading, isError: tracksError, refetch: refetchTracks } = useQuery({
    queryKey: ["top", "tracks", timeRange, { page, pageSize }],
    queryFn: () => getTopTracks({ time_range: timeRange, page, page_size: pageSize }),
    enabled: activeTab === "tracks",
  });

  const { data: artistData, isLoading: artistsLoading, isError: artistsError, refetch: refetchArtists } = useQuery({
    queryKey: ["top", "artists", timeRange, { page, pageSize }],
    queryFn: () => getTopArtists({ time_range: timeRange, page, page_size: pageSize }),
    enabled: activeTab === "artists",
  });

  function handleTabChange(tab: "tracks" | "artists") {
    setActiveTab(tab);
    setPage(1);
  }

  function handleTimeRangeChange(range: TimeRange) {
    setTimeRange(range);
    setPage(1);
  }

  const isError = activeTab === "tracks" ? tracksError : artistsError;
  const refetch = activeTab === "tracks" ? refetchTracks : refetchArtists;

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
      <div className="flex items-center justify-between px-4 py-3 border-b border-gray-800 shrink-0">
        <div className="flex gap-1">
          {(["tracks", "artists"] as const).map(tab => (
            <button
              key={tab}
              onClick={() => { handleTabChange(tab); }}
              className={`px-3 py-1.5 text-sm rounded transition-colors capitalize ${
                activeTab === tab
                  ? "bg-[#1DB954] text-black font-medium"
                  : "text-gray-400 hover:text-white hover:bg-gray-800"
              }`}
            >
              {tab}
            </button>
          ))}
        </div>
        <div className="flex gap-1">
          {(Object.entries(TIME_RANGE_LABELS) as [TimeRange, string][]).map(([range, label]) => (
            <button
              key={range}
              onClick={() => { handleTimeRangeChange(range); }}
              className={`px-3 py-1.5 text-xs rounded transition-colors ${
                timeRange === range
                  ? "bg-gray-700 text-white"
                  : "text-gray-500 hover:text-white hover:bg-gray-800"
              }`}
            >
              {label}
            </button>
          ))}
        </div>
      </div>

      {activeTab === "tracks" ? (
        <>
          <DataTable
            data={trackData?.items ?? []}
            columns={trackColumns}
            sorting={[]}
            onSortingChange={() => undefined}
            isLoading={tracksLoading}
          />
          <Pagination
            page={page}
            pageSize={pageSize}
            total={trackData?.total ?? 0}
            onPageChange={setPage}
          />
        </>
      ) : (
        <>
          <DataTable
            data={artistData?.items ?? []}
            columns={artistColumns}
            sorting={[]}
            onSortingChange={() => undefined}
            isLoading={artistsLoading}
          />
          <Pagination
            page={page}
            pageSize={pageSize}
            total={artistData?.total ?? 0}
            onPageChange={setPage}
          />
        </>
      )}
    </div>
  );
}
