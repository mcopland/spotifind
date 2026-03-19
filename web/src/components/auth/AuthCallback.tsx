import { useQueryClient } from "@tanstack/react-query";
import { useEffect } from "react";
import { useNavigate } from "react-router-dom";

export default function AuthCallback() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  useEffect(() => {
    queryClient.invalidateQueries({ queryKey: ["me"] });
    navigate("/", { replace: true });
  }, [navigate, queryClient]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-[#0f0f0f]">
      <p className="text-gray-400">Completing login...</p>
    </div>
  );
}
