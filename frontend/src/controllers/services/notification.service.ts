import { toast } from 'sonner';

/**
 * Тонкая обёртка над sonner. Если в будущем поменяем библиотеку
 * (react-hot-toast, react-toastify), правки будут только тут.
 */
export const notificationService = {
  success: (message: string, description?: string) => {
    toast.success(message, { description });
  },
  error: (message: string, description?: string) => {
    toast.error(message, { description });
  },
  info: (message: string, description?: string) => {
    toast.info(message, { description });
  },
  warning: (message: string, description?: string) => {
    toast.warning(message, { description });
  },
  /** Promise-toast: автоматически меняет состояние по результату промиса */
  promise: <T>(
    promise: Promise<T>,
    messages: { loading: string; success: string; error: string },
  ) => {
    return toast.promise(promise, messages);
  },
};
