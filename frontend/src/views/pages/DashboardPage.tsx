import { Link } from 'react-router-dom';
import {
  FolderKanban,
  ListTodo,
  Clock,
  CheckCircle2,
  AlertCircle,
  ArrowRight,
} from 'lucide-react';

import { useAuthStore } from '@controllers/stores/auth.store';
import { useProjectsList } from '@controllers/hooks/useProjects';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@views/components/ui/card';
import { Progress } from '@views/components/ui/progress';
import { Badge } from '@views/components/ui/badge';
import { Skeleton } from '@views/components/ui/skeleton';
import { Button } from '@views/components/ui/button';
import { ROUTES, buildPath } from '@config/routes';
import { PROJECT_STATUSES } from '@utils/constants';
import { formatRelative, formatDate } from '@utils/formatters';

export function DashboardPage() {
  const user = useAuthStore((s) => s.user);
  const { data: projectsPage, isLoading } = useProjectsList({ pageSize: 6 });

  const projects = projectsPage?.items ?? [];

  // Агрегаты для карточек статистики
  const stats = projects.reduce(
    (acc, p) => {
      acc.total += 1;
      acc.todo += p.tasksCount.todo;
      acc.inProgress += p.tasksCount.inProgress;
      acc.done += p.tasksCount.done;
      acc.blocked += p.tasksCount.blocked;
      if (p.deadline && new Date(p.deadline).getTime() < Date.now() + 7 * 24 * 3600 * 1000) {
        acc.upcomingDeadlines += 1;
      }
      return acc;
    },
    { total: 0, todo: 0, inProgress: 0, done: 0, blocked: 0, upcomingDeadlines: 0 },
  );

  return (
    <div className="space-y-6">
      {/* Приветствие */}
      <div>
        <h1 className="text-2xl font-bold tracking-tight md:text-3xl">
          Здравствуйте, {user?.fullName.split(' ')[1] ?? user?.fullName ?? 'студент'} 👋
        </h1>
        <p className="text-sm text-muted-foreground">
          Сегодня {formatDate(new Date().toISOString())}. Вот общая картина по вашим проектам.
        </p>
      </div>

      {/* Карточки статистики */}
      <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
        <StatCard
          icon={<FolderKanban className="h-5 w-5" />}
          label="Проектов"
          value={isLoading ? '—' : stats.total}
        />
        <StatCard
          icon={<ListTodo className="h-5 w-5 text-blue-500" />}
          label="В работе"
          value={isLoading ? '—' : stats.inProgress}
        />
        <StatCard
          icon={<CheckCircle2 className="h-5 w-5 text-green-500" />}
          label="Готово"
          value={isLoading ? '—' : stats.done}
        />
        <StatCard
          icon={<Clock className="h-5 w-5 text-orange-500" />}
          label="Дедлайны на неделе"
          value={isLoading ? '—' : stats.upcomingDeadlines}
          highlight={stats.upcomingDeadlines > 0}
        />
      </div>

      {/* Активные проекты */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle className="text-xl">Активные проекты</CardTitle>
            <CardDescription>Последние обновления по вашим проектам</CardDescription>
          </div>
          <Button asChild variant="ghost" size="sm">
            <Link to={ROUTES.PROJECTS}>
              Все проекты <ArrowRight className="h-4 w-4" />
            </Link>
          </Button>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-3">
              {Array.from({ length: 3 }).map((_, i) => (
                <Skeleton key={i} className="h-24 w-full" />
              ))}
            </div>
          ) : projects.length === 0 ? (
            <div className="py-8 text-center text-sm text-muted-foreground">
              У вас пока нет активных проектов
            </div>
          ) : (
            <div className="grid gap-3 md:grid-cols-2">
              {projects.slice(0, 6).map((p) => (
                <Link
                  key={p.id}
                  to={buildPath(ROUTES.PROJECT_DETAIL, { id: p.id })}
                  className="group rounded-lg border p-4 transition-all hover:border-primary hover:shadow-md"
                >
                  <div className="mb-2 flex items-start justify-between gap-2">
                    <h3 className="font-medium leading-tight transition-colors group-hover:text-primary">
                      {p.name}
                    </h3>
                    <Badge variant="secondary">{PROJECT_STATUSES[p.status]}</Badge>
                  </div>
                  <p className="mb-3 line-clamp-2 text-sm text-muted-foreground">
                    {p.description}
                  </p>
                  <div className="space-y-1.5">
                    <div className="flex items-center justify-between text-xs text-muted-foreground">
                      <span>Прогресс</span>
                      <span className="font-medium text-foreground">{p.progress}%</span>
                    </div>
                    <Progress value={p.progress} />
                  </div>
                  {p.deadline && (
                    <p className="mt-2 text-xs text-muted-foreground">
                      Дедлайн: {formatRelative(p.deadline)}
                    </p>
                  )}
                </Link>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Быстрые действия */}
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <QuickAction
          title="Создать проект"
          description="Начните новый учебный проект для группы"
          icon={<FolderKanban className="h-5 w-5" />}
          to={ROUTES.PROJECTS}
        />
        <QuickAction
          title="Просмотр групп"
          description="Управляйте составом и проектами групп"
          icon={<AlertCircle className="h-5 w-5" />}
          to={ROUTES.GROUPS}
        />
      </div>
    </div>
  );
}

function StatCard({
  icon,
  label,
  value,
  highlight,
}: {
  icon: React.ReactNode;
  label: string;
  value: string | number;
  highlight?: boolean;
}) {
  return (
    <Card className={highlight ? 'border-orange-300 bg-orange-50/50 dark:bg-orange-950/20' : ''}>
      <CardContent className="flex items-center gap-3 p-4">
        <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-muted">
          {icon}
        </div>
        <div>
          <p className="text-xs text-muted-foreground">{label}</p>
          <p className="text-xl font-bold leading-tight">{value}</p>
        </div>
      </CardContent>
    </Card>
  );
}

function QuickAction({
  title,
  description,
  icon,
  to,
}: {
  title: string;
  description: string;
  icon: React.ReactNode;
  to: string;
}) {
  return (
    <Link
      to={to}
      className="flex items-center gap-4 rounded-xl border bg-card p-4 transition-all hover:border-primary hover:shadow-sm"
    >
      <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10 text-primary">
        {icon}
      </div>
      <div className="flex-1">
        <h3 className="font-semibold">{title}</h3>
        <p className="text-sm text-muted-foreground">{description}</p>
      </div>
      <ArrowRight className="h-4 w-4 text-muted-foreground" />
    </Link>
  );
}
