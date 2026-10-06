import type { Account } from '../types/transactions';
import type { Transfer } from '../types/transfers';
import { formatDate, formatMoney } from '../utils/format';

export function TransferList({ items, accounts, disabled, onEdit, onDelete }: { items: Transfer[]; accounts: Account[]; disabled: boolean; onEdit: (transfer: Transfer) => void; onDelete: (transfer: Transfer) => void }) {
  return <ul className="transfer-list" aria-label="Transferencias">{items.map(transfer => {
    const from = accounts.find(account => account.id === transfer.fromAccountId);
    const to = accounts.find(account => account.id === transfer.toAccountId);
    return <li key={transfer.id}><div className="transfer-detail"><div className="transfer-route"><div><small>Origen</small><strong>{from?.name ?? 'Cuenta no disponible'}</strong>{from && !from.active && <small>Inactiva</small>}</div><span aria-hidden="true">→</span><div><small>Destino</small><strong>{to?.name ?? 'Cuenta no disponible'}</strong>{to && !to.active && <small>Inactiva</small>}</div></div><p>{transfer.description || 'Sin descripción'}</p><time dateTime={transfer.date}>{formatDate(transfer.date)}</time></div><div className="transfer-value"><strong>{formatMoney(transfer.amount, from?.currency ?? '')}</strong>{from && to && from.currency !== to.currency && <small>Destino: {formatMoney(transfer.amount, to.currency)} · Sin conversión</small>}</div><div className="transaction-actions"><button className="button quiet" disabled={disabled} aria-label={`Editar transferencia de ${from?.name ?? 'origen'} a ${to?.name ?? 'destino'}`} onClick={() => onEdit(transfer)}>Editar</button><button className="button text-danger" disabled={disabled} aria-label={`Eliminar transferencia de ${from?.name ?? 'origen'} a ${to?.name ?? 'destino'}`} onClick={() => onDelete(transfer)}>Eliminar</button></div></li>;
  })}</ul>;
}
