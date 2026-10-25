import { getRuntimeTranslator } from '@/i18n/runtime';

export function isLogStorageDisabledMessage(message: string): boolean {
  const normalized = message.toLowerCase();
  return (
    message.includes('ClickHouse') ||
    (['en', 'zh-CN'] as const).some((locale) =>
      normalized.includes(
        getRuntimeTranslator(locale)('disabled_f95ea7f4').toLowerCase(),
      ),
    )
  );
}
