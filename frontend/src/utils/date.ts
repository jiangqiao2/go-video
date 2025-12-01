import dayjs, { Dayjs } from 'dayjs';

type PublishedTimeInput = string | number | Date | Dayjs | null | undefined;

interface FormatPublishedTimeOptions {
  fallbackFormat?: string;
}

export const formatPublishedTime = (
  input?: PublishedTimeInput,
  options?: FormatPublishedTimeOptions,
): string => {
  const normalizeToDayjs = (value: PublishedTimeInput): Dayjs | null => {
    if (!value) return null;
    if (dayjs.isDayjs(value)) return value;
    if (value instanceof Date) return dayjs(value);
    if (typeof value === 'number') {
      const num = value > 0 && value < 1e12 ? value * 1000 : value; // 秒级转毫秒
      return dayjs(num);
    }
    if (typeof value === 'string') {
      const trimmed = value.trim();
      if (/^-?\d+$/.test(trimmed)) {
        const raw = Number(trimmed);
        const num = trimmed.length <= 10 ? raw * 1000 : raw; // 10位内按秒处理
        return dayjs(num);
      }
      return dayjs(trimmed);
    }
    return dayjs(value);
  };

  const publishedAt = normalizeToDayjs(input);
  if (!publishedAt || !publishedAt.isValid()) {
    return '--';
  }

  const fallbackFormat = options?.fallbackFormat ?? 'YYYY-MM-DD HH:mm';
  const now = dayjs();

  if (publishedAt.isAfter(now)) {
    return publishedAt.format(fallbackFormat);
  }

  const diffHours = now.diff(publishedAt, 'hour', true);
  const isSameDay = now.isSame(publishedAt, 'day');

  if (isSameDay && diffHours < 6) {
    const diffMinutes = now.diff(publishedAt, 'minute');

    if (diffMinutes <= 0) {
      return '刚刚发布';
    }

    if (diffMinutes < 60) {
      return `${diffMinutes}分钟前`;
    }

    return `${Math.max(Math.floor(diffHours), 1)}小时前`;
  }

  return publishedAt.format(fallbackFormat);
};

export default formatPublishedTime;
