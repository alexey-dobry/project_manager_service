/**
 * Сервис управления JWT-токенами.
 *
 * Выбор хранилища: localStorage.
 *
 * Альтернативы и почему отвергнуты для курсовой:
 *  - httpOnly cookie: безопаснее против XSS, но требует CORS-настройки
 *    бэкенда с credentials и настроенного домена. Усложняет курсовой проект.
 *  - sessionStorage: токен теряется при закрытии вкладки — UX хуже.
 *
 * Известные риски localStorage:
 *  - Уязвимость к XSS: любой JS на странице может прочитать токен.
 *    Митигируем санитизацией пользовательского контента и CSP-заголовками.
 *  - Не передаётся между поддоменами автоматически.
 *
 * Для production-системы рекомендуется перейти на httpOnly cookie
 * с CSRF-защитой.
 */

const ACCESS_TOKEN_KEY = 'studentpm_access_token';
const REFRESH_TOKEN_KEY = 'studentpm_refresh_token';

class TokenService {
  getAccessToken(): string | null {
    try {
      return localStorage.getItem(ACCESS_TOKEN_KEY);
    } catch {
      return null;
    }
  }

  getRefreshToken(): string | null {
    try {
      return localStorage.getItem(REFRESH_TOKEN_KEY);
    } catch {
      return null;
    }
  }

  setTokens(accessToken: string, refreshToken: string): void {
    try {
      localStorage.setItem(ACCESS_TOKEN_KEY, accessToken);
      localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken);
    } catch (error) {
      console.error('Failed to persist tokens:', error);
    }
  }

  clearTokens(): void {
    try {
      localStorage.removeItem(ACCESS_TOKEN_KEY);
      localStorage.removeItem(REFRESH_TOKEN_KEY);
    } catch {
      /* no-op */
    }
  }

  isAuthenticated(): boolean {
    return Boolean(this.getAccessToken());
  }

  /**
   * Простейшая декодировка payload без проверки подписи.
   * Для UI-логики (например, показать дату истечения) — этого достаточно;
   * безопасность обеспечивает бэкенд.
   */
  decodePayload<T = Record<string, unknown>>(token: string): T | null {
    try {
      const [, payload] = token.split('.');
      if (!payload) return null;
      const decoded = atob(payload.replace(/-/g, '+').replace(/_/g, '/'));
      return JSON.parse(decoded) as T;
    } catch {
      return null;
    }
  }
}

export const tokenService = new TokenService();
