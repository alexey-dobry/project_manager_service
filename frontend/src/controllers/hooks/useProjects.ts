import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { projectsApi } from '@models/api/projects.api';
import { notificationService } from '@controllers/services/notification.service';
import type {
  ProjectFilters,
  CreateProjectDto,
  UpdateProjectDto,
} from '@models/types/project.types';
import type { PaginationParams } from '@models/types/api.types';

const KEYS = {
  all: ['projects'] as const,
  list: (params?: unknown) => ['projects', 'list', params] as const,
  detail: (id: string) => ['projects', 'detail', id] as const,
};

/** Список проектов с фильтрами и пагинацией */
export function useProjectsList(params?: PaginationParams & ProjectFilters) {
  return useQuery({
    queryKey: KEYS.list(params),
    queryFn: () => projectsApi.list(params),
    staleTime: 30_000,
  });
}

/** Детальный проект */
export function useProject(id: string | undefined) {
  return useQuery({
    queryKey: KEYS.detail(id ?? ''),
    queryFn: () => projectsApi.getById(id as string),
    enabled: Boolean(id),
  });
}

/** Создание проекта */
export function useCreateProject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (dto: CreateProjectDto) => projectsApi.create(dto),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.all });
      notificationService.success('Проект создан');
    },
    onError: (e: { message?: string }) =>
      notificationService.error('Не удалось создать проект', e.message),
  });
}

/** Обновление проекта */
export function useUpdateProject(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (dto: UpdateProjectDto) => projectsApi.update(id, dto),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.detail(id) });
      qc.invalidateQueries({ queryKey: KEYS.all });
      notificationService.success('Проект обновлён');
    },
    onError: (e: { message?: string }) =>
      notificationService.error('Не удалось обновить проект', e.message),
  });
}

/** Удаление проекта */
export function useDeleteProject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => projectsApi.remove(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.all });
      notificationService.success('Проект удалён');
    },
    onError: (e: { message?: string }) =>
      notificationService.error('Не удалось удалить проект', e.message),
  });
}
