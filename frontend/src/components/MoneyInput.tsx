import { useId } from 'react';

export function MoneyInput({ value, onChange, currency }: { value: string; onChange: (value: string) => void; currency?: string }) {
  const helpId = useId();
  return <label className="amount-field">Monto <span>{currency || 'Seleccioná una cuenta para ver la moneda'}</span><input autoFocus inputMode="decimal" value={value} onChange={event => onChange(event.target.value)} placeholder="18500,50" aria-describedby={helpId} required /><small id={helpId}>Sin separadores de miles · Hasta 2 decimales</small></label>;
}
