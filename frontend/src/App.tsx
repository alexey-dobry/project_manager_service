import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Toaster } from 'sonner';

import { ROUTES } from '@config/routes';
import { ProtectedRoute } from '@views/components/auth/ProtectedRoute';
import { Layout } from '@views/components/layout/Layout';
import { ErrorBoundary } from '@views/components/common/ErrorBoundary';

import { LoginPage } from '@views/pages/LoginPage';
import { RegisterPage } from '@views/pages/RegisterPage';
import { DashboardPage } from '@views/pages/DashboardPage';
import { GroupsPage } from '@views/pages/GroupsPage';
import { GroupDetailPage } from '@views/pages/GroupDetailPage';
import { ProjectsPage } from '@views/pages/ProjectsPage';
import { ProjectDetailPage } from '@views/pages/ProjectDetailPage';
import { ProfilePage } from '@views/pages/ProfilePage';
import { NotFoundPage } from '@views/pages/NotFoundPage';

/**
 * Конфигурация TanStack Query.
 * - retry: 1 — лишний раз не штурмуем падающий бэк, но даём один шанс
 * - staleTime: 30s — данные считаются свежими полминуты, фоновых рефетчей меньше
 * - refetchOnWindowFocus: false — обычно раздражает в админках
 */
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      staleTime: 30_000,
      refetchOnWindowFocus: false,
    },
    mutations: {
      retry: 0,
    },
  },
});

function App() {
  return (
    <ErrorBoundary>
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <Routes>
            {/* Публичные маршруты */}
            <Route path={ROUTES.LOGIN} element={<LoginPage />} />
            <Route path={ROUTES.REGISTER} element={<RegisterPage />} />

            {/* Защищённые маршруты под общим Layout */}
            <Route
              element={
                <ProtectedRoute>
                  <Layout />
                </ProtectedRoute>
              }
            >
              <Route path={ROUTES.HOME} element={<Navigate to={ROUTES.DASHBOARD} replace />} />
              <Route path={ROUTES.DASHBOARD} element={<DashboardPage />} />
              <Route path={ROUTES.GROUPS} element={<GroupsPage />} />
              <Route path={ROUTES.GROUP_DETAIL} element={<GroupDetailPage />} />
              <Route path={ROUTES.PROJECTS} element={<ProjectsPage />} />
              <Route path={ROUTES.PROJECT_DETAIL} element={<ProjectDetailPage />} />
              <Route path={ROUTES.PROFILE} element={<ProfilePage />} />
            </Route>

            <Route path="*" element={<NotFoundPage />} />
          </Routes>
        </BrowserRouter>

        <Toaster position="top-right" richColors closeButton />
      </QueryClientProvider>
    </ErrorBoundary>
  );
}

export default App;
