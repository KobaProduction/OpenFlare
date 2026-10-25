'use client';

import { createTranslator, type useTranslations } from 'next-intl';
import english from '@/messages/fragments/runtime.en.json';
import chinese from '@/messages/fragments/runtime.zh-CN.json';
import { resolveClientLocale } from './client';
import type { AppLocale } from './config';

export type RuntimeTranslator = ReturnType<typeof useTranslations>;
type RuntimeKey = keyof typeof english.runtime;

const translators = {
  en: createTranslator({
    locale: 'en',
    messages: english,
    namespace: 'runtime',
  }),
  'zh-CN': createTranslator({
    locale: 'zh-CN',
    messages: chinese,
    namespace: 'runtime',
  }),
};

export function getRuntimeTranslator(locale: AppLocale) {
  return translators[locale];
}

// Imperative browser code (network errors, form utilities) has no React hook
// context. Resolve the current preference when creating a message, not at import.
export const runtimeText = Object.assign(
  (key: RuntimeKey, values?: Record<string, string | number>) =>
    translators[resolveClientLocale()](key, values),
  {
    raw: (key: RuntimeKey): string =>
      translators[resolveClientLocale()].raw(key),
  },
);
