import type { Money } from './dashboard';

export type TransactionType = 'income' | 'expense';
export interface Transaction {
  id: number;
  type: TransactionType;
  amount: Money;
  description: string;
  categoryId: number;
  accountId: number;
  date: string;
  createdAt: string;
  updatedAt: string;
}
export type TransactionInput = Pick<Transaction, 'type' | 'amount' | 'description' | 'categoryId' | 'accountId' | 'date'>;
export interface TransactionQuery {
  type?: TransactionType | '';
  accountId?: string;
  categoryId?: string;
  from?: string;
  to?: string;
  search?: string;
  sortBy?: 'id' | 'amount' | 'date' | 'createdAt' | 'updatedAt' | 'type';
  sortOrder?: 'asc' | 'desc';
  page?: number;
  pageSize?: number;
}
export interface Pagination { page: number; pageSize: number; totalItems: number; totalPages: number }
export interface TransactionPage { items: Transaction[]; pagination: Pagination }
export interface Account { id: number; name: string; type: string; currency: string; initialBalance: Money; active: boolean; createdAt: string; updatedAt: string }
export interface Category { id: number; name: string; type: TransactionType; createdAt: string; updatedAt: string }
