import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { groupsApi } from '@models/api/groups.api';
import { notificationService } from '@controllers/services/notification.service';
import type {
  CreateGroupDto,
  GroupFilters,
  UpdateGroupDto,
} from '@models/types/group.types';
import type { PaginationParams } from '@models/types/api.types';

const KEYS = {
  all: ['groups'] as const,
  list: (params?: unknown) => ['groups', 'list', params] as const,
  detail: (id: string) => ['groups', 'detail', id] as const,
};

/**
 * Список групп с пагинацией.
 *
 * @param params  — фильтры и пагинация
 * @param enabled — если false, запрос не отправляется (полезно на странице
 *                  регистрации, где пользователь ещё не авторизован)
 */
export function useGroupsList(
  params?: PaginationParams & GroupFilters,
  enabled = true,
) {
  return useQuery({
    queryKey: KEYS.list(params),
    queryFn: () => groupsApi.list(params),
    staleTime: 60_000,
    // Не делаем запрос, если enabled=false, а также не повторяем при 401 —
    // ошибка авторизации не означает "нужно попробовать снова".
    enabled,
    retry: (failureCount, error: { status?: number }) => {
      if (error?.status === 401 || error?.status === 403) return false;
      return failureCount < 2;
    },
  });
}

export function useGroup(id: string | undefined) {
  return useQuery({
    queryKey: KEYS.detail(id ?? ''),
    queryFn: () => groupsApi.getById(id as string),
    enabled: Boolean(id),
  });
}

export function useCreateGroup() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (dto: CreateGroupDto) => groupsApi.create(dto),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.all });
      notificationService.success('Группа создана');
    },
    onError: (e: { message?: string }) =>
      notificationService.error('Не удалось создать группу', e.message),
  });
}

export function useUpdateGroup(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (dto: UpdateGroupDto) => groupsApi.update(id, dto),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.detail(id) });
      qc.invalidateQueries({ queryKey: KEYS.all });
      notificationService.success('Группа обновлена');
    },
  });
}

export function useAddMember(groupId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (userId: string) => groupsApi.addMember(groupId, userId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.detail(groupId) });
      notificationService.success('Участник добавлен');
    },
  });
}

export function useRemoveMember(groupId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (userId: string) => groupsApi.removeMember(groupId, userId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.detail(groupId) });
      notificationService.success('Участник удалён');
    },
  });
}
