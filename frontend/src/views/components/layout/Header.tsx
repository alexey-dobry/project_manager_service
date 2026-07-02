import { Menu, LogOut, Sun, Moon, Monitor } from 'lucide-react';
import { Button } from '@views/components/ui/button';
import { Avatar, AvatarFallback, AvatarImage } from '@views/components/ui/avatar';
import { useAuth } from '@controllers/hooks/useAuth';
import { useUIStore } from '@controllers/stores/ui.store';
import { getInitials } from '@utils/formatters';
import { USER_ROLES } from '@utils/constants';

export function Header() {
  const { user, logout } = useAuth();
  const { toggleSidebar, theme, setTheme } = useUIStore();

  const cycleTheme = () => {
    const next = theme === 'light' ? 'dark' : theme === 'dark' ? 'system' : 'light';
    setTheme(next);
  };

  const ThemeIcon = theme === 'light' ? Sun : theme === 'dark' ? Moon : Monitor;

  return (
    <header className="sticky top-0 z-30 flex h-16 items-center justify-between gap-4 border-b bg-card/80 px-4 backdrop-blur md:px-6">
      <Button variant="ghost" size="icon" onClick={toggleSidebar} aria-label="Меню">
        <Menu className="h-5 w-5" />
      </Button>

      <div className="flex items-center gap-3">
        <Button
          variant="ghost"
          size="icon"
          onClick={cycleTheme}
          title={`Тема: ${theme}`}
          aria-label="Переключить тему"
        >
          <ThemeIcon className="h-4 w-4" />
        </Button>

        {user && (
          <>
            <div className="hidden text-right md:block">
              <p className="text-sm font-medium leading-tight">{user.fullName}</p>
              <p className="text-xs text-muted-foreground">{USER_ROLES[user.role]}</p>
            </div>
            <Avatar className="h-9 w-9">
              {user.avatarUrl && <AvatarImage src={user.avatarUrl} alt={user.fullName} />}
              <AvatarFallback>{getInitials(user.fullName)}</AvatarFallback>
            </Avatar>
            <Button variant="ghost" size="icon" onClick={() => logout()} aria-label="Выйти">
              <LogOut className="h-4 w-4" />
            </Button>
          </>
        )}
      </div>
    </header>
  );
}
