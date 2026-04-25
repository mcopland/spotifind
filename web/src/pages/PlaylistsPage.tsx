import { useQuery } from "@tanstack/react-query";
import { getPlaylists } from "../api/playlists";
import Badge from "../components/primitives/Badge";
import DataTable, { type ColumnDef } from "../components/table/DataTable";
import type { Playlist } from "../types";

function go(path: string, state?: Record<string, string>) {
  window.history.pushState(state ?? null, "", path);
  window.dispatchEvent(new PopStateEvent("popstate"));
}

const COVER_COLORS = [
  "oklch(0.7 0.12 280)",
  "oklch(0.7 0.12 160)",
  "oklch(0.7 0.12 40)",
  "oklch(0.7 0.12 220)",
  "oklch(0.7 0.12 320)",
];

function CoverPlaceholder({ name, size = 24 }: { name: string; size?: number }) {
  const idx = name.charCodeAt(0) % COVER_COLORS.length;
  return (
    <div
      style={{
        width: size,
        height: size,
        borderRadius: 4,
        background: `linear-gradient(135deg, ${COVER_COLORS[idx]}, var(--bg-3))`,
        border: "1px solid var(--hair)",
        display: "inline-grid",
        placeItems: "center",
        color: "white",
        fontFamily: "var(--font-mono)",
        fontSize: size * 0.4,
        flexShrink: 0,
      }}
    >
      {name.slice(0, 1).toUpperCase()}
    </div>
  );
}

const columns: ColumnDef<Playlist>[] = [
  {
    id: "name",
    header: "Playlist",
    accessorKey: "name",
    cell: ({ row }) => (
      <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
        {row.original.image_url ? (
          <img
            src={row.original.image_url}
            alt=""
            style={{ width: 24, height: 24, borderRadius: 3, objectFit: "cover", flexShrink: 0 }}
          />
        ) : (
          <CoverPlaceholder name={row.original.name} />
        )}
        <span
          onClick={() => { go(`/playlists/${row.original.spotify_id}`, { name: row.original.name }); }}
          style={{
            fontWeight: 500,
            color: "var(--fg)",
            cursor: "pointer",
            borderBottom: "1px solid transparent",
          }}
          onMouseEnter={(e) => {
            (e.currentTarget as HTMLElement).style.color = "var(--acc-ink)";
            (e.currentTarget as HTMLElement).style.borderBottomColor = "var(--acc)";
          }}
          onMouseLeave={(e) => {
            (e.currentTarget as HTMLElement).style.color = "var(--fg)";
            (e.currentTarget as HTMLElement).style.borderBottomColor = "transparent";
          }}
        >
          {row.original.name}
        </span>
      </div>
    ),
  },
  {
    id: "owner",
    header: "Owner",
    accessorKey: "owner_id",
    cell: ({ getValue }) => (
      <span style={{ fontFamily: "var(--font-mono)", fontSize: 11.5, color: "var(--fg-2)" }}>
        {getValue() as string}
      </span>
    ),
  },
  {
    id: "track_count",
    header: "Tracks",
    accessorKey: "track_count",
    numeric: true,
    cell: ({ getValue }) => (
      <span style={{ fontFamily: "var(--font-mono)", fontSize: 11.5, color: "var(--fg-1)" }}>
        {(getValue() as number | undefined)?.toLocaleString() ?? "--"}
      </span>
    ),
  },
  {
    id: "is_public",
    header: "Visibility",
    accessorKey: "is_public",
    cell: ({ row }) => (
      <Badge variant={row.original.is_public ? "ok" : "default"}>
        {row.original.is_public ? "Public" : "Private"}
      </Badge>
    ),
  },
  {
    id: "collaborative",
    header: "Type",
    accessorKey: "collaborative",
    cell: ({ getValue }) =>
      getValue() ? <Badge variant="info">Collab</Badge> : null,
  },
];

export default function PlaylistsPage() {
  const { data: playlists = [], isLoading, isError, refetch } = useQuery({
    queryKey: ["playlists"],
    queryFn: getPlaylists,
  });

  if (isError) {
    return (
      <div
        style={{
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          height: 240,
          gap: 12,
        }}
      >
        <span style={{ color: "var(--err)" }}>Failed to load playlists.</span>
        <button
          onClick={() => void refetch()}
          style={{
            padding: "4px 12px",
            background: "var(--acc)",
            color: "black",
            borderRadius: "var(--radius-sm)",
            fontSize: 12,
            fontWeight: 500,
          }}
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <>
      {/* Page header */}
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
        <h1
          style={{
            margin: 0,
            fontSize: 18,
            fontWeight: 600,
            letterSpacing: "-0.015em",
            display: "flex",
            alignItems: "center",
            gap: 8,
            fontFamily: "var(--font-ui)",
          }}
        >
          Playlists
          <span
            style={{
              fontFamily: "var(--font-mono)",
              color: "var(--fg-2)",
              fontSize: 13,
              fontWeight: 400,
            }}
          >
            · {playlists.length.toLocaleString()}
          </span>
        </h1>
        <div style={{ marginTop: 2, fontSize: 12, color: "var(--fg-2)" }}>
          Your Spotify playlists and collaborative collections.
        </div>
      </div>

      {/* Table */}
      <div style={{ overflow: "auto" }}>
        <DataTable data={playlists} columns={columns} isLoading={isLoading} />
      </div>
    </>
  );
}
