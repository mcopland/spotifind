import { useQuery } from "@tanstack/react-query";
import { Columns2, Search, SlidersHorizontal } from "lucide-react";
import { type ReactNode, type SetStateAction, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { getGenres } from "../api/sync";
import { getTracks, getTrackStats } from "../api/tracks";
import { getPlaylists } from "../api/playlists";
import Badge from "../components/primitives/Badge";
import FilterChip from "../components/primitives/FilterChip";
import Popover, { PopoverGroup, PopoverOption } from "../components/primitives/Popover";
import DataTable, { type ColumnDef, type SortingState } from "../components/table/DataTable";
import Pagination from "../components/table/Pagination";
import { useFilterStore } from "../stores/filterStore";
import type { Track } from "../types";
import { fmtMs, relDate } from "../utils/format";

const PITCH_CLASSES = ["C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"];

function EnergyBar({ value }: { value: number }) {
  const pct = Math.round(value * 100);
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
          style={{ width: `${String(pct)}%`, height: "100%", background: "var(--acc)" }}
        />
      </div>
      <span style={{ fontFamily: "var(--font-mono)", fontSize: 11.5, color: "var(--fg-1)" }}>
        {pct}
      </span>
    </div>
  );
}

function ArtistLinks({
  artists,
  navigate,
}: {
  artists: Track["artists"];
  navigate: ReturnType<typeof useNavigate>;
}) {
  return (
    <>
      {artists.map((a, i) => (
        <span key={a.spotify_id}>
          {i > 0 && ", "}
          <span
            onClick={() => { void navigate(`/artists/${a.spotify_id}`); }}
            style={{ cursor: "pointer", borderBottom: "1px solid transparent" }}
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
    </>
  );
}

// Maps column id to the API sort_by key
const COL_SORT_KEYS: Record<string, string> = {
  name: "name",
  album: "album",
  duration: "duration",
  added: "saved_at",
  popularity: "popularity",
  bpm: "tempo",
  energy: "energy",
  danceability: "danceability",
};

const SORT_KEY_TO_COL: Record<string, string> = Object.fromEntries(
  Object.entries(COL_SORT_KEYS).map(([col, key]) => [key, col]),
);

interface TrackColumnSpec {
  id: string;
  label: string;
  defaultVisible: boolean;
  colDef: ColumnDef<Track & { index: number }>;
}

function buildRegistry(
  navigate: ReturnType<typeof useNavigate>,
): TrackColumnSpec[] {
  return [
    {
      id: "num",
      label: "#",
      defaultVisible: true,
      colDef: {
        id: "num",
        header: "#",
        numeric: true,
        cell: ({ row }) => (
          <span style={{ fontFamily: "var(--font-mono)", fontSize: 11.5, color: "var(--fg-3)" }}>
            {(row.original as Track & { index: number }).index}
          </span>
        ),
      },
    },
    {
      id: "track",
      label: "Track",
      defaultVisible: true,
      colDef: {
        id: "name",
        header: "Track",
        accessorKey: "name",
        enableSorting: true,
        cell: ({ row }) => (
          <div style={{ display: "flex", alignItems: "center", gap: 8, minWidth: 0 }}>
            {row.original.album?.image_url ? (
              <img
                src={row.original.album.image_url}
                alt=""
                width={32}
                height={32}
                style={{ borderRadius: 4, objectFit: "cover", flexShrink: 0 }}
              />
            ) : (
              <div style={{ width: 32, height: 32, borderRadius: 4, background: "var(--bg-3)", flexShrink: 0 }} />
            )}
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
              </div>
            </div>
          </div>
        ),
      },
    },
    {
      id: "artist",
      label: "Artist",
      defaultVisible: true,
      colDef: {
        id: "artist",
        header: "Artist",
        cell: ({ row }) => (
          <span style={{ fontSize: 12, color: "var(--fg-2)" }}>
            <ArtistLinks artists={row.original.artists} navigate={navigate} />
          </span>
        ),
      },
    },
    {
      id: "album",
      label: "Album",
      defaultVisible: true,
      colDef: {
        id: "album",
        header: "Album",
        accessorFn: (r) => r.album?.name ?? "",
        enableSorting: true,
        cell: ({ row }) => (
          <span
            onClick={() => {
              if (row.original.album)
                void navigate(`/albums/${row.original.album.spotify_id}`);
            }}
            style={{
              color: "var(--fg-2)",
              cursor: "pointer",
              borderBottom: "1px solid transparent",
              fontSize: 12,
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
        ),
      },
    },
    {
      id: "genre",
      label: "Genre",
      defaultVisible: true,
      colDef: {
        id: "genre",
        header: "Genre",
        cell: ({ row }) => {
          const genre = row.original.artists[0]?.genres?.[0];
          if (!genre) return <span style={{ color: "var(--fg-3)" }}>--</span>;
          return <Badge>{genre}</Badge>;
        },
      },
    },
    {
      id: "length",
      label: "Length",
      defaultVisible: true,
      colDef: {
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
    },
    {
      id: "added",
      label: "Added",
      defaultVisible: true,
      colDef: {
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
    },
    {
      id: "popularity",
      label: "Popularity",
      defaultVisible: false,
      colDef: {
        id: "popularity",
        header: "Popularity",
        accessorKey: "popularity",
        enableSorting: true,
        numeric: true,
        cell: ({ row }) => <EnergyBar value={row.original.popularity / 100} />,
      },
    },
    {
      id: "explicit",
      label: "Explicit",
      defaultVisible: false,
      colDef: {
        id: "explicit",
        header: "Explicit",
        cell: ({ row }) =>
          row.original.explicit
            ? <Badge style={{ fontSize: 9, height: 14, padding: "0 4px" }}>E</Badge>
            : <span style={{ color: "var(--fg-3)" }}>--</span>,
      },
    },
    {
      id: "bpm",
      label: "BPM",
      defaultVisible: false,
      colDef: {
        id: "bpm",
        header: "BPM",
        numeric: true,
        enableSorting: true,
        cell: ({ row }) =>
          row.original.tempo != null
            ? (
              <span style={{ fontFamily: "var(--font-mono)", fontSize: 11.5, color: "var(--fg-1)" }}>
                {Math.round(row.original.tempo)}
              </span>
            )
            : <span style={{ color: "var(--fg-3)" }}>--</span>,
      },
    },
    {
      id: "key",
      label: "Key",
      defaultVisible: false,
      colDef: {
        id: "key",
        header: "Key",
        cell: ({ row }) =>
          row.original.key != null
            ? (
              <span style={{ fontFamily: "var(--font-mono)", fontSize: 11.5, color: "var(--fg-1)" }}>
                {PITCH_CLASSES[row.original.key] ?? "--"}
              </span>
            )
            : <span style={{ color: "var(--fg-3)" }}>--</span>,
      },
    },
    {
      id: "mode",
      label: "Mode",
      defaultVisible: false,
      colDef: {
        id: "mode",
        header: "Mode",
        cell: ({ row }) =>
          row.original.mode === 1
            ? <span style={{ fontSize: 11.5, color: "var(--fg-1)" }}>Major</span>
            : row.original.mode === 0
              ? <span style={{ fontSize: 11.5, color: "var(--fg-1)" }}>Minor</span>
              : <span style={{ color: "var(--fg-3)" }}>--</span>,
      },
    },
    {
      id: "time_sig",
      label: "Time sig.",
      defaultVisible: false,
      colDef: {
        id: "time_sig",
        header: "Time sig.",
        numeric: true,
        cell: ({ row }) =>
          row.original.time_signature != null
            ? (
              <span style={{ fontFamily: "var(--font-mono)", fontSize: 11.5, color: "var(--fg-1)" }}>
                {row.original.time_signature}/4
              </span>
            )
            : <span style={{ color: "var(--fg-3)" }}>--</span>,
      },
    },
    {
      id: "energy",
      label: "Energy",
      defaultVisible: false,
      colDef: {
        id: "energy",
        header: "Energy",
        numeric: true,
        enableSorting: true,
        cell: ({ row }) =>
          row.original.energy != null
            ? <EnergyBar value={row.original.energy} />
            : <span style={{ color: "var(--fg-3)" }}>--</span>,
      },
    },
    {
      id: "danceability",
      label: "Danceability",
      defaultVisible: false,
      colDef: {
        id: "danceability",
        header: "Danceability",
        numeric: true,
        enableSorting: true,
        cell: ({ row }) =>
          row.original.danceability != null
            ? <EnergyBar value={row.original.danceability} />
            : <span style={{ color: "var(--fg-3)" }}>--</span>,
      },
    },
    {
      id: "valence",
      label: "Valence",
      defaultVisible: false,
      colDef: {
        id: "valence",
        header: "Valence",
        numeric: true,
        cell: ({ row }) =>
          row.original.valence != null
            ? <EnergyBar value={row.original.valence} />
            : <span style={{ color: "var(--fg-3)" }}>--</span>,
      },
    },
    {
      id: "acousticness",
      label: "Acousticness",
      defaultVisible: false,
      colDef: {
        id: "acousticness",
        header: "Acousticness",
        numeric: true,
        cell: ({ row }) =>
          row.original.acousticness != null
            ? <EnergyBar value={row.original.acousticness} />
            : <span style={{ color: "var(--fg-3)" }}>--</span>,
      },
    },
    {
      id: "instrumentalness",
      label: "Instrumental",
      defaultVisible: false,
      colDef: {
        id: "instrumentalness",
        header: "Instrumental",
        numeric: true,
        cell: ({ row }) =>
          row.original.instrumentalness != null
            ? <EnergyBar value={row.original.instrumentalness} />
            : <span style={{ color: "var(--fg-3)" }}>--</span>,
      },
    },
    {
      id: "liveness",
      label: "Liveness",
      defaultVisible: false,
      colDef: {
        id: "liveness",
        header: "Liveness",
        numeric: true,
        cell: ({ row }) =>
          row.original.liveness != null
            ? <EnergyBar value={row.original.liveness} />
            : <span style={{ color: "var(--fg-3)" }}>--</span>,
      },
    },
    {
      id: "speechiness",
      label: "Speechiness",
      defaultVisible: false,
      colDef: {
        id: "speechiness",
        header: "Speechiness",
        numeric: true,
        cell: ({ row }) =>
          row.original.speechiness != null
            ? <EnergyBar value={row.original.speechiness} />
            : <span style={{ color: "var(--fg-3)" }}>--</span>,
      },
    },
    {
      id: "loudness",
      label: "Loudness",
      defaultVisible: false,
      colDef: {
        id: "loudness",
        header: "Loudness",
        numeric: true,
        cell: ({ row }) =>
          row.original.loudness != null
            ? (
              <span style={{ fontFamily: "var(--font-mono)", fontSize: 11.5, color: "var(--fg-1)" }}>
                {row.original.loudness.toFixed(1)} dB
              </span>
            )
            : <span style={{ color: "var(--fg-3)" }}>--</span>,
      },
    },
  ];
}

function buildVisibleColumns(
  registry: TrackColumnSpec[],
  visibility: Record<string, boolean>,
  order: string[],
): ColumnDef<Track & { index: number }>[] {
  const byId = new Map(registry.map((s) => [s.id, s]));
  return order
    .filter((id) => visibility[id])
    .map((id) => byId.get(id))
    .filter((s): s is TrackColumnSpec => s !== undefined)
    .map((s) => s.colDef);
}

function NumberInput({
  placeholder,
  value,
  onChange,
}: {
  placeholder: string;
  value: number | undefined;
  onChange: (v: number | undefined) => void;
}) {
  return (
    <input
      type="number"
      placeholder={placeholder}
      value={value ?? ""}
      onChange={(e) => { onChange(e.target.value ? Number(e.target.value) : undefined); }}
      style={{
        width: 72,
        height: 26,
        padding: "0 6px",
        border: "1px solid var(--hair)",
        borderRadius: "var(--radius-sm)",
        background: "var(--bg)",
        color: "var(--fg)",
        fontSize: 11.5,
      }}
    />
  );
}

function RangeRow({
  label,
  minVal,
  maxVal,
  onMinChange,
  onMaxChange,
  minPlaceholder,
  maxPlaceholder,
  children,
}: {
  label: string;
  minVal: number | undefined;
  maxVal: number | undefined;
  onMinChange: (v: number | undefined) => void;
  onMaxChange: (v: number | undefined) => void;
  minPlaceholder?: string;
  maxPlaceholder?: string;
  children?: ReactNode;
}) {
  return (
    <div style={{ padding: "4px 8px", display: "flex", alignItems: "center", gap: 8 }}>
      <span style={{ fontSize: 11.5, color: "var(--fg-1)", minWidth: 100 }}>{label}</span>
      {children ?? (
        <>
          <NumberInput placeholder={minPlaceholder ?? "Min"} value={minVal} onChange={onMinChange} />
          <span style={{ color: "var(--fg-3)", fontSize: 11 }}>–</span>
          <NumberInput placeholder={maxPlaceholder ?? "Max"} value={maxVal} onChange={onMaxChange} />
        </>
      )}
    </div>
  );
}

export default function TracksPage() {
  const navigate = useNavigate();
  const [localSearch, setLocalSearch] = useState(() => useFilterStore.getState().tracksSearch);
  const searchRef = useRef<ReturnType<typeof setTimeout>>(null);
  const [columnPickerOpen, setColumnPickerOpen] = useState(false);
  const columnPickerRef = useRef<HTMLButtonElement>(null);
  const [moreFiltersOpen, setMoreFiltersOpen] = useState(false);
  const moreFiltersRef = useRef<HTMLButtonElement>(null);

  const store = useFilterStore();
  const {
    tracksSearch, setTracksSearch,
    genres, setGenres,
    yearMin, setYearMin,
    yearMax, setYearMax,
    popularityMin, setPopularityMin,
    popularityMax, setPopularityMax,
    durationMin, setDurationMin,
    durationMax, setDurationMax,
    explicit, setExplicit,
    playlistId, setPlaylistId,
    savedAtMin, setSavedAtMin,
    savedAtMax, setSavedAtMax,
    artistPopularityMin, setArtistPopularityMin,
    artistPopularityMax, setArtistPopularityMax,
    artistFollowersMin, setArtistFollowersMin,
    artistFollowersMax, setArtistFollowersMax,
    tempoMin, setTempoMin,
    tempoMax, setTempoMax,
    energyMin, setEnergyMin,
    energyMax, setEnergyMax,
    danceabilityMin, setDanceabilityMin,
    danceabilityMax, setDanceabilityMax,
    valenceMin, setValenceMin,
    valenceMax, setValenceMax,
    acousticnessMin, setAcousticnessMin,
    acousticnessMax, setAcousticnessMax,
    instrumentalnessMin, setInstrumentalnessMin,
    instrumentalnessMax, setInstrumentalnessMax,
    livenessMin, setLivenessMin,
    livenessMax, setLivenessMax,
    speechinessMin, setSpeechinessMin,
    speechinessMax, setSpeechinessMax,
    loudnessMin, setLoudnessMin,
    loudnessMax, setLoudnessMax,
    keys, setKeys,
    mode, setMode,
    timeSignatures, setTimeSignatures,
    page, setPage,
    pageSize,
    sortBy, setSortBy,
    sortDir, setSortDir,
    tracksColumnVisibility,
    tracksColumnOrder,
    setTracksColumnVisibility,
    resetTracksColumns,
  } = store;

  const { data: genreOptions = [] } = useQuery({
    queryKey: ["genres"],
    queryFn: getGenres,
  });

  const { data: playlists = [] } = useQuery({
    queryKey: ["playlists"],
    queryFn: getPlaylists,
  });

  const { data: stats } = useQuery({
    queryKey: ["track-stats"],
    queryFn: getTrackStats,
    staleTime: 60_000,
  });

  const sorting: SortingState = sortBy
    ? [{ id: SORT_KEY_TO_COL[sortBy] ?? sortBy, desc: sortDir === "desc" }]
    : [];

  function handleSortingChange(newSorting: SetStateAction<SortingState>) {
    const resolved = typeof newSorting === "function" ? newSorting(sorting) : newSorting;
    if (resolved.length === 0) {
      setSortBy("");
      setSortDir("asc");
    } else {
      const { id, desc } = resolved[0];
      setSortBy(COL_SORT_KEYS[id] ?? id);
      setSortDir(desc ? "desc" : "asc");
    }
  }

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: [
      "tracks",
      {
        tracksSearch, genres, yearMin, yearMax, popularityMin, popularityMax,
        durationMin, durationMax, explicit, playlistId, savedAtMin, savedAtMax,
        artistPopularityMin, artistPopularityMax, artistFollowersMin, artistFollowersMax,
        tempoMin, tempoMax, energyMin, energyMax,
        danceabilityMin, danceabilityMax, valenceMin, valenceMax,
        acousticnessMin, acousticnessMax, instrumentalnessMin, instrumentalnessMax,
        livenessMin, livenessMax, speechinessMin, speechinessMax,
        loudnessMin, loudnessMax, keys, mode, timeSignatures,
        page, pageSize, sortBy, sortDir,
      },
    ],
    queryFn: () =>
      getTracks({
        search: tracksSearch,
        genres,
        year_min: yearMin,
        year_max: yearMax,
        popularity_min: popularityMin,
        popularity_max: popularityMax,
        duration_min: durationMin,
        duration_max: durationMax,
        explicit,
        playlist: playlistId,
        saved_at_min: savedAtMin,
        saved_at_max: savedAtMax,
        artist_popularity_min: artistPopularityMin,
        artist_popularity_max: artistPopularityMax,
        artist_followers_min: artistFollowersMin,
        artist_followers_max: artistFollowersMax,
        tempo_min: tempoMin,
        tempo_max: tempoMax,
        energy_min: energyMin,
        energy_max: energyMax,
        danceability_min: danceabilityMin,
        danceability_max: danceabilityMax,
        valence_min: valenceMin,
        valence_max: valenceMax,
        acousticness_min: acousticnessMin,
        acousticness_max: acousticnessMax,
        instrumentalness_min: instrumentalnessMin,
        instrumentalness_max: instrumentalnessMax,
        liveness_min: livenessMin,
        liveness_max: livenessMax,
        speechiness_min: speechinessMin,
        speechiness_max: speechinessMax,
        loudness_min: loudnessMin,
        loudness_max: loudnessMax,
        keys,
        mode,
        time_signatures: timeSignatures,
        page,
        page_size: pageSize,
        sort_by: sortBy || undefined,
        sort_dir: sortDir,
      }),
  });

  function handleSearch(e: React.ChangeEvent<HTMLInputElement>) {
    const val = e.target.value;
    setLocalSearch(val);
    if (searchRef.current) clearTimeout(searchRef.current);
    searchRef.current = setTimeout(() => { setTracksSearch(val); }, 300);
  }

  const registry = buildRegistry(navigate);

  const rows = (data?.items ?? []).map((r, i) => ({
    ...r,
    index: (page - 1) * pageSize + i + 1,
  })) as (Track & { index: number })[];

  const columns = buildVisibleColumns(
    registry,
    tracksColumnVisibility,
    tracksColumnOrder,
  );

  const moreFilterCount = [
    playlistId, savedAtMin, savedAtMax,
    artistPopularityMin, artistPopularityMax, artistFollowersMin, artistFollowersMax,
    tempoMin, tempoMax, energyMin, energyMax,
    danceabilityMin, danceabilityMax, valenceMin, valenceMax,
    acousticnessMin, acousticnessMax, instrumentalnessMin, instrumentalnessMax,
    livenessMin, livenessMax, speechinessMin, speechinessMax,
    loudnessMin, loudnessMax,
    mode,
  ].filter((v) => v !== undefined).length + keys.length + timeSignatures.length;

  const hasInlineFilters =
    genres.length > 0 ||
    yearMin !== undefined ||
    yearMax !== undefined ||
    popularityMin !== undefined ||
    popularityMax !== undefined ||
    durationMin !== undefined ||
    durationMax !== undefined ||
    explicit !== undefined;

  function clearAllFilters() {
    setGenres([]);
    setYearMin(undefined);
    setYearMax(undefined);
    setPopularityMin(undefined);
    setPopularityMax(undefined);
    setDurationMin(undefined);
    setDurationMax(undefined);
    setExplicit(undefined);
    setPlaylistId(undefined);
    setSavedAtMin(undefined);
    setSavedAtMax(undefined);
    setArtistPopularityMin(undefined);
    setArtistPopularityMax(undefined);
    setArtistFollowersMin(undefined);
    setArtistFollowersMax(undefined);
    setTempoMin(undefined);
    setTempoMax(undefined);
    setEnergyMin(undefined);
    setEnergyMax(undefined);
    setDanceabilityMin(undefined);
    setDanceabilityMax(undefined);
    setValenceMin(undefined);
    setValenceMax(undefined);
    setAcousticnessMin(undefined);
    setAcousticnessMax(undefined);
    setInstrumentalnessMin(undefined);
    setInstrumentalnessMax(undefined);
    setLivenessMin(undefined);
    setLivenessMax(undefined);
    setSpeechinessMin(undefined);
    setSpeechinessMax(undefined);
    setLoudnessMin(undefined);
    setLoudnessMax(undefined);
    setKeys([]);
    setMode(undefined);
    setTimeSignatures([]);
    setLocalSearch("");
    setTracksSearch("");
  }

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

  const inputStyle: React.CSSProperties = {
    width: 80,
    height: 26,
    padding: "0 6px",
    border: "1px solid var(--hair)",
    borderRadius: "var(--radius-sm)",
    background: "var(--bg)",
    color: "var(--fg)",
    fontSize: 11.5,
  };

  const toggleBtnStyle = (active: boolean): React.CSSProperties => ({
    height: 22,
    padding: "0 7px",
    border: `1px solid ${active ? "var(--acc)" : "var(--hair)"}`,
    borderRadius: 3,
    background: active ? "var(--acc-soft)" : "var(--bg)",
    color: active ? "var(--acc-ink)" : "var(--fg-2)",
    fontSize: 11,
    cursor: "pointer",
    fontFamily: "var(--font-mono)",
  });

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
              placeholder={stats?.year_min != null ? String(stats.year_min) : "From"}
              value={yearMin ?? ""}
              onChange={(e) => { setYearMin(e.target.value ? Number(e.target.value) : undefined); }}
              style={inputStyle}
            />
            <input
              type="number"
              placeholder={stats?.year_max != null ? String(stats.year_max) : "To"}
              value={yearMax ?? ""}
              onChange={(e) => { setYearMax(e.target.value ? Number(e.target.value) : undefined); }}
              style={inputStyle}
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

        {/* Duration chip */}
        <FilterChip
          label="Duration"
          applied={durationMin !== undefined || durationMax !== undefined}
          value={
            durationMin !== undefined || durationMax !== undefined
              ? `${durationMin !== undefined ? fmtMs(durationMin) : "0:00"}–${durationMax !== undefined ? fmtMs(durationMax) : "∞"}`
              : undefined
          }
          onRemove={() => { setDurationMin(undefined); setDurationMax(undefined); }}
        >
          <PopoverGroup>Duration (seconds)</PopoverGroup>
          <div style={{ padding: "6px 8px", display: "flex", gap: 8 }}>
            <input
              type="number"
              placeholder={stats?.duration_min != null ? String(Math.round(stats.duration_min / 1000)) : "From"}
              value={durationMin !== undefined ? Math.round(durationMin / 1000) : ""}
              onChange={(e) => { setDurationMin(e.target.value ? Number(e.target.value) * 1000 : undefined); }}
              style={inputStyle}
            />
            <input
              type="number"
              placeholder={stats?.duration_max != null ? String(Math.round(stats.duration_max / 1000)) : "To"}
              value={durationMax !== undefined ? Math.round(durationMax / 1000) : ""}
              onChange={(e) => { setDurationMax(e.target.value ? Number(e.target.value) * 1000 : undefined); }}
              style={inputStyle}
            />
          </div>
        </FilterChip>

        {/* Date Added chip */}
        <FilterChip
          label="Added"
          applied={savedAtMin !== undefined || savedAtMax !== undefined}
          value={
            savedAtMin !== undefined || savedAtMax !== undefined
              ? [savedAtMin?.slice(0, 10), savedAtMax?.slice(0, 10)].filter(Boolean).join("–")
              : undefined
          }
          onRemove={() => { setSavedAtMin(undefined); setSavedAtMax(undefined); }}
        >
          <PopoverGroup>Date added</PopoverGroup>
          <div style={{ padding: "6px 8px", display: "flex", flexDirection: "column", gap: 6 }}>
            <label style={{ fontSize: 11, color: "var(--fg-2)" }}>
              From
              <input
                type="date"
                value={savedAtMin ? savedAtMin.slice(0, 10) : ""}
                onChange={(e) => { setSavedAtMin(e.target.value ? `${e.target.value}T00:00:00Z` : undefined); }}
                style={{ ...inputStyle, display: "block", width: 140, marginTop: 2 }}
              />
            </label>
            <label style={{ fontSize: 11, color: "var(--fg-2)" }}>
              To
              <input
                type="date"
                value={savedAtMax ? savedAtMax.slice(0, 10) : ""}
                onChange={(e) => { setSavedAtMax(e.target.value ? `${e.target.value}T23:59:59Z` : undefined); }}
                style={{ ...inputStyle, display: "block", width: 140, marginTop: 2 }}
              />
            </label>
          </div>
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

        {/* More filters */}
        <button
          ref={moreFiltersRef}
          onClick={() => { setMoreFiltersOpen((o) => !o); }}
          style={{
            display: "inline-flex",
            alignItems: "center",
            gap: 5,
            height: 24,
            padding: "0 9px",
            border: `1px solid ${moreFilterCount > 0 ? "color-mix(in oklch, var(--acc) 40%, var(--hair))" : "var(--hair-strong)"}`,
            borderStyle: moreFilterCount > 0 ? "solid" : "dashed",
            borderRadius: 999,
            background: moreFilterCount > 0 ? "var(--acc-soft)" : "var(--bg)",
            color: moreFilterCount > 0 ? "var(--acc-ink)" : "var(--fg-1)",
            fontSize: 11.5,
            cursor: "pointer",
          }}
        >
          <SlidersHorizontal size={11} />
          More filters
          {moreFilterCount > 0 && (
            <span
              style={{
                display: "inline-flex",
                alignItems: "center",
                justifyContent: "center",
                width: 16,
                height: 16,
                borderRadius: "50%",
                background: "var(--acc)",
                color: "black",
                fontSize: 9,
                fontWeight: 700,
              }}
            >
              {moreFilterCount}
            </span>
          )}
        </button>

        <Popover
          anchor={moreFiltersRef}
          open={moreFiltersOpen}
          onClose={() => { setMoreFiltersOpen(false); }}
        >
          <div style={{ minWidth: 360, maxHeight: 480, overflowY: "auto" }}>
            <PopoverGroup>Track</PopoverGroup>
            {/* Playlist */}
            <div style={{ padding: "4px 8px" }}>
              <span style={{ fontSize: 11, color: "var(--fg-2)", display: "block", marginBottom: 3 }}>Playlist</span>
              <select
                value={playlistId ?? ""}
                onChange={(e) => { setPlaylistId(e.target.value || undefined); }}
                style={{
                  width: "100%",
                  height: 26,
                  padding: "0 6px",
                  border: "1px solid var(--hair)",
                  borderRadius: "var(--radius-sm)",
                  background: "var(--bg)",
                  color: "var(--fg)",
                  fontSize: 11.5,
                }}
              >
                <option value="">Any playlist</option>
                {playlists.map((p) => (
                  <option key={p.spotify_id} value={p.spotify_id}>{p.name}</option>
                ))}
              </select>
            </div>

            <PopoverGroup>Audio features</PopoverGroup>
            {/* Tempo */}
            <RangeRow
              label="Tempo (BPM)"
              minVal={tempoMin} maxVal={tempoMax}
              onMinChange={setTempoMin} onMaxChange={setTempoMax}
              minPlaceholder={stats?.tempo_min != null ? String(Math.round(stats.tempo_min)) : undefined}
              maxPlaceholder={stats?.tempo_max != null ? String(Math.round(stats.tempo_max)) : undefined}
            />
            {/* Loudness */}
            <RangeRow
              label="Loudness (dB)"
              minVal={loudnessMin} maxVal={loudnessMax}
              onMinChange={setLoudnessMin} onMaxChange={setLoudnessMax}
              minPlaceholder={stats?.loudness_min != null ? String(Math.round(stats.loudness_min)) : "-60"}
              maxPlaceholder={stats?.loudness_max != null ? String(Math.round(stats.loudness_max)) : "0"}
            />
            {/* 0-1 float filters */}
            {([
              ["Energy", energyMin, energyMax, setEnergyMin, setEnergyMax, stats?.energy_min, stats?.energy_max],
              ["Danceability", danceabilityMin, danceabilityMax, setDanceabilityMin, setDanceabilityMax, stats?.danceability_min, stats?.danceability_max],
              ["Valence", valenceMin, valenceMax, setValenceMin, setValenceMax, stats?.valence_min, stats?.valence_max],
              ["Acousticness", acousticnessMin, acousticnessMax, setAcousticnessMin, setAcousticnessMax, stats?.acousticness_min, stats?.acousticness_max],
              ["Instrumental.", instrumentalnessMin, instrumentalnessMax, setInstrumentalnessMin, setInstrumentalnessMax, stats?.instrumentalness_min, stats?.instrumentalness_max],
              ["Liveness", livenessMin, livenessMax, setLivenessMin, setLivenessMax, stats?.liveness_min, stats?.liveness_max],
              ["Speechiness", speechinessMin, speechinessMax, setSpeechinessMin, setSpeechinessMax, stats?.speechiness_min, stats?.speechiness_max],
            ] as const).map(([label, minV, maxV, onMin, onMax, statMin, statMax]) => (
              <RangeRow
                key={label}
                label={label}
                minVal={minV !== undefined ? Math.round(minV * 100) : undefined}
                maxVal={maxV !== undefined ? Math.round(maxV * 100) : undefined}
                onMinChange={(v) => { onMin(v !== undefined ? v / 100 : undefined); }}
                onMaxChange={(v) => { onMax(v !== undefined ? v / 100 : undefined); }}
                minPlaceholder={statMin != null ? `${String(Math.round(statMin * 100))}%` : "0%"}
                maxPlaceholder={statMax != null ? `${String(Math.round(statMax * 100))}%` : "100%"}
              />
            ))}
            {/* Key */}
            <div style={{ padding: "4px 8px" }}>
              <span style={{ fontSize: 11, color: "var(--fg-2)", display: "block", marginBottom: 4 }}>Key</span>
              <div style={{ display: "flex", flexWrap: "wrap", gap: 4 }}>
                {PITCH_CLASSES.map((pc, i) => (
                  <button
                    key={pc}
                    onClick={() => {
                      setKeys(keys.includes(i) ? keys.filter((k) => k !== i) : [...keys, i]);
                    }}
                    style={toggleBtnStyle(keys.includes(i))}
                  >
                    {pc}
                  </button>
                ))}
              </div>
            </div>
            {/* Mode */}
            <div style={{ padding: "4px 8px" }}>
              <span style={{ fontSize: 11, color: "var(--fg-2)", display: "block", marginBottom: 4 }}>Mode</span>
              <div style={{ display: "flex", gap: 4 }}>
                {[["Any", undefined], ["Major", 1], ["Minor", 0]].map(([label, val]) => (
                  <button
                    key={String(label)}
                    onClick={() => { setMode(mode === val ? undefined : (val as number | undefined)); }}
                    style={toggleBtnStyle(mode === val)}
                  >
                    {String(label)}
                  </button>
                ))}
              </div>
            </div>
            {/* Time signature */}
            <div style={{ padding: "4px 8px" }}>
              <span style={{ fontSize: 11, color: "var(--fg-2)", display: "block", marginBottom: 4 }}>Time signature</span>
              <div style={{ display: "flex", gap: 4 }}>
                {[3, 4, 5, 6, 7].map((ts) => (
                  <button
                    key={ts}
                    onClick={() => {
                      setTimeSignatures(
                        timeSignatures.includes(ts)
                          ? timeSignatures.filter((t) => t !== ts)
                          : [...timeSignatures, ts],
                      );
                    }}
                    style={toggleBtnStyle(timeSignatures.includes(ts))}
                  >
                    {ts}/4
                  </button>
                ))}
              </div>
            </div>

            <PopoverGroup>Artists</PopoverGroup>
            <RangeRow
              label="Artist popularity"
              minVal={artistPopularityMin} maxVal={artistPopularityMax}
              onMinChange={setArtistPopularityMin} onMaxChange={setArtistPopularityMax}
              minPlaceholder={stats?.artist_popularity_min != null ? String(stats.artist_popularity_min) : "0"}
              maxPlaceholder={stats?.artist_popularity_max != null ? String(stats.artist_popularity_max) : "100"}
            />
            <RangeRow
              label="Followers"
              minVal={artistFollowersMin} maxVal={artistFollowersMax}
              onMinChange={setArtistFollowersMin} onMaxChange={setArtistFollowersMax}
              minPlaceholder={stats?.artist_followers_min != null ? String(stats.artist_followers_min) : undefined}
              maxPlaceholder={stats?.artist_followers_max != null ? String(stats.artist_followers_max) : undefined}
            />

            {moreFilterCount > 0 && (
              <div style={{ padding: "6px 8px", borderTop: "1px solid var(--hair)", marginTop: 4 }}>
                <button
                  onClick={() => {
                    setPlaylistId(undefined);
                    setSavedAtMin(undefined);
                    setSavedAtMax(undefined);
                    setArtistPopularityMin(undefined);
                    setArtistPopularityMax(undefined);
                    setArtistFollowersMin(undefined);
                    setArtistFollowersMax(undefined);
                    setTempoMin(undefined);
                    setTempoMax(undefined);
                    setEnergyMin(undefined);
                    setEnergyMax(undefined);
                    setDanceabilityMin(undefined);
                    setDanceabilityMax(undefined);
                    setValenceMin(undefined);
                    setValenceMax(undefined);
                    setAcousticnessMin(undefined);
                    setAcousticnessMax(undefined);
                    setInstrumentalnessMin(undefined);
                    setInstrumentalnessMax(undefined);
                    setLivenessMin(undefined);
                    setLivenessMax(undefined);
                    setSpeechinessMin(undefined);
                    setSpeechinessMax(undefined);
                    setLoudnessMin(undefined);
                    setLoudnessMax(undefined);
                    setKeys([]);
                    setMode(undefined);
                    setTimeSignatures([]);
                  }}
                  style={{
                    fontSize: 11.5,
                    color: "var(--fg-3)",
                    cursor: "pointer",
                    textDecoration: "underline",
                    background: "none",
                    border: "none",
                  }}
                >
                  Clear advanced filters
                </button>
              </div>
            )}
          </div>
        </Popover>

        {(hasInlineFilters || moreFilterCount > 0) && (
          <button
            onClick={clearAllFilters}
            style={{
              fontSize: 11.5,
              color: "var(--fg-3)",
              cursor: "pointer",
              textDecoration: "underline",
              background: "none",
              border: "none",
            }}
          >
            Clear all
          </button>
        )}

        <div style={{ flex: 1 }} />

        {/* Column picker */}
        <button
          ref={columnPickerRef}
          onClick={() => { setColumnPickerOpen((o) => !o); }}
          style={{
            display: "inline-flex",
            alignItems: "center",
            gap: 5,
            height: 24,
            padding: "0 9px",
            border: "1px solid var(--hair)",
            borderRadius: "var(--radius-sm)",
            background: "var(--bg)",
            color: "var(--fg-1)",
            fontSize: 11.5,
            cursor: "pointer",
          }}
        >
          <Columns2 size={11} />
          Columns
        </button>

        <Popover
          anchor={columnPickerRef}
          open={columnPickerOpen}
          onClose={() => { setColumnPickerOpen(false); }}
          align="end"
        >
          <div style={{ minWidth: 200 }}>
            <PopoverGroup>Columns</PopoverGroup>
            {registry.map((spec) => (
              <label
                key={spec.id}
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: 8,
                  padding: "5px 8px",
                  cursor: "pointer",
                  fontSize: 12,
                  color: "var(--fg-1)",
                  borderRadius: 3,
                }}
                onMouseEnter={(e) => {
                  (e.currentTarget as HTMLElement).style.background = "var(--bg-2)";
                }}
                onMouseLeave={(e) => {
                  (e.currentTarget as HTMLElement).style.background = "none";
                }}
              >
                <input
                  type="checkbox"
                  checked={tracksColumnVisibility[spec.id] ?? spec.defaultVisible}
                  onChange={(e) => { setTracksColumnVisibility(spec.id, e.target.checked); }}
                  style={{ cursor: "pointer" }}
                />
                {spec.label}
              </label>
            ))}
            <div style={{ borderTop: "1px solid var(--hair)", padding: "5px 8px", marginTop: 2 }}>
              <button
                onClick={resetTracksColumns}
                style={{
                  fontSize: 11.5,
                  color: "var(--fg-3)",
                  cursor: "pointer",
                  textDecoration: "underline",
                  background: "none",
                  border: "none",
                }}
              >
                Reset to defaults
              </button>
            </div>
          </div>
        </Popover>
      </div>

      {/* Table */}
      <div style={{ overflow: "auto" }}>
        <DataTable
          data={rows as unknown as Track[]}
          columns={columns as ColumnDef<Track>[]}
          sorting={sorting}
          onSortingChange={handleSortingChange}
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
