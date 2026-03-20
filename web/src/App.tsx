import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import AuthCallback from "./components/auth/AuthCallback";
import LoginPage from "./components/auth/LoginPage";
import AppLayout from "./components/layout/AppLayout";
import ErrorBoundary from "./components/ErrorBoundary";
import { useAuth } from "./hooks/useAuth";
import AlbumsPage from "./pages/AlbumsPage";
import ArtistsPage from "./pages/ArtistsPage";
import PlaylistsPage from "./pages/PlaylistsPage";
import RecentlyPlayedPage from "./pages/RecentlyPlayedPage";
import TopChartsPage from "./pages/TopChartsPage";
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
      <div className="min-h-screen flex items-center justify-center bg-[#0f0f0f]">
        <div className="w-6 h-6 border-2 border-[#1DB954] border-t-transparent rounded-full animate-spin" />
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
        <Route index element={<ErrorBoundary><TracksPage /></ErrorBoundary>} />
        <Route path="albums" element={<ErrorBoundary><AlbumsPage /></ErrorBoundary>} />
        <Route path="artists" element={<ErrorBoundary><ArtistsPage /></ErrorBoundary>} />
        <Route path="playlists" element={<ErrorBoundary><PlaylistsPage /></ErrorBoundary>} />
        <Route path="recently-played" element={<ErrorBoundary><RecentlyPlayedPage /></ErrorBoundary>} />
        <Route path="top" element={<ErrorBoundary><TopChartsPage /></ErrorBoundary>} />
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
