import { useState } from 'react';
import type { Account, Transaction } from '../types/transactions';
import { formatDate, formatMoney } from '../utils/format';
import { Dialog } from './Dialog';

export function TransactionDeleteDialog({ transaction, accounts, onDelete, onDismiss }: { transaction: Transaction; accounts: Account[]; onDelete: () => Promise<void>; onDismiss: () => void }) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  async function confirm() {
    if (busy) return;
    setBusy(true); setError('');
    try { await onDelete(); } catch (cause) { setError(cause instanceof Error ? cause.message : 'No se pudo eliminar el movimiento.'); } finally { setBusy(false); }
  }
  return <Dialog title="Eliminar movimiento" busy={busy} onDismiss={onDismiss}><div className="delete-details"><p>¿Eliminar este {transaction.type === 'income' ? 'ingreso' : 'gasto'}?</p><h3>{transaction.description || 'Sin descripción'}</h3><strong>{formatMoney(transaction.amount, accounts.find(account => account.id === transaction.accountId)?.currency ?? '')}</strong><small>{formatDate(transaction.date)}</small><p>Esta acción no se puede deshacer.</p></div>{error && <p className="field-error" role="alert">{error}</p>}<footer className="dialog-actions"><button className="button quiet" disabled={busy} onClick={onDismiss}>Cancelar</button><button className="button danger" disabled={busy} onClick={confirm}>{busy ? 'Eliminando…' : 'Eliminar movimiento'}</button></footer></Dialog>;
}
