import { useQuery } from "@tanstack/react-query";
import { RefreshCw } from "lucide-react";
import { getStats } from "../api/sync";
import { useSyncStatus } from "../hooks/useSyncStatus";

const PIPELINE_STEPS = [
  { key: "tracks", label: "Tracks", description: "Saved & playlist tracks" },
  { key: "albums", label: "Albums", description: "Album metadata" },
  { key: "artists", label: "Artists", description: "Artist profiles & genres" },
  { key: "playlists", label: "Playlists", description: "Your playlists" },
];

function StepBadge({ status }: { status: "done" | "running" | "pending" | "error" }) {
  const styles: Record<string, React.CSSProperties> = {
    done: { background: "color-mix(in oklch, var(--ok) 15%, var(--bg))", color: "var(--ok)", border: "1px solid color-mix(in oklch, var(--ok) 30%, transparent)" },
    running: { background: "color-mix(in oklch, var(--acc) 15%, var(--bg))", color: "var(--acc-ink)", border: "1px solid color-mix(in oklch, var(--acc) 30%, transparent)" },
    pending: { background: "var(--bg-2)", color: "var(--fg-3)", border: "1px solid var(--hair)" },
    error: { background: "color-mix(in oklch, var(--err) 15%, var(--bg))", color: "var(--err)", border: "1px solid color-mix(in oklch, var(--err) 30%, transparent)" },
  };
  const labels = { done: "Done", running: "Running", pending: "Pending", error: "Error" };
  return (
    <span
      style={{
        ...styles[status],
        display: "inline-block",
        fontSize: 10,
        borderRadius: 4,
        padding: "2px 7px",
        fontWeight: 500,
        fontFamily: "var(--font-mono)",
      }}
    >
      {labels[status]}
    </span>
  );
}

function formatDuration(start?: string, end?: string): string {
  if (!start || !end) return "--";
  const ms = new Date(end).getTime() - new Date(start).getTime();
  if (ms < 1000) return `${String(ms)}ms`;
  return `${(ms / 1000).toFixed(1)}s`;
}

