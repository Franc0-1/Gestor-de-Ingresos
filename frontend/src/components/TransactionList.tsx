import type { Account, Category, Transaction } from '../types/transactions';
import { formatDate, formatMoney } from '../utils/format';

export function TransactionList({ items, accounts, categories, disabled, onEdit, onDelete }: { items: Transaction[]; accounts: Account[]; categories: Category[]; disabled: boolean; onEdit: (transaction: Transaction) => void; onDelete: (transaction: Transaction) => void }) {
  return <ul className="transaction-list" aria-label="Movimientos">{items.map(transaction => {
    const account = accounts.find(item => item.id === transaction.accountId);
    const category = categories.find(item => item.id === transaction.categoryId);
    const income = transaction.type === 'income';
    return <li key={transaction.id}><div className={`transaction-marker ${transaction.type}`}><span aria-hidden="true">{income ? '↗' : '↘'}</span><small>{income ? 'Ingreso' : 'Gasto'}</small></div><div className="transaction-description"><strong>{transaction.description || 'Sin descripción'}</strong><div><span>{category?.name || `Categoría ${transaction.categoryId}`}</span><span aria-hidden="true"> · </span><span>{account?.name || `Cuenta ${transaction.accountId}`}{account && !account.active ? ' (inactiva)' : ''}</span></div><time dateTime={transaction.date}>{formatDate(transaction.date)}</time></div><strong className={`transaction-amount ${transaction.type}`}>{income ? '+ ' : '− '}{formatMoney(transaction.amount, account?.currency ?? '')}</strong><div className="transaction-actions"><button className="button quiet" disabled={disabled} aria-label={`Editar ${transaction.description || `movimiento ${transaction.id}`}`} onClick={() => onEdit(transaction)}>Editar</button><button className="button text-danger" disabled={disabled} aria-label={`Eliminar ${transaction.description || `movimiento ${transaction.id}`}`} onClick={() => onDelete(transaction)}>Eliminar</button></div></li>;
  })}</ul>;
}
