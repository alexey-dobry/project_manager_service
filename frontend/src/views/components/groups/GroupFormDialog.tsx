import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Loader2 } from 'lucide-react';

import { groupSchema, type GroupFormValues } from '@models/schemas/project.schema';
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
import { FACULTIES } from '@utils/constants';

interface Props {
  open: boolean;
  onOpenChange: (v: boolean) => void;
  onSubmit: (values: GroupFormValues) => void;
  loading?: boolean;
  initialValues?: Partial<GroupFormValues>;
  mode?: 'create' | 'edit';
}

export function GroupFormDialog({
  open,
  onOpenChange,
  onSubmit,
  loading,
  initialValues,
  mode = 'create',
}: Props) {
  const {
    register,
    handleSubmit,
    setValue,
    reset,
    formState: { errors },
  } = useForm<GroupFormValues>({
    resolver: zodResolver(groupSchema),
    defaultValues: {
      name: '',
      description: '',
      faculty: '',
      course: 1,
      ...initialValues,
    },
  });

  // Сброс при открытии с новыми initialValues
  useEffect(() => {
    if (open) reset({ name: '', description: '', faculty: '', course: 1, ...initialValues });
  }, [open, initialValues, reset]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{mode === 'create' ? 'Новая группа' : 'Редактирование группы'}</DialogTitle>
          <DialogDescription>
            Укажите основные данные группы — её можно будет дополнить позже.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
          <div className="space-y-2">
            <Label htmlFor="name">Название</Label>
            <Input id="name" placeholder="ИВТ-21-1" {...register('name')} />
            {errors.name && <p className="text-xs text-destructive">{errors.name.message}</p>}
          </div>

          <div className="space-y-2">
            <Label htmlFor="description">Описание (необязательно)</Label>
            <Textarea id="description" rows={3} {...register('description')} />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-2">
              <Label>Факультет</Label>
              <Select
                defaultValue={initialValues?.faculty}
                onValueChange={(v) => setValue('faculty', v)}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Выберите" />
                </SelectTrigger>
                <SelectContent>
                  {FACULTIES.map((f) => (
                    <SelectItem key={f} value={f}>
                      {f}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {errors.faculty && (
                <p className="text-xs text-destructive">{errors.faculty.message}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="course">Курс</Label>
              <Input
                id="course"
                type="number"
                min={1}
                max={6}
                {...register('course', { valueAsNumber: true })}
              />
              {errors.course && (
                <p className="text-xs text-destructive">{errors.course.message}</p>
              )}
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
