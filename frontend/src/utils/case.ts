/**
 * Утилиты конверсии ключей объектов между camelCase (фронт) и snake_case (бэкенд).
 *
 * Зачем: бэкенд на Go использует JSON-теги в snake_case (`json:"full_name"`),
 * фронт — идиоматический camelCase (`fullName`). Чтобы не мапить руками каждый
 * DTO, делаем универсальную конверсию в interceptor'ах axios:
 *  - req body / params: camelCase → snake_case
 *  - res body: snake_case → camelCase
 *
 * Не трогаем строки, primitives, Date, File/Blob и массивы примитивов.
 */

type Plain = Record<string, unknown>;

const toSnake = (s: string): string =>
  s.replace(/([A-Z])/g, (_, c) => `_${c.toLowerCase()}`).replace(/^_/, '');

const toCamel = (s: string): string =>
  s.replace(/_([a-z0-9])/g, (_, c) => c.toUpperCase());

/** true для plain-объектов; false для Date, File, Blob, FormData, Array, null, primitive */
const isPlainObject = (v: unknown): v is Plain => {
  if (v === null || typeof v !== 'object') return false;
  if (Array.isArray(v)) return false;
  if (v instanceof Date) return false;
  if (typeof File !== 'undefined' && v instanceof File) return false;
  if (typeof Blob !== 'undefined' && v instanceof Blob) return false;
  if (typeof FormData !== 'undefined' && v instanceof FormData) return false;
  // Защита от классов с собственным toJSON
  const proto = Object.getPrototypeOf(v);
  return proto === Object.prototype || proto === null;
};

const transformKeys = (input: unknown, mapKey: (k: string) => string): unknown => {
  if (Array.isArray(input)) return input.map((v) => transformKeys(v, mapKey));
  if (isPlainObject(input)) {
    const out: Plain = {};
    for (const [k, v] of Object.entries(input)) {
      out[mapKey(k)] = transformKeys(v, mapKey);
    }
    return out;
  }
  return input;
};

export const keysToSnake = (input: unknown): unknown => transformKeys(input, toSnake);
export const keysToCamel = (input: unknown): unknown => transformKeys(input, toCamel);