export default function SyncPage() {
  const { syncJob, isRunning, startSync } = useSyncStatus();
  const { data: stats } = useQuery({ queryKey: ["stats"], queryFn: getStats });

  // Determine which step is currently active based on entity_type
  function stepStatus(key: string): "done" | "running" | "pending" | "error" {
    if (!syncJob || syncJob.status === "none") return "pending";
    if (syncJob.status === "failed" && syncJob.entity_type === key) return "error";
    if (syncJob.status === "completed") return "done";
    const order = PIPELINE_STEPS.map((s) => s.key);
    const currentIdx = order.indexOf(syncJob.entity_type);
    const stepIdx = order.indexOf(key);
    if (stepIdx < currentIdx) return "done";
    if (stepIdx === currentIdx && isRunning) return "running";
    return "pending";
  }

  const lastSynced = syncJob?.finished_at
    ? new Date(syncJob.finished_at).toLocaleString(undefined, {
        month: "short",
        day: "numeric",
        hour: "2-digit",
        minute: "2-digit",
      })
    : "--";

  return (
    <div style={{ padding: "20px 24px", maxWidth: 800, margin: "0 auto" }}>
      {/* Header */}
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          marginBottom: 20,
        }}
      >
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
            Sync
          </h1>
          <div style={{ marginTop: 2, fontSize: 12, color: "var(--fg-2)" }}>
            Sync your Spotify library to the local database.
          </div>
        </div>
        <button
          onClick={() => { void startSync(); }}
          disabled={isRunning}
          style={{
            display: "inline-flex",
            alignItems: "center",
            gap: 6,
            padding: "6px 14px",
            background: isRunning ? "var(--bg-2)" : "var(--acc)",
            color: isRunning ? "var(--fg-3)" : "black",
            border: "none",
            borderRadius: "var(--radius-sm)",
            fontSize: 12,
            fontWeight: 600,
            cursor: isRunning ? "not-allowed" : "pointer",
          }}
        >
          <RefreshCw size={13} style={{ animation: isRunning ? "spin 1s linear infinite" : "none" }} />
          {isRunning ? "Syncing…" : "Sync now"}
        </button>
        <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
      </div>

      {/* Status card */}
      <div
        style={{
          background: "var(--bg)",
          border: "1px solid var(--hair)",
          borderRadius: "var(--radius)",
          padding: "14px 16px",
          marginBottom: 16,
          display: "flex",
          gap: 32,
        }}
      >
        <div>
          <div style={{ fontSize: 11, color: "var(--fg-3)", marginBottom: 2 }}>Status</div>
          <StepBadge
            status={
              !syncJob || syncJob.status === "none"
                ? "pending"
                : syncJob.status === "completed"
                  ? "done"
                  : syncJob.status === "failed"
                    ? "error"
                    : "running"
            }
          />
        </div>
        <div>
          <div style={{ fontSize: 11, color: "var(--fg-3)", marginBottom: 2 }}>Last synced</div>
          <span style={{ fontSize: 12, fontFamily: "var(--font-mono)", color: "var(--fg-1)" }}>
            {lastSynced}
          </span>
        </div>
        {syncJob && syncJob.status !== "none" && (
          <div>
            <div style={{ fontSize: 11, color: "var(--fg-3)", marginBottom: 2 }}>Duration</div>
            <span style={{ fontSize: 12, fontFamily: "var(--font-mono)", color: "var(--fg-1)" }}>
              {formatDuration(syncJob.started_at, syncJob.finished_at)}
            </span>
          </div>
        )}
        {syncJob?.error && (
          <div style={{ flex: 1 }}>
            <div style={{ fontSize: 11, color: "var(--fg-3)", marginBottom: 2 }}>Error</div>
            <span style={{ fontSize: 11, fontFamily: "var(--font-mono)", color: "var(--err)" }}>
              {syncJob.error}
            </span>
          </div>
        )}
      </div>

      {/* Pipeline steps */}
      <div
        style={{
          background: "var(--bg)",
          border: "1px solid var(--hair)",
          borderRadius: "var(--radius)",
          overflow: "hidden",
          marginBottom: 16,
        }}
      >
        <div
          style={{
            padding: "10px 16px",
            borderBottom: "1px solid var(--hair)",
            fontSize: 11,
            fontWeight: 600,
            color: "var(--fg)",
          }}
        >
          Pipeline
        </div>
        {PIPELINE_STEPS.map((step, i) => {
          const status = stepStatus(step.key);
          const count =
            step.key === "tracks"
              ? stats?.tracks
              : step.key === "albums"
                ? stats?.albums
                : step.key === "artists"
                  ? stats?.artists
                  : stats?.playlists;
          return (
            <div
              key={step.key}
              style={{
                display: "flex",
                alignItems: "center",
                padding: "12px 16px",
                borderBottom: i < PIPELINE_STEPS.length - 1 ? "1px solid var(--hair)" : undefined,
                gap: 12,
              }}
            >
              <div
                style={{
                  width: 28,
                  height: 28,
                  borderRadius: "50%",
                  background:
                    status === "done"
                      ? "color-mix(in oklch, var(--ok) 20%, var(--bg))"
                      : status === "running"
                        ? "color-mix(in oklch, var(--acc) 20%, var(--bg))"
                        : "var(--bg-2)",
                  border: `1px solid ${status === "done" ? "var(--ok)" : status === "running" ? "var(--acc)" : "var(--hair)"}`,
                  display: "grid",
                  placeItems: "center",
                  fontSize: 11,
                  fontFamily: "var(--font-mono)",
                  color: status === "done" ? "var(--ok)" : status === "running" ? "var(--acc-ink)" : "var(--fg-3)",
                  flexShrink: 0,
                }}
              >
                {status === "done" ? "✓" : i + 1}
              </div>
              <div style={{ flex: 1 }}>
                <div style={{ fontSize: 12, fontWeight: 500, color: "var(--fg)" }}>{step.label}</div>
                <div style={{ fontSize: 11, color: "var(--fg-3)" }}>{step.description}</div>
              </div>
              {count != null && (
                <span style={{ fontFamily: "var(--font-mono)", fontSize: 12, color: "var(--fg-2)" }}>
                  {count.toLocaleString()}
                </span>
              )}
              <StepBadge status={status} />
            </div>
          );
        })}
      </div>

      {/* Progress bar when running */}
      {isRunning && syncJob && syncJob.total_items > 0 && (
        <div
          style={{
            background: "var(--bg)",
            border: "1px solid var(--hair)",
            borderRadius: "var(--radius)",
            padding: "14px 16px",
          }}
        >
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              marginBottom: 6,
              fontSize: 11,
              color: "var(--fg-2)",
            }}
          >
            <span>{syncJob.entity_type}</span>
            <span style={{ fontFamily: "var(--font-mono)" }}>
              {syncJob.synced_items} / {syncJob.total_items}
            </span>
          </div>
          <div
            style={{ height: 4, background: "var(--bg-2)", borderRadius: 2, overflow: "hidden" }}
          >
            <div
              style={{
                width: `${String(Math.round((syncJob.synced_items / syncJob.total_items) * 100))}%`,
                height: "100%",
                background: "var(--acc)",
                borderRadius: 2,
                transition: "width 0.3s ease",
              }}
            />
          </div>
        </div>
      )}
    </div>
  );
}
