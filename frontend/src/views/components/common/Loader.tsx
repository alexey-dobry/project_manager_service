import { Loader2 } from 'lucide-react';
import { cn } from '@utils/cn';

interface LoaderProps {
  className?: string;
  size?: number;
  /** Отображать на всю секцию по центру */
  full?: boolean;
  label?: string;
}

export function Loader({ className, size = 24, full = false, label }: LoaderProps) {
  const spinner = (
    <Loader2 className={cn('animate-spin text-muted-foreground', className)} size={size} />
  );

  if (full) {
    return (
      <div className="flex h-full min-h-[200px] flex-col items-center justify-center gap-2">
        {spinner}
        {label && <p className="text-sm text-muted-foreground">{label}</p>}
      </div>
    );
  }

  return spinner;
}
