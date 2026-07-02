import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { tasksApi } from '@models/api/tasks.api';
import { notificationService } from '@controllers/services/notification.service';
import type {
  Task,
  CreateTaskDto,
  UpdateTaskDto,
  MoveTaskDto,
  CreateCommentDto,
} from '@models/types/task.types';

const KEYS = {
  byProject: (projectId: string) => ['tasks', 'project', projectId] as const,
  detail: (id: string) => ['tasks', 'detail', id] as const,
  comments: (taskId: string) => ['tasks', taskId, 'comments'] as const,
};

/** Список задач проекта (для Kanban) */
export function useProjectTasks(projectId: string | undefined) {
  return useQuery({
    queryKey: KEYS.byProject(projectId ?? ''),
    queryFn: () => tasksApi.listByProject(projectId as string),
    enabled: Boolean(projectId),
    staleTime: 10_000,
  });
}

export function useCreateTask(projectId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (dto: CreateTaskDto) => tasksApi.create(dto),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.byProject(projectId) });
      notificationService.success('Задача создана');
    },
    onError: (e: { message?: string }) =>
      notificationService.error('Не удалось создать задачу', e.message),
  });
}

export function useUpdateTask(projectId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, dto }: { id: string; dto: UpdateTaskDto }) =>
      tasksApi.update(id, dto),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.byProject(projectId) });
    },
  });
}

/**
 * Перемещение задачи между колонками с оптимистичным апдейтом.
 *
 * Это ядро UX Kanban-доски: пользователь должен сразу видеть
 * перемещённую карточку, не дожидаясь сетевого ответа. Если
 * запрос упадёт — мы откатим состояние через onError.
 */
export function useMoveTask(projectId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (dto: MoveTaskDto) => tasksApi.move(dto),

    onMutate: async (dto) => {
      const queryKey = KEYS.byProject(projectId);
      // 1) Отменяем летящие запросы, чтобы не перетёрли наш оптимизм
      await qc.cancelQueries({ queryKey });

      // 2) Снимаем снапшот для отката
      const previous = qc.getQueryData<Task[]>(queryKey);

      // 3) Оптимистично применяем изменение
      if (previous) {
        const updated = previous.map((t) =>
          t.id === dto.taskId
            ? { ...t, status: dto.newStatus, order: dto.newOrder }
            : t,
        );
        qc.setQueryData<Task[]>(queryKey, updated);
      }

      return { previous };
    },

    onError: (err: { message?: string }, _dto, context) => {
      // Откат к снапшоту при ошибке
      if (context?.previous) {
        qc.setQueryData(KEYS.byProject(projectId), context.previous);
      }
      notificationService.error('Не удалось переместить задачу', err.message);
    },

    onSettled: () => {
      qc.invalidateQueries({ queryKey: KEYS.byProject(projectId) });
    },
  });
}

export function useDeleteTask(projectId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => tasksApi.remove(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.byProject(projectId) });
      notificationService.success('Задача удалена');
    },
  });
}

// ============= Комментарии =============
export function useTaskComments(taskId: string | undefined) {
  return useQuery({
    queryKey: KEYS.comments(taskId ?? ''),
    queryFn: () => tasksApi.listComments(taskId as string),
    enabled: Boolean(taskId),
  });
}

export function useCreateComment(taskId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (dto: CreateCommentDto) => tasksApi.createComment(dto),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.comments(taskId) });
    },
  });
}
