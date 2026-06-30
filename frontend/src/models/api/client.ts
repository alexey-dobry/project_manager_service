import axios, {
  AxiosError,
  AxiosInstance,
  AxiosResponse,
  InternalAxiosRequestConfig,
} from 'axios';
import { tokenService } from '@controllers/services/token.service';
import type { ApiError } from '../types/api.types';

/**
 * Базовый axios-инстанс.
 *
 * Исправления относительно оригинала:
 * 1. VITE_API_PREFIX по умолчанию '' (пусто). Бэкенд-gateway использует
 *    пути /auth/register, /groups, /projects — без префикса /api/v1.
 *    В prod nginx режет /api/ → бэк получает /auth/register, /groups и т.д.
 *    VITE_API_URL для prod должен быть '/api', VITE_API_PREFIX=''.
 *
 * 2. Request interceptor конвертирует тело и query-параметры camelCase→snake_case.
 *    Бэкенд на Go ждёт snake_case (json:"full_name", json:"refresh_token" и т.д.).
 *
 * 3. Response interceptor конвертирует snake_case→camelCase, чтобы фронт
 *    мог работать с привычным JS-стилем (accessToken, createdAt, ...).
 *
 * 4. Нормализация ошибки читает { "error": { "code", "message" } } —
 *    именно такой формат отдаёт бэкенд student-pm.
 */

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
// ВАЖНО: пусто. Бэкенд не использует /api/v1 — маршруты /auth/*, /groups/*, ...
const API_PREFIX = import.meta.env.VITE_API_PREFIX ?? '';
const API_TIMEOUT = Number(import.meta.env.VITE_API_TIMEOUT) || 15000;

// ─── camelCase ↔ snake_case ────────────────────────────────────────────────────
// Встроено прямо здесь, чтобы не зависеть от отдельного модуля.

const toSnake = (s: string) =>
  s.replace(/([A-Z])/g, (_, c: string) => `_${c.toLowerCase()}`);

const toCamel = (s: string) =>
  s.replace(/_([a-z0-9])/g, (_, c: string) => c.toUpperCase());

const isPlain = (v: unknown): v is Record<string, unknown> => {
  if (v === null || typeof v !== 'object' || Array.isArray(v)) return false;
  if (v instanceof Date) return false;
  if (typeof File !== 'undefined' && v instanceof File) return false;
  if (typeof Blob !== 'undefined' && v instanceof Blob) return false;
  if (typeof FormData !== 'undefined' && v instanceof FormData) return false;
  return true;
};

const mapKeys = (val: unknown, fn: (k: string) => string): unknown => {
  if (Array.isArray(val)) return val.map((v) => mapKeys(v, fn));
  if (isPlain(val)) {
    const out: Record<string, unknown> = {};
    for (const [k, v] of Object.entries(val)) out[fn(k)] = mapKeys(v, fn);
    return out;
  }
  return val;
};

// ─── axios instance ────────────────────────────────────────────────────────────

export const apiClient: AxiosInstance = axios.create({
  baseURL: `${API_URL}${API_PREFIX}`,
  timeout: API_TIMEOUT,
  headers: {
    'Content-Type': 'application/json',
    Accept: 'application/json',
  },
});

// ─── request: Bearer + camelCase → snake_case ─────────────────────────────────

apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = tokenService.getAccessToken();
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    if (config.data != null) config.data = mapKeys(config.data, toSnake);
    if (config.params) config.params = mapKeys(config.params, toSnake);
    return config;
  },
  (err: AxiosError) => Promise.reject(err),
);

// ─── response: snake_case → camelCase + refresh-очередь + нормализация ошибок ─

interface QueuedRequest {
  resolve: (token: string) => void;
  reject: (error: unknown) => void;
}

let isRefreshing = false;
let failedQueue: QueuedRequest[] = [];

const processQueue = (error: unknown, token: string | null = null) => {
  failedQueue.forEach(({ resolve, reject }) =>
    error ? reject(error) : token ? resolve(token) : reject(new Error('no token')),
  );
  failedQueue = [];
};

// Формат ошибки бэкенда: { "error": { "code": "...", "message": "..." } }
interface BackendError {
  error?: { code?: string; message?: string; details?: Record<string, unknown> };
}

apiClient.interceptors.response.use(
  (response: AxiosResponse) => {
    if (response.data != null) response.data = mapKeys(response.data, toCamel);
    return response;
  },
  async (error: AxiosError<BackendError>) => {
    const original = error.config as InternalAxiosRequestConfig & { _retry?: boolean };

    if (error.response?.status === 401 && original && !original._retry) {
      if (isRefreshing) {
        return new Promise<string>((resolve, reject) =>
          failedQueue.push({ resolve, reject }),
        ).then((token) => {
          if (original.headers) original.headers.Authorization = `Bearer ${token}`;
          return apiClient(original);
        });
      }

      original._retry = true;
      isRefreshing = true;

      try {
        const refreshToken = tokenService.getRefreshToken();
        if (!refreshToken) throw new Error('No refresh token');

        // "голый" axios — не попадаем в рекурсию интерцептора.
        // Бэк ждёт { refresh_token: "..." } и отдаёт { access_token, refresh_token }.
        const { data } = await axios.post<{
          access_token: string;
          refresh_token: string;
        }>(`${API_URL}${API_PREFIX}/auth/refresh`, { refresh_token: refreshToken });

        tokenService.setTokens(data.access_token, data.refresh_token);
        processQueue(null, data.access_token);
        if (original.headers)
          original.headers.Authorization = `Bearer ${data.access_token}`;
        return apiClient(original);
      } catch (refreshErr) {
        processQueue(refreshErr);
        tokenService.clearTokens();
        if (window.location.pathname !== '/login') window.location.href = '/login';
        return Promise.reject(refreshErr);
      } finally {
        isRefreshing = false;
      }
    }

    // Нормализуем ошибку в ApiError.
    // Читаем { "error": { "code", "message" } } — формат student-pm бэкенда.
    const nested = error.response?.data?.error;
    const apiError: ApiError = {
      code: nested?.code ?? 'UNKNOWN_ERROR',
      message: nested?.message ?? error.message ?? 'Произошла ошибка',
      details: nested?.details,
      status: error.response?.status,
    };
    return Promise.reject(apiError);
  },
);

/** Хелпер: разворачивает { data: T } обёртку, если она есть. */
export const unwrap = <T>(response: AxiosResponse<T | { data: T }>): T => {
  const body = response.data as T | { data: T };
  if (body && typeof body === 'object' && 'data' in body) return (body as { data: T }).data;
  return body as T;
};
