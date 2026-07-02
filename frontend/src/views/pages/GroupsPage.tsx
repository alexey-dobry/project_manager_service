import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Search, Plus, Users, GraduationCap } from 'lucide-react';

import { useGroupsList, useCreateGroup } from '@controllers/hooks/useGroups';
import { Button } from '@views/components/ui/button';
import { Input } from '@views/components/ui/input';
import { Card, CardContent } from '@views/components/ui/card';
import { Skeleton } from '@views/components/ui/skeleton';
import { EmptyState } from '@views/components/common/EmptyState';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@views/components/ui/select';
import { GroupFormDialog } from '@views/components/groups/GroupFormDialog';
import { useAuthStore } from '@controllers/stores/auth.store';
import { ROUTES, buildPath } from '@config/routes';
import { FACULTIES } from '@utils/constants';

export function GroupsPage() {
  const user = useAuthStore((s) => s.user);
  const canCreate = user?.role === 'admin' || user?.role === 'teacher';

  const [search, setSearch] = useState('');
  const [faculty, setFaculty] = useState<string>('');
  const [course, setCourse] = useState<string>('');
  const [isCreateOpen, setCreateOpen] = useState(false);

  const { data, isLoading } = useGroupsList({
    search: search || undefined,
    faculty: faculty || undefined,
    course: course ? Number(course) : undefined,
    pageSize: 50,
  });

  const createMut = useCreateGroup();

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight md:text-3xl">Группы</h1>
          <p className="text-sm text-muted-foreground">
            Студенческие группы и их проекты
          </p>
        </div>
        {canCreate && (
          <Button onClick={() => setCreateOpen(true)}>
            <Plus className="h-4 w-4" />
            Создать группу
          </Button>
        )}
      </div>

      {/* Фильтры */}
      <div className="flex flex-col gap-3 md:flex-row">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Поиск групп..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-9"
          />
        </div>
        <Select value={faculty} onValueChange={(v) => setFaculty(v === 'all' ? '' : v)}>
          <SelectTrigger className="md:w-72">
            <SelectValue placeholder="Все факультеты" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">Все факультеты</SelectItem>
            {FACULTIES.map((f) => (
              <SelectItem key={f} value={f}>
                {f}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Select value={course} onValueChange={(v) => setCourse(v === 'all' ? '' : v)}>
          <SelectTrigger className="md:w-32">
            <SelectValue placeholder="Курс" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">Все курсы</SelectItem>
            {[1, 2, 3, 4, 5, 6].map((c) => (
              <SelectItem key={c} value={String(c)}>
                {c} курс
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* Сетка групп */}
      {isLoading ? (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <Skeleton key={i} className="h-40 w-full" />
          ))}
        </div>
      ) : !data?.items.length ? (
        <EmptyState
          icon={Users}
          title="Группы не найдены"
          description="Попробуйте сбросить фильтры или создайте новую группу."
        />
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {data.items.map((group) => (
            <Link
              key={group.id}
              to={buildPath(ROUTES.GROUP_DETAIL, { id: group.id })}
              className="group block"
            >
              <Card className="h-full transition-all hover:border-primary hover:shadow-md">
                <CardContent className="p-5">
                  <div className="mb-3 flex items-center gap-2">
                    <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
                      <GraduationCap className="h-5 w-5" />
                    </div>
                    <h3 className="truncate text-lg font-semibold transition-colors group-hover:text-primary">
                      {group.name}
                    </h3>
                  </div>
                  <p className="line-clamp-2 text-sm text-muted-foreground">
                    {group.description || group.faculty}
                  </p>
                  <div className="mt-4 flex items-center gap-4 text-xs text-muted-foreground">
                    <span>{group.course} курс</span>
                    <span>·</span>
                    <span>{group.membersCount} участников</span>
                    <span>·</span>
                    <span>{group.projectsCount} проектов</span>
                  </div>
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      )}

      {/* Диалог создания */}
      <GroupFormDialog
        open={isCreateOpen}
        onOpenChange={setCreateOpen}
        onSubmit={(values) => {
          createMut.mutate(values, {
            onSuccess: () => setCreateOpen(false),
          });
        }}
        loading={createMut.isPending}
      />
    </div>
  );
}
