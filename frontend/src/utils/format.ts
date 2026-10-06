import type { Money, Period } from '../types/dashboard';

export function formatMoney(amount: Money, currency: string): string {
  let formatter: Intl.NumberFormat;
  try {
    formatter = new Intl.NumberFormat('es-AR', { style: 'currency', currency, currencyDisplay: 'code' });
  } catch {
    formatter = new Intl.NumberFormat('es-AR', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
  }
  const digits = formatter.resolvedOptions().maximumFractionDigits ?? 2;
  const scale = 10n ** BigInt(digits);
  const negative = amount < 0n;
  const absolute = negative ? -amount : amount;
  const major = absolute / scale;
  const fraction = (absolute % scale).toString().padStart(digits, '0');
  const parts = formatter.formatToParts(negative ? -(major || 1n) : major);
  const formatted = parts.map(part => {
    if (part.type === 'fraction') return fraction;
    if (part.type === 'integer' && major === 0n) return '0';
    return part.value;
  }).join('');
  return currency ? formatted : `${formatted} · moneda sin identificar`;
}

export function formatDate(iso: string): string {
  return new Intl.DateTimeFormat('es-AR', { day: '2-digit', month: 'short', year: 'numeric', timeZone: 'UTC' }).format(new Date(iso));
}

export function formatPeriod(period: Period): string {
  const from = new Date(`${period.from}T00:00:00Z`);
  const to = new Date(`${period.to}T00:00:00Z`);
  if (from.getUTCDate() === 1 && from.getUTCMonth() === to.getUTCMonth() && from.getUTCFullYear() === to.getUTCFullYear() && to.getUTCDate() === new Date(Date.UTC(to.getUTCFullYear(), to.getUTCMonth() + 1, 0)).getUTCDate()) {
    return new Intl.DateTimeFormat('es-AR', { month: 'long', year: 'numeric', timeZone: 'UTC' }).format(from);
  }
  return `${formatDate(from.toISOString())} — ${formatDate(to.toISOString())}`;
}

export function proportion(amount: Money, total: Money): number {
  return total > 0n ? Number(amount * 10000n / total) / 100 : 0;
}
