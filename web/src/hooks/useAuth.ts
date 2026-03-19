import { useQuery } from "@tanstack/react-query";
import { getMe } from "../api/auth";

export function useAuth() {
  const {
    data: user,
    isLoading,
    error,
  } = useQuery({
    queryKey: ["me"],
    queryFn: getMe,
    retry: false,
  });

  return { user, isLoading, isAuthenticated: !!user && !error };
}
