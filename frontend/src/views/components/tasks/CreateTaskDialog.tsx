import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Loader2 } from 'lucide-react';

import { taskSchema, type TaskFormValues } from '@models/schemas/project.schema';
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
import { TASK_PRIORITIES } from '@utils/constants';

interface Props {
  open: boolean;
  onOpenChange: (v: boolean) => void;
  onSubmit: (values: TaskFormValues) => void;
  loading?: boolean;
}

export function CreateTaskDialog({ open, onOpenChange, onSubmit, loading }: Props) {
  const {
    register,
    handleSubmit,
    setValue,
    reset,
    formState: { errors },
  } = useForm<TaskFormValues>({
    resolver: zodResolver(taskSchema),
    defaultValues: { title: '', description: '', priority: 'medium' },
  });

  useEffect(() => {
    if (open) reset({ title: '', description: '', priority: 'medium' });
  }, [open, reset]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Новая задача</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
          <div className="space-y-2">
            <Label htmlFor="title">Заголовок</Label>
            <Input id="title" {...register('title')} autoFocus />
            {errors.title && <p className="text-xs text-destructive">{errors.title.message}</p>}
          </div>

          <div className="space-y-2">
            <Label htmlFor="description">Описание</Label>
            <Textarea id="description" rows={4} {...register('description')} />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-2">
              <Label>Приоритет</Label>
              <Select
                defaultValue="medium"
                onValueChange={(v) => setValue('priority', v as TaskFormValues['priority'])}
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
              <Label htmlFor="dueDate">Дедлайн</Label>
              <Input id="dueDate" type="date" {...register('dueDate')} />
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
              Создать
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
