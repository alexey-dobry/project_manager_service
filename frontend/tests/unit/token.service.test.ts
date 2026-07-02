import { describe, expect, it, beforeEach } from 'vitest';
import { tokenService } from '@/controllers/services/token.service';

describe('tokenService', () => {
  beforeEach(() => localStorage.clear());

  it('возвращает null когда токенов нет', () => {
    expect(tokenService.getAccessToken()).toBeNull();
    expect(tokenService.getRefreshToken()).toBeNull();
    expect(tokenService.isAuthenticated()).toBe(false);
  });

  it('сохраняет и читает пару токенов', () => {
    tokenService.setTokens('access-1', 'refresh-1');
    expect(tokenService.getAccessToken()).toBe('access-1');
    expect(tokenService.getRefreshToken()).toBe('refresh-1');
    expect(tokenService.isAuthenticated()).toBe(true);
  });

  it('очищает токены', () => {
    tokenService.setTokens('a', 'b');
    tokenService.clearTokens();
    expect(tokenService.getAccessToken()).toBeNull();
    expect(tokenService.isAuthenticated()).toBe(false);
  });

  it('декодирует payload JWT', () => {
    // header.payload.signature — собираем фейковый JWT
    const payload = { sub: 'user-1', role: 'student' };
    const encoded = btoa(JSON.stringify(payload));
    const fakeJwt = `header.${encoded}.signature`;

    const decoded = tokenService.decodePayload<typeof payload>(fakeJwt);
    expect(decoded).toEqual(payload);
  });

  it('возвращает null для невалидного токена', () => {
    expect(tokenService.decodePayload('not-a-jwt')).toBeNull();
    expect(tokenService.decodePayload('')).toBeNull();
  });
});
