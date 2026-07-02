import { apiClient, unwrap } from './client';
import type { PaginatedResponse, PaginationParams } from '../types/api.types';
import type {
  Group,
  GroupWithMembers,
  CreateGroupDto,
  UpdateGroupDto,
  GroupFilters,
} from '../types/group.types';

export const groupsApi = {
  list: async (
    params?: PaginationParams & GroupFilters,
  ): Promise<PaginatedResponse<Group>> => {
    const res = await apiClient.get<PaginatedResponse<Group>>('/groups', { params });
    return unwrap(res);
  },

  getById: async (id: string): Promise<GroupWithMembers> => {
    const res = await apiClient.get<GroupWithMembers>(`/groups/${id}`);
    return unwrap(res);
  },

  create: async (dto: CreateGroupDto): Promise<Group> => {
    const res = await apiClient.post<Group>('/groups', dto);
    return unwrap(res);
  },

  update: async (id: string, dto: UpdateGroupDto): Promise<Group> => {
    const res = await apiClient.patch<Group>(`/groups/${id}`, dto);
    return unwrap(res);
  },

  remove: async (id: string): Promise<void> => {
    await apiClient.delete(`/groups/${id}`);
  },

  addMember: async (groupId: string, userId: string): Promise<void> => {
    await apiClient.post(`/groups/${groupId}/members`, { userId });
  },

  removeMember: async (groupId: string, userId: string): Promise<void> => {
    await apiClient.delete(`/groups/${groupId}/members/${userId}`);
  },
};
