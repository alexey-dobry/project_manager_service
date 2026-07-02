import { useState, type FormEvent } from 'react';
import { Loader2 } from 'lucide-react';

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

interface Props {
  open: boolean;
  onOpenChange: (v: boolean) => void;
  /** Резолвится/реджектится по итогу поиска+добавления — диалог сам решает,
   *  что показать (ошибку под полем или закрыться при успехе). */
  onSubmit: (email: string) => Promise<void>;
  loading?: boolean;
}

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export function AddMemberDialog({ open, onOpenChange, onSubmit, loading }: Props) {
  const [email, setEmail] = useState('');
  const [error, setError] = useState<string | null>(null);

  const close = (v: boolean) => {
    if (!v) {
      setEmail('');
      setError(null);
    }
    onOpenChange(v);
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    const trimmed = email.trim();
    if (!EMAIL_RE.test(trimmed)) {
      setError('Введите корректный email');
      return;
    }
    setError(null);
    try {
      await onSubmit(trimmed);
      close(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось добавить участника');
    }
  };

  return (
    <Dialog open={open} onOpenChange={close}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Добавить участника</DialogTitle>
          <DialogDescription>
            Укажите email пользователя — он должен быть уже зарегистрирован в системе.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4" noValidate>
          <div className="space-y-2">
            <Label htmlFor="member-email">Email</Label>
            <Input
              id="member-email"
              type="email"
              placeholder="student@university.edu"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              autoFocus
            />
            {error && <p className="text-xs text-destructive">{error}</p>}
          </div>

          <DialogFooter className="gap-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => close(false)}
              disabled={loading}
            >
              Отмена
            </Button>
            <Button type="submit" disabled={loading}>
              {loading && <Loader2 className="h-4 w-4 animate-spin" />}
              Добавить
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
