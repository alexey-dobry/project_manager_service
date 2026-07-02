import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

/**
 * Утилита для условного объединения tailwind-классов.
 * Используется в shadcn/ui-компонентах: разрешает конфликты классов
 * (например, `px-2` поверх `px-4`).
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}
