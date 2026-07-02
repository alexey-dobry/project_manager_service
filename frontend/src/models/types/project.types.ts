import type { BaseEntity } from './api.types';

export type ProjectStatus = 'planning' | 'active' | 'on_hold' | 'completed' | 'archived';

export interface Project extends BaseEntity {
  name: string;
  description: string;
  status: ProjectStatus;
  groupId: string;
  /** Создатель проекта */
  ownerId: string;
  /** Дедлайн проекта (ISO date) */
  deadline?: string;
  /** Прогресс 0..100, рассчитывается на бэкенде */
  progress: number;
  /** Количество задач по статусам */
  tasksCount: {
    total: number;
    todo: number;
    inProgress: number;
    done: number;
    blocked: number;
  };
}

export interface CreateProjectDto {
  name: string;
  description: string;
  groupId: string;
  deadline?: string;
}

export interface UpdateProjectDto extends Partial<CreateProjectDto> {
  status?: ProjectStatus;
}

export interface ProjectFilters {
  status?: ProjectStatus;
  groupId?: string;
  search?: string;
}
