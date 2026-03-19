import { RefreshCw } from "lucide-react";
import { useSyncStatus } from "../../hooks/useSyncStatus";

export default function SyncButton() {
  const { syncJob, isRunning, startSync } = useSyncStatus();

  const progress =
    syncJob && syncJob.total_items > 0
      ? Math.round((syncJob.synced_items / syncJob.total_items) * 100)
      : null;

  return (
    <div className="flex items-center gap-3">
      {isRunning && (
        <div className="flex items-center gap-2 text-sm text-gray-400">
          <div className="w-32 h-1.5 bg-gray-700 rounded-full overflow-hidden">
            <div
              className="h-full bg-[#1DB954] transition-all duration-500"
              style={{ width: progress != null ? `${progress}%` : "30%" }}
            />
          </div>
          <span>{progress != null ? `${progress}%` : "syncing..."}</span>
        </div>
      )}
      <button
        onClick={startSync}
        disabled={isRunning}
        className="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-[#1DB954] text-black font-medium rounded-full hover:bg-[#1ed760] transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
      >
        <RefreshCw className={`w-3.5 h-3.5 ${isRunning ? "animate-spin" : ""}`} />
        {isRunning ? "Syncing" : "Sync"}
      </button>
    </div>
  );
}
