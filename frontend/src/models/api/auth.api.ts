import { apiClient, unwrap } from './client';
import type {
  LoginDto,
  LoginResponse,
  RegisterDto,
  User,
  ChangePasswordDto,
  UpdateUserDto,
} from '../types/user.types';

/**
 * Тонкий слой над axios — никакого состояния, никаких побочных эффектов.
 * Только функции «запрос → данные».
 */
export const authApi = {
  login: async (dto: LoginDto): Promise<LoginResponse> => {
    const res = await apiClient.post<LoginResponse>('/auth/login', dto);
    return unwrap(res);
  },

  register: async (dto: RegisterDto): Promise<LoginResponse> => {
    const res = await apiClient.post<LoginResponse>('/auth/register', dto);
    return unwrap(res);
  },

  logout: async (): Promise<void> => {
    await apiClient.post('/auth/logout');
  },

  /** Получить текущего пользователя по токену */
  me: async (): Promise<User> => {
    const res = await apiClient.get<User>('/auth/me');
    return unwrap(res);
  },

  refresh: async (refreshToken: string): Promise<LoginResponse> => {
    const res = await apiClient.post<LoginResponse>('/auth/refresh', { refreshToken });
    return unwrap(res);
  },

  changePassword: async (dto: ChangePasswordDto): Promise<void> => {
    await apiClient.post('/auth/change-password', dto);
  },

  updateProfile: async (dto: UpdateUserDto): Promise<User> => {
    const res = await apiClient.patch<User>('/auth/me', dto);
    return unwrap(res);
  },
};
