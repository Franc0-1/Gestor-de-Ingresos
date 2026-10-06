import type { Money } from './dashboard';
import type { Pagination, TransactionQuery } from './transactions';

export interface Transfer {
  id: number;
  fromAccountId: number;
  toAccountId: number;
  amount: Money;
  description: string;
  date: string;
  createdAt: string;
  updatedAt: string;
}
export type TransferInput = Pick<Transfer, 'fromAccountId' | 'toAccountId' | 'amount' | 'description' | 'date'>;
export interface TransferQuery extends Pick<TransactionQuery, 'from' | 'to' | 'search' | 'sortOrder' | 'page' | 'pageSize'> {
  fromAccountId?: string;
  toAccountId?: string;
  sortBy?: 'id' | 'amount' | 'date' | 'createdAt' | 'updatedAt';
}
export interface TransferPage { items: Transfer[]; pagination: Pagination }
