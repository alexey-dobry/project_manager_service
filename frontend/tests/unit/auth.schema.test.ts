import { describe, expect, it } from 'vitest';
import { loginSchema, registerSchema } from '@/models/schemas/auth.schema';

describe('loginSchema', () => {
  it('пропускает корректные данные', () => {
    const result = loginSchema.safeParse({
      email: 'user@example.com',
      password: 'pass1234',
    });
    expect(result.success).toBe(true);
  });

  it('падает на невалидном email', () => {
    const result = loginSchema.safeParse({ email: 'not-email', password: 'pass1234' });
    expect(result.success).toBe(false);
  });

  it('падает на коротком пароле', () => {
    const result = loginSchema.safeParse({ email: 'a@b.c', password: '123' });
    expect(result.success).toBe(false);
  });
});

describe('registerSchema', () => {
  const valid = {
    email: 'user@example.com',
    password: 'StrongPass1',
    confirmPassword: 'StrongPass1',
    fullName: 'Иванов Иван',
    role: 'student' as const,
  };

  it('пропускает корректные данные', () => {
    expect(registerSchema.safeParse(valid).success).toBe(true);
  });

  it('падает если пароли не совпадают', () => {
    const result = registerSchema.safeParse({ ...valid, confirmPassword: 'Other1234' });
    expect(result.success).toBe(false);
  });

  it('требует цифру в пароле', () => {
    const result = registerSchema.safeParse({
      ...valid,
      password: 'NoDigitsAtAll',
      confirmPassword: 'NoDigitsAtAll',
    });
    expect(result.success).toBe(false);
  });

  it('требует заглавную букву', () => {
    const result = registerSchema.safeParse({
      ...valid,
      password: 'allsmall1',
      confirmPassword: 'allsmall1',
    });
    expect(result.success).toBe(false);
  });
});
