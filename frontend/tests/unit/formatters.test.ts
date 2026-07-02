import { describe, expect, it } from 'vitest';
import { getInitials, truncate, stringToColor } from '@/utils/formatters';

describe('getInitials', () => {
  it('возвращает инициалы из ФИО', () => {
    expect(getInitials('Иванов Иван Иванович')).toBe('ИИ');
    expect(getInitials('John Doe')).toBe('JD');
  });

  it('обрабатывает одно слово', () => {
    expect(getInitials('Madonna')).toBe('M');
  });

  it('возвращает ? для пустой строки', () => {
    expect(getInitials('')).toBe('?');
    expect(getInitials(undefined)).toBe('?');
  });
});

describe('truncate', () => {
  it('не трогает короткие строки', () => {
    expect(truncate('short', 60)).toBe('short');
  });

  it('обрезает длинные строки', () => {
    expect(truncate('a'.repeat(100), 10)).toMatch(/^a+…$/);
    expect(truncate('a'.repeat(100), 10).length).toBe(10);
  });
});

describe('stringToColor', () => {
  it('возвращает один и тот же цвет для одной строки', () => {
    expect(stringToColor('hello')).toBe(stringToColor('hello'));
  });

  it('возвращает hex-цвет', () => {
    expect(stringToColor('test')).toMatch(/^#[0-9a-f]{6}$/i);
  });
});
