import { useQuery } from "@tanstack/react-query";
import { getGenres, getStats } from "../api/sync";
import { getTopArtists } from "../api/top";
import StatCard from "../components/primitives/StatCard";
import HourOfWeek from "../components/viz/HourOfWeek";
import ListenPulse from "../components/viz/ListenPulse";
import YearHeatmap from "../components/viz/YearHeatmap";

const SAMPLE_SPARK = [40, 55, 38, 70, 62, 80, 74, 90, 68, 85, 92, 88];

function SampleLabel() {
  return (
    <span
      style={{
        display: "inline-block",
        fontSize: 9,
        color: "var(--fg-3)",
        background: "var(--bg-2)",
        border: "1px solid var(--hair)",
        borderRadius: 3,
        padding: "1px 5px",
        fontFamily: "var(--font-mono)",
      }}
    >
      sample data
    </span>
  );
}

const COVER_COLORS = [
  "oklch(0.7 0.12 280)",
  "oklch(0.7 0.12 160)",
  "oklch(0.7 0.12 40)",
  "oklch(0.7 0.12 220)",
  "oklch(0.7 0.12 320)",
];

function AvatarPlaceholder({ name, size = 32 }: { name: string; size?: number }) {
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
        fontSize: size * 0.38,
        flexShrink: 0,
      }}
    >
      {name.slice(0, 1).toUpperCase()}
    </div>
  );
}

function Card({ children, style }: { children: React.ReactNode; style?: React.CSSProperties }) {
  return (
    <div
      style={{
        background: "var(--bg)",
        border: "1px solid var(--hair)",
        borderRadius: "var(--radius)",
        overflow: "hidden",
        ...style,
      }}
    >
      {children}
    </div>
  );
}

function CardHeader({ title, action }: { title: string; action?: React.ReactNode }) {
  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        padding: "12px 16px",
        borderBottom: "1px solid var(--hair)",
      }}
    >
      <span style={{ fontSize: 12, fontWeight: 600, color: "var(--fg)" }}>{title}</span>
      {action}
    </div>
  );
}

// Deterministic genre proportions for the top 6 genres
const GENRE_WIDTHS = [38, 22, 15, 11, 8, 6];

