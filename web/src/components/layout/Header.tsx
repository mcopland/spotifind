import { Search } from "lucide-react";
import { useAuth } from "../../hooks/useAuth";
import { useSyncStatus } from "../../hooks/useSyncStatus";
import Breadcrumbs from "./Breadcrumbs";
import ThemeToggle from "./ThemeToggle";

export default function Header() {
  const { user } = useAuth();
  const { syncJob, isRunning } = useSyncStatus();

  const initials = user?.display_name
    ? user.display_name
        .split(" ")
        .map((w) => w[0])
        .join("")
        .slice(0, 2)
        .toUpperCase()
    : "?";

  function syncLabel() {
    if (!syncJob || syncJob.status === "none") return null;
    if (isRunning) return "Syncing…";
    if (syncJob.status === "completed") return "Synced";
    if (syncJob.status === "failed") return "Sync failed";
    return null;
  }

  const label = syncLabel();

  return (
    <header
      style={{
        gridColumn: "1 / -1",
        display: "flex",
        alignItems: "center",
        gap: 10,
        padding: "0 12px",
        borderBottom: "1px solid var(--hair)",
        background: "var(--bg-1)",
        height: 40,
        position: "relative",
        zIndex: 5,
      }}
    >
      {/* Brand */}
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: 7,
          fontWeight: 600,
          fontSize: 13,
          letterSpacing: "-0.01em",
          flexShrink: 0,
        }}
      >
        <div
          style={{
            width: 14,
            height: 14,
            borderRadius: 3,
            background: `
              radial-gradient(circle at 30% 30%, var(--acc) 0 35%, transparent 36%),
              radial-gradient(circle at 70% 70%, var(--acc-ink) 0 30%, transparent 31%),
              var(--acc-soft)`,
            border: "1px solid var(--hair-strong)",
          }}
        />
        <span>SpotiFind</span>
        <span style={{ color: "var(--fg-2)", fontWeight: 400 }}>
          · your library, as a database
        </span>
      </div>

      <div style={{ width: 16 }} />

      <Breadcrumbs />

      <div style={{ flex: 1 }} />

      {/* Cmd+K search placeholder */}
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: 6,
          padding: "4px 8px",
          border: "1px solid var(--hair)",
          borderRadius: "var(--radius-sm)",
          color: "var(--fg-2)",
          background: "var(--bg)",
          minWidth: 220,
          fontSize: 12,
          cursor: "text",
        }}
      >
        <Search size={12} style={{ color: "var(--fg-3)" }} />
        <span>Search tracks, artists, albums…</span>
        <kbd
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 10,
            color: "var(--fg-3)",
            marginLeft: "auto",
            border: "1px solid var(--hair)",
            padding: "0 4px",
            borderRadius: 3,
          }}
        >
          ⌘K
        </kbd>
      </div>

      {/* Sync pill */}
      {label && (
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 6,
            padding: "3px 8px",
            fontSize: 11,
            color: "var(--fg-1)",
            border: "1px solid var(--hair)",
            borderRadius: 999,
            background: "var(--bg)",
            fontFamily: "var(--font-mono)",
          }}
        >
          <span
            style={{
              width: 6,
              height: 6,
              borderRadius: "50%",
              background: isRunning ? "var(--warn)" : "var(--ok)",
              boxShadow: isRunning
                ? "0 0 0 3px color-mix(in oklch, var(--warn) 15%, transparent)"
                : "0 0 0 3px color-mix(in oklch, var(--ok) 15%, transparent)",
            }}
          />
          {label}
        </div>
      )}

      <ThemeToggle />

      {/* Avatar */}
      <div
        style={{
          width: 22,
          height: 22,
          borderRadius: "50%",
          background: "var(--acc-soft)",
          color: "var(--acc-ink)",
          fontSize: 10,
          fontWeight: 600,
          display: "grid",
          placeItems: "center",
          border: "1px solid var(--hair)",
          flexShrink: 0,
          overflow: "hidden",
        }}
        title={user?.display_name}
      >
        {user?.avatar_url ? (
          <img
            src={user.avatar_url}
            alt=""
            style={{ width: "100%", height: "100%", objectFit: "cover" }}
          />
        ) : (
          initials
        )}
      </div>
    </header>
  );
}
