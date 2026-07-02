import { describe, expect, it, beforeEach } from 'vitest';
import { useAuthStore } from '@/controllers/stores/auth.store';
import type { User } from '@/models/types/user.types';

const mockUser: User = {
  id: '1',
  email: 'a@b.c',
  fullName: 'Тест Тестов',
  role: 'student',
  groupIds: [],
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-01T00:00:00Z',
};

describe('useAuthStore', () => {
  beforeEach(() => {
    useAuthStore.setState({ user: null, isAuthenticated: false, isLoading: false });
    localStorage.clear();
  });

  it('начальное состояние — неавторизованный', () => {
    const { user, isAuthenticated } = useAuthStore.getState();
    expect(user).toBeNull();
    expect(isAuthenticated).toBe(false);
  });

  it('setUser выставляет isAuthenticated', () => {
    useAuthStore.getState().setUser(mockUser);
    expect(useAuthStore.getState().user).toEqual(mockUser);
    expect(useAuthStore.getState().isAuthenticated).toBe(true);
  });

  it('setUser(null) сбрасывает аутентификацию', () => {
    useAuthStore.getState().setUser(mockUser);
    useAuthStore.getState().setUser(null);
    expect(useAuthStore.getState().isAuthenticated).toBe(false);
  });

  it('reset сбрасывает всё', () => {
    useAuthStore.getState().setUser(mockUser);
    useAuthStore.getState().setLoading(true);
    useAuthStore.getState().reset();

    const state = useAuthStore.getState();
    expect(state.user).toBeNull();
    expect(state.isAuthenticated).toBe(false);
    expect(state.isLoading).toBe(false);
  });
});