export default function DashboardPage() {
  const { data: stats } = useQuery({ queryKey: ["stats"], queryFn: getStats });
  const { data: topArtistsData } = useQuery({
    queryKey: ["top-artists", "medium_term"],
    queryFn: () => getTopArtists({ time_range: "medium_term", page_size: 6 }),
  });
  const { data: genres = [] } = useQuery({ queryKey: ["genres"], queryFn: getGenres });

  const topArtists = topArtistsData?.items ?? [];
  const displayGenres = genres.slice(0, 6);

  return (
    <div style={{ padding: "20px 24px", maxWidth: 1200, margin: "0 auto" }}>
      {/* KPI row */}
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "repeat(5, 1fr)",
          gap: 12,
          marginBottom: 20,
        }}
      >
        <StatCard
          label="Tracks"
          value={stats ? stats.tracks.toLocaleString() : "…"}
          sparkData={SAMPLE_SPARK}
        />
        <StatCard
          label="Artists"
          value={stats ? stats.artists.toLocaleString() : "…"}
          sparkData={SAMPLE_SPARK.slice().reverse()}
        />
        <StatCard
          label="Playlists"
          value={stats ? stats.playlists.toLocaleString() : "…"}
        />
        <StatCard label="Plays" value="--" delta="sample" deltaUp={true} />
        <StatCard label="Avg session" value="--" delta="sample" deltaUp={false} />
      </div>

      {/* Listening visualizations */}
      <Card style={{ marginBottom: 20 }}>
        <CardHeader title="Listening activity" action={<SampleLabel />} />
        <div style={{ padding: 16, display: "flex", flexDirection: "column", gap: 20 }}>
          <div>
            <div style={{ fontSize: 11, color: "var(--fg-2)", marginBottom: 8, fontWeight: 500 }}>
              Past year
            </div>
            <YearHeatmap />
          </div>
          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 24 }}>
            <div>
              <div style={{ fontSize: 11, color: "var(--fg-2)", marginBottom: 8, fontWeight: 500 }}>
                Monthly plays
              </div>
              <ListenPulse width={320} height={90} />
            </div>
            <div>
              <div style={{ fontSize: 11, color: "var(--fg-2)", marginBottom: 8, fontWeight: 500 }}>
                Hour of week
              </div>
              <HourOfWeek />
            </div>
          </div>
        </div>
      </Card>

      {/* Bottom 3-column grid */}
      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr 1fr", gap: 12 }}>
        {/* Top artists */}
        <Card>
          <CardHeader
            title="Top artists"
            action={
              <span style={{ fontSize: 10, color: "var(--fg-3)", fontFamily: "var(--font-mono)" }}>
                6mo
              </span>
            }
          />
          <div>
            {topArtists.length === 0
              ? Array.from({ length: 5 }, (_, i) => (
                <div
                  key={i}
                  style={{
                    display: "flex",
                    alignItems: "center",
                    gap: 10,
                    padding: "8px 16px",
                    borderBottom: i < 4 ? "1px solid var(--hair)" : undefined,
                  }}
                >
                  <div
                    style={{ width: 32, height: 32, borderRadius: "50%", background: "var(--bg-2)" }}
                  />
                  <div style={{ flex: 1 }}>
                    <div
                      style={{
                        height: 9,
                        width: 80 + i * 10,
                        background: "var(--bg-2)",
                        borderRadius: 3,
                        marginBottom: 4,
                      }}
                    />
                    <div
                      style={{ height: 7, width: 50, background: "var(--bg-2)", borderRadius: 3 }}
                    />
                  </div>
                </div>
              ))
              : topArtists.map((artist, i) => (
                <div
                  key={artist.spotify_id}
                  style={{
                    display: "flex",
                    alignItems: "center",
                    gap: 10,
                    padding: "8px 16px",
                    borderBottom: i < topArtists.length - 1 ? "1px solid var(--hair)" : undefined,
                  }}
                >
                  <span
                    style={{
                      fontFamily: "var(--font-mono)",
                      fontSize: 10,
                      color: "var(--fg-3)",
                      width: 14,
                      textAlign: "right",
                      flexShrink: 0,
                    }}
                  >
                    {i + 1}
                  </span>
                  {artist.image_url ? (
                    <img
                      src={artist.image_url}
                      alt=""
                      style={{
                        width: 32,
                        height: 32,
                        borderRadius: "50%",
                        objectFit: "cover",
                        flexShrink: 0,
                      }}
                    />
                  ) : (
                    <AvatarPlaceholder name={artist.name} />
                  )}
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div
                      style={{
                        fontSize: 12,
                        fontWeight: 500,
                        color: "var(--fg)",
                        overflow: "hidden",
                        textOverflow: "ellipsis",
                        whiteSpace: "nowrap",
                      }}
                    >
                      {artist.name}
                    </div>
                    <div
                      style={{
                        fontSize: 10,
                        color: "var(--fg-3)",
                        overflow: "hidden",
                        textOverflow: "ellipsis",
                        whiteSpace: "nowrap",
                      }}
                    >
                      {artist.genres?.[0] ?? ""}
                    </div>
                  </div>
                </div>
              ))}
          </div>
        </Card>

        {/* Genre mix */}
        <Card>
          <CardHeader title="Genre mix" action={<SampleLabel />} />
          <div style={{ padding: "12px 16px" }}>
            {displayGenres.length === 0 ? (
              <div style={{ color: "var(--fg-3)", fontSize: 12, padding: "12px 0" }}>
                No genres found.
              </div>
            ) : (
              displayGenres.map((genre, i) => (
                <div key={genre} style={{ marginBottom: 10 }}>
                  <div
                    style={{
                      display: "flex",
                      justifyContent: "space-between",
                      marginBottom: 3,
                      fontSize: 11,
                    }}
                  >
                    <span style={{ color: "var(--fg-1)" }}>{genre}</span>
                    <span
                      style={{
                        fontFamily: "var(--font-mono)",
                        color: "var(--fg-3)",
                        fontSize: 10,
                      }}
                    >
                      {GENRE_WIDTHS[i] ?? 4}%
                    </span>
                  </div>
                  <div
                    style={{
                      height: 4,
                      background: "var(--bg-2)",
                      borderRadius: 2,
                      overflow: "hidden",
                    }}
                  >
                    <div
                      style={{
                        width: `${String(GENRE_WIDTHS[i] ?? 4)}%`,
                        height: "100%",
                        background: `color-mix(in oklch, var(--acc) ${String(90 - i * 12)}%, var(--bg-3))`,
                        borderRadius: 2,
                      }}
                    />
                  </div>
                </div>
              ))
            )}
          </div>
        </Card>

        {/* Needs attention (placeholder) */}
        <Card>
          <CardHeader title="Needs attention" />
          <div style={{ padding: "12px 16px" }}>
            {[
              { label: "Duplicate tracks", count: "--", color: "var(--warn)" },
              { label: "Orphaned albums", count: "--", color: "var(--warn)" },
              { label: "Missing artwork", count: "--", color: "var(--info)" },
              { label: "Low quality tracks", count: "--", color: "var(--fg-3)" },
            ].map(({ label, count, color }, i, arr) => (
              <div
                key={label}
                style={{
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "space-between",
                  padding: "8px 0",
                  borderBottom: i < arr.length - 1 ? "1px solid var(--hair)" : undefined,
                }}
              >
                <span style={{ fontSize: 12, color: "var(--fg-1)" }}>{label}</span>
                <span
                  style={{ fontFamily: "var(--font-mono)", fontSize: 12, color, fontWeight: 500 }}
                >
                  {count}
                </span>
              </div>
            ))}
            <div
              style={{
                marginTop: 10,
                fontSize: 10,
                color: "var(--fg-3)",
                fontFamily: "var(--font-mono)",
              }}
            >
              backend support coming soon
            </div>
          </div>
        </Card>
      </div>
    </div>
  );
}
