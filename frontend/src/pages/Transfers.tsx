import { useEffect, useRef, useState } from 'react';
import { EmptyState } from '../components/EmptyState';
import { ErrorState } from '../components/ErrorState';
import { LoadingState } from '../components/LoadingState';
import { Pagination } from '../components/Pagination';
import { TransferDeleteDialog } from '../components/TransferDeleteDialog';
import { TransferFilters } from '../components/TransferFilters';
import { TransferForm } from '../components/TransferForm';
import { TransferList } from '../components/TransferList';
import { useTransfers } from '../hooks/useTransfers';
import { createTransfer, deleteTransfer, getTransfer, updateTransfer } from '../services/api';
import type { Transfer, TransferInput, TransferQuery } from '../types/transfers';

export function Transfers() {
  const [query, setQuery] = useState<TransferQuery>({ page: 1, pageSize: 20 });
  const { page, accounts, loading, error, retry } = useTransfers(query);
  const [form, setForm] = useState<{ transfer?: Transfer } | null>(null);
  const [deleting, setDeleting] = useState<Transfer | null>(null);
  const [notice, setNotice] = useState('');
  const [actionError, setActionError] = useState('');
  const [editing, setEditing] = useState(false);
  const editRequest = useRef<AbortController | null>(null);
  useEffect(() => () => editRequest.current?.abort(), []);
  const filtered = Boolean(query.fromAccountId || query.toAccountId || query.from || query.to || query.search);
  useEffect(() => {
    if (!loading && !error && page && page.pagination.totalPages > 0 && page.pagination.page > page.pagination.totalPages) setQuery(current => ({ ...current, page: page.pagination.totalPages }));
  }, [page, loading, error]);
  async function edit(transfer: Transfer) {
    editRequest.current?.abort();
    const controller = new AbortController(); editRequest.current = controller;
    setEditing(true); setActionError('');
    try { const latest = await getTransfer(transfer.id, controller.signal); if (!controller.signal.aborted) setForm({ transfer: latest }); }
    catch (cause) { if (!controller.signal.aborted) setActionError(cause instanceof Error ? cause.message : 'No se pudo cargar la transferencia.'); }
    finally { if (!controller.signal.aborted) setEditing(false); }
  }
  async function save(input: TransferInput) {
    if (form?.transfer) await updateTransfer(form.transfer.id, input); else await createTransfer(input);
    setNotice(form?.transfer ? 'Transferencia actualizada.' : 'Transferencia creada.');
    setForm(null); setQuery(current => ({ ...current, page: form?.transfer ? current.page : 1 }));
  }
  async function remove() {
    if (!deleting) return;
    await deleteTransfer(deleting.id);
    setDeleting(null); setNotice('Transferencia eliminada.');
    setQuery(current => ({ ...current, page: page?.items.length === 1 && (current.page ?? 1) > 1 ? (current.page ?? 1) - 1 : current.page }));
  }
  function newTransfer() { setActionError(''); setForm({}); }
  return <>
    <div className="page-heading"><div><span className="eyebrow">Entre tus cuentas</span><h1>Tu dinero, donde lo necesitás<span>.</span></h1><p>Organizá tus fondos con transferencias claras de origen a destino.</p></div><button className="button primary new-movement" disabled={loading || Boolean(error) || editing} onClick={newTransfer}>+ Nueva transferencia</button></div>
    {notice && <div className="success-notice" role="status"><span>{notice}</span><button className="button quiet" aria-label="Cerrar confirmación" onClick={() => setNotice('')}>Cerrar ×</button></div>}
    {actionError && <p className="action-error" role="alert">{actionError}</p>}
    {editing && <p role="status" className="form-note">Cargando la transferencia para editar…</p>}
    <TransferFilters query={query} accounts={accounts} onApply={setQuery} />
    <section className="transactions-results" aria-label="Resultado de transferencias" aria-busy={loading}>
      {loading ? <LoadingState /> : error ? <ErrorState message={error} retry={retry} /> : page && <>
        {page.items.length ? <TransferList items={page.items} accounts={accounts} disabled={editing} onEdit={edit} onDelete={setDeleting} /> : <div className="transactions-empty"><EmptyState title={filtered ? 'No hay transferencias que coincidan con los filtros.' : 'Todavía no registraste transferencias entre cuentas.'} message={filtered ? 'Probá ampliar el período o limpiar los filtros.' : 'Mové dinero entre tus cuentas y mantené todo organizado.'} /><button className="button primary" disabled={editing} onClick={filtered ? () => setQuery({ page: 1, pageSize: query.pageSize }) : newTransfer}>{filtered ? 'Limpiar filtros' : '+ Nueva transferencia'}</button></div>}
        <Pagination pagination={page.pagination} disabled={loading} onPage={value => setQuery(current => ({ ...current, page: value }))} onSize={value => setQuery(current => ({ ...current, pageSize: value, page: 1 }))} />
      </>}
    </section>
    {form && <TransferForm transfer={form.transfer} accounts={accounts} onSave={save} onDismiss={() => setForm(null)} />}
    {deleting && <TransferDeleteDialog transfer={deleting} accounts={accounts} onDelete={remove} onDismiss={() => setDeleting(null)} />}
  </>;
}
