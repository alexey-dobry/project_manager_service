import { useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { ArrowLeft, CalendarClock, Pencil, Users } from 'lucide-react';

import { useProject, useUpdateProject } from '@controllers/hooks/useProjects';
import { useProjectTasks } from '@controllers/hooks/useTasks';
import { useGroup } from '@controllers/hooks/useGroups';
import { Button } from '@views/components/ui/button';
import { Badge } from '@views/components/ui/badge';
import { Progress } from '@views/components/ui/progress';
import { Card, CardContent } from '@views/components/ui/card';
import { Loader } from '@views/components/common/Loader';
import { KanbanBoard } from '@views/components/projects/KanbanBoard';
import { ProjectFormDialog } from '@views/components/projects/ProjectFormDialog';
import { ROUTES, buildPath } from '@config/routes';
import { PROJECT_STATUSES, PROJECT_STATUS_COLORS } from '@utils/constants';
import { formatDate } from '@utils/formatters';

export function ProjectDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [editOpen, setEditOpen] = useState(false);

  const { data: project, isLoading: loadingProject } = useProject(id);
  const { data: tasks, isLoading: loadingTasks } = useProjectTasks(id);
  const { data: group } = useGroup(project?.groupId);

  const updateMut = useUpdateProject(id ?? '');

  if (loadingProject) return <Loader full label="Загрузка проекта..." />;
  if (!project) {
    return (
      <div className="py-12 text-center text-muted-foreground">Проект не найден</div>
    );
  }

  return (
    <div className="space-y-6">
      <Button asChild variant="ghost" size="sm">
        <Link to={ROUTES.PROJECTS}>
          <ArrowLeft className="h-4 w-4" /> К списку проектов
        </Link>
      </Button>

      {/* Шапка проекта */}
      <Card>
        <CardContent className="p-6">
          <div className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
            <div className="min-w-0 flex-1">
              <div className="mb-2 flex items-center gap-2">
                <Badge
                  className={PROJECT_STATUS_COLORS[project.status]}
                  variant="outline"
                >
                  {PROJECT_STATUSES[project.status]}
                </Badge>
                {group && (
                  <Link
                    to={buildPath(ROUTES.GROUP_DETAIL, { id: group.id })}
                    className="inline-flex items-center gap-1 text-xs text-muted-foreground hover:text-primary"
                  >
                    <Users className="h-3.5 w-3.5" />
                    {group.name}
                  </Link>
                )}
              </div>
              <h1 className="text-2xl font-bold tracking-tight md:text-3xl">{project.name}</h1>
              <p className="mt-2 max-w-prose text-sm text-muted-foreground">
                {project.description}
              </p>

              <div className="mt-4 flex flex-wrap gap-4 text-xs text-muted-foreground">
                {project.deadline && (
                  <span className="inline-flex items-center gap-1">
                    <CalendarClock className="h-3.5 w-3.5" />
                    Дедлайн: {formatDate(project.deadline)}
                  </span>
                )}
                <span>
                  Задач: {project.tasksCount.total} (готово: {project.tasksCount.done})
                </span>
              </div>

              <div className="mt-4 max-w-md">
                <div className="mb-1 flex items-center justify-between text-xs">
                  <span className="text-muted-foreground">Прогресс</span>
                  <span className="font-medium">{project.progress}%</span>
                </div>
                <Progress value={project.progress} />
              </div>
            </div>

            <Button variant="outline" onClick={() => setEditOpen(true)}>
              <Pencil className="h-4 w-4" /> Редактировать
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Kanban */}
      {loadingTasks ? (
        <Loader full label="Загрузка задач..." />
      ) : (
        <KanbanBoard
          projectId={project.id}
          groupId={project.groupId}
          tasks={tasks ?? []}
        />
      )}

      <ProjectFormDialog
        open={editOpen}
        onOpenChange={setEditOpen}
        mode="edit"
        initialValues={{
          name: project.name,
          description: project.description,
          groupId: project.groupId,
          deadline: project.deadline?.split('T')[0],
        }}
        onSubmit={(values) => {
          updateMut.mutate(values, { onSuccess: () => setEditOpen(false) });
        }}
        loading={updateMut.isPending}
      />
    </div>
  );
}
