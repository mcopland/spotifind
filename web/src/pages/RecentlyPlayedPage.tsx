import { useQuery } from "@tanstack/react-query";
import { getRecentlyPlayed } from "../api/recently-played";
import DataTable, { type ColumnDef } from "../components/table/DataTable";
import Pagination from "../components/table/Pagination";
import { useFilterStore } from "../stores/filterStore";
import type { RecentlyPlayedTrack } from "../types";
import { fmtMs, relDate } from "../utils/format";

const COVER_COLORS = [
  "oklch(0.7 0.12 280)",
  "oklch(0.7 0.12 160)",
  "oklch(0.7 0.12 40)",
  "oklch(0.7 0.12 220)",
  "oklch(0.7 0.12 320)",
];

function CoverPlaceholder({ name, size = 22 }: { name: string; size?: number }) {
  const idx = name.charCodeAt(0) % COVER_COLORS.length;
  return (
    <div
      style={{
        width: size,
        height: size,
        borderRadius: 3,
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

const columns: ColumnDef<RecentlyPlayedTrack>[] = [
  {
    id: "played_at",
    header: "When",
    accessorKey: "played_at",
    cell: ({ row }) => (
      <span style={{ fontFamily: "var(--font-mono)", fontSize: 11.5, color: "var(--fg-2)" }}>
        {relDate(row.original.played_at)}
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
          <CoverPlaceholder name={row.original.name} />
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
            style={{
              fontSize: 11,
              color: "var(--fg-2)",
              overflow: "hidden",
              textOverflow: "ellipsis",
            }}
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
];

export default function RecentlyPlayedPage() {
  const { page, pageSize, setPage } = useFilterStore();

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ["recently-played", { page, pageSize }],
    queryFn: () => getRecentlyPlayed({ page, page_size: pageSize }),
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
        <span style={{ color: "var(--err)" }}>Failed to load history.</span>
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
          History
          <span
            style={{
              fontFamily: "var(--font-mono)",
              color: "var(--fg-2)",
              fontSize: 13,
              fontWeight: 400,
            }}
          >
            · {data?.total.toLocaleString() ?? "…"} plays
          </span>
        </h1>
        <div style={{ marginTop: 2, fontSize: 12, color: "var(--fg-2)" }}>
          Your recently played tracks from Spotify.
        </div>
      </div>

      {/* Table */}
      <div style={{ overflow: "auto" }}>
        <DataTable
          data={data?.items ?? []}
          columns={columns}
          isLoading={isLoading}
        />
        <Pagination
          page={page}
          pageSize={pageSize}
          total={data?.total ?? 0}
          onPageChange={setPage}
        />
      </div>
    </>
  );
}
