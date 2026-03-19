import { useQueryClient } from "@tanstack/react-query";
import { LogOut, Search } from "lucide-react";
import { useRef } from "react";
import { useNavigate } from "react-router-dom";
import { logout } from "../../api/auth";
import { useAuth } from "../../hooks/useAuth";
import { useFilterStore } from "../../stores/filterStore";
import SyncButton from "../sync/SyncButton";

export default function Header() {
  const { user } = useAuth();
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const setSearch = useFilterStore(s => s.setSearch);
  const search = useFilterStore(s => s.search);
  const debounceRef = useRef<ReturnType<typeof setTimeout>>(null);

  function handleSearch(e: React.ChangeEvent<HTMLInputElement>) {
    const val = e.target.value;
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => { setSearch(val); }, 300);
    e.target.value = val;
  }

  async function handleLogout() {
    await logout();
    queryClient.clear();
    void navigate("/login");
  }

  return (
    <header className="flex items-center justify-between px-4 h-14 border-b border-gray-800 bg-[#0f0f0f] sticky top-0 z-50">
      <span className="text-lg font-bold text-white">SpotiFind</span>

      <div className="flex-1 max-w-md mx-6">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
          <input
            type="text"
            placeholder="Search..."
            defaultValue={search}
            onChange={handleSearch}
            className="w-full pl-9 pr-4 py-1.5 bg-gray-900 border border-gray-700 rounded-full text-sm text-white placeholder-gray-500 focus:outline-none focus:border-gray-500"
          />
        </div>
      </div>

      <div className="flex items-center gap-3">
        <SyncButton />
        {user && (
          <div className="flex items-center gap-2">
            {user.avatar_url && (
              <img src={user.avatar_url} alt="" className="w-7 h-7 rounded-full" />
            )}
            <span className="text-sm text-gray-300">{user.display_name}</span>
          </div>
        )}
        <button
          onClick={() => { void handleLogout(); }}
          className="p-1.5 text-gray-500 hover:text-white transition-colors"
          title="Logout"
        >
          <LogOut className="w-4 h-4" />
        </button>
      </div>
    </header>
  );
}
