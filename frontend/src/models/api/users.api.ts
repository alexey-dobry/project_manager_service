import { apiClient, unwrap } from './client';
import type { PaginatedResponse, PaginationParams } from '../types/api.types';
import type { User } from '../types/user.types';

export const usersApi = {
  list: async (params?: PaginationParams & { search?: string }): Promise<PaginatedResponse<User>> => {
    const res = await apiClient.get<PaginatedResponse<User>>('/users', { params });
    return unwrap(res);
  },

  getById: async (id: string): Promise<User> => {
    const res = await apiClient.get<User>(`/users/${id}`);
    return unwrap(res);
  },

  searchByEmail: async (email: string): Promise<User[]> => {
    const res = await apiClient.get<User[]>('/users/search', { params: { email } });
    return unwrap(res);
  },
};
