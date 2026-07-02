import type { BaseEntity } from './api.types';
import type { User } from './user.types';

export interface Group extends BaseEntity {
  name: string;
  description?: string;
  faculty: string;
  course: number; // 1..6
  /** ID пользователя-лидера группы */
  leaderId: string;
  membersCount: number;
  projectsCount: number;
}

/** Расширенная версия с предзагруженными участниками */
export interface GroupWithMembers extends Group {
  leader: User;
  members: User[];
}

export interface CreateGroupDto {
  name: string;
  description?: string;
  faculty: string;
  course: number;
}

export interface UpdateGroupDto extends Partial<CreateGroupDto> {
  leaderId?: string;
}

export interface GroupFilters {
  faculty?: string;
  course?: number;
  search?: string;
}
