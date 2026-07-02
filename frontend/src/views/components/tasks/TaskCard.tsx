import { CalendarClock, MessageSquare, AlertCircle } from 'lucide-react';
import { cn } from '@utils/cn';
import { Avatar, AvatarFallback, AvatarImage } from '@views/components/ui/avatar';
import { Badge } from '@views/components/ui/badge';
import { PRIORITY_COLORS, TASK_PRIORITIES } from '@utils/constants';
import { formatDate, getInitials } from '@utils/formatters';
import type { Task } from '@models/types/task.types';

interface Props {
  task: Task;
  onClick?: () => void;
  isDragging?: boolean;
}

export function TaskCard({ task, onClick, isDragging }: Props) {
  const overdue =
    task.dueDate &&
    task.status !== 'done' &&
    new Date(task.dueDate).getTime() < Date.now();

  return (
    <div
      onClick={onClick}
      className={cn('kanban-card cursor-pointer', isDragging && 'dragging')}
    >
      <div className="mb-2 flex items-start justify-between gap-2">
        <h4 className="line-clamp-2 text-sm font-medium leading-tight">{task.title}</h4>
        <Badge
          variant="outline"
          className={cn('shrink-0 border-0 text-[10px]', PRIORITY_COLORS[task.priority])}
        >
          {TASK_PRIORITIES[task.priority]}
        </Badge>
      </div>

      {task.description && (
        <p className="mb-2 line-clamp-2 text-xs text-muted-foreground">{task.description}</p>
      )}

      {task.labels && task.labels.length > 0 && (
        <div className="mb-2 flex flex-wrap gap-1">
          {task.labels.slice(0, 3).map((l) => (
            <Badge key={l} variant="secondary" className="text-[10px]">
              {l}
            </Badge>
          ))}
        </div>
      )}

      <div className="flex items-center justify-between text-xs text-muted-foreground">
        <div className="flex items-center gap-2">
          {task.dueDate && (
            <span
              className={cn(
                'inline-flex items-center gap-1',
                overdue && 'font-medium text-destructive',
              )}
            >
              {overdue ? (
                <AlertCircle className="h-3 w-3" />
              ) : (
                <CalendarClock className="h-3 w-3" />
              )}
              {formatDate(task.dueDate, 'd MMM')}
            </span>
          )}
          {task.commentsCount > 0 && (
            <span className="inline-flex items-center gap-1">
              <MessageSquare className="h-3 w-3" />
              {task.commentsCount}
            </span>
          )}
        </div>

        {task.assignee && (
          <Avatar className="h-6 w-6">
            {task.assignee.avatarUrl && (
              <AvatarImage src={task.assignee.avatarUrl} alt={task.assignee.fullName} />
            )}
            <AvatarFallback className="text-[10px]">
              {getInitials(task.assignee.fullName)}
            </AvatarFallback>
          </Avatar>
        )}
      </div>
    </div>
  );
}
