import { useThemeStore } from "../stores/themeStore";

export function useTheme() {
  const { theme, toggle, setTheme } = useThemeStore();
  return { theme, toggle, setTheme };
}
