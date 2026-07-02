import { apiClient, unwrap } from './client';
import type {
  Task,
  CreateTaskDto,
  UpdateTaskDto,
  MoveTaskDto,
  Comment,
  CreateCommentDto,
} from '../types/task.types';

export const tasksApi = {
  /** Получить все задачи проекта (для Kanban-доски) */
  listByProject: async (projectId: string): Promise<Task[]> => {
    const res = await apiClient.get<Task[]>(`/projects/${projectId}/tasks`);
    return unwrap(res);
  },

  getById: async (id: string): Promise<Task> => {
    const res = await apiClient.get<Task>(`/tasks/${id}`);
    return unwrap(res);
  },

  create: async (dto: CreateTaskDto): Promise<Task> => {
    const res = await apiClient.post<Task>('/tasks', dto);
    return unwrap(res);
  },

  update: async (id: string, dto: UpdateTaskDto): Promise<Task> => {
    const res = await apiClient.patch<Task>(`/tasks/${id}`, dto);
    return unwrap(res);
  },

  /**
   * Атомарное перемещение задачи между колонками.
   * Бэкенд внутри транзакции пересчитает order остальных задач.
   */
  move: async (dto: MoveTaskDto): Promise<Task> => {
    const res = await apiClient.post<Task>('/tasks/move', dto);
    return unwrap(res);
  },

  remove: async (id: string): Promise<void> => {
    await apiClient.delete(`/tasks/${id}`);
  },

  // -------- Комментарии --------
  listComments: async (taskId: string): Promise<Comment[]> => {
    const res = await apiClient.get<Comment[]>(`/tasks/${taskId}/comments`);
    return unwrap(res);
  },

  createComment: async (dto: CreateCommentDto): Promise<Comment> => {
    const res = await apiClient.post<Comment>(`/tasks/${dto.taskId}/comments`, {
      text: dto.text,
    });
    return unwrap(res);
  },

  removeComment: async (taskId: string, commentId: string): Promise<void> => {
    await apiClient.delete(`/tasks/${taskId}/comments/${commentId}`);
  },
};
