import { Link } from 'react-router-dom';
import { Home, Compass } from 'lucide-react';
import { Button } from '@views/components/ui/button';
import { ROUTES } from '@config/routes';

export function NotFoundPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-6">
      <div className="max-w-md text-center">
        <div className="mx-auto mb-6 flex h-16 w-16 items-center justify-center rounded-2xl bg-muted">
          <Compass className="h-8 w-8 text-muted-foreground" />
        </div>
        <h1 className="text-6xl font-bold tracking-tighter">404</h1>
        <h2 className="mt-2 text-xl font-semibold">Страница не найдена</h2>
        <p className="mt-3 text-sm text-muted-foreground">
          Возможно, страница была удалена или вы перешли по неверной ссылке.
        </p>
        <Button asChild className="mt-6">
          <Link to={ROUTES.DASHBOARD}>
            <Home className="h-4 w-4" /> На главную
          </Link>
        </Button>
      </div>
    </div>
  );
}
