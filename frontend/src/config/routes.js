/**
 * Декларативное описание маршрутов в одном месте.
 * Удобно для:
 *  - построения navbar/sidebar
 *  - программного редиректа (route paths не разбросаны по магическим строкам)
 *  - генерации breadcrumbs
 */

export const ROUTES = {
  HOME: '/',
  LOGIN: '/login',
  REGISTER: '/register',
  DASHBOARD: '/dashboard',
  GROUPS: '/groups',
  GROUP_DETAIL: '/groups/:id',
  PROJECTS: '/projects',
  PROJECT_DETAIL: '/projects/:id',
  PROFILE: '/profile',
  NOT_FOUND: '*',
};

/**
 * Подстановка параметров: routePath('/groups/:id', { id: 'abc' }) → '/groups/abc'
 */
export const buildPath = (template, params = {}) =>
  Object.entries(params).reduce(
    (acc, [key, value]) => acc.replace(`:${key}`, encodeURIComponent(value)),
    template,
  );

/**
 * Маршруты для основной навигации (sidebar).
 * Поле roles — какие роли видят пункт.
 */
export const NAV_ITEMS = [
  {
    path: ROUTES.DASHBOARD,
    label: 'Дашборд',
    icon: 'LayoutDashboard',
    roles: ['admin', 'teacher', 'student'],
  },
  {
    path: ROUTES.GROUPS,
    label: 'Группы',
    icon: 'Users',
    roles: ['admin', 'teacher', 'student'],
  },
  {
    path: ROUTES.PROJECTS,
    label: 'Проекты',
    icon: 'FolderKanban',
    roles: ['admin', 'teacher', 'student'],
  },
  {
    path: ROUTES.PROFILE,
    label: 'Профиль',
    icon: 'UserCircle',
    roles: ['admin', 'teacher', 'student'],
  },
];
