import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import AuthCallback from "./components/auth/AuthCallback";
import LoginPage from "./components/auth/LoginPage";
import AppLayout from "./components/layout/AppLayout";
import ErrorBoundary from "./components/ErrorBoundary";
import { useAuth } from "./hooks/useAuth";
import AlbumDetailPage from "./pages/AlbumDetailPage";
import ArtistDetailPage from "./pages/ArtistDetailPage";
import DashboardPage from "./pages/DashboardPage";
import PlaylistDetailPage from "./pages/PlaylistDetailPage";
import PlaylistsPage from "./pages/PlaylistsPage";
import RecentlyPlayedPage from "./pages/RecentlyPlayedPage";
import SettingsPage from "./pages/SettingsPage";
import SyncPage from "./pages/SyncPage";
import TracksPage from "./pages/TracksPage";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { staleTime: 30_000, retry: 1 },
  },
});

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth();
  if (isLoading) {
    return (
      <div
        style={{
          minHeight: "100vh",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          background: "var(--bg)",
        }}
      >
        <div
          style={{
            width: 24,
            height: 24,
            border: "2px solid var(--acc)",
            borderTopColor: "transparent",
            borderRadius: "50%",
            animation: "spin 0.7s linear infinite",
          }}
        />
        <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
      </div>
    );
  }
  if (!isAuthenticated) return <Navigate to="/login" replace />;
  return <>{children}</>;
}

function AppRoutes() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/callback" element={<AuthCallback />} />
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <AppLayout />
          </ProtectedRoute>
        }
      >
        <Route index element={<ErrorBoundary><DashboardPage /></ErrorBoundary>} />
        <Route path="library" element={<ErrorBoundary><TracksPage /></ErrorBoundary>} />
        <Route path="artists/:id" element={<ErrorBoundary><ArtistDetailPage /></ErrorBoundary>} />
        <Route path="albums/:id" element={<ErrorBoundary><AlbumDetailPage /></ErrorBoundary>} />
        <Route path="playlists" element={<ErrorBoundary><PlaylistsPage /></ErrorBoundary>} />
        <Route path="playlists/:id" element={<ErrorBoundary><PlaylistDetailPage /></ErrorBoundary>} />
        <Route path="history" element={<ErrorBoundary><RecentlyPlayedPage /></ErrorBoundary>} />
        <Route path="sync" element={<ErrorBoundary><SyncPage /></ErrorBoundary>} />
        <Route path="settings" element={<ErrorBoundary><SettingsPage /></ErrorBoundary>} />
      </Route>
    </Routes>
  );
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <AppRoutes />
      </BrowserRouter>
    </QueryClientProvider>
  );
}
