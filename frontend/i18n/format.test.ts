import { afterEach, describe, expect, it } from 'vitest';
import { formatDateTime, getFormattingLocale } from './format';

afterEach(() => {
  document.documentElement.lang = '';
});

describe('locale-aware formatting', () => {
  it('uses the rendered language for dates and preserves invalid input', () => {
    document.documentElement.lang = 'en';
    expect(getFormattingLocale()).toBe('en');
    expect(formatDateTime('2026-01-15T00:00:00Z')).toContain('01/15/2026');
    document.documentElement.lang = 'zh-CN';
    expect(formatDateTime('2026-01-15T00:00:00Z')).toContain('2026/01/15');
    expect(formatDateTime('自定义文本')).toBe('自定义文本');
  });
});
