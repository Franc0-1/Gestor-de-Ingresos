import { useState } from 'react';
import type { FormEvent } from 'react';
import type { Account } from '../types/transactions';
import type { Transfer, TransferInput } from '../types/transfers';
import { moneyInputValue, todayDate } from '../utils/money';
import { transferInput } from '../utils/transfer';
import { Dialog } from './Dialog';
import { MoneyInput } from './MoneyInput';

export function TransferForm({ transfer, accounts, onSave, onDismiss }: { transfer?: Transfer; accounts: Account[]; onSave: (input: TransferInput) => Promise<void>; onDismiss: () => void }) {
  const active = accounts.filter(account => account.active);
  const [fields, setFields] = useState({
    fromAccountId: transfer && active.some(account => account.id === transfer.fromAccountId) ? String(transfer.fromAccountId) : '',
    toAccountId: transfer && active.some(account => account.id === transfer.toAccountId) ? String(transfer.toAccountId) : '',
    amount: transfer ? moneyInputValue(transfer.amount) : '', description: transfer?.description ?? '', date: transfer?.date.slice(0, 10) ?? todayDate(),
  });
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const from = active.find(account => String(account.id) === fields.fromAccountId);
  const to = active.find(account => String(account.id) === fields.toAccountId);
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); if (busy) return;
    setError('');
    try {
      const input = transferInput(fields, accounts);
      if (fields.date === transfer?.date.slice(0, 10)) input.date = transfer.date;
      setBusy(true); await onSave(input);
    } catch (cause) { setError(cause instanceof Error ? cause.message : 'No se pudo guardar la transferencia.'); }
    finally { setBusy(false); }
  }
  return <Dialog title={transfer ? 'Editar transferencia' : 'Nueva transferencia'} busy={busy} onDismiss={onDismiss}><form className="transaction-form" onSubmit={submit} noValidate><fieldset disabled={busy}><legend className="sr-only">Datos de la transferencia</legend>
    <div className="form-grid"><label>Cuenta origen<select required value={fields.fromAccountId} onChange={event => setFields({ ...fields, fromAccountId: event.target.value })}><option value="">Seleccioná origen</option>{active.map(account => <option key={account.id} value={account.id} disabled={String(account.id) === fields.toAccountId}>{account.name} · {account.currency}</option>)}</select></label><label>Cuenta destino<select required value={fields.toAccountId} onChange={event => setFields({ ...fields, toAccountId: event.target.value })}><option value="">Seleccioná destino</option>{active.map(account => <option key={account.id} value={account.id} disabled={String(account.id) === fields.fromAccountId}>{account.name} · {account.currency}</option>)}</select></label></div>
    <MoneyInput value={fields.amount} onChange={amount => setFields({ ...fields, amount })} currency={from?.currency} />
    {from && to && from.currency !== to.currency && <p className="form-note">Las cuentas tienen monedas diferentes ({from.currency} → {to.currency}). Se registra el mismo importe en ambas cuentas, sin conversión.</p>}
    <label>Descripción <span className="optional">Opcional</span><input value={fields.description} onChange={event => setFields({ ...fields, description: event.target.value })} placeholder="Mover dinero para gastos" /></label>
    <label>Fecha<input type="date" required value={fields.date} onChange={event => setFields({ ...fields, date: event.target.value })} /></label>
    {transfer && (!fields.fromAccountId || !fields.toAccountId) && <p className="form-note">Una cuenta original ya no está activa o no existe. Elegí cuentas activas para guardar.</p>}
    {active.length < 2 && <p className="form-note">Necesitás al menos dos cuentas activas para transferir dinero.</p>}
  </fieldset>{error && <p className="field-error" role="alert">{error}</p>}<footer className="dialog-actions"><button type="button" className="button quiet" disabled={busy} onClick={onDismiss}>Cancelar</button><button className="button primary" disabled={busy || active.length < 2}>{busy ? 'Guardando…' : transfer ? 'Guardar cambios' : 'Crear transferencia'}</button></footer></form></Dialog>;
}
