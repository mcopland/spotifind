import { Moon, Sun } from "lucide-react";
import { useTheme } from "../../hooks/useTheme";

export default function ThemeToggle() {
  const { theme, toggle } = useTheme();
  return (
    <button
      onClick={toggle}
      title="Toggle theme"
      aria-label="Toggle theme"
      style={{
        width: 28,
        height: 24,
        display: "grid",
        placeItems: "center",
        border: "1px solid var(--hair)",
        borderRadius: 999,
        background: "var(--bg)",
        color: "var(--fg-1)",
        cursor: "pointer",
      }}
    >
      {theme === "dark" ? <Sun size={13} /> : <Moon size={13} />}
    </button>
  );
}
