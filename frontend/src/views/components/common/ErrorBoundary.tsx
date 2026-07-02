import { Component, type ErrorInfo, type ReactNode } from 'react';
import { AlertTriangle, RefreshCw } from 'lucide-react';
import { Button } from '@views/components/ui/button';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

/**
 * Класс-компонент: error boundary в React работает только через классы.
 * Хуков-аналога пока нет.
 */
export class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false, error: null };

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    // В проде здесь можно слать в Sentry / собственный logger
    console.error('ErrorBoundary поймал ошибку:', error, info);
  }

  handleReset = () => {
    this.setState({ hasError: false, error: null });
    window.location.reload();
  };

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) return this.props.fallback;

      return (
        <div className="flex min-h-screen items-center justify-center bg-background p-6">
          <div className="max-w-md rounded-xl border bg-card p-8 text-center shadow-sm">
            <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-destructive/10">
              <AlertTriangle className="h-6 w-6 text-destructive" />
            </div>
            <h2 className="mt-4 text-lg font-semibold">Что-то пошло не так</h2>
            <p className="mt-2 text-sm text-muted-foreground">
              {this.state.error?.message || 'Произошла непредвиденная ошибка.'}
            </p>
            <Button onClick={this.handleReset} className="mt-6">
              <RefreshCw className="h-4 w-4" />
              Перезагрузить
            </Button>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
