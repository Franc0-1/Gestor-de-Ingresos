import { useState } from 'react';
import type { Account } from '../types/transactions';
import type { Transfer } from '../types/transfers';
import { formatDate, formatMoney } from '../utils/format';
import { Dialog } from './Dialog';

export function TransferDeleteDialog({ transfer, accounts, onDelete, onDismiss }: { transfer: Transfer; accounts: Account[]; onDelete: () => Promise<void>; onDismiss: () => void }) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const from = accounts.find(account => account.id === transfer.fromAccountId);
  const to = accounts.find(account => account.id === transfer.toAccountId);
  async function confirm() {
    if (busy) return;
    setBusy(true); setError('');
    try { await onDelete(); } catch (cause) { setError(cause instanceof Error ? cause.message : 'No se pudo eliminar la transferencia.'); } finally { setBusy(false); }
  }
  return <Dialog title="Eliminar transferencia" busy={busy} onDismiss={onDismiss}><div className="delete-details"><p>¿Eliminar la transferencia de {from?.name ?? 'cuenta no disponible'} a {to?.name ?? 'cuenta no disponible'}?</p><strong>{formatMoney(transfer.amount, from?.currency ?? '')}</strong>{from && to && from.currency !== to.currency && <p>Destino: {formatMoney(transfer.amount, to.currency)} · Sin conversión</p>}<h3>{transfer.description || 'Sin descripción'}</h3><small>{formatDate(transfer.date)}</small><p>Esta acción no se puede deshacer.</p></div>{error && <p className="field-error" role="alert">{error}</p>}<footer className="dialog-actions"><button className="button quiet" disabled={busy} onClick={onDismiss}>Cancelar</button><button className="button danger" disabled={busy} onClick={confirm}>{busy ? 'Eliminando…' : 'Eliminar transferencia'}</button></footer></Dialog>;
}
