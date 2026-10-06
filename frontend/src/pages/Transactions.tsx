import { useEffect, useRef, useState } from 'react';
import { EmptyState } from '../components/EmptyState';
import { ErrorState } from '../components/ErrorState';
import { LoadingState } from '../components/LoadingState';
import { Pagination } from '../components/Pagination';
import { TransactionDeleteDialog } from '../components/TransactionDeleteDialog';
import { TransactionFilters } from '../components/TransactionFilters';
import { TransactionForm } from '../components/TransactionForm';
import { TransactionList } from '../components/TransactionList';
import { useTransactions } from '../hooks/useTransactions';
import { createTransaction, deleteTransaction, getTransaction, updateTransaction } from '../services/api';
import type { Transaction, TransactionInput, TransactionQuery } from '../types/transactions';

export function Transactions() {
  const [query, setQuery] = useState<TransactionQuery>({ page: 1, pageSize: 20 });
  const { page, accounts, categories, loading, error, retry } = useTransactions(query);
  const [form, setForm] = useState<{ transaction?: Transaction } | null>(null);
  const [deleting, setDeleting] = useState<Transaction | null>(null);
  const [notice, setNotice] = useState('');
  const [actionError, setActionError] = useState('');
  const [editing, setEditing] = useState(false);
  const editRequest = useRef<AbortController | null>(null);
  useEffect(() => () => editRequest.current?.abort(), []);
  const filtered = Boolean(query.type || query.accountId || query.categoryId || query.from || query.to || query.search);
  useEffect(() => {
    if (!loading && !error && page && page.pagination.totalPages > 0 && page.pagination.page > page.pagination.totalPages) setQuery(current => ({ ...current, page: page.pagination.totalPages }));
  }, [page, loading, error]);
  async function edit(transaction: Transaction) {
    editRequest.current?.abort();
    const controller = new AbortController(); editRequest.current = controller;
    setEditing(true); setActionError('');
    try { const latest = await getTransaction(transaction.id, controller.signal); if (!controller.signal.aborted) setForm({ transaction: latest }); }
    catch (cause) { if (!controller.signal.aborted) setActionError(cause instanceof Error ? cause.message : 'No se pudo cargar el movimiento.'); }
    finally { if (!controller.signal.aborted) setEditing(false); }
  }
  async function save(input: TransactionInput) {
    if (form?.transaction) await updateTransaction(form.transaction.id, input); else await createTransaction(input);
    setNotice(form?.transaction ? 'Movimiento actualizado.' : 'Movimiento creado.');
    setForm(null); setQuery(current => ({ ...current, page: form?.transaction ? current.page : 1 }));
  }
  async function remove() {
    if (!deleting) return;
    await deleteTransaction(deleting.id);
    setDeleting(null); setNotice('Movimiento eliminado.');
    setQuery(current => ({ ...current, page: page?.items.length === 1 && (current.page ?? 1) > 1 ? (current.page ?? 1) - 1 : current.page }));
  }
  function newMovement() { setActionError(''); setForm({}); }
  return <>
    <div className="page-heading"><div><span className="eyebrow">Tu dinero en movimiento</span><h1>Entradas y salidas<span>.</span></h1><p>Cada movimiento cuenta. Registralo con claridad.</p></div><button className="button primary new-movement" disabled={loading || Boolean(error)} onClick={newMovement}>+ Nuevo movimiento</button></div>
    {notice && <div className="success-notice" role="status"><span>{notice}</span><button className="button quiet" aria-label="Cerrar confirmación" onClick={() => setNotice('')}>Cerrar ×</button></div>}
    {actionError && <p className="action-error" role="alert">{actionError}</p>}
    {editing && <p role="status" className="form-note">Cargando el movimiento para editar…</p>}
    <TransactionFilters query={query} accounts={accounts} categories={categories} onApply={setQuery} />
    <section className="transactions-results" aria-label="Resultado de movimientos" aria-busy={loading}>
      {loading ? <LoadingState /> : error ? <ErrorState message={error} retry={retry} /> : page && <>
        {page.items.length ? <TransactionList items={page.items} accounts={accounts} categories={categories} disabled={editing} onEdit={edit} onDelete={setDeleting} /> : <div className="transactions-empty"><EmptyState title={filtered ? 'No hay movimientos que coincidan con estos filtros.' : 'Todavía no registraste movimientos.'} message={filtered ? 'Probá ampliar el período o limpiar los filtros.' : 'Agregá tu primer ingreso o gasto para comenzar.'} /><button className="button primary" onClick={filtered ? () => setQuery({ page: 1, pageSize: query.pageSize }) : newMovement}>{filtered ? 'Limpiar filtros' : '+ Nuevo movimiento'}</button></div>}
        <Pagination pagination={page.pagination} disabled={loading} onPage={value => setQuery(current => ({ ...current, page: value }))} onSize={value => setQuery(current => ({ ...current, pageSize: value, page: 1 }))} />
      </>}
    </section>
    {form && <TransactionForm transaction={form.transaction} accounts={accounts} categories={categories} onSave={save} onDismiss={() => setForm(null)} />}
    {deleting && <TransactionDeleteDialog transaction={deleting} accounts={accounts} onDelete={remove} onDismiss={() => setDeleting(null)} />}
  </>;
}
