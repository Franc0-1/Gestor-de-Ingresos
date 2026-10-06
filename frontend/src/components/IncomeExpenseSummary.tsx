import type { CurrencySummary } from '../types/dashboard';
import { formatMoney, proportion } from '../utils/format';
import { EmptyState } from './EmptyState';

export function IncomeExpenseSummary({ summaries }: { summaries: CurrencySummary[] }) {
  return <section className="period-summary" aria-labelledby="period-summary-title"><div className="section-heading"><div><span className="eyebrow">El período, en perspectiva</span><h2 id="period-summary-title">Lo que entra. Lo que sale.</h2></div><span className="subtle-note">Dentro del período</span></div>{!summaries.length ? <EmptyState title="Todavía sin importes" message="Los ingresos y gastos del período se verán aquí." /> : summaries.map(summary => {
    const total = summary.periodIncome + summary.periodExpenses;
    return <div className="currency-flow" key={summary.currency}><span className="flow-currency">{summary.currency || 'Moneda sin identificar'}</span><div className="flow-numbers"><div className="metric income"><span><b aria-hidden="true">↗</b> Ingresos</span><strong>{formatMoney(summary.periodIncome, summary.currency)}</strong></div><div className="metric expense"><span><b aria-hidden="true">↘</b> Gastos</span><strong>{formatMoney(summary.periodExpenses, summary.currency)}</strong></div><div className="metric savings"><span>Ahorro neto</span><strong className={summary.netSavings < 0n ? 'negative' : ''}>{formatMoney(summary.netSavings, summary.currency)}</strong></div></div><div className="flow-bar" aria-hidden="true"><span style={{ width: `${proportion(summary.periodIncome, total)}%` }} /><span style={{ width: `${proportion(summary.periodExpenses, total)}%` }} /></div><p className="flow-caption">{total > 0n ? `${proportion(summary.periodExpenses, total).toLocaleString('es-AR', { maximumFractionDigits: 1 })}% del flujo corresponde a gastos.` : 'No hay ingresos ni gastos en este período.'} Las transferencias no se cuentan como ingresos o gastos.</p></div>;
  })}</section>;
}
