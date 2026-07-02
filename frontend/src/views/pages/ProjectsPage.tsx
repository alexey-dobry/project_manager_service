import { useState } from 'react';
import { Plus, Search, FolderKanban } from 'lucide-react';

import { useCreateProject, useProjectsList } from '@controllers/hooks/useProjects';
import { useGroupsList } from '@controllers/hooks/useGroups';
import { Button } from '@views/components/ui/button';
import { Input } from '@views/components/ui/input';
import { Skeleton } from '@views/components/ui/skeleton';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@views/components/ui/select';
import { EmptyState } from '@views/components/common/EmptyState';
import { ProjectCard } from '@views/components/projects/ProjectCard';
import { ProjectFormDialog } from '@views/components/projects/ProjectFormDialog';
import { PROJECT_STATUSES } from '@utils/constants';
import type { ProjectStatus } from '@models/types/project.types';

export function ProjectsPage() {
  const [search, setSearch] = useState('');
  const [status, setStatus] = useState<ProjectStatus | ''>('');
  const [groupId, setGroupId] = useState<string>('');
  const [createOpen, setCreateOpen] = useState(false);

  const { data: groups } = useGroupsList({ pageSize: 100 });
  const { data, isLoading } = useProjectsList({
    search: search || undefined,
    status: status || undefined,
    groupId: groupId || undefined,
    pageSize: 50,
  });

  const createMut = useCreateProject();

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight md:text-3xl">Проекты</h1>
          <p className="text-sm text-muted-foreground">
            Все учебные проекты ваших групп
          </p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="h-4 w-4" /> Создать проект
        </Button>
      </div>

      {/* Фильтры */}
      <div className="flex flex-col gap-3 md:flex-row">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Поиск проектов..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-9"
          />
        </div>
        <Select
          value={status}
          onValueChange={(v) => setStatus(v === 'all' ? '' : (v as ProjectStatus))}
        >
          <SelectTrigger className="md:w-48">
            <SelectValue placeholder="Все статусы" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">Все статусы</SelectItem>
            {Object.entries(PROJECT_STATUSES).map(([k, v]) => (
              <SelectItem key={k} value={k}>
                {v}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Select value={groupId} onValueChange={(v) => setGroupId(v === 'all' ? '' : v)}>
          <SelectTrigger className="md:w-56">
            <SelectValue placeholder="Все группы" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">Все группы</SelectItem>
            {groups?.items.map((g) => (
              <SelectItem key={g.id} value={g.id}>
                {g.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* Сетка проектов */}
      {isLoading ? (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <Skeleton key={i} className="h-52 w-full" />
          ))}
        </div>
      ) : !data?.items.length ? (
        <EmptyState
          icon={FolderKanban}
          title="Проектов пока нет"
          description="Создайте первый проект для своей группы."
          action={
            <Button onClick={() => setCreateOpen(true)}>
              <Plus className="h-4 w-4" /> Создать проект
            </Button>
          }
        />
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {data.items.map((p) => (
            <ProjectCard key={p.id} project={p} />
          ))}
        </div>
      )}

      <ProjectFormDialog
        open={createOpen}
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
