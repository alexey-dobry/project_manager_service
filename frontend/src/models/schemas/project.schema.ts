import { z } from 'zod';

export const projectSchema = z.object({
  name: z
    .string()
    .min(3, 'Название проекта должно быть от 3 символов')
    .max(120, 'Слишком длинное название'),
  description: z
    .string()
    .min(10, 'Опишите проект подробнее (минимум 10 символов)')
    .max(2000, 'Описание слишком длинное'),
  groupId: z.string().min(1, 'Выберите группу'),
  deadline: z.string().optional(),
});

export type ProjectFormValues = z.infer<typeof projectSchema>;

export const taskSchema = z.object({
  title: z
    .string()
    .min(3, 'Заголовок задачи должен быть от 3 символов')
    .max(150, 'Заголовок слишком длинный'),
  description: z.string().max(5000, 'Описание слишком длинное').optional(),
  priority: z.enum(['low', 'medium', 'high', 'critical']),
  assigneeId: z.string().optional(),
  dueDate: z.string().optional(),
});

export type TaskFormValues = z.infer<typeof taskSchema>;

export const groupSchema = z.object({
  name: z.string().min(2, 'Название группы').max(80),
  description: z.string().max(500).optional(),
  faculty: z.string().min(2, 'Укажите факультет').max(120),
  course: z.coerce
    .number()
    .int()
    .min(1, 'Курс от 1')
    .max(6, 'Курс до 6'),
});

export type GroupFormValues = z.infer<typeof groupSchema>;

export const commentSchema = z.object({
  text: z
    .string()
    .min(1, 'Сообщение не может быть пустым')
    .max(2000, 'Слишком длинное сообщение'),
});

export type CommentFormValues = z.infer<typeof commentSchema>;
