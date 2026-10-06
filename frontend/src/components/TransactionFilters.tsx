import type { Account, Category, TransactionQuery } from '../types/transactions';
import { ListFilters } from './ListFilters';

export function TransactionFilters({ query, accounts, categories, onApply }: { query: TransactionQuery; accounts: Account[]; categories: Category[]; onApply: (query: TransactionQuery) => void }) {
  return <ListFilters query={query} onApply={onApply} summary="Tipo, cuenta, categoría y fechas" includeType>{(draft, update) => <>
    <label>Tipo<select value={draft.type ?? ''} onChange={event => { const type = event.target.value as TransactionQuery['type']; update({ ...draft, type, categoryId: categories.some(category => String(category.id) === draft.categoryId && (!type || category.type === type)) ? draft.categoryId : '' }); }}><option value="">Todos</option><option value="income">Ingresos</option><option value="expense">Gastos</option></select></label>
    <label>Cuenta<select value={draft.accountId ?? ''} onChange={event => update({ ...draft, accountId: event.target.value })}><option value="">Todas las cuentas</option>{accounts.map(account => <option key={account.id} value={account.id}>{account.name}{!account.active ? ' · Inactiva' : ''}</option>)}</select></label>
    <label>Categoría<select value={draft.categoryId ?? ''} onChange={event => update({ ...draft, categoryId: event.target.value })}><option value="">Todas las categorías</option>{categories.filter(category => !draft.type || category.type === draft.type).map(category => <option key={category.id} value={category.id}>{category.name}</option>)}</select></label>
  </>}</ListFilters>;
}
