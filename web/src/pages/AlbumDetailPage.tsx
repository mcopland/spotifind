import { useQuery } from "@tanstack/react-query";
import { ChevronLeft } from "lucide-react";
import { useLocation, useParams } from "react-router-dom";
import { getTracks } from "../api/tracks";
import DataTable, { type ColumnDef } from "../components/table/DataTable";
import Pagination from "../components/table/Pagination";
import { useFilterStore } from "../stores/filterStore";
import type { Track } from "../types";
import { fmtMs } from "../utils/format";

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

function CoverPlaceholder({ name, size = 56 }: { name: string; size?: number }) {
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

const trackColumns: ColumnDef<Track>[] = [
  {
    id: "track_number",
    header: "#",
    accessorKey: "track_number",
    numeric: true,
    cell: ({ getValue }) => (
      <span style={{ fontFamily: "var(--font-mono)", fontSize: 11.5, color: "var(--fg-3)" }}>
        {getValue() as number}
      </span>
    ),
  },
  {
    id: "name",
    header: "Track",
    accessorKey: "name",
    cell: ({ row }) => (
      <div style={{ minWidth: 0 }}>
        <div style={{ fontWeight: 500, color: "var(--fg)", overflow: "hidden", textOverflow: "ellipsis" }}>
          {row.original.name}
        </div>
        <div style={{ fontSize: 11, color: "var(--fg-2)", overflow: "hidden", textOverflow: "ellipsis" }}>
          {row.original.artists.map((a) => a.name).join(", ")}
        </div>
      </div>
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
    id: "popularity",
    header: "Popularity",
    accessorKey: "popularity",
    numeric: true,
    cell: ({ row }) => (
      <span style={{ fontFamily: "var(--font-mono)", fontSize: 11.5, color: "var(--fg-1)" }}>
        {row.original.popularity}
      </span>
    ),
  },
];

export default function AlbumDetailPage() {
  const { id } = useParams<{ id: string }>();
  const location = useLocation();
  const nameFromState = (location.state as Record<string, string> | null)?.name;
  const { page, pageSize, setPage } = useFilterStore();

  // No direct /api/albums/:id endpoint — search by album name as workaround
  const { data: tracksData, isLoading } = useQuery({
    queryKey: ["album-tracks", nameFromState, page, pageSize],
    queryFn: () => getTracks({ search: nameFromState, page, page_size: pageSize }),
    enabled: !!nameFromState,
  });

  // Get album metadata from the first track
  const albumMeta = tracksData?.items[0]?.album;
  const displayName = albumMeta?.name ?? nameFromState ?? id ?? "Album";

  if (!nameFromState) {
    return (
      <div
        style={{
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          height: 240,
          gap: 8,
          color: "var(--fg-3)",
          fontSize: 13,
        }}
      >
        <span>Navigate to an album from the Library view.</span>
        <button
          onClick={() => { go("/library"); }}
          style={{
            marginTop: 4,
            padding: "4px 12px",
            background: "var(--acc)",
            color: "black",
            borderRadius: "var(--radius-sm)",
            fontSize: 12,
            fontWeight: 500,
            cursor: "pointer",
          }}
        >
          Go to Library
        </button>
      </div>
    );
  }

  return (
    <>
      {/* Back + hero */}
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
          onClick={() => { go("/library"); }}
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
          <ChevronLeft size={13} /> Library
        </button>
        <div style={{ display: "flex", alignItems: "flex-start", gap: 14 }}>
          {albumMeta?.image_url ? (
            <img
              src={albumMeta.image_url}
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
              {albumMeta?.artists && albumMeta.artists.length > 0 && (
                <span>{albumMeta.artists.map((a) => a.name).join(", ")}</span>
              )}
              {albumMeta?.release_year && (
                <span style={{ fontFamily: "var(--font-mono)" }}>{albumMeta.release_year}</span>
              )}
              {albumMeta?.total_tracks && (
                <span style={{ fontFamily: "var(--font-mono)" }}>
                  {albumMeta.total_tracks} tracks
                </span>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Track table */}
      <div style={{ overflow: "auto" }}>
        <DataTable
          data={tracksData?.items ?? []}
          columns={trackColumns}
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
