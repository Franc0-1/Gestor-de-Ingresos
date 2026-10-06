import { useCallback, useEffect, useState } from 'react';
import { getAccounts, getTransfers } from '../services/api';
import type { Account } from '../types/transactions';
import type { TransferPage, TransferQuery } from '../types/transfers';

export function useTransfers(query: TransferQuery) {
  const [page, setPage] = useState<TransferPage | null>(null);
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [attempt, setAttempt] = useState(0);
  const retry = useCallback(() => setAttempt(value => value + 1), []);
  useEffect(() => {
    const controller = new AbortController();
    setLoading(true); setError('');
    Promise.all([getTransfers(query, controller.signal), getAccounts(controller.signal)])
      .then(([result, allAccounts]) => { if (!controller.signal.aborted) { setPage(result); setAccounts(allAccounts); } })
      .catch((cause: unknown) => { if (!controller.signal.aborted) setError(cause instanceof Error ? cause.message : 'No se pudieron cargar las transferencias.'); })
      .finally(() => { if (!controller.signal.aborted) setLoading(false); });
    return () => controller.abort();
  }, [query, attempt]);
  return { page, accounts, loading, error, retry };
}
