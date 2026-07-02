import { Link } from 'react-router-dom';
import { CalendarClock, ListChecks } from 'lucide-react';
import { Card, CardContent } from '@views/components/ui/card';
import { Badge } from '@views/components/ui/badge';
import { Progress } from '@views/components/ui/progress';
import { ROUTES, buildPath } from '@config/routes';
import { PROJECT_STATUSES } from '@utils/constants';
import { formatDate } from '@utils/formatters';
import type { Project } from '@models/types/project.types';

interface Props {
  project: Project;
}

export function ProjectCard({ project }: Props) {
  return (
    <Link
      to={buildPath(ROUTES.PROJECT_DETAIL, { id: project.id })}
      className="group block transition-transform hover:-translate-y-0.5"
    >
      <Card className="h-full transition-all hover:border-primary hover:shadow-md">
        <CardContent className="flex h-full flex-col p-5">
          <div className="mb-3 flex items-start justify-between gap-2">
            <h3 className="line-clamp-2 font-semibold leading-tight transition-colors group-hover:text-primary">
              {project.name}
            </h3>
            <Badge variant="secondary" className="shrink-0">
              {PROJECT_STATUSES[project.status]}
            </Badge>
          </div>

          <p className="mb-4 line-clamp-2 text-sm text-muted-foreground">
            {project.description}
          </p>

          <div className="mt-auto space-y-3">
            <div>
              <div className="mb-1 flex items-center justify-between text-xs">
                <span className="text-muted-foreground">Прогресс</span>
                <span className="font-medium">{project.progress}%</span>
              </div>
              <Progress value={project.progress} />
            </div>

            <div className="flex items-center justify-between text-xs text-muted-foreground">
              <span className="inline-flex items-center gap-1">
                <ListChecks className="h-3.5 w-3.5" />
                {project.tasksCount.total} задач
              </span>
              {project.deadline && (
                <span className="inline-flex items-center gap-1">
                  <CalendarClock className="h-3.5 w-3.5" />
                  {formatDate(project.deadline, 'd MMM')}
                </span>
              )}
            </div>
          </div>
        </CardContent>
      </Card>
    </Link>
  );
}
