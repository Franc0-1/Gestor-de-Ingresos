import type { Account } from '../types/transactions';
import type { TransferQuery } from '../types/transfers';
import { ListFilters } from './ListFilters';

export function TransferFilters({ query, accounts, onApply }: { query: TransferQuery; accounts: Account[]; onApply: (query: TransferQuery) => void }) {
  return <ListFilters query={query} onApply={onApply} summary="Origen, destino y fechas">{(draft, update) => <>
    <label>Cuenta origen<select value={draft.fromAccountId ?? ''} onChange={event => update({ ...draft, fromAccountId: event.target.value })}><option value="">Todos los orígenes</option>{accounts.map(account => <option key={account.id} value={account.id}>{account.name} · {account.currency}{!account.active ? ' · Inactiva' : ''}</option>)}</select></label>
    <label>Cuenta destino<select value={draft.toAccountId ?? ''} onChange={event => update({ ...draft, toAccountId: event.target.value })}><option value="">Todos los destinos</option>{accounts.map(account => <option key={account.id} value={account.id}>{account.name} · {account.currency}{!account.active ? ' · Inactiva' : ''}</option>)}</select></label>
  </>}</ListFilters>;
}
