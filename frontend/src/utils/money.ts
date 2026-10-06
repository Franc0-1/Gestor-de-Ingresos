export const MAX_AMOUNT = 9223372036854775807n;

export function parseMoneyInput(text: string): bigint {
  const value = text.trim();
  if (!/^\d+(?:[,.]\d{1,2})?$/.test(value)) throw new Error('Ingresá un monto válido, con un máximo de 2 decimales.');
  const [whole = '0', fraction = ''] = value.split(/[,.]/);
  const amount = BigInt(whole) * 100n + BigInt(fraction.padEnd(2, '0'));
  if (amount <= 0n) throw new Error('El monto debe ser mayor que cero.');
  if (amount > MAX_AMOUNT) throw new Error('El monto supera el máximo permitido.');
  return amount;
}

export function moneyInputValue(amount: bigint): string {
  return `${amount / 100n},${(amount % 100n).toString().padStart(2, '0')}`;
}

export function todayDate(): string {
  const now = new Date();
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`;
}
