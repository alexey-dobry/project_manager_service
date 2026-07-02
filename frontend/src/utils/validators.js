/**
 * Дополнительные валидаторы поверх Zod — для быстрых проверок
 * прямо в обработчиках без построения полной схемы.
 */

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export const isEmail = (value) => EMAIL_RE.test(String(value ?? '').trim());

export const isStrongPassword = (value) => {
  if (typeof value !== 'string' || value.length < 8) return false;
  if (!/[A-ZА-Я]/.test(value)) return false;
  if (!/[0-9]/.test(value)) return false;
  return true;
};

export const isUUID = (value) =>
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(String(value));

/** Дедлайн должен быть в будущем */
export const isFutureDate = (iso) => {
  if (!iso) return true; // дедлайн необязателен
  const date = new Date(iso);
  return !Number.isNaN(date.getTime()) && date.getTime() > Date.now();
};
