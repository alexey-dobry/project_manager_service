import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import type { User } from '@models/types/user.types';
import { tokenService } from '@controllers/services/token.service';

/**
 * Глобальное состояние аутентификации.
 *
 * ВАЖНО: isAuthenticated НЕ персистируется в localStorage.
 * Источник истины — tokenService.isAuthenticated() (наличие access-токена).
 * Если персистировать isAuthenticated, то после протухания/удаления токена
 * стор остаётся в состоянии isAuthenticated=true → RegisterPage редиректит
 * на /dashboard, хотя пользователь по факту не авторизован.
 */
interface AuthState {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;

  setUser: (user: User | null) => void;
  setLoading: (loading: boolean) => void;
  reset: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      isLoading: false,
      // При первом рендере синхронно проверяем наличие токена.
      isAuthenticated: tokenService.isAuthenticated(),

      setUser: (user) =>
        set({
          user,
          // Если user=null — считаем не авторизованным.
          // Если user задан — дополнительно проверяем, что токен реально есть.
          isAuthenticated: Boolean(user) && tokenService.isAuthenticated(),
        }),

      setLoading: (isLoading) => set({ isLoading }),

      reset: () =>
        set({
          user: null,
          isAuthenticated: false,
          isLoading: false,
        }),
    }),
    {
      name: 'studentpm-auth-store',
      storage: createJSONStorage(() => localStorage),
      // Персистируем только user — НЕ isAuthenticated.
      // isAuthenticated при гидрации вычислится из tokenService выше.
      partialize: (state) => ({ user: state.user }),
      // После загрузки из localStorage пересчитываем isAuthenticated
      // на основе реального наличия токена.
      onRehydrateStorage: () => (state) => {
        if (state) {
          state.isAuthenticated =
            Boolean(state.user) && tokenService.isAuthenticated();
        }
      },
    },
  ),
);
