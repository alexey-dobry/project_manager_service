import { Link, Navigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Loader2, GraduationCap } from 'lucide-react';

import { registerSchema, type RegisterFormValues } from '@models/schemas/auth.schema';
import { useAuth } from '@controllers/hooks/useAuth';
import { useGroupsList } from '@controllers/hooks/useGroups';
import { tokenService } from '@controllers/services/token.service';
import { Button } from '@views/components/ui/button';
import { Input } from '@views/components/ui/input';
import { Label } from '@views/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@views/components/ui/select';
import { ROUTES } from '@config/routes';
import { USER_ROLES } from '@utils/constants';

export function RegisterPage() {
  const { register: registerUser, isRegistering } = useAuth();

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    formState: { errors },
  } = useForm<RegisterFormValues>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      email: '',
      password: '',
      confirmPassword: '',
      fullName: '',
      role: 'student',
      groupId: '',
    },
  });

  const role = watch('role');

  // Загружаем список групп только если:
  // 1. роль = student (поле группы реально видно)
  // 2. пользователь авторизован (иначе запрос без токена → 401 → редирект на /login)
  // На странице регистрации пользователь НЕ авторизован, поэтому enabled=false.
  const isLoggedIn = tokenService.isAuthenticated();
  const { data: groupsData } = useGroupsList(
    { pageSize: 100 },
    role === 'student' && isLoggedIn,
  );

  // Если токен реально есть — редиректим на dashboard.
  // Используем tokenService, а НЕ useAuthStore.isAuthenticated, потому что
  // стор может содержать устаревший isAuthenticated=true из прошлой сессии.
  if (tokenService.isAuthenticated()) {
    return <Navigate to={ROUTES.DASHBOARD} replace />;
  }

  const onSubmit = (values: RegisterFormValues) => {
    registerUser({
      email: values.email,
      password: values.password,
      fullName: values.fullName,
      role: values.role,
      groupId: values.groupId || undefined,
    });
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-background via-muted/30 to-background p-4 py-10">
      <div className="w-full max-w-md animate-fade-in rounded-2xl border bg-card p-8 shadow-xl">
        <div className="mb-6 text-center">
          <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-primary text-primary-foreground">
            <GraduationCap className="h-7 w-7" />
          </div>
          <h1 className="text-2xl font-bold tracking-tight">Регистрация</h1>
          <p className="mt-2 text-sm text-muted-foreground">Создайте аккаунт в StudentPM</p>
        </div>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
          <div className="space-y-2">
            <Label htmlFor="fullName">ФИО</Label>
            <Input id="fullName" placeholder="Иванов Иван Иванович" {...register('fullName')} />
            {errors.fullName && (
              <p className="text-xs text-destructive">{errors.fullName.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="email">Email</Label>
            <Input id="email" type="email" placeholder="you@example.com" {...register('email')} />
            {errors.email && <p className="text-xs text-destructive">{errors.email.message}</p>}
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-2">
              <Label htmlFor="password">Пароль</Label>
              <Input
                id="password"
                type="password"
                autoComplete="new-password"
                {...register('password')}
              />
              {errors.password && (
                <p className="text-xs text-destructive">{errors.password.message}</p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="confirmPassword">Повтор</Label>
              <Input
                id="confirmPassword"
                type="password"
                autoComplete="new-password"
                {...register('confirmPassword')}
              />
              {errors.confirmPassword && (
                <p className="text-xs text-destructive">{errors.confirmPassword.message}</p>
              )}
            </div>
          </div>

          <div className="space-y-2">
            <Label>Роль</Label>
            <Select
              defaultValue="student"
              onValueChange={(v) => setValue('role', v as RegisterFormValues['role'])}
            >
              <SelectTrigger>
                <SelectValue placeholder="Выберите роль" />
              </SelectTrigger>
              <SelectContent>
                {Object.entries(USER_ROLES).map(([value, label]) => (
                  <SelectItem key={value} value={value}>
                    {label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {errors.role && <p className="text-xs text-destructive">{errors.role.message}</p>}
          </div>

          {role === 'student' && isLoggedIn && (
            <div className="space-y-2">
              <Label>Группа</Label>
              <Select onValueChange={(v) => setValue('groupId', v)}>
                <SelectTrigger>
                  <SelectValue placeholder="Выберите группу (опционально)" />
                </SelectTrigger>
                <SelectContent>
                  {groupsData?.items.map((g) => (
                    <SelectItem key={g.id} value={g.id}>
                      {g.name} · {g.faculty}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          )}

          <Button type="submit" className="w-full" disabled={isRegistering}>
            {isRegistering && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            {isRegistering ? 'Регистрация...' : 'Зарегистрироваться'}
          </Button>
        </form>

        <p className="mt-6 text-center text-sm text-muted-foreground">
          Уже есть аккаунт?{' '}
          <Link to={ROUTES.LOGIN} className="font-medium text-primary hover:underline">
            Войти
          </Link>
        </p>
      </div>
    </div>
  );
}
