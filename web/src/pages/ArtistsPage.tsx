import { useQuery } from "@tanstack/react-query";
import { Search } from "lucide-react";
import { useRef, useState } from "react";
import { getArtists } from "../api/artists";
import { getGenres } from "../api/sync";
import Badge from "../components/primitives/Badge";
import FilterChip from "../components/primitives/FilterChip";
import { PopoverGroup, PopoverOption } from "../components/primitives/Popover";
import DataTable, { type ColumnDef, type SortingState } from "../components/table/DataTable";
import Pagination from "../components/table/Pagination";
import { useFilterStore } from "../stores/filterStore";
import type { Artist } from "../types";
import { fmtK } from "../utils/format";

const COVER_COLORS = [
  "oklch(0.7 0.12 280)",
  "oklch(0.7 0.12 160)",
  "oklch(0.7 0.12 40)",
  "oklch(0.7 0.12 220)",
  "oklch(0.7 0.12 320)",
];

function AvatarPlaceholder({ name, size = 24 }: { name: string; size?: number }) {
  const idx = name.charCodeAt(0) % COVER_COLORS.length;
  return (
    <div
      style={{
        width: size,
        height: size,
        borderRadius: "50%",
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
        <div style={{ width: `${String(value)}%`, height: "100%", background: "var(--acc)" }} />
      </div>
      <span style={{ fontFamily: "var(--font-mono)", fontSize: 11.5, color: "var(--fg-1)" }}>
        {value}
      </span>
    </div>
  );
}

type ArtistRow = Artist & { index: number };

function buildColumns(): ColumnDef<ArtistRow>[] {
  return [
    {
      id: "num",
      header: "#",
      numeric: true,
      cell: ({ row }) => (
        <span style={{ fontFamily: "var(--font-mono)", fontSize: 11.5, color: "var(--fg-3)" }}>
          {row.original.index}
        </span>
      ),
    },
    {
      id: "name",
      header: "Artist",
      accessorKey: "name",
      enableSorting: true,
      cell: ({ row }) => (
        <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
          {row.original.image_url ? (
            <img
              src={row.original.image_url}
              alt=""
              style={{ width: 24, height: 24, borderRadius: "50%", objectFit: "cover", flexShrink: 0 }}
            />
          ) : (
            <AvatarPlaceholder name={row.original.name} />
          )}
          <span style={{ fontWeight: 500, color: "var(--fg)" }}>{row.original.name}</span>
        </div>
      ),
    },
    {
      id: "genre",
      header: "Genre",
      cell: ({ row }) => {
        const genre = row.original.genres?.[0];
        if (!genre) return <span style={{ color: "var(--fg-3)" }}>--</span>;
        return <Badge>{genre}</Badge>;
      },
    },
    {
      id: "popularity",
      header: "Popularity",
      accessorKey: "popularity",
      enableSorting: true,
      numeric: true,
      cell: ({ row }) => <EnergyBar value={row.original.popularity ?? 0} />,
    },
    {
      id: "followers",
      header: "Followers",
      accessorKey: "followers",
      enableSorting: true,
      numeric: true,
      cell: ({ row }) => (
        <span style={{ fontFamily: "var(--font-mono)", fontSize: 11.5, color: "var(--fg-1)" }}>
          {row.original.followers != null ? fmtK(row.original.followers) : "--"}
        </span>
      ),
    },
  ];
}

const columns = buildColumns();

export default function ArtistsPage() {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [localSearch, setLocalSearch] = useState("");
  const searchRef = useRef<ReturnType<typeof setTimeout>>(null);

  const { search, setSearch, genres, setGenres, page, pageSize, setPage } = useFilterStore();

  const { data: genreOptions = [] } = useQuery({
    queryKey: ["genres"],
    queryFn: getGenres,
  });

  const sortBy = sorting[0]?.id === "name"
    ? "name"
    : sorting[0]?.id === "popularity"
      ? "popularity"
      : sorting[0]?.id === "followers"
        ? "followers"
        : undefined;
  const sortDir = sorting[0] ? (sorting[0].desc ? "desc" : "asc") : undefined;

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ["artists", { search, genres, page, pageSize, sortBy, sortDir }],
    queryFn: () =>
      getArtists({ search, genres, page, page_size: pageSize, sort_by: sortBy, sort_dir: sortDir }),
  });

  function handleSearch(e: React.ChangeEvent<HTMLInputElement>) {
    const val = e.target.value;
    setLocalSearch(val);
    if (searchRef.current) clearTimeout(searchRef.current);
    searchRef.current = setTimeout(() => { setSearch(val); }, 300);
  }

  const rows = (data?.items ?? []).map((r, i) => ({
    ...r,
    index: (page - 1) * pageSize + i + 1,
  })) as ArtistRow[];

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
        <span style={{ color: "var(--err)" }}>Failed to load artists.</span>
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
          Artists
          <span
            style={{
              fontFamily: "var(--font-mono)",
              color: "var(--fg-2)",
              fontSize: 13,
              fontWeight: 400,
            }}
          >
            · {data?.total.toLocaleString() ?? "…"}
          </span>
        </h1>
        <div style={{ marginTop: 2, fontSize: 12, color: "var(--fg-2)" }}>
          All artists across your saved tracks and followed artists.
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
            minWidth: 240,
          }}
        >
          <Search size={12} style={{ color: "var(--fg-3)" }} />
          <input
            placeholder="Search artists…"
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

        {(genres.length > 0 || search) && (
          <button
            onClick={() => {
              setGenres([]);
              setLocalSearch("");
              setSearch("");
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
          data={rows as unknown as ArtistRow[]}
          columns={columns as unknown as ColumnDef<ArtistRow>[]}
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
