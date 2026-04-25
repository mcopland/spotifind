import { useQuery } from "@tanstack/react-query";
import { ChevronLeft } from "lucide-react";
import { useLocation, useParams } from "react-router-dom";
import { getArtists } from "../api/artists";
import { getTracks } from "../api/tracks";
import Badge from "../components/primitives/Badge";
import DataTable, { type ColumnDef } from "../components/table/DataTable";
import Pagination from "../components/table/Pagination";
import { useFilterStore } from "../stores/filterStore";
import type { Track } from "../types";
import { fmtMs, fmtK } from "../utils/format";

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

function AvatarPlaceholder({ name, size = 80 }: { name: string; size?: number }) {
  const idx = name.charCodeAt(0) % COVER_COLORS.length;
  return (
    <div
      style={{
        width: size,
        height: size,
        borderRadius: "50%",
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

function EnergyBar({ value }: { value: number }) {
  return (
    <div style={{ display: "inline-flex", alignItems: "center", gap: 6 }}>
      <div style={{ width: 42, height: 3, background: "var(--bg-2)", borderRadius: 2, overflow: "hidden" }}>
        <div style={{ width: `${String(value)}%`, height: "100%", background: "var(--acc)" }} />
      </div>
      <span style={{ fontFamily: "var(--font-mono)", fontSize: 11.5, color: "var(--fg-1)" }}>{value}</span>
    </div>
  );
}

const trackColumns: ColumnDef<Track>[] = [
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
          <div style={{ width: 22, height: 22, borderRadius: 3, background: "var(--bg-2)", flexShrink: 0 }} />
        )}
        <span style={{ fontWeight: 500, color: "var(--fg)", overflow: "hidden", textOverflow: "ellipsis" }}>
          {row.original.name}
        </span>
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
    id: "popularity",
    header: "Popularity",
    accessorKey: "popularity",
    numeric: true,
    cell: ({ row }) => <EnergyBar value={row.original.popularity} />,
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

export default function ArtistDetailPage() {
  const { id } = useParams<{ id: string }>();
  const location = useLocation();
  const nameFromState = (location.state as Record<string, string> | null)?.name;
  const { page, pageSize, setPage } = useFilterStore();

  // Look up the artist by searching for their name (no direct /api/artists/:id endpoint)
  const { data: artistData } = useQuery({
    queryKey: ["artist-lookup", nameFromState],
    queryFn: () => getArtists({ search: nameFromState, page_size: 1 }),
    enabled: !!nameFromState,
  });

  const artist = artistData?.items[0];
  const displayName = artist?.name ?? nameFromState ?? id ?? "Artist";

  const { data: tracksData, isLoading } = useQuery({
    queryKey: ["artist-tracks", nameFromState, page, pageSize],
    queryFn: () => getTracks({ search: nameFromState, page, page_size: pageSize }),
    enabled: !!nameFromState,
  });

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
        <span>Navigate to an artist from the Library view.</span>
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
          {artist?.image_url ? (
            <img
              src={artist.image_url}
              alt=""
              style={{ width: 56, height: 56, borderRadius: "50%", objectFit: "cover", flexShrink: 0 }}
            />
          ) : (
            <AvatarPlaceholder name={displayName} size={56} />
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
            <div style={{ marginTop: 4, display: "flex", flexWrap: "wrap", gap: 6 }}>
              {artist?.genres?.slice(0, 4).map((g) => (
                <Badge key={g}>{g}</Badge>
              ))}
            </div>
            <div style={{ marginTop: 6, fontSize: 12, color: "var(--fg-2)", display: "flex", gap: 14 }}>
              {artist?.followers != null && (
                <span>
                  <span style={{ fontFamily: "var(--font-mono)" }}>{fmtK(artist.followers)}</span> followers
                </span>
              )}
              {artist?.popularity != null && (
                <span>
                  <span style={{ fontFamily: "var(--font-mono)" }}>{artist.popularity}</span> popularity
                </span>
              )}
              {tracksData && (
                <span>
                  <span style={{ fontFamily: "var(--font-mono)" }}>{tracksData.total.toLocaleString()}</span> tracks in library
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
