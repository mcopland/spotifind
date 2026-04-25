import { useLocation, useParams } from "react-router-dom";

const ROUTE_LABELS: Record<string, string> = {
  "": "Dashboard",
  library: "Library",
  artists: "Artists",
  playlists: "Playlists",
  history: "History",
  "query-builder": "Query builder",
  saved: "Saved queries",
  sync: "Sync",
  settings: "Settings",
};

export default function Breadcrumbs() {
  const { pathname } = useLocation();
  const params = useParams();

  const segments = pathname.split("/").filter(Boolean);

  // Build crumb list
  const crumbs: string[] = ["Home"];

  if (segments.length === 0) {
    crumbs.push("Dashboard");
  } else {
    const first = segments[0];
    const label = ROUTE_LABELS[first];
    if (label) crumbs.push(label);

    // Entity detail pages
    if (segments.length > 1 && segments[1]) {
      const id = params.id ?? segments[1];
      // Use a short form of the ID as fallback; pages can override via document.title
      crumbs.push(id);
    }
  }

  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        gap: 6,
        color: "var(--fg-2)",
        fontSize: 12,
      }}
    >
      {crumbs.map((c, i) => (
        <span key={i} style={{ display: "flex", alignItems: "center", gap: 6 }}>
          {i > 0 && (
            <span style={{ color: "var(--fg-3)" }}>/</span>
          )}
          {i === crumbs.length - 1 ? (
            <b style={{ color: "var(--fg)", fontWeight: 500 }}>{c}</b>
          ) : (
            <span>{c}</span>
          )}
        </span>
      ))}
    </div>
  );
}
