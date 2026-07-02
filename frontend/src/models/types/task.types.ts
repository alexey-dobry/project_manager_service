import type { BaseEntity } from './api.types';
import type { User } from './user.types';

export type TaskStatus = 'todo' | 'in_progress' | 'done' | 'blocked';
export type TaskPriority = 'low' | 'medium' | 'high' | 'critical';

export interface Task extends BaseEntity {
  title: string;
  description?: string;
  status: TaskStatus;
  priority: TaskPriority;
  projectId: string;
  /** ID назначенного исполнителя */
  assigneeId?: string;
  /** Денормализованный объект исполнителя (если бэк отдаёт сразу) */
  assignee?: User;
  /** Дедлайн задачи */
  dueDate?: string;
  /** Порядок внутри колонки на доске */
  order: number;
  commentsCount: number;
  /** Метки/теги */
  labels?: string[];
}

export interface CreateTaskDto {
  title: string;
  description?: string;
  projectId: string;
  priority?: TaskPriority;
  assigneeId?: string;
  dueDate?: string;
}

export interface UpdateTaskDto extends Partial<CreateTaskDto> {
  status?: TaskStatus;
  order?: number;
}

/**
 * DTO для перемещения задачи на доске.
 * Отдельная команда — атомарность изменения колонки и порядка.
 */
export interface MoveTaskDto {
  taskId: string;
  newStatus: TaskStatus;
  newOrder: number;
}

export interface Comment extends BaseEntity {
  taskId: string;
  authorId: string;
  author?: User;
  text: string;
}

export interface CreateCommentDto {
  taskId: string;
  text: string;
}
