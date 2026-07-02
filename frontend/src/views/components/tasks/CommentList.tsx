import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Loader2, Send } from 'lucide-react';

import { commentSchema, type CommentFormValues } from '@models/schemas/project.schema';
import { useTaskComments, useCreateComment } from '@controllers/hooks/useTasks';
import { Avatar, AvatarFallback, AvatarImage } from '@views/components/ui/avatar';
import { Textarea } from '@views/components/ui/textarea';
import { Button } from '@views/components/ui/button';
import { Loader } from '@views/components/common/Loader';
import { formatRelative, getInitials } from '@utils/formatters';

interface Props {
  taskId: string;
}

export function CommentList({ taskId }: Props) {
  const { data: comments, isLoading } = useTaskComments(taskId);
  const createMut = useCreateComment(taskId);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CommentFormValues>({
    resolver: zodResolver(commentSchema),
    defaultValues: { text: '' },
  });

  const onSubmit = (values: CommentFormValues) => {
    createMut.mutate(
      { taskId, text: values.text },
      {
        onSuccess: () => reset({ text: '' }),
      },
    );
  };

  return (
    <div className="space-y-4">
      <h3 className="text-sm font-semibold">Комментарии</h3>

      {isLoading ? (
        <Loader full />
      ) : !comments?.length ? (
        <p className="rounded-lg bg-muted/30 p-4 text-center text-sm text-muted-foreground">
          Комментариев пока нет
        </p>
      ) : (
        <div className="space-y-3">
          {comments.map((c) => (
            <div key={c.id} className="flex gap-3">
              <Avatar className="h-8 w-8 shrink-0">
                {c.author?.avatarUrl && <AvatarImage src={c.author.avatarUrl} />}
                <AvatarFallback className="text-xs">
                  {getInitials(c.author?.fullName ?? '?')}
                </AvatarFallback>
              </Avatar>
              <div className="min-w-0 flex-1 rounded-lg bg-muted/40 p-3">
                <div className="mb-1 flex items-baseline justify-between gap-2">
                  <span className="text-sm font-medium">
                    {c.author?.fullName ?? 'Пользователь'}
                  </span>
                  <span className="text-xs text-muted-foreground">
                    {formatRelative(c.createdAt)}
                  </span>
                </div>
                <p className="whitespace-pre-wrap break-words text-sm">{c.text}</p>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Форма комментария */}
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-2">
        <Textarea
          placeholder="Оставьте комментарий..."
          rows={2}
          {...register('text')}
          aria-invalid={!!errors.text}
        />
        {errors.text && <p className="text-xs text-destructive">{errors.text.message}</p>}
        <div className="flex justify-end">
          <Button type="submit" size="sm" disabled={createMut.isPending}>
            {createMut.isPending ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <Send className="h-4 w-4" />
            )}
            Отправить
          </Button>
        </div>
      </form>
    </div>
  );
}
