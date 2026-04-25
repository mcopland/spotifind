import { useQuery, useQueryClient } from "@tanstack/react-query";
import { LogOut, RefreshCw } from "lucide-react";
import { logout } from "../api/auth";
import { getStats } from "../api/sync";
import { useAuth } from "../hooks/useAuth";
import { useSyncStatus } from "../hooks/useSyncStatus";

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div
      style={{
        background: "var(--bg)",
        border: "1px solid var(--hair)",
        borderRadius: "var(--radius)",
        overflow: "hidden",
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
        {title}
      </div>
      <div style={{ padding: "14px 16px" }}>{children}</div>
    </div>
  );
}

function Row({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        padding: "6px 0",
        borderBottom: "1px solid var(--hair)",
        fontSize: 12,
      }}
    >
      <span style={{ color: "var(--fg-2)" }}>{label}</span>
      <span style={{ fontFamily: "var(--font-mono)", color: "var(--fg-1)" }}>{value}</span>
    </div>
  );
}

export default function SettingsPage() {
  const { user } = useAuth();
  const { data: stats } = useQuery({ queryKey: ["stats"], queryFn: getStats });
  const { syncJob, isRunning, startSync } = useSyncStatus();
  const queryClient = useQueryClient();

  async function handleLogout() {
    try {
      await logout();
      void queryClient.invalidateQueries({ queryKey: ["me"] });
    } catch {
      // ignore — redirect will happen from axios interceptor on 401
    }
  }

  const lastSynced = syncJob?.finished_at
    ? new Date(syncJob.finished_at).toLocaleString(undefined, {
        month: "short",
        day: "numeric",
        year: "numeric",
        hour: "2-digit",
        minute: "2-digit",
      })
    : "Never";

  return (
    <div style={{ padding: "20px 24px", maxWidth: 720, margin: "0 auto" }}>
      <div style={{ marginBottom: 20 }}>
        <h1
          style={{
            margin: 0,
            fontSize: 18,
            fontWeight: 600,
            letterSpacing: "-0.015em",
            fontFamily: "var(--font-ui)",
          }}
        >
          Settings
        </h1>
        <div style={{ marginTop: 2, fontSize: 12, color: "var(--fg-2)" }}>
          Account, database, and sync configuration.
        </div>
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 12, marginBottom: 12 }}>
        {/* Account */}
        <Section title="Account">
          {user?.avatar_url && (
            <div style={{ marginBottom: 12, display: "flex", alignItems: "center", gap: 10 }}>
              <img
                src={user.avatar_url}
                alt=""
                style={{ width: 36, height: 36, borderRadius: "50%", objectFit: "cover" }}
              />
              <div>
                <div style={{ fontSize: 13, fontWeight: 500, color: "var(--fg)" }}>
                  {user.display_name}
                </div>
                <div style={{ fontSize: 11, color: "var(--fg-3)" }}>{user.email}</div>
              </div>
            </div>
          )}
          <Row label="Spotify ID" value={user?.spotify_id ?? "--"} />
          <Row
            label="Last sync"
            value={
              user?.last_synced_at
                ? new Date(user.last_synced_at).toLocaleDateString()
                : "--"
            }
          />
        </Section>

        {/* Database */}
        <Section title="Database">
          <Row label="Tracks" value={stats?.tracks.toLocaleString() ?? "--"} />
          <Row label="Albums" value={stats?.albums.toLocaleString() ?? "--"} />
          <Row label="Artists" value={stats?.artists.toLocaleString() ?? "--"} />
          <Row label="Playlists" value={stats?.playlists.toLocaleString() ?? "--"} />
        </Section>
      </div>

      {/* Sync */}
      <Section title="Sync">
        <div
          style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            marginBottom: 12,
          }}
        >
          <div>
            <div style={{ fontSize: 12, color: "var(--fg-2)" }}>Last synced: {lastSynced}</div>
            {syncJob?.status && syncJob.status !== "none" && (
              <div style={{ fontSize: 11, color: "var(--fg-3)", marginTop: 2 }}>
                Status:{" "}
                <span
                  style={{
                    fontFamily: "var(--font-mono)",
                    color:
                      syncJob.status === "completed"
                        ? "var(--ok)"
                        : syncJob.status === "failed"
                          ? "var(--err)"
                          : "var(--acc-ink)",
                  }}
                >
                  {syncJob.status}
                </span>
              </div>
            )}
          </div>
          <button
            onClick={() => { void startSync(); }}
            disabled={isRunning}
            style={{
              display: "inline-flex",
              alignItems: "center",
              gap: 6,
              padding: "5px 12px",
              background: isRunning ? "var(--bg-2)" : "var(--acc)",
              color: isRunning ? "var(--fg-3)" : "black",
              border: "none",
              borderRadius: "var(--radius-sm)",
              fontSize: 12,
              fontWeight: 600,
              cursor: isRunning ? "not-allowed" : "pointer",
            }}
          >
            <RefreshCw
              size={12}
              style={{ animation: isRunning ? "spin 1s linear infinite" : "none" }}
            />
            {isRunning ? "Syncing…" : "Sync now"}
          </button>
          <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
        </div>
      </Section>

      {/* Danger zone */}
      <div
        style={{
          marginTop: 24,
          background: "var(--bg)",
          border: "1px solid color-mix(in oklch, var(--err) 30%, var(--hair))",
          borderRadius: "var(--radius)",
          overflow: "hidden",
        }}
      >
        <div
          style={{
            padding: "10px 16px",
            borderBottom: "1px solid color-mix(in oklch, var(--err) 20%, var(--hair))",
            fontSize: 11,
            fontWeight: 600,
            color: "var(--err)",
          }}
        >
          Danger zone
        </div>
        <div
          style={{
            padding: "14px 16px",
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
          }}
        >
          <div>
            <div style={{ fontSize: 12, fontWeight: 500, color: "var(--fg)" }}>Sign out</div>
            <div style={{ fontSize: 11, color: "var(--fg-3)" }}>
              Log out of your Spotify account.
            </div>
          </div>
          <button
            onClick={() => { void handleLogout(); }}
            style={{
              display: "inline-flex",
              alignItems: "center",
              gap: 6,
              padding: "5px 12px",
              background: "color-mix(in oklch, var(--err) 10%, var(--bg))",
              color: "var(--err)",
              border: "1px solid color-mix(in oklch, var(--err) 30%, var(--hair))",
              borderRadius: "var(--radius-sm)",
              fontSize: 12,
              fontWeight: 500,
              cursor: "pointer",
            }}
          >
            <LogOut size={12} /> Sign out
          </button>
        </div>
      </div>
    </div>
  );
}
