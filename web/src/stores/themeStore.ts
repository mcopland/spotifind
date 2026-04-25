import { create } from "zustand";

type Theme = "light" | "dark";

interface ThemeState {
  theme: Theme;
  toggle: () => void;
  setTheme: (t: Theme) => void;
}

function applyTheme(theme: Theme) {
  document.documentElement.dataset.theme = theme;
  localStorage.setItem("spotifind.theme", theme);
}

const stored = localStorage.getItem("spotifind.theme") as Theme | null;
const initial: Theme = stored ?? "light";
applyTheme(initial);

export const useThemeStore = create<ThemeState>((set) => ({
  theme: initial,
  toggle: () => {
    set((state) => {
      const next: Theme = state.theme === "light" ? "dark" : "light";
      applyTheme(next);
      return { theme: next };
    });
  },
  setTheme: (t) => {
    set(() => {
      applyTheme(t);
      return { theme: t };
    });
  },
}));
