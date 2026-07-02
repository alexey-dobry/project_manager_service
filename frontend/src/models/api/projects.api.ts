import { apiClient, unwrap } from './client';
import type { PaginatedResponse, PaginationParams } from '../types/api.types';
import type {
  Project,
  CreateProjectDto,
  UpdateProjectDto,
  ProjectFilters,
} from '../types/project.types';

export const projectsApi = {
  list: async (
    params?: PaginationParams & ProjectFilters,
  ): Promise<PaginatedResponse<Project>> => {
    const res = await apiClient.get<PaginatedResponse<Project>>('/projects', { params });
    return unwrap(res);
  },

  getById: async (id: string): Promise<Project> => {
    const res = await apiClient.get<Project>(`/projects/${id}`);
    return unwrap(res);
  },

  create: async (dto: CreateProjectDto): Promise<Project> => {
    const res = await apiClient.post<Project>('/projects', dto);
    return unwrap(res);
  },

  update: async (id: string, dto: UpdateProjectDto): Promise<Project> => {
    const res = await apiClient.patch<Project>(`/projects/${id}`, dto);
    return unwrap(res);
  },

  remove: async (id: string): Promise<void> => {
    await apiClient.delete(`/projects/${id}`);
  },
};
