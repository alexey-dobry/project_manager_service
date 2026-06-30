import { Link, Navigate } from 'react-router-dom';
import { GraduationCap } from 'lucide-react';
import { LoginForm } from '@views/components/auth/LoginForm';
import { tokenService } from '@controllers/services/token.service';
import { ROUTES } from '@config/routes';

export function LoginPage() {
  // Используем tokenService напрямую — не стор — чтобы избежать редиректа
  // из-за устаревшего isAuthenticated=true в localStorage после протухания токена.
  if (tokenService.isAuthenticated()) {
    return <Navigate to={ROUTES.DASHBOARD} replace />;
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-background via-muted/30 to-background p-4">
      <div className="w-full max-w-md animate-fade-in rounded-2xl border bg-card p-8 shadow-xl">
        <div className="mb-8 text-center">
          <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-primary text-primary-foreground">
            <GraduationCap className="h-7 w-7" />
          </div>
          <h1 className="text-2xl font-bold tracking-tight">Вход в StudentPM</h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Управление проектами для студенческих групп
          </p>
        </div>

        <LoginForm />

        <p className="mt-6 text-center text-sm text-muted-foreground">
          Нет аккаунта?{' '}
          <Link to={ROUTES.REGISTER} className="font-medium text-primary hover:underline">
            Зарегистрироваться
          </Link>
        </p>
      </div>
    </div>
  );
}
