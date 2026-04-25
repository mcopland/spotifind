import { useQuery } from "@tanstack/react-query";
import { ChevronLeft } from "lucide-react";
import { useLocation, useParams } from "react-router-dom";
import { getPlaylistTracks, getPlaylists } from "../api/playlists";
import DataTable, { type ColumnDef } from "../components/table/DataTable";
import Pagination from "../components/table/Pagination";
import { useFilterStore } from "../stores/filterStore";
import type { Track } from "../types";
import { fmtMs, relDate } from "../utils/format";

function go(path: string) {
  window.history.pushState(null, "", path);
  window.dispatchEvent(new PopStateEvent("popstate"));
}

const COVER_COLORS = [
  "oklch(0.7 0.12 280)",
  "oklch(0.7 0.12 160)",
  "oklch(0.7 0.12 40)",
  "oklch(0.7 0.12 220)",
  "oklch(0.7 0.12 320)",
];

function CoverPlaceholder({ name, size = 80 }: { name: string; size?: number }) {
  const idx = name.charCodeAt(0) % COVER_COLORS.length;
  return (
    <div
      style={{
        width: size,
        height: size,
        borderRadius: 6,
        background: `linear-gradient(135deg, ${COVER_COLORS[idx]}, var(--bg-3))`,
        border: "1px solid var(--hair)",
        display: "grid",
        placeItems: "center",
        color: "white",
        fontFamily: "var(--font-mono)",
        fontSize: size * 0.35,
        flexShrink: 0,
      }}
    >
      {name.slice(0, 1).toUpperCase()}
    </div>
  );
}

type TrackRow = Track & { num: number };

const columns: ColumnDef<TrackRow>[] = [
  {
    id: "num",
    header: "#",
    numeric: true,
    cell: ({ row }) => (
      <span style={{ fontFamily: "var(--font-mono)", fontSize: 11.5, color: "var(--fg-3)" }}>
        {row.original.num}
      </span>
    ),
  },
  {
    id: "name",
    header: "Track",
    accessorKey: "name",
    cell: ({ row }) => (
      <div style={{ display: "flex", alignItems: "center", gap: 8, minWidth: 0 }}>
        {row.original.album?.image_url ? (
          <img
            src={row.original.album.image_url}
            alt=""
            style={{ width: 22, height: 22, borderRadius: 3, objectFit: "cover", flexShrink: 0 }}
          />
        ) : (
          <div
            style={{
              width: 22,
              height: 22,
              borderRadius: 3,
              background: "var(--bg-2)",
              flexShrink: 0,
            }}
          />
        )}
        <div style={{ minWidth: 0 }}>
          <div
            style={{
              fontWeight: 500,
              color: "var(--fg)",
              overflow: "hidden",
              textOverflow: "ellipsis",
            }}
          >
            {row.original.name}
          </div>
          <div
            style={{ fontSize: 11, color: "var(--fg-2)", overflow: "hidden", textOverflow: "ellipsis" }}
          >
            {row.original.artists.map((a) => a.name).join(", ")}
          </div>
        </div>
      </div>
    ),
  },
  {
    id: "album",
    header: "Album",
    accessorFn: (r) => r.album?.name ?? "",
    cell: ({ row }) => (
      <span style={{ color: "var(--fg-2)", fontSize: 12 }}>{row.original.album?.name ?? "--"}</span>
    ),
  },
  {
    id: "duration",
    header: "Length",
    accessorKey: "duration_ms",
    numeric: true,
    cell: ({ row }) => (
      <span style={{ fontFamily: "var(--font-mono)", fontSize: 11.5, color: "var(--fg-1)" }}>
        {fmtMs(row.original.duration_ms)}
      </span>
    ),
  },
  {
    id: "added",
    header: "Added",
    accessorKey: "saved_at",
    cell: ({ row }) => (
      <span style={{ fontFamily: "var(--font-mono)", fontSize: 11.5, color: "var(--fg-2)" }}>
        {row.original.saved_at ? relDate(row.original.saved_at) : "--"}
      </span>
    ),
  },
];

export default function PlaylistDetailPage() {
  const { id } = useParams<{ id: string }>();
  const location = useLocation();
  const nameFromState = (location.state as Record<string, string> | null)?.name;
  const { page, pageSize, setPage } = useFilterStore();

  const { data: playlists = [] } = useQuery({
    queryKey: ["playlists"],
    queryFn: getPlaylists,
  });

  const playlist = playlists.find((p) => p.spotify_id === id);
  const displayName = playlist?.name ?? nameFromState ?? id ?? "Playlist";

  const { data: tracksData, isLoading } = useQuery({
    queryKey: ["playlist-tracks", id, page, pageSize],
    queryFn: () => getPlaylistTracks(id ?? "", { page, page_size: pageSize }),
    enabled: !!id,
  });

  const rows: TrackRow[] = (tracksData?.items ?? []).map((t, i) => ({
    ...t,
    num: (page - 1) * pageSize + i + 1,
  }));

  return (
    <>
      {/* Back + header */}
      <div
        style={{
          borderBottom: "1px solid var(--hair)",
          padding: "14px 20px",
          background: "var(--bg)",
          position: "sticky",
          top: 0,
          zIndex: 2,
        }}
      >
        <button
          onClick={() => { go("/playlists"); }}
          style={{
            display: "inline-flex",
            alignItems: "center",
            gap: 4,
            fontSize: 11,
            color: "var(--fg-3)",
            cursor: "pointer",
            marginBottom: 10,
          }}
          onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.color = "var(--fg-1)"; }}
          onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.color = "var(--fg-3)"; }}
        >
          <ChevronLeft size={13} /> Playlists
        </button>
        <div style={{ display: "flex", alignItems: "flex-start", gap: 14 }}>
          {playlist?.image_url ? (
            <img
              src={playlist.image_url}
              alt=""
              style={{ width: 56, height: 56, borderRadius: 5, objectFit: "cover", flexShrink: 0 }}
            />
          ) : (
            <CoverPlaceholder name={displayName} size={56} />
          )}
          <div>
            <h1
              style={{
                margin: 0,
                fontSize: 18,
                fontWeight: 600,
                letterSpacing: "-0.015em",
                fontFamily: "var(--font-ui)",
              }}
            >
              {displayName}
            </h1>
            <div style={{ marginTop: 4, fontSize: 12, color: "var(--fg-2)", display: "flex", gap: 12 }}>
              {playlist?.owner_id && <span>by {playlist.owner_id}</span>}
              <span style={{ fontFamily: "var(--font-mono)" }}>
                {tracksData?.total.toLocaleString() ?? playlist?.track_count?.toLocaleString() ?? "…"} tracks
              </span>
              {playlist?.is_public !== undefined && (
                <span style={{ color: playlist.is_public ? "var(--ok)" : "var(--fg-3)" }}>
                  {playlist.is_public ? "Public" : "Private"}
                </span>
              )}
            </div>
            {playlist?.description && (
              <div
                style={{
                  marginTop: 4,
                  fontSize: 11,
                  color: "var(--fg-3)",
                  maxWidth: 480,
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                  whiteSpace: "nowrap",
                }}
              >
                {playlist.description}
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Track table */}
      <div style={{ overflow: "auto" }}>
        <DataTable
          data={rows as unknown as TrackRow[]}
          columns={columns as unknown as ColumnDef<TrackRow>[]}
          isLoading={isLoading}
        />
        <Pagination
          page={page}
          pageSize={pageSize}
          total={tracksData?.total ?? 0}
          onPageChange={setPage}
        />
      </div>
    </>
  );
}
