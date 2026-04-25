import { Outlet } from "react-router-dom";
import { useTheme } from "../../hooks/useTheme";
import Header from "./Header";
import TabBar from "./TabBar";

export default function AppLayout() {
  const { theme } = useTheme();

  return (
    <div
      data-theme={theme}
      style={{
        display: "grid",
        gridTemplateRows: "40px 40px 1fr",
        height: "100vh",
        width: "100vw",
        background: "var(--bg)",
        color: "var(--fg)",
      }}
    >
      <Header />
      <TabBar />
      <main
        style={{
          minWidth: 0,
          minHeight: 0,
          overflow: "auto",
          background: "var(--bg)",
        }}
      >
        <Outlet />
      </main>
    </div>
  );
}
