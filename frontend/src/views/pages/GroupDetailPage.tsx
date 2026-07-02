import { useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { ArrowLeft, GraduationCap, Pencil, UserMinus, UserPlus } from 'lucide-react';

import { useGroup, useRemoveMember, useUpdateGroup } from '@controllers/hooks/useGroups';
import { useProjectsList } from '@controllers/hooks/useProjects';
import { useAuthStore } from '@controllers/stores/auth.store';
import { Card, CardContent, CardHeader, CardTitle } from '@views/components/ui/card';
import { Button } from '@views/components/ui/button';
import { Avatar, AvatarFallback, AvatarImage } from '@views/components/ui/avatar';
import { Badge } from '@views/components/ui/badge';
import { Loader } from '@views/components/common/Loader';
import { ConfirmDialog } from '@views/components/common/ConfirmDialog';
import { GroupFormDialog } from '@views/components/groups/GroupFormDialog';
import { ROUTES, buildPath } from '@config/routes';
import { getInitials } from '@utils/formatters';
import { USER_ROLES, PROJECT_STATUSES } from '@utils/constants';

export function GroupDetailPage() {
  const { id } = useParams<{ id: string }>();
  const user = useAuthStore((s) => s.user);

  const { data: group, isLoading } = useGroup(id);
  const { data: projects } = useProjectsList({ groupId: id, pageSize: 50 });

  const [editOpen, setEditOpen] = useState(false);
  const [memberToRemove, setMemberToRemove] = useState<string | null>(null);

  const updateMut = useUpdateGroup(id ?? '');
  const removeMemberMut = useRemoveMember(id ?? '');

  if (isLoading) return <Loader full label="Загрузка группы..." />;
  if (!group) {
    return (
      <div className="py-12 text-center text-muted-foreground">Группа не найдена</div>
    );
  }

  const canEdit = user?.role === 'admin' || user?.id === group.leaderId;

  return (
    <div className="space-y-6">
      <Button asChild variant="ghost" size="sm">
        <Link to={ROUTES.GROUPS}>
          <ArrowLeft className="h-4 w-4" /> Назад к группам
        </Link>
      </Button>

      {/* Шапка */}
      <Card>
        <CardContent className="p-6">
          <div className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
            <div className="flex items-start gap-4">
              <div className="flex h-14 w-14 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary">
                <GraduationCap className="h-7 w-7" />
              </div>
              <div>
                <h1 className="text-2xl font-bold tracking-tight">{group.name}</h1>
                <p className="text-sm text-muted-foreground">
                  {group.faculty} · {group.course} курс
                </p>
                {group.description && (
                  <p className="mt-2 max-w-prose text-sm">{group.description}</p>
                )}
              </div>
            </div>
            {canEdit && (
              <Button variant="outline" onClick={() => setEditOpen(true)}>
                <Pencil className="h-4 w-4" /> Редактировать
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      <div className="grid gap-6 lg:grid-cols-3">
        {/* Участники */}
        <Card className="lg:col-span-1">
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="text-lg">Участники ({group.members.length})</CardTitle>
            {canEdit && (
              <Button size="sm" variant="outline">
                <UserPlus className="h-4 w-4" />
              </Button>
            )}
          </CardHeader>
          <CardContent className="space-y-2">
            {/* Лидер выделяем сверху */}
            <MemberRow
              user={group.leader}
              isLeader
              canRemove={false}
            />
            {group.members
              .filter((m) => m.id !== group.leaderId)
              .map((m) => (
                <MemberRow
                  key={m.id}
                  user={m}
                  canRemove={canEdit}
                  onRemove={() => setMemberToRemove(m.id)}
                />
              ))}
          </CardContent>
        </Card>

        {/* Проекты группы */}
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle className="text-lg">Проекты группы</CardTitle>
          </CardHeader>
          <CardContent>
            {!projects?.items.length ? (
              <p className="py-8 text-center text-sm text-muted-foreground">
                У группы пока нет проектов
              </p>
            ) : (
              <div className="space-y-2">
                {projects.items.map((p) => (
                  <Link
                    key={p.id}
                    to={buildPath(ROUTES.PROJECT_DETAIL, { id: p.id })}
                    className="block rounded-lg border p-3 transition-colors hover:bg-muted/30"
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div className="min-w-0">
                        <h4 className="truncate font-medium">{p.name}</h4>
                        <p className="line-clamp-1 text-xs text-muted-foreground">
                          {p.description}
                        </p>
                      </div>
                      <Badge variant="secondary">{PROJECT_STATUSES[p.status]}</Badge>
                    </div>
                    <p className="mt-2 text-xs text-muted-foreground">
                      Прогресс: {p.progress}% · {p.tasksCount.total} задач
                    </p>
                  </Link>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Диалог редактирования */}
      {canEdit && (
        <GroupFormDialog
          open={editOpen}
          onOpenChange={setEditOpen}
          mode="edit"
          initialValues={{
            name: group.name,
            description: group.description,
            faculty: group.faculty,
            course: group.course,
          }}
          onSubmit={(values) => {
            updateMut.mutate(values, { onSuccess: () => setEditOpen(false) });
          }}
          loading={updateMut.isPending}
        />
      )}

      {/* Подтверждение удаления участника */}
      <ConfirmDialog
        open={!!memberToRemove}
        onOpenChange={(o) => !o && setMemberToRemove(null)}
        title="Удалить участника?"
        description="Участник потеряет доступ к группе и её проектам."
        confirmText="Удалить"
        destructive
        loading={removeMemberMut.isPending}
        onConfirm={() => {
          if (memberToRemove)
            removeMemberMut.mutate(memberToRemove, {
              onSuccess: () => setMemberToRemove(null),
            });
        }}
      />
    </div>
  );
}

function MemberRow({
  user,
  isLeader,
  canRemove,
  onRemove,
}: {
  user: { id: string; fullName: string; role: 'admin' | 'teacher' | 'student'; avatarUrl?: string };
  isLeader?: boolean;
  canRemove: boolean;
  onRemove?: () => void;
}) {
  return (
    <div className="flex items-center gap-3 rounded-lg p-2 hover:bg-muted/30">
      <Avatar className="h-9 w-9">
        {user.avatarUrl && <AvatarImage src={user.avatarUrl} alt={user.fullName} />}
        <AvatarFallback>{getInitials(user.fullName)}</AvatarFallback>
      </Avatar>
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium">{user.fullName}</p>
        <p className="truncate text-xs text-muted-foreground">{USER_ROLES[user.role]}</p>
      </div>
      {isLeader && <Badge variant="default">Лидер</Badge>}
      {canRemove && (
        <Button size="icon" variant="ghost" onClick={onRemove} aria-label="Удалить участника">
          <UserMinus className="h-4 w-4 text-destructive" />
        </Button>
      )}
    </div>
  );
}
