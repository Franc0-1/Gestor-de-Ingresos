import { useState } from 'react';
import { AccountBalances } from '../components/AccountBalances';
import { BalanceSummary } from '../components/BalanceSummary';
import { CategoryBreakdown } from '../components/CategoryBreakdown';
import { ErrorState } from '../components/ErrorState';
import { IncomeExpenseSummary } from '../components/IncomeExpenseSummary';
import { LoadingState } from '../components/LoadingState';
import { PeriodSelector } from '../components/PeriodSelector';
import { RecentMovements } from '../components/RecentMovements';
import { useDashboard } from '../hooks/useDashboard';
import type { PeriodFilter } from '../types/dashboard';
import { formatPeriod } from '../utils/format';

export function Dashboard() {
  const [filter, setFilter] = useState<PeriodFilter>({});
  const { data, loading, error, retry } = useDashboard(filter);
  return <><div className="page-heading"><div><span className="eyebrow">Tu panorama financiero</span><h1>Todo en su lugar<span>.</span></h1><p>{data ? formatPeriod(data.period) : 'Una mirada clara a tus finanzas personales.'}</p></div><PeriodSelector period={data?.period} onChange={setFilter} disabled={loading} /></div><div aria-busy={loading}>{loading ? <LoadingState /> : error ? <ErrorState message={error} retry={retry} /> : data && <><div className="dashboard-overview"><BalanceSummary balances={data.balancesByCurrency} accountCount={data.accounts.length} /><IncomeExpenseSummary summaries={data.summaryByCurrency} /></div><div className="detail-layout"><div className="category-column"><CategoryBreakdown title="En qué gastaste" kind="expense" categories={data.expensesByCategory} /><CategoryBreakdown title="De dónde vino" kind="income" categories={data.incomeByCategory} /></div><AccountBalances accounts={data.accounts} /></div><RecentMovements movements={data.recentMovements} accounts={data.accounts} /></>}</div></>;
}
