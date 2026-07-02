/**
 * Общие типы API: пагинация, ошибки, обёртки ответов.
 * Эти контракты должны соответствовать тому, что отдаёт API Gateway.
 */

/** Унифицированный формат ошибки от бэкенда */
export interface ApiError {
  code: string;
  message: string;
  details?: Record<string, unknown>;
  /** HTTP-статус, проставляется клиентом */
  status?: number;
}

/** Стандартная обёртка успешного ответа */
export interface ApiResponse<T> {
  data: T;
  meta?: {
    requestId?: string;
    timestamp?: string;
  };
}

/** Параметры запроса со страничной выдачей */
export interface PaginationParams {
  page?: number;
  pageSize?: number;
  sort?: string;
  order?: 'asc' | 'desc';
}

/** Ответ со страничной выдачей */
export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

/** Шаблонные поля любой сущности */
export interface BaseEntity {
  id: string;
  createdAt: string;
  updatedAt: string;
}

/** Параметры авторизации запроса */
export interface AuthTokens {
  accessToken: string;
  refreshToken: string;
  expiresAt?: number;
}
