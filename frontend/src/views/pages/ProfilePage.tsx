import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation } from '@tanstack/react-query';
import { Loader2 } from 'lucide-react';

import {
  changePasswordSchema,
  type ChangePasswordFormValues,
} from '@models/schemas/auth.schema';
import { authApi } from '@models/api/auth.api';
import { useAuthStore } from '@controllers/stores/auth.store';
import { notificationService } from '@controllers/services/notification.service';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@views/components/ui/card';
import { Avatar, AvatarFallback, AvatarImage } from '@views/components/ui/avatar';
import { Button } from '@views/components/ui/button';
import { Input } from '@views/components/ui/input';
import { Label } from '@views/components/ui/label';
import { Badge } from '@views/components/ui/badge';
import { useGroupsList } from '@controllers/hooks/useGroups';
import { getInitials } from '@utils/formatters';
import { USER_ROLES } from '@utils/constants';

export function ProfilePage() {
  const user = useAuthStore((s) => s.user);
  const setUser = useAuthStore((s) => s.setUser);
  const { data: groups } = useGroupsList({ pageSize: 100 });

  // ---- Профиль ----
  const updateMut = useMutation({
    mutationFn: authApi.updateProfile,
    onSuccess: (u) => {
      setUser(u);
      notificationService.success('Профиль обновлён');
    },
    onError: (e: { message?: string }) => notificationService.error('Ошибка', e.message),
  });

  const profileForm = useForm({
    defaultValues: {
      fullName: user?.fullName ?? '',
      email: user?.email ?? '',
      department: user?.department ?? '',
    },
  });

  // ---- Смена пароля ----
  const passwordMut = useMutation({
    mutationFn: authApi.changePassword,
    onSuccess: () => {
      notificationService.success('Пароль обновлён');
      passwordForm.reset();
    },
    onError: (e: { message?: string }) =>
      notificationService.error('Не удалось сменить пароль', e.message),
  });

  const passwordForm = useForm<ChangePasswordFormValues>({
    resolver: zodResolver(changePasswordSchema),
    defaultValues: { currentPassword: '', newPassword: '', confirmPassword: '' },
  });

  if (!user) return null;

  const userGroups = groups?.items.filter((g) => user.groupIds?.includes(g.id)) ?? [];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight md:text-3xl">Профиль</h1>
        <p className="text-sm text-muted-foreground">
          Управляйте данными аккаунта и безопасностью
        </p>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        {/* Карточка пользователя */}
        <Card className="lg:col-span-1">
          <CardContent className="flex flex-col items-center p-6 text-center">
            <Avatar className="h-24 w-24">
              {user.avatarUrl && <AvatarImage src={user.avatarUrl} alt={user.fullName} />}
              <AvatarFallback className="text-2xl">{getInitials(user.fullName)}</AvatarFallback>
            </Avatar>
            <h2 className="mt-4 text-lg font-semibold">{user.fullName}</h2>
            <p className="text-sm text-muted-foreground">{user.email}</p>
            <Badge className="mt-3" variant="secondary">
              {USER_ROLES[user.role]}
            </Badge>

            {userGroups.length > 0 && (
              <div className="mt-6 w-full text-left">
                <p className="mb-2 text-xs font-medium uppercase text-muted-foreground">
                  Состою в группах
                </p>
                <div className="flex flex-wrap gap-1.5">
                  {userGroups.map((g) => (
                    <Badge key={g.id} variant="outline" className="font-normal">
                      {g.name}
                    </Badge>
                  ))}
                </div>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Редактирование данных */}
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle className="text-lg">Личные данные</CardTitle>
            <CardDescription>Обновите ФИО, email и кафедру</CardDescription>
          </CardHeader>
          <CardContent>
            <form
              onSubmit={profileForm.handleSubmit((v) => updateMut.mutate(v))}
              className="space-y-4"
            >
              <div className="space-y-2">
                <Label htmlFor="fullName">ФИО</Label>
                <Input id="fullName" {...profileForm.register('fullName')} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="email">Email</Label>
                <Input id="email" type="email" {...profileForm.register('email')} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="department">Кафедра</Label>
                <Input id="department" {...profileForm.register('department')} />
              </div>

              <div className="flex justify-end">
                <Button type="submit" disabled={updateMut.isPending}>
                  {updateMut.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
                  Сохранить
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>

        {/* Смена пароля */}
        <Card className="lg:col-span-2 lg:col-start-2">
          <CardHeader>
            <CardTitle className="text-lg">Смена пароля</CardTitle>
            <CardDescription>
              Используйте надёжный пароль (минимум 8 символов, цифра, заглавная буква)
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form
              onSubmit={passwordForm.handleSubmit((v) =>
                passwordMut.mutate({
                  currentPassword: v.currentPassword,
                  newPassword: v.newPassword,
                }),
              )}
              className="space-y-4"
              noValidate
            >
              <div className="space-y-2">
                <Label htmlFor="currentPassword">Текущий пароль</Label>
                <Input
                  id="currentPassword"
                  type="password"
                  autoComplete="current-password"
                  {...passwordForm.register('currentPassword')}
                />
                {passwordForm.formState.errors.currentPassword && (
                  <p className="text-xs text-destructive">
                    {passwordForm.formState.errors.currentPassword.message}
                  </p>
                )}
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-2">
                  <Label htmlFor="newPassword">Новый пароль</Label>
                  <Input
                    id="newPassword"
                    type="password"
                    autoComplete="new-password"
                    {...passwordForm.register('newPassword')}
                  />
                  {passwordForm.formState.errors.newPassword && (
                    <p className="text-xs text-destructive">
                      {passwordForm.formState.errors.newPassword.message}
                    </p>
                  )}
                </div>
                <div className="space-y-2">
                  <Label htmlFor="confirmPassword">Повторите</Label>
                  <Input
                    id="confirmPassword"
                    type="password"
                    autoComplete="new-password"
                    {...passwordForm.register('confirmPassword')}
                  />
                  {passwordForm.formState.errors.confirmPassword && (
                    <p className="text-xs text-destructive">
                      {passwordForm.formState.errors.confirmPassword.message}
                    </p>
                  )}
                </div>
              </div>

              <div className="flex justify-end">
                <Button type="submit" disabled={passwordMut.isPending}>
                  {passwordMut.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
                  Сменить пароль
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
