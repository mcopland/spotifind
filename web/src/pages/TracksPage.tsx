import { useQuery } from "@tanstack/react-query";
import { Plus, Search } from "lucide-react";
import { useRef, useState } from "react";
import { getTracks } from "../api/tracks";
import { getGenres } from "../api/sync";
import Badge from "../components/primitives/Badge";
import FilterChip from "../components/primitives/FilterChip";
import { PopoverGroup, PopoverOption } from "../components/primitives/Popover";
import DataTable, { type ColumnDef, type SortingState } from "../components/table/DataTable";
import Pagination from "../components/table/Pagination";
import { useFilterStore } from "../stores/filterStore";
import type { Track } from "../types";
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

function EnergyBar({ value }: { value: number }) {
  return (
    <div style={{ display: "inline-flex", alignItems: "center", gap: 6 }}>
      <div
        style={{
          width: 42,
          height: 3,
          background: "var(--bg-2)",
          borderRadius: 2,
          overflow: "hidden",
        }}
      >
        <div
          style={{ width: `${String(value)}%`, height: "100%", background: "var(--acc)" }}
        />
      </div>
      <span style={{ fontFamily: "var(--font-mono)", fontSize: 11.5, color: "var(--fg-1)" }}>
        {value}
      </span>
    </div>
  );
}

function buildColumns(go: (path: string, state?: Record<string, string>) => void): ColumnDef<Track>[] {
  return [
    {
      id: "num",
      header: "#",
      numeric: true,
      cell: ({ row }) => {
        // row index is not available directly; use a counter
        return (
          <span
            style={{ fontFamily: "var(--font-mono)", fontSize: 11.5, color: "var(--fg-3)" }}
          >
            {(row as { original: Track; index?: number }).index ?? ""}
          </span>
        );
      },
    },
    {
      id: "name",
      header: "Track",
      accessorKey: "name",
      enableSorting: true,
      cell: ({ row }) => (
        <div style={{ display: "flex", alignItems: "center", gap: 8, minWidth: 0 }}>
          <CoverPlaceholder name={row.original.name} />
          <div style={{ minWidth: 0 }}>
            <div
              style={{
                color: "var(--fg)",
                fontWeight: 500,
                overflow: "hidden",
                textOverflow: "ellipsis",
              }}
            >
              {row.original.name}
              {row.original.explicit && (
                <Badge style={{ marginLeft: 6, fontSize: 9, height: 14, padding: "0 4px" }}>
                  E
                </Badge>
              )}
            </div>
            <div
              style={{
                color: "var(--fg-2)",
                fontSize: 11,
                overflow: "hidden",
                textOverflow: "ellipsis",
              }}
            >
              {row.original.artists.map((a, i) => (
                <span key={a.spotify_id}>
                  {i > 0 && ", "}
                  <span
                    onClick={() => { go(`/artists/${a.spotify_id}`, { name: a.name }); }}
                    style={{
                      cursor: "pointer",
                      borderBottom: "1px solid transparent",
                    }}
                    onMouseEnter={(e) => {
                      (e.currentTarget as HTMLElement).style.color = "var(--acc-ink)";
                      (e.currentTarget as HTMLElement).style.borderBottomColor = "var(--acc)";
                    }}
                    onMouseLeave={(e) => {
                      (e.currentTarget as HTMLElement).style.color = "";
                      (e.currentTarget as HTMLElement).style.borderBottomColor = "transparent";
                    }}
                  >
                    {a.name}
                  </span>
                </span>
              ))}
            </div>
          </div>
        </div>
      ),
    },
    {
      id: "album",
      header: "Album",
      accessorFn: (r) => r.album?.name ?? "",
      enableSorting: true,
      cell: ({ row }) => (
        <span style={{ display: "flex", alignItems: "center", gap: 4 }}>
          <span
            onClick={() => {
              if (row.original.album)
                go(`/albums/${row.original.album.spotify_id}`, { name: row.original.album.name });
            }}
            style={{
              color: "var(--fg-2)",
              cursor: "pointer",
              borderBottom: "1px solid transparent",
            }}
            onMouseEnter={(e) => {
              (e.currentTarget as HTMLElement).style.color = "var(--acc-ink)";
              (e.currentTarget as HTMLElement).style.borderBottomColor = "var(--acc)";
            }}
            onMouseLeave={(e) => {
              (e.currentTarget as HTMLElement).style.color = "var(--fg-2)";
              (e.currentTarget as HTMLElement).style.borderBottomColor = "transparent";
            }}
          >
            {row.original.album?.name}
          </span>
          {row.original.album?.release_year && (
            <span style={{ color: "var(--fg-3)", fontSize: 11 }}>
              · {row.original.album.release_year}
            </span>
          )}
        </span>
      ),
    },
    {
      id: "genre",
      header: "Genre",
      cell: ({ row }) => {
        const genre = row.original.artists[0]?.genres?.[0];
        if (!genre) return <span style={{ color: "var(--fg-3)" }}>--</span>;
        return <Badge>{genre}</Badge>;
      },
    },
    {
      id: "bpm",
      header: "BPM",
      numeric: true,
      cell: () => (
        <span style={{ fontFamily: "var(--font-mono)", fontSize: 11.5, color: "var(--fg-3)" }}>
          --
        </span>
      ),
    },
    {
      id: "key",
      header: "Key",
      cell: () => (
        <span style={{ fontFamily: "var(--font-mono)", fontSize: 10.5, color: "var(--fg-3)" }}>
          --
        </span>
      ),
    },
    {
      id: "energy",
      header: "Energy",
      numeric: true,
      cell: ({ row }) => <EnergyBar value={row.original.popularity} />,
    },
    {
      id: "duration",
      header: "Length",
      accessorKey: "duration_ms",
      enableSorting: true,
      numeric: true,
      cell: ({ row }) => (
        <span style={{ fontFamily: "var(--font-mono)", fontSize: 11.5, color: "var(--fg-1)" }}>
          {fmtMs(row.original.duration_ms)}
        </span>
      ),
    },
    {
      id: "plays",
      header: "Plays",
      numeric: true,
      cell: () => (
        <span style={{ fontFamily: "var(--font-mono)", fontSize: 11.5, color: "var(--fg-3)" }}>
          --
        </span>
      ),
    },
    {
      id: "added",
      header: "Added",
      accessorKey: "saved_at",
      enableSorting: true,
      cell: ({ row }) => (
        <span style={{ fontFamily: "var(--font-mono)", fontSize: 11.5, color: "var(--fg-2)" }}>
          {row.original.saved_at ? relDate(row.original.saved_at) : "--"}
        </span>
      ),
    },
    {
      id: "source",
      header: "Source",
      cell: () => (
        <Badge variant="info">library</Badge>
      ),
    },
  ];
}

