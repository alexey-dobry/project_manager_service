import { z } from 'zod';

/**
 * Zod-схемы валидации форм аутентификации.
 * Используются и в react-hook-form (через @hookform/resolvers/zod),
 * и для статического вывода TypeScript-типов через z.infer<...>.
 */

export const loginSchema = z.object({
  email: z.string().min(1, 'Введите email').email('Некорректный email'),
  password: z.string().min(6, 'Пароль должен содержать минимум 6 символов'),
});

export type LoginFormValues = z.infer<typeof loginSchema>;

export const registerSchema = z
  .object({
    email: z.string().min(1, 'Введите email').email('Некорректный email'),
    password: z
      .string()
      .min(8, 'Минимум 8 символов')
      .regex(/[A-ZА-Я]/, 'Должна быть хотя бы одна заглавная буква')
      .regex(/[0-9]/, 'Должна быть хотя бы одна цифра'),
    confirmPassword: z.string().min(8, 'Подтвердите пароль'),
    fullName: z
      .string()
      .min(2, 'Введите ФИО')
      .max(100, 'Слишком длинное ФИО'),
    role: z.enum(['admin', 'teacher', 'student'], {
      errorMap: () => ({ message: 'Выберите роль' }),
    }),
    groupId: z.string().optional(),
  })
  .refine((d) => d.password === d.confirmPassword, {
    path: ['confirmPassword'],
    message: 'Пароли не совпадают',
  });

export type RegisterFormValues = z.infer<typeof registerSchema>;

export const changePasswordSchema = z
  .object({
    currentPassword: z.string().min(1, 'Введите текущий пароль'),
    newPassword: z
      .string()
      .min(8, 'Минимум 8 символов')
      .regex(/[A-ZА-Я]/, 'Должна быть хотя бы одна заглавная буква')
      .regex(/[0-9]/, 'Должна быть хотя бы одна цифра'),
    confirmPassword: z.string(),
  })
  .refine((d) => d.newPassword === d.confirmPassword, {
    path: ['confirmPassword'],
    message: 'Пароли не совпадают',
  });

export type ChangePasswordFormValues = z.infer<typeof changePasswordSchema>;
