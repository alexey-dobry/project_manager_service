import { NavLink } from 'react-router-dom';
import {
  LayoutDashboard,
  Users,
  FolderKanban,
  UserCircle,
  GraduationCap,
  type LucideIcon,
} from 'lucide-react';
import { cn } from '@utils/cn';
import { useAuthStore } from '@controllers/stores/auth.store';
import { useUIStore } from '@controllers/stores/ui.store';
import { ROUTES } from '@config/routes';
import type { UserRole } from '@models/types/user.types';

const ICONS: Record<string, LucideIcon> = {
  LayoutDashboard,
  Users,
  FolderKanban,
  UserCircle,
};

interface NavItem {
  path: string;
  label: string;
  icon: keyof typeof ICONS;
  roles: UserRole[];
}

const NAV: NavItem[] = [
  { path: ROUTES.DASHBOARD, label: 'Дашборд', icon: 'LayoutDashboard', roles: ['admin', 'teacher', 'student'] },
  { path: ROUTES.GROUPS, label: 'Группы', icon: 'Users', roles: ['admin', 'teacher', 'student'] },
  { path: ROUTES.PROJECTS, label: 'Проекты', icon: 'FolderKanban', roles: ['admin', 'teacher', 'student'] },
  { path: ROUTES.PROFILE, label: 'Профиль', icon: 'UserCircle', roles: ['admin', 'teacher', 'student'] },
];

export function Sidebar() {
  const user = useAuthStore((s) => s.user);
  const isOpen = useUIStore((s) => s.isSidebarOpen);

  const visibleNav = NAV.filter((item) => !user || item.roles.includes(user.role));

  return (
    <aside
      className={cn(
        'sticky top-0 flex h-screen shrink-0 flex-col border-r bg-card transition-all duration-200',
        isOpen ? 'w-64' : 'w-16',
      )}
    >
      {/* Логотип */}
      <div className="flex h-16 items-center gap-2 border-b px-4">
        <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-primary text-primary-foreground">
          <GraduationCap className="h-5 w-5" />
        </div>
        {isOpen && (
          <span className="truncate text-base font-semibold tracking-tight">StudentPM</span>
        )}
      </div>

      <nav className="flex-1 space-y-1 px-2 py-4">
        {visibleNav.map((item) => {
          const Icon = ICONS[item.icon];
          return (
            <NavLink
              key={item.path}
              to={item.path}
              end={item.path === ROUTES.DASHBOARD}
              className={({ isActive }) =>
                cn(
                  'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
                  isActive
                    ? 'bg-primary text-primary-foreground'
                    : 'text-muted-foreground hover:bg-accent hover:text-foreground',
                  !isOpen && 'justify-center',
                )
              }
              title={!isOpen ? item.label : undefined}
            >
              <Icon className="h-5 w-5 shrink-0" />
              {isOpen && <span>{item.label}</span>}
            </NavLink>
          );
        })}
      </nav>

      {isOpen && user && (
        <div className="border-t p-4 text-xs text-muted-foreground">
          <p className="truncate">Вы вошли как</p>
          <p className="truncate font-semibold text-foreground">{user.fullName}</p>
        </div>
      )}
    </aside>
  );
}
