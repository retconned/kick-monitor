import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import { BrowserRouter, Routes, Route, Navigate } from "react-router";
import LandingPage from "./pages/LandingPage.tsx";
import ContactPage from "./pages/ContactPage.tsx";
import ProfilePage from "./pages/ProfilePage.tsx";
import { ThemeProvider } from "./components/theme-provider.tsx";
import {
    // useQuery,
    // useMutation,
    // useQueryClient,
    QueryClient,
    QueryClientProvider,
} from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";

import { AuthProvider } from "./components/auth-context.tsx";
import { useAuth } from "./hooks/useAuth.tsx";
import StreamDetailsPage from "./pages/StreamPage.tsx";
import AdminPage from "./pages/AdminPage.tsx";
import AuthPage from "./pages/AuthPage.tsx";

const queryClient = new QueryClient({
    defaultOptions: {
        queries: {
            staleTime: 5 * 60 * 1000, // 5 minutes
            // cacheTime: 10 * 60 * 1000, // 10 minutes
            retry: 3,
            retryDelay: (attemptIndex) =>
                Math.min(1000 * 2 ** attemptIndex, 30000),
        },
    },
});

function ProtectedRoute({ children }: { children: React.ReactNode }) {
    const { isAuthenticated, isLoading } = useAuth();

    if (isLoading) {
        // You can show a global spinner here
        return <div>Loading...</div>;
    }

    if (!isAuthenticated) {
        return <Navigate to="/auth" replace />;
    }

    return <>{children}</>;
}

createRoot(document.getElementById("root")!).render(
    <StrictMode>
        <QueryClientProvider client={queryClient}>
            <BrowserRouter>
                <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
                    <AuthProvider>
                        <Routes>
                            <Route path="/" element={<LandingPage />} />
                            <Route path="/auth" element={<AuthPage />} />

                            <Route path="/contact" element={<ContactPage />} />
                            <Route path="profile">
                                <Route
                                    path=":username"
                                    element={<ProfilePage />}
                                />
                            </Route>
                            <Route path="stream">
                                <Route
                                    path=":id"
                                    element={<StreamDetailsPage />}
                                />
                            </Route>
                            <Route
                                path="/admin"
                                element={
                                    <ProtectedRoute>
                                        <AdminPage />
                                    </ProtectedRoute>
                                }
                            />
                        </Routes>
                    </AuthProvider>
                </ThemeProvider>
            </BrowserRouter>
            <ReactQueryDevtools initialIsOpen={false} />
        </QueryClientProvider>
    </StrictMode>,
);
