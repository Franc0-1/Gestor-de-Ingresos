import type { AccountBalance, DashboardMovement } from '../types/dashboard';
import { formatDate, formatMoney } from '../utils/format';
import { EmptyState } from './EmptyState';

const kinds = { income: { label: 'Ingreso', symbol: '↗' }, expense: { label: 'Gasto', symbol: '↘' }, transfer: { label: 'Transferencia', symbol: '⇄' } };

export function RecentMovements({ movements, accounts }: { movements: DashboardMovement[]; accounts: AccountBalance[] }) {
  return <section className="movements-panel" aria-labelledby="movements-title"><div className="section-heading"><div><span className="eyebrow">Tu actividad</span><h2 id="movements-title">Últimos movimientos</h2></div><span className="subtle-note">Hasta 10 · En el período</span></div>{!movements.length ? <EmptyState title="Un período tranquilo" message="Todavía no hay movimientos en estas fechas." /> : <ul className="movement-list">{movements.map(movement => {
    const kind = kinds[movement.kind];
    const account = accounts.find(item => item.accountId === movement.accountId);
    const destination = accounts.find(item => item.accountId === movement.toAccountId);
    return <li key={`${movement.kind}-${movement.id}`}><span className={`movement-symbol ${movement.kind}`} aria-hidden="true">{kind.symbol}</span><div className="movement-description"><strong>{movement.description || kind.label}</strong><small>{account?.accountName || `Cuenta ${movement.accountId}`}{movement.kind === 'transfer' ? ` → ${destination?.accountName || `Cuenta ${movement.toAccountId}`}` : ''}</small></div><time dateTime={movement.date}>{formatDate(movement.date)}</time><span className={`kind-label ${movement.kind}`}>{kind.label}</span><strong className={`movement-amount ${movement.kind}`}>{movement.kind === 'income' ? '+ ' : movement.kind === 'expense' ? '− ' : ''}{formatMoney(movement.amount, account?.currency ?? '')}</strong></li>;
  })}</ul>}</section>;
}
