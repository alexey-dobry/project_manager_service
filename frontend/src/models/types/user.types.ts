import type { BaseEntity } from './api.types';

/**
 * Роли в системе. Должны полностью совпадать со строковыми
 * значениями, используемыми бэкендом в JWT claims.
 */
export type UserRole = 'admin' | 'teacher' | 'student';

export interface User extends BaseEntity {
  email: string;
  fullName: string;
  role: UserRole;
  /** ID групп, в которых состоит пользователь */
  groupIds: string[];
  avatarUrl?: string;
  /** Кафедра/факультет (для преподавателей и студентов) */
  department?: string;
}

/** DTO для регистрации */
export interface RegisterDto {
  email: string;
  password: string;
  fullName: string;
  role: UserRole;
  groupId?: string;
}

/** DTO для входа */
export interface LoginDto {
  email: string;
  password: string;
}

/** Ответ на /auth/login */
export interface LoginResponse {
  accessToken: string;
  refreshToken: string;
  user: User;
}

/** Обновление профиля */
export interface UpdateUserDto {
  fullName?: string;
  email?: string;
  department?: string;
  avatarUrl?: string;
}

/** Смена пароля */
export interface ChangePasswordDto {
  currentPassword: string;
  newPassword: string;
}
