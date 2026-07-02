import { format, formatDistanceToNow, isValid, parseISO } from 'date-fns';
import { ru } from 'date-fns/locale';

/**
 * Форматирование дат, чисел, имён и т.п.
 * Чистый JS — функции без сайд-эффектов и зависимостей от React.
 */

export const formatDate = (iso, fmt = 'd MMMM yyyy') => {
  if (!iso) return '—';
  const date = typeof iso === 'string' ? parseISO(iso) : iso;
  return isValid(date) ? format(date, fmt, { locale: ru }) : '—';
};

export const formatDateTime = (iso) => formatDate(iso, 'd MMM yyyy, HH:mm');

export const formatRelative = (iso) => {
  if (!iso) return '';
  const date = typeof iso === 'string' ? parseISO(iso) : iso;
  if (!isValid(date)) return '';
  return formatDistanceToNow(date, { addSuffix: true, locale: ru });
};

/** ФИО в инициалы: "Иванов Иван Иванович" → "ИИ" */
export const getInitials = (fullName) => {
  if (!fullName) return '?';
  return fullName
    .trim()
    .split(/\s+/)
    .slice(0, 2)
    .map((p) => p[0]?.toUpperCase() ?? '')
    .join('');
};

/** Безопасное укорачивание строки */
export const truncate = (str, maxLen = 60) => {
  if (!str) return '';
  return str.length > maxLen ? `${str.slice(0, maxLen - 1)}…` : str;
};

/** Псевдо-уникальный цвет по строке (для аватарок-плейсхолдеров) */
export const stringToColor = (str) => {
  if (!str) return '#94a3b8';
  let hash = 0;
  for (let i = 0; i < str.length; i++) hash = str.charCodeAt(i) + ((hash << 5) - hash);
  const palette = [
    '#ef4444', '#f97316', '#eab308', '#22c55e',
    '#06b6d4', '#3b82f6', '#8b5cf6', '#ec4899',
  ];
  return palette[Math.abs(hash) % palette.length];
};
