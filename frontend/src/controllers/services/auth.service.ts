import { authApi } from '@models/api/auth.api';
import type { LoginDto, RegisterDto, User } from '@models/types/user.types';
import { tokenService } from './token.service';

/**
 * Сервис уровня бизнес-логики аутентификации.
 * Объединяет API-вызовы с управлением токенами и не зависит от React.
 */
class AuthService {
  async login(dto: LoginDto): Promise<User> {
    const { accessToken, refreshToken, user } = await authApi.login(dto);
    tokenService.setTokens(accessToken, refreshToken);
    return user;
  }

  async register(dto: RegisterDto): Promise<User> {
    const { accessToken, refreshToken, user } = await authApi.register(dto);
    tokenService.setTokens(accessToken, refreshToken);
    return user;
  }

  async logout(): Promise<void> {
    try {
      await authApi.logout();
    } catch {
      // Игнорируем ошибки сервера при логауте — главное очистить токены
    } finally {
      tokenService.clearTokens();
    }
  }

  async getCurrentUser(): Promise<User | null> {
    if (!tokenService.isAuthenticated()) return null;
    try {
      return await authApi.me();
    } catch {
      return null;
    }
  }
}

export const authService = new AuthService();
