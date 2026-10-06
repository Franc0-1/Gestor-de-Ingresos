import type { CategoryTotal } from '../types/dashboard';
import { formatMoney, proportion } from '../utils/format';
import { EmptyState } from './EmptyState';

export function CategoryBreakdown({ title, kind, categories }: { title: string; kind: 'income' | 'expense'; categories: CategoryTotal[] }) {
  const currencies = [...new Set(categories.map(category => category.currency))];
  return <section className={`category-panel ${kind}`} aria-labelledby={`${kind}-category-title`}><div className="section-heading"><div><span className="eyebrow">Por categoría</span><h2 id={`${kind}-category-title`}>{title}</h2></div><span className="category-kind">{kind === 'expense' ? 'Salidas' : 'Entradas'}</span></div>{!categories.length ? <EmptyState title={kind === 'expense' ? 'Sin gastos en el período' : 'Sin ingresos en el período'} message="Elegí otro rango para consultar más movimientos." /> : currencies.map(currency => {
    const items = categories.filter(category => category.currency === currency);
    const total = items.reduce((sum, item) => sum + item.amount, 0n);
    return <div className="category-group" key={currency}>{currencies.length > 1 && <h3>{currency || 'Moneda sin identificar'}</h3>}<ul className="category-list">{items.map(category => <li key={`${category.categoryId}-${currency}`}><div><span>{category.categoryName || `Categoría ${category.categoryId}`}</span><strong>{formatMoney(category.amount, currency)}</strong></div><div className="category-track" aria-hidden="true"><span style={{ width: `${proportion(category.amount, total)}%` }} /></div></li>)}</ul></div>;
  })}</section>;
}
