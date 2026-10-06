import { useEffect, useState } from 'react';
import type { FormEvent } from 'react';
import type { Period, PeriodFilter } from '../types/dashboard';

export function PeriodSelector({ period, onChange, disabled }: { period?: Period; onChange: (filter: PeriodFilter) => void; disabled: boolean }) {
  const [from, setFrom] = useState(period?.from ?? '');
  const [to, setTo] = useState(period?.to ?? '');
  const [error, setError] = useState('');
  useEffect(() => { if (period) { setFrom(period.from); setTo(period.to); } }, [period]);
  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!from || !to || from > to) { setError('Elegí un rango válido: desde debe ser anterior o igual a hasta.'); return; }
    setError('');
    onChange({ from, to });
  }
  return <form className="period-selector" onSubmit={submit} aria-label="Seleccionar período"><div className="date-fields"><label>Desde<input type="date" value={from} onChange={event => setFrom(event.target.value)} required disabled={disabled} aria-describedby={error ? 'period-error' : undefined} /></label><span aria-hidden="true">—</span><label>Hasta<input type="date" value={to} onChange={event => setTo(event.target.value)} required disabled={disabled} aria-describedby={error ? 'period-error' : undefined} /></label></div><button type="submit" className="button primary" disabled={disabled}>Aplicar</button><button type="button" className="button quiet" disabled={disabled} onClick={() => { setError(''); onChange({}); }}>Este mes</button>{error && <p className="field-error" id="period-error" role="alert">{error}</p>}</form>;
}