export default function TracksPage() {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [localSearch, setLocalSearch] = useState(() => useFilterStore.getState().tracksSearch);
  const searchRef = useRef<ReturnType<typeof setTimeout>>(null);

  const {
    tracksSearch,
    setTracksSearch,
    genres,
    setGenres,
    yearMin,
    setYearMin,
    yearMax,
    setYearMax,
    popularityMin,
    setPopularityMin,
    popularityMax,
    setPopularityMax,
    explicit,
    setExplicit,
    page,
    pageSize,
    setPage,
  } = useFilterStore();

  const { data: genreOptions = [] } = useQuery({
    queryKey: ["genres"],
    queryFn: getGenres,
  });

  const sortBy = sorting[0]?.id === "name"
    ? "name"
    : sorting[0]?.id === "album"
      ? "album"
      : sorting[0]?.id === "duration"
        ? "duration"
        : sorting[0]?.id === "added"
          ? "saved_at"
          : undefined;
  const sortDir = sorting[0] ? (sorting[0].desc ? "desc" : "asc") : undefined;

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: [
      "tracks",
      { tracksSearch, genres, yearMin, yearMax, popularityMin, popularityMax, explicit, page, pageSize, sortBy, sortDir },
    ],
    queryFn: () =>
      getTracks({
        search: tracksSearch,
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

  function handleSearch(e: React.ChangeEvent<HTMLInputElement>) {
    const val = e.target.value;
    setLocalSearch(val);
    if (searchRef.current) clearTimeout(searchRef.current);
    searchRef.current = setTimeout(() => { setTracksSearch(val); }, 300);
  }

  // navigation helper using window.history (avoids prop drilling)
  function go(path: string, state?: Record<string, string>) {
    window.history.pushState(state ?? null, "", path);
    window.dispatchEvent(new PopStateEvent("popstate"));
  }

  const columns = buildColumns(go);

  // Inject row index into data for the # column
  const rows = (data?.items ?? []).map((r, i) => ({
    ...r,
    index: (page - 1) * pageSize + i + 1,
  })) as (Track & { index: number })[];

  const hasFilters =
    genres.length > 0 ||
    yearMin !== undefined ||
    yearMax !== undefined ||
    popularityMin !== undefined ||
    popularityMax !== undefined ||
    explicit !== undefined;

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
        <span style={{ color: "var(--err)" }}>Failed to load tracks.</span>
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
        <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
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
            Library
            <span
              style={{
                fontFamily: "var(--font-mono)",
                color: "var(--fg-2)",
                fontSize: 13,
                fontWeight: 400,
              }}
            >
              · {data?.total.toLocaleString() ?? "…"} tracks
            </span>
          </h1>
          <div style={{ flex: 1 }} />
          <button
            style={{
              display: "inline-flex",
              alignItems: "center",
              gap: 6,
              padding: "4px 9px",
              height: 26,
              border: "1px solid var(--hair)",
              borderRadius: "var(--radius-sm)",
              background: "var(--bg)",
              color: "var(--fg-1)",
              fontSize: 12,
            }}
          >
            <Plus size={11} /> New from query
          </button>
        </div>
        <div style={{ marginTop: 2, fontSize: 12, color: "var(--fg-2)" }}>
          Every track you've saved, added to a playlist, or recently played.
        </div>
      </div>

      {/* Toolbar */}
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: 6,
          flexWrap: "wrap",
          padding: "8px 20px",
          borderBottom: "1px solid var(--hair)",
          background: "var(--bg-1)",
        }}
      >
        {/* Search */}
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 6,
            height: 26,
            padding: "0 8px",
            border: "1px solid var(--hair)",
            borderRadius: "var(--radius-sm)",
            background: "var(--bg)",
            minWidth: 260,
          }}
        >
          <Search size={12} style={{ color: "var(--fg-3)" }} />
          <input
            placeholder="Search tracks, artists, albums…"
            value={localSearch}
            onChange={handleSearch}
            style={{
              border: "none",
              background: "none",
              outline: "none",
              width: "100%",
              fontSize: 12,
              color: "var(--fg)",
            }}
          />
        </div>

        <div style={{ width: 1, height: 18, background: "var(--hair)", margin: "0 4px" }} />

        {/* Genre chip */}
        <FilterChip
          label="Genre"
          applied={genres.length > 0}
          value={genres.length > 0 ? genres[0] : undefined}
          onRemove={() => { setGenres([]); }}
        >
          <PopoverGroup>Genre</PopoverGroup>
          {genreOptions.map((g) => (
            <PopoverOption
              key={g}
              onClick={() => {
                setGenres(genres.includes(g) ? genres.filter((x) => x !== g) : [...genres, g]);
              }}
            >
              {genres.includes(g) ? "✓ " : ""}{g}
            </PopoverOption>
          ))}
        </FilterChip>

        {/* Year chip */}
        <FilterChip
          label="Year"
          applied={yearMin !== undefined || yearMax !== undefined}
          value={
            yearMin !== undefined && yearMax !== undefined
              ? `${String(yearMin)}–${String(yearMax)}`
              : yearMin !== undefined
                ? `≥ ${String(yearMin)}`
                : yearMax !== undefined
                  ? `≤ ${String(yearMax)}`
                  : undefined
          }
          onRemove={() => { setYearMin(undefined); setYearMax(undefined); }}
        >
          <PopoverGroup>Release year</PopoverGroup>
          <div style={{ padding: "6px 8px", display: "flex", gap: 8 }}>
            <input
              type="number"
              placeholder="From"
              defaultValue={yearMin}
              onChange={(e) => { setYearMin(e.target.value ? Number(e.target.value) : undefined); }}
              style={{
                width: 80,
                height: 28,
                padding: "0 8px",
                border: "1px solid var(--hair)",
                borderRadius: "var(--radius-sm)",
                background: "var(--bg)",
                color: "var(--fg)",
                fontSize: 12,
              }}
            />
            <input
              type="number"
              placeholder="To"
              defaultValue={yearMax}
              onChange={(e) => { setYearMax(e.target.value ? Number(e.target.value) : undefined); }}
              style={{
                width: 80,
                height: 28,
                padding: "0 8px",
                border: "1px solid var(--hair)",
                borderRadius: "var(--radius-sm)",
                background: "var(--bg)",
                color: "var(--fg)",
                fontSize: 12,
              }}
            />
          </div>
        </FilterChip>

        {/* Popularity chip */}
        <FilterChip
          label="Popularity"
          applied={popularityMin !== undefined || popularityMax !== undefined}
          value={
            popularityMin !== undefined || popularityMax !== undefined
              ? `${String(popularityMin ?? 0)}–${String(popularityMax ?? 100)}`
              : undefined
          }
          onRemove={() => { setPopularityMin(undefined); setPopularityMax(undefined); }}
        >
          <PopoverGroup>Popularity</PopoverGroup>
          {[
            { label: "High (≥ 70)", min: 70, max: undefined },
            { label: "Medium (40–70)", min: 40, max: 70 },
            { label: "Low (< 40)", min: undefined, max: 40 },
          ].map((opt) => (
            <PopoverOption
              key={opt.label}
              onClick={() => {
                setPopularityMin(opt.min);
                setPopularityMax(opt.max);
              }}
            >
              {opt.label}
            </PopoverOption>
          ))}
        </FilterChip>

        {/* Explicit chip */}
        <FilterChip
          label="Explicit"
          applied={explicit !== undefined}
          value={explicit !== undefined ? String(explicit) : undefined}
          onRemove={() => { setExplicit(undefined); }}
        >
          <PopoverGroup>Explicit</PopoverGroup>
          <PopoverOption onClick={() => { setExplicit(true); }}>Yes</PopoverOption>
          <PopoverOption onClick={() => { setExplicit(false); }}>No</PopoverOption>
        </FilterChip>

        {hasFilters && (
          <button
            onClick={() => {
              setGenres([]);
              setYearMin(undefined);
              setYearMax(undefined);
              setPopularityMin(undefined);
              setPopularityMax(undefined);
              setExplicit(undefined);
              setLocalSearch("");
              setTracksSearch("");
            }}
            style={{
              fontSize: 11.5,
              color: "var(--fg-3)",
              cursor: "pointer",
              textDecoration: "underline",
            }}
          >
            Clear filters
          </button>
        )}
      </div>

      {/* Table */}
      <div style={{ overflow: "auto" }}>
        <DataTable
          data={rows as unknown as Track[]}
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
    </>
  );
}
