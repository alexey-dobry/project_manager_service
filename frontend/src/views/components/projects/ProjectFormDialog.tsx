import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Loader2 } from 'lucide-react';

import { projectSchema, type ProjectFormValues } from '@models/schemas/project.schema';
import { useGroupsList } from '@controllers/hooks/useGroups';
import {
  Dialog,
  DialogContent,
  DialogDescription,
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

interface Props {
  open: boolean;
  onOpenChange: (v: boolean) => void;
  onSubmit: (values: ProjectFormValues) => void;
  loading?: boolean;
  initialValues?: Partial<ProjectFormValues>;
  mode?: 'create' | 'edit';
}

export function ProjectFormDialog({
  open,
  onOpenChange,
  onSubmit,
  loading,
  initialValues,
  mode = 'create',
}: Props) {
  const { data: groupsData } = useGroupsList({ pageSize: 100 });

  const {
    register,
    handleSubmit,
    setValue,
    reset,
    formState: { errors },
  } = useForm<ProjectFormValues>({
    resolver: zodResolver(projectSchema),
    defaultValues: {
      name: '',
      description: '',
      groupId: '',
      deadline: '',
      ...initialValues,
    },
  });

  useEffect(() => {
    if (open) {
      reset({
        name: '',
        description: '',
        groupId: '',
        deadline: '',
        ...initialValues,
      });
    }
  }, [open, initialValues, reset]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {mode === 'create' ? 'Новый проект' : 'Редактирование проекта'}
          </DialogTitle>
          <DialogDescription>
            Заполните основные параметры. Задачи можно будет добавить на странице проекта.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
          <div className="space-y-2">
            <Label htmlFor="name">Название</Label>
            <Input id="name" placeholder="Курсовой проект..." {...register('name')} />
            {errors.name && <p className="text-xs text-destructive">{errors.name.message}</p>}
          </div>

          <div className="space-y-2">
            <Label htmlFor="description">Описание</Label>
            <Textarea
              id="description"
              rows={4}
              placeholder="Что предстоит сделать..."
              {...register('description')}
            />
            {errors.description && (
              <p className="text-xs text-destructive">{errors.description.message}</p>
            )}
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-2">
              <Label>Группа</Label>
              <Select
                defaultValue={initialValues?.groupId}
                onValueChange={(v) => setValue('groupId', v)}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Выберите группу" />
                </SelectTrigger>
                <SelectContent>
                  {groupsData?.items.map((g) => (
                    <SelectItem key={g.id} value={g.id}>
                      {g.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {errors.groupId && (
                <p className="text-xs text-destructive">{errors.groupId.message}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="deadline">Дедлайн</Label>
              <Input id="deadline" type="date" {...register('deadline')} />
            </div>
          </div>

          <DialogFooter className="gap-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={loading}
            >
              Отмена
            </Button>
            <Button type="submit" disabled={loading}>
              {loading && <Loader2 className="h-4 w-4 animate-spin" />}
              {mode === 'create' ? 'Создать' : 'Сохранить'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
