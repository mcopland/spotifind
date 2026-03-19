import { useQuery, useQueryClient } from "@tanstack/react-query";
import { getSyncStatus, triggerSync } from "../api/sync";

export function useSyncStatus() {
  const queryClient = useQueryClient();

  const { data: syncJob, isLoading } = useQuery({
    queryKey: ["sync-status"],
    queryFn: getSyncStatus,
    refetchInterval: query => {
      const status = query.state.data?.status;
      if (status === "running" || status === "pending") return 2000;
      return false;
    },
  });

  const isRunning = syncJob?.status === "running" || syncJob?.status === "pending";

  async function startSync() {
    await triggerSync();
    queryClient.invalidateQueries({ queryKey: ["sync-status"] });
  }

  function onSyncComplete() {
    queryClient.invalidateQueries({ queryKey: ["tracks"] });
    queryClient.invalidateQueries({ queryKey: ["albums"] });
    queryClient.invalidateQueries({ queryKey: ["artists"] });
    queryClient.invalidateQueries({ queryKey: ["playlists"] });
    queryClient.invalidateQueries({ queryKey: ["stats"] });
    queryClient.invalidateQueries({ queryKey: ["genres"] });
  }

  return { syncJob, isLoading, isRunning, startSync, onSyncComplete };
}
