// Monetary int64 values are decoded as bigint to retain the backend's precision.
export type Money = bigint;
export interface Period { from: string; to: string }
export interface PeriodFilter { from?: string; to?: string }
export interface DashboardSummary {
  totalBalance?: Money;
  currency?: string;
  periodIncome?: Money;
  periodExpenses?: Money;
  netSavings?: Money;
}
export interface CurrencyBalance { currency: string; amount: Money }
export interface CurrencySummary { currency: string; periodIncome: Money; periodExpenses: Money; netSavings: Money }
export interface AccountBalance { accountId: number; accountName: string; currency: string; balance: Money }
export interface CategoryTotal { categoryId: number; categoryName: string; currency: string; amount: Money }
export interface DashboardMovement {
  id: number;
  kind: 'income' | 'expense' | 'transfer';
  amount: Money;
  description: string;
  date: string;
  accountId: number;
  toAccountId?: number;
}
export interface DashboardResponse {
  period: Period;
  summary: DashboardSummary;
  balancesByCurrency: CurrencyBalance[];
  summaryByCurrency: CurrencySummary[];
  accounts: AccountBalance[];
  expensesByCategory: CategoryTotal[];
  incomeByCategory: CategoryTotal[];
  recentMovements: DashboardMovement[];
}
