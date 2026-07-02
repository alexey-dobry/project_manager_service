import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Loader2, Trash2 } from 'lucide-react';

import { taskSchema, type TaskFormValues } from '@models/schemas/project.schema';
import { useUpdateTask, useDeleteTask } from '@controllers/hooks/useTasks';
import { useGroup } from '@controllers/hooks/useGroups';
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@views/components/ui/dialog';
import { Button } from '@views/components/ui/button';
import { Input } from '@views/components/ui/input';
import { Label } from '@views/components/ui/label';
import { Textarea } from '@views/components/ui/textarea';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@views/components/ui/select';
import { CommentList } from './CommentList';
import { ConfirmDialog } from '@views/components/common/ConfirmDialog';
import { TASK_PRIORITIES } from '@utils/constants';
import type { Task } from '@models/types/task.types';

interface Props {
  task: Task | null;
  projectId: string;
  groupId: string;
  open: boolean;
  onOpenChange: (v: boolean) => void;
}

/**
 * Модалка редактирования задачи + просмотра/добавления комментариев.
 * mode редактирования включается на тот же диалог, чтобы не плодить экраны.
 */
export function TaskModal({ task, projectId, groupId, open, onOpenChange }: Props) {
  const [confirmDelete, setConfirmDelete] = useState(false);
  const updateMut = useUpdateTask(projectId);
  const deleteMut = useDeleteTask(projectId);
  const { data: group } = useGroup(groupId);

  const {
    register,
    handleSubmit,
    setValue,
    formState: { errors, isDirty },
  } = useForm<TaskFormValues>({
    resolver: zodResolver(taskSchema),
    values: task
      ? {
          title: task.title,
          description: task.description ?? '',
          priority: task.priority,
          assigneeId: task.assigneeId ?? '',
          dueDate: task.dueDate ?? '',
        }
      : undefined,
  });

  if (!task) return null;

  const onSubmit = (values: TaskFormValues) => {
    updateMut.mutate(
      {
        id: task.id,
        dto: {
          title: values.title,
          description: values.description,
          priority: values.priority,
          assigneeId: values.assigneeId || undefined,
          dueDate: values.dueDate || undefined,
        },
      },
      { onSuccess: () => onOpenChange(false) },
    );
  };

  return (
    <>
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Задача</DialogTitle>
          </DialogHeader>

          <div className="grid gap-6 md:grid-cols-[1fr_280px]">
            {/* Левая колонка: форма */}
            <form
              id="task-form"
              onSubmit={handleSubmit(onSubmit)}
              className="space-y-4"
              noValidate
            >
              <div className="space-y-2">
                <Label htmlFor="title">Заголовок</Label>
                <Input id="title" {...register('title')} />
                {errors.title && (
                  <p className="text-xs text-destructive">{errors.title.message}</p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="description">Описание</Label>
                <Textarea id="description" rows={5} {...register('description')} />
              </div>
            </form>

            {/* Правая колонка: метаданные */}
            <div className="space-y-4">
              <div className="space-y-2">
                <Label>Приоритет</Label>
                <Select
                  defaultValue={task.priority}
                  onValueChange={(v) =>
                    setValue('priority', v as TaskFormValues['priority'], {
                      shouldDirty: true,
                    })
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {Object.entries(TASK_PRIORITIES).map(([k, v]) => (
                      <SelectItem key={k} value={k}>
                        {v}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label>Исполнитель</Label>
                <Select
                  defaultValue={task.assigneeId ?? ''}
                  onValueChange={(v) => setValue('assigneeId', v, { shouldDirty: true })}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Не назначен" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="">Не назначен</SelectItem>
                    {group?.members.map((m) => (
                      <SelectItem key={m.id} value={m.id}>
                        {m.fullName}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label htmlFor="dueDate">Дедлайн</Label>
                <Input id="dueDate" type="date" {...register('dueDate')} />
              </div>
            </div>
          </div>

          {/* Комментарии */}
          <div className="border-t pt-4">
            <CommentList taskId={task.id} />
          </div>

          <DialogFooter className="gap-2 sm:justify-between">
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => setConfirmDelete(true)}
              className="text-destructive hover:text-destructive"
            >
              <Trash2 className="h-4 w-4" /> Удалить
            </Button>
            <div className="flex gap-2">
              <Button variant="outline" onClick={() => onOpenChange(false)}>
                Отмена
              </Button>
              <Button
                type="submit"
                form="task-form"
                disabled={!isDirty || updateMut.isPending}
              >
                {updateMut.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
                Сохранить
              </Button>
            </div>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        open={confirmDelete}
        onOpenChange={setConfirmDelete}
        title="Удалить задачу?"
        description="Действие нельзя будет отменить. Все комментарии тоже будут удалены."
        confirmText="Удалить"
        destructive
        loading={deleteMut.isPending}
        onConfirm={() => {
          deleteMut.mutate(task.id, {
            onSuccess: () => {
              setConfirmDelete(false);
              onOpenChange(false);
            },
          });
        }}
      />
    </>
  );
}
