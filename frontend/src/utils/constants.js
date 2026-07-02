/**
 * Текстовые константы и словари — на чистом JavaScript.
 *
 * Почему JS, а не TS: эти данные не нуждаются в проверке типов
 * (только строки), а отсутствие типизации делает их универсальнее
 * для импорта из любых JSON/i18n-файлов в будущем.
 */

export const USER_ROLES = {
  admin: 'Администратор',
  teacher: 'Преподаватель',
  student: 'Студент',
};

export const TASK_STATUSES = {
  todo: 'К выполнению',
  in_progress: 'В работе',
  done: 'Готово',
  blocked: 'Заблокировано',
};

export const TASK_STATUS_ORDER = ['todo', 'in_progress', 'done', 'blocked'];

export const TASK_PRIORITIES = {
  low: 'Низкий',
  medium: 'Средний',
  high: 'Высокий',
  critical: 'Критичный',
};

export const PRIORITY_COLORS = {
  low: 'bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300',
  medium: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300',
  high: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-300',
  critical: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300',
};

export const PROJECT_STATUSES = {
  planning: 'Планирование',
  active: 'Активный',
  on_hold: 'Приостановлен',
  completed: 'Завершён',
  archived: 'В архиве',
};

export const PROJECT_STATUS_COLORS = {
  planning: 'bg-slate-100 text-slate-700',
  active: 'bg-green-100 text-green-700',
  on_hold: 'bg-yellow-100 text-yellow-700',
  completed: 'bg-blue-100 text-blue-700',
  archived: 'bg-gray-100 text-gray-700',
};

export const STATUS_COLUMN_COLORS = {
  todo: '#94a3b8',
  in_progress: '#3b82f6',
  done: '#22c55e',
  blocked: '#ef4444',
};

export const FACULTIES = [
  'Информатики и вычислительной техники',
  'Прикладной математики',
  'Радиотехники',
  'Экономики и менеджмента',
  'Иностранных языков',
  'Гуманитарный',
];
