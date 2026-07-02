import { describe, expect, it, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

import { LoginForm } from '@/views/components/auth/LoginForm';

// Мокаем сервис: тестируем форму, а не сетевой слой
vi.mock('@/controllers/services/auth.service', () => ({
  authService: {
    login: vi.fn().mockResolvedValue({
      id: '1',
      email: 'a@b.c',
      fullName: 'Test',
      role: 'student',
    }),
  },
}));

// Заглушка react-router navigate
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
  return {
    ...actual,
    useNavigate: () => vi.fn(),
  };
});

function renderForm() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>
        <LoginForm />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('LoginForm', () => {
  beforeEach(() => localStorage.clear());

  it('рендерит поля email и пароль', () => {
    renderForm();
    expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/пароль/i)).toBeInTheDocument();
  });

  it('показывает ошибку валидации при пустом email', async () => {
    renderForm();
    const user = userEvent.setup();
    await user.click(screen.getByRole('button', { name: /войти/i }));

    expect(await screen.findByText(/введите email/i)).toBeInTheDocument();
  });

  it('показывает ошибку при коротком пароле', async () => {
    renderForm();
    const user = userEvent.setup();

    await user.type(screen.getByLabelText(/email/i), 'a@b.c');
    await user.type(screen.getByLabelText(/пароль/i), '12');
    await user.click(screen.getByRole('button', { name: /войти/i }));

    expect(await screen.findByText(/минимум 6 символов/i)).toBeInTheDocument();
  });
});
