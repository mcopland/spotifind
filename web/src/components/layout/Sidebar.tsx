import { useQuery } from "@tanstack/react-query";
import { Clock, Disc, ListMusic, Mic2, Music, TrendingUp } from "lucide-react";
import { NavLink } from "react-router-dom";
import { getStats } from "../../api/sync";

const nav = [
  { to: "/", label: "Tracks", icon: Music },
  { to: "/albums", label: "Albums", icon: Disc },
  { to: "/artists", label: "Artists", icon: Mic2 },
  { to: "/playlists", label: "Playlists", icon: ListMusic },
  { to: "/recently-played", label: "Recently Played", icon: Clock },
  { to: "/top", label: "Top Charts", icon: TrendingUp },
];

export default function Sidebar() {
  const { data: stats } = useQuery({ queryKey: ["stats"], queryFn: getStats });

  return (
    <nav className="w-48 border-r border-gray-800 pt-4 shrink-0">
      {nav.map(({ to, label, icon: Icon }) => {
        const count = stats
          ? {
              "/": stats.tracks,
              "/albums": stats.albums,
              "/artists": stats.artists,
              "/playlists": stats.playlists,
            }[to]
          : undefined;

        return (
          <NavLink
            key={to}
            to={to}
            end={to === "/"}
            className={({ isActive }) =>
              `flex items-center justify-between px-4 py-2.5 text-sm transition-colors ${
                isActive
                  ? "text-white bg-gray-800"
                  : "text-gray-400 hover:text-white hover:bg-gray-900"
              }`
            }
          >
            <div className="flex items-center gap-2.5">
              <Icon className="w-4 h-4" />
              {label}
            </div>
            {count != null && (
              <span className="text-xs text-gray-600">{count.toLocaleString()}</span>
            )}
          </NavLink>
        );
      })}
    </nav>
  );
}
