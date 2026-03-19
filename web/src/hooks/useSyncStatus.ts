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
    void queryClient.invalidateQueries({ queryKey: ["sync-status"] });
  }

  function onSyncComplete() {
    void queryClient.invalidateQueries({ queryKey: ["tracks"] });
    void queryClient.invalidateQueries({ queryKey: ["albums"] });
    void queryClient.invalidateQueries({ queryKey: ["artists"] });
    void queryClient.invalidateQueries({ queryKey: ["playlists"] });
    void queryClient.invalidateQueries({ queryKey: ["stats"] });
    void queryClient.invalidateQueries({ queryKey: ["genres"] });
  }

  return { syncJob, isLoading, isRunning, startSync, onSyncComplete };
}
