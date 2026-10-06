import { useCallback, useEffect, useState } from 'react';
import { getAccounts, getCategories, getTransactions } from '../services/api';
import type { Account, Category, TransactionPage, TransactionQuery } from '../types/transactions';

export function useTransactions(query: TransactionQuery) {
  const [page, setPage] = useState<TransactionPage | null>(null);
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [attempt, setAttempt] = useState(0);
  const retry = useCallback(() => setAttempt(value => value + 1), []);
  useEffect(() => {
    const controller = new AbortController();
    setLoading(true); setError('');
    Promise.all([getTransactions(query, controller.signal), getAccounts(controller.signal), getCategories(controller.signal)])
      .then(([result, allAccounts, allCategories]) => {
        if (!controller.signal.aborted) { setPage(result); setAccounts(allAccounts); setCategories(allCategories); }
      })
      .catch((cause: unknown) => { if (!controller.signal.aborted) setError(cause instanceof Error ? cause.message : 'No se pudieron cargar los movimientos.'); })
      .finally(() => { if (!controller.signal.aborted) setLoading(false); });
    return () => controller.abort();
  }, [query, attempt]);
  return { page, accounts, categories, loading, error, retry };
}
