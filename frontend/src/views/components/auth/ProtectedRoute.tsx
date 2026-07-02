import type { ReactNode } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { tokenService } from '@controllers/services/token.service';
import { useAuthBootstrap } from '@controllers/hooks/useAuth';
import { useAuthStore } from '@controllers/stores/auth.store';
import { Loader } from '@views/components/common/Loader';
import { ROUTES } from '@config/routes';
import type { UserRole } from '@models/types/user.types';

interface ProtectedRouteProps {
  children: ReactNode;
  /** Список ролей, которым разрешён доступ. Если не указан — любой авторизованный. */
  allowedRoles?: UserRole[];
}

/**
 * Защита маршрутов:
 *  1. Нет токена  → редирект на /login (с сохранением пути для возврата)
 *  2. Токен есть, но профиль ещё грузится → лоадер
 *  3. Не подходящая роль → редирект на /dashboard
 */
export function ProtectedRoute({ children, allowedRoles }: ProtectedRouteProps) {
  const location = useLocation();
  const { isLoading } = useAuthBootstrap();
  const user = useAuthStore((s) => s.user);

  if (!tokenService.isAuthenticated()) {
    return <Navigate to={ROUTES.LOGIN} state={{ from: location }} replace />;
  }

  if (isLoading) {
    return <Loader full label="Загружаем сессию..." />;
  }

  if (allowedRoles && user && !allowedRoles.includes(user.role)) {
    return <Navigate to={ROUTES.DASHBOARD} replace />;
  }

  return <>{children}</>;
}
