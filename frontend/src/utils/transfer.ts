import type { Account } from '../types/transactions';
import type { TransferInput } from '../types/transfers';
import { parseMoneyInput } from './money.ts';

export function transferInput(fields: { fromAccountId: string; toAccountId: string; amount: string; description: string; date: string }, accounts: Account[]): TransferInput {
  const from = accounts.find(account => String(account.id) === fields.fromAccountId && account.active);
  const to = accounts.find(account => String(account.id) === fields.toAccountId && account.active);
  if (!from) throw new Error('Seleccioná una cuenta de origen activa.');
  if (!to) throw new Error('Seleccioná una cuenta de destino activa.');
  if (from.id === to.id) throw new Error('Las cuentas deben ser diferentes.');
  const amount = parseMoneyInput(fields.amount);
  const date = `${fields.date}T00:00:00Z`;
  const parsed = new Date(date);
  if (!/^\d{4}-\d{2}-\d{2}$/.test(fields.date) || Number.isNaN(parsed.getTime()) || parsed.toISOString().slice(0, 10) !== fields.date) throw new Error('Seleccioná una fecha válida.');
  return { fromAccountId: from.id, toAccountId: to.id, amount, description: fields.description.trim(), date };
}
