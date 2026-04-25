import {
  BarChart2,
  Clock,
  Grid,
  List,
  RefreshCw,
  Save,
  Settings,
  Sliders,
  Table,
} from "lucide-react";
import { NavLink } from "react-router-dom";

interface Tab {
  id: string;
  to: string;
  label: string;
  icon: React.ReactNode;
  disabled?: boolean;
}

const TABS: Tab[] = [
  { id: "dashboard", to: "/", label: "Dashboard", icon: <Grid size={13} /> },
  { id: "library", to: "/library", label: "Library", icon: <Table size={13} /> },
  {
    id: "query-builder",
    to: "/query-builder",
    label: "Query builder",
    icon: <Sliders size={13} />,
  },
  { id: "history", to: "/history", label: "History", icon: <Clock size={13} /> },
  { id: "artists", to: "/artists", label: "Artists", icon: <BarChart2 size={13} /> },
  { id: "playlists", to: "/playlists", label: "Playlists", icon: <List size={13} /> },
  {
    id: "saved",
    to: "/saved",
    label: "Saved",
    icon: <Save size={13} />,
  },
  { id: "sync", to: "/sync", label: "Sync", icon: <RefreshCw size={13} /> },
  { id: "settings", to: "/settings", label: "Settings", icon: <Settings size={13} /> },
];

export default function TabBar() {
  return (
    <nav
      style={{
        display: "flex",
        alignItems: "center",
        gap: 0,
        padding: "0 12px",
        borderBottom: "1px solid var(--hair)",
        background: "var(--bg)",
        overflowX: "auto",
        height: 40,
      }}
    >
      {TABS.map((tab) => (
        <NavLink
          key={tab.id}
          to={tab.to}
          end={tab.to === "/"}
          style={({ isActive }) => ({
            display: "flex",
            alignItems: "center",
            gap: 6,
            padding: "0 12px",
            height: "100%",
            fontSize: 12,
            color: isActive ? "var(--fg)" : "var(--fg-1)",
            borderBottom: `2px solid ${isActive ? "var(--acc)" : "transparent"}`,
            fontWeight: isActive ? 500 : 400,
            whiteSpace: "nowrap",
            textDecoration: "none",
            opacity: tab.disabled ? 0.45 : 1,
            pointerEvents: tab.disabled ? "none" : "auto",
            transition: "color 80ms",
          })}
        >
          {tab.icon}
          {tab.label}
        </NavLink>
      ))}
    </nav>
  );
}
