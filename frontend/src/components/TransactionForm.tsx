import { useState } from 'react';
import type { FormEvent } from 'react';
import type { Account, Category, Transaction, TransactionInput, TransactionType } from '../types/transactions';
import { moneyInputValue, parseMoneyInput, todayDate } from '../utils/money';
import { Dialog } from './Dialog';
import { MoneyInput } from './MoneyInput';

export function TransactionForm({ transaction, accounts, categories, onSave, onDismiss }: { transaction?: Transaction; accounts: Account[]; categories: Category[]; onSave: (input: TransactionInput) => Promise<void>; onDismiss: () => void }) {
  const [type, setType] = useState<TransactionType>(transaction?.type ?? 'expense');
  const [amount, setAmount] = useState(transaction ? moneyInputValue(transaction.amount) : '');
  const [description, setDescription] = useState(transaction?.description ?? '');
  const [categoryId, setCategoryId] = useState(transaction ? String(transaction.categoryId) : '');
  const [accountId, setAccountId] = useState(transaction && accounts.some(account => account.id === transaction.accountId && account.active) ? String(transaction.accountId) : '');
  const [date, setDate] = useState(transaction?.date.slice(0, 10) ?? todayDate());
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  const activeAccounts = accounts.filter(account => account.active);
  const availableCategories = categories.filter(category => category.type === type);
  function changeType(value: TransactionType) {
    setType(value);
    if (!categories.some(category => String(category.id) === categoryId && category.type === value)) setCategoryId('');
  }
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); if (busy) return;
    setError('');
    try {
      const minorAmount = parseMoneyInput(amount);
      if (!activeAccounts.some(account => String(account.id) === accountId)) throw new Error('Seleccioná una cuenta activa.');
      if (!availableCategories.some(category => String(category.id) === categoryId)) throw new Error('Seleccioná una categoría del tipo correspondiente.');
      if (!/^\d{4}-\d{2}-\d{2}$/.test(date) || Number.isNaN(Date.parse(`${date}T00:00:00Z`))) throw new Error('Seleccioná una fecha válida.');
      setBusy(true);
      await onSave({ type, amount: minorAmount, description: description.trim(), categoryId: Number(categoryId), accountId: Number(accountId), date: date === transaction?.date.slice(0, 10) ? transaction.date : `${date}T00:00:00Z` });
    } catch (cause) { setError(cause instanceof Error ? cause.message : 'No se pudo guardar el movimiento.'); }
    finally { setBusy(false); }
  }
  return <Dialog title={transaction ? 'Editar movimiento' : 'Nuevo movimiento'} busy={busy} onDismiss={onDismiss}>
    <form className="transaction-form" onSubmit={submit} noValidate>
      <fieldset disabled={busy}><legend className="sr-only">Datos del movimiento</legend>
        <div className="type-switch" role="group" aria-label="Tipo de movimiento"><button type="button" aria-pressed={type === 'income'} className={type === 'income' ? 'selected income' : ''} onClick={() => changeType('income')}>↗ Ingreso</button><button type="button" aria-pressed={type === 'expense'} className={type === 'expense' ? 'selected expense' : ''} onClick={() => changeType('expense')}>↘ Gasto</button></div>
        <MoneyInput value={amount} onChange={setAmount} currency={accounts.find(account => String(account.id) === accountId)?.currency} />
        <label>Descripción <span className="optional">Opcional</span><input value={description} onChange={event => setDescription(event.target.value)} placeholder="¿En qué consistió este movimiento?" /></label>
        <div className="form-grid"><label>Cuenta<select value={accountId} onChange={event => setAccountId(event.target.value)} required><option value="">Seleccioná una cuenta</option>{activeAccounts.map(account => <option key={account.id} value={account.id}>{account.name} · {account.currency}</option>)}</select></label><label>Categoría<select value={categoryId} onChange={event => setCategoryId(event.target.value)} required><option value="">Seleccioná una categoría</option>{availableCategories.map(category => <option key={category.id} value={category.id}>{category.name}</option>)}</select></label></div>
        <label>Fecha<input type="date" value={date} onChange={event => setDate(event.target.value)} required /></label>
        {transaction && !accountId && <p className="form-note">La cuenta original no está activa o ya no existe. Elegí una cuenta activa para guardar.</p>}
        {!activeAccounts.length && <p className="form-note">Necesitás al menos una cuenta activa para registrar un movimiento.</p>}
        {!availableCategories.length && <p className="form-note">No hay categorías para este tipo de movimiento. Creá una desde la API antes de continuar.</p>}
      </fieldset>
      {error && <p className="field-error" role="alert">{error}</p>}
      <footer className="dialog-actions"><button type="button" className="button quiet" disabled={busy} onClick={onDismiss}>Cancelar</button><button className="button primary" disabled={busy || !activeAccounts.length || !availableCategories.length}>{busy ? 'Guardando…' : transaction ? 'Guardar cambios' : 'Crear movimiento'}</button></footer>
    </form>
  </Dialog>;
}
