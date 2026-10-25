import english from '@/messages/fragments/builtin-metadata.en.json';
import chinese from '@/messages/fragments/builtin-metadata.zh-CN.json';

function localizedDefault(
  value: string,
  en: string | undefined,
  zh: string | undefined,
  locale: string,
): string {
  // Only known defaults are translated. Custom and unknown text stays intact.
  if (!en || !zh || (value !== en && value !== zh)) return value;
  return locale === 'zh-CN' ? zh : en;
}

export function localizeConfigDescription(
  key: string,
  value: string,
  locale: string,
): string {
  const en = english.builtinMetadata.configDescriptions as Record<
    string,
    string
  >;
  const zh = chinese.builtinMetadata.configDescriptions as Record<
    string,
    string
  >;
  return localizedDefault(value, en[key], zh[key], locale);
}

export function localizeTemplateMetadata(
  key: string,
  isSystem: boolean,
  field: 'name' | 'description',
  value: string,
  locale: string,
): string {
  if (!isSystem) return value;
  const en = english.builtinMetadata.templates as Record<
    string,
    { name: string; description: string }
  >;
  const zh = chinese.builtinMetadata.templates as typeof en;
  return localizedDefault(value, en[key]?.[field], zh[key]?.[field], locale);
}
