import { useEffect, useState } from 'react';
import type { FormEvent, ReactNode } from 'react';
import type { TransactionQuery } from '../types/transactions';

type ListQuery = Pick<TransactionQuery, 'search' | 'from' | 'to' | 'sortBy' | 'sortOrder' | 'page' | 'pageSize'>;

export function ListFilters<Q extends ListQuery>({ query, onApply, summary, children, includeType = false }: { query: Q; onApply: (query: Q) => void; summary: string; children: (draft: Q, update: (draft: Q) => void) => ReactNode; includeType?: boolean }) {
  const [draft, setDraft] = useState(query);
  const [error, setError] = useState('');
  useEffect(() => { setDraft(query); }, [query]);
  function apply(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (draft.from && draft.to && draft.from > draft.to) { setError('La fecha desde debe ser anterior o igual a hasta.'); return; }
    setError(''); onApply({ ...draft, search: draft.search?.trim(), page: 1 });
  }
  return <details className="transaction-filters" open><summary>Buscar y filtrar <span>{summary}</span></summary><form onSubmit={apply}>
    <label className="search-field">Buscar por descripción<input type="search" value={draft.search ?? ''} onChange={event => setDraft({ ...draft, search: event.target.value })} placeholder="Buscar una descripción…" /></label>
    <div className="filter-grid">{children(draft, setDraft)}
      <label>Desde<input type="date" value={draft.from ?? ''} onChange={event => setDraft({ ...draft, from: event.target.value })} /></label>
      <label>Hasta<input type="date" value={draft.to ?? ''} onChange={event => setDraft({ ...draft, to: event.target.value })} /></label>
      <label>Ordenar por<select value={draft.sortBy ?? 'date'} onChange={event => setDraft({ ...draft, sortBy: event.target.value as Q['sortBy'] })}><option value="date">Fecha</option><option value="amount">Monto</option><option value="id">ID</option><option value="createdAt">Fecha de creación</option><option value="updatedAt">Última edición</option>{includeType && <option value="type">Tipo</option>}</select></label>
      <label>Dirección<select value={draft.sortOrder ?? 'desc'} onChange={event => setDraft({ ...draft, sortOrder: event.target.value as Q['sortOrder'] })}><option value="desc">Descendente</option><option value="asc">Ascendente</option></select></label>
    </div>
    {error && <p className="field-error" role="alert">{error}</p>}
    <div className="filter-actions"><button className="button primary">Aplicar filtros</button><button type="button" className="button quiet" onClick={() => { setError(''); onApply({ page: 1, pageSize: query.pageSize ?? 20 } as Q); }}>Limpiar filtros</button></div>
  </form></details>;
}
