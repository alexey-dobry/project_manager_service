/**
 * Точка доступа к env-переменным.
 * Если переменная отсутствует — отдаём осмысленное значение по умолчанию.
 */

const getEnv = (key, fallback = '') => {
  const value = import.meta.env[key];
  if (value === undefined || value === '') return fallback;
  return value;
};

export const ENV = {
  API_URL: getEnv('VITE_API_URL', 'http://localhost:8080'),
  API_PREFIX: getEnv('VITE_API_PREFIX', '/api/v1'),
  APP_ENV: getEnv('VITE_APP_ENV', 'development'),
  API_TIMEOUT: Number(getEnv('VITE_API_TIMEOUT', 15000)),
  IS_DEV: getEnv('VITE_APP_ENV', 'development') === 'development',
  IS_PROD: getEnv('VITE_APP_ENV', 'development') === 'production',
};
