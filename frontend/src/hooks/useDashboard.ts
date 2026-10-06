import { useCallback, useEffect, useState } from 'react';
import { ApiError, getDashboard } from '../services/api';
import type { DashboardResponse, PeriodFilter } from '../types/dashboard';

export function useDashboard(filter: PeriodFilter) {
  const [data, setData] = useState<DashboardResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [attempt, setAttempt] = useState(0);
  const retry = useCallback(() => setAttempt(value => value + 1), []);

  useEffect(() => {
    const controller = new AbortController();
    setLoading(true);
    setError(null);
    getDashboard({ from: filter.from, to: filter.to }, controller.signal)
      .then(result => { if (!controller.signal.aborted) setData(result); })
      .catch((cause: unknown) => {
        if (!controller.signal.aborted) setError(cause instanceof ApiError ? cause.message : 'No pudimos conectar con el servidor. Comprobá que esté iniciado e intentá de nuevo.');
      })
      .finally(() => { if (!controller.signal.aborted) setLoading(false); });
    return () => controller.abort();
  }, [filter.from, filter.to, attempt]);

  return { data, loading, error, retry };
}
