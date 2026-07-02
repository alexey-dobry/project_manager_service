import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useMutation, useQuery } from '@tanstack/react-query';
import { authService } from '@controllers/services/auth.service';
import { notificationService } from '@controllers/services/notification.service';
import { useAuthStore } from '@controllers/stores/auth.store';
import { tokenService } from '@controllers/services/token.service';
import type { LoginDto, RegisterDto } from '@models/types/user.types';

/**
 * Контроллер аутентификации.
 * Связывает View (формы) ↔ Service (бизнес-логика) ↔ Store (состояние).
 *
 * Принцип: компоненты вызывают только хуки отсюда, не лезут напрямую
 * в API или сервисы.
 */
export function useAuth() {
  const navigate = useNavigate();
  const { user, isAuthenticated, setUser, reset } = useAuthStore();

  // ---- Login ----
  const loginMutation = useMutation({
    mutationFn: (dto: LoginDto) => authService.login(dto),
    onSuccess: (user) => {
      setUser(user);
      notificationService.success('Вход выполнен', `Добро пожаловать, ${user.fullName}`);
      navigate('/dashboard');
    },
    onError: (error: { message?: string }) => {
      notificationService.error('Ошибка входа', error.message ?? 'Проверьте email и пароль');
    },
  });

  // ---- Register ----
  const registerMutation = useMutation({
    mutationFn: (dto: RegisterDto) => authService.register(dto),
    onSuccess: (user) => {
      setUser(user);
      notificationService.success('Регистрация прошла успешно');
      navigate('/dashboard');
    },
    onError: (error: { message?: string }) => {
      notificationService.error('Ошибка регистрации', error.message);
    },
  });

  // ---- Logout ----
  const logoutMutation = useMutation({
    mutationFn: () => authService.logout(),
    onSuccess: () => {
      reset();
      navigate('/login');
      notificationService.info('Вы вышли из системы');
    },
  });

  return {
    user,
    isAuthenticated,
    login: loginMutation.mutate,
    register: registerMutation.mutate,
    logout: logoutMutation.mutate,
    isLoggingIn: loginMutation.isPending,
    isRegistering: registerMutation.isPending,
  };
}

/**
 * Хук для запуска проверки сессии при монтировании App.
 * Если токен есть — пытается получить профиль.
 */
export function useAuthBootstrap() {
  const setUser = useAuthStore((s) => s.setUser);

  const query = useQuery({
    queryKey: ['auth', 'me'],
    queryFn: () => authService.getCurrentUser(),
    enabled: tokenService.isAuthenticated(),
    staleTime: 1000 * 60 * 5,
    retry: false,
  });

  useEffect(() => {
    if (query.data) setUser(query.data);
    if (query.isError) setUser(null);
  }, [query.data, query.isError, setUser]);

  return { isLoading: query.isLoading };
}
