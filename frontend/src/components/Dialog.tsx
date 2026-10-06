import { useEffect, useRef } from 'react';
import type { ReactNode } from 'react';

export function Dialog({ title, busy, onDismiss, children }: { title: string; busy: boolean; onDismiss: () => void; children: ReactNode }) {
  const ref = useRef<HTMLDialogElement>(null);
  useEffect(() => {
    const previous = document.activeElement;
    const dialog = ref.current;
    dialog?.showModal();
    return () => { dialog?.close(); if (previous instanceof HTMLElement && previous.isConnected) previous.focus(); };
  }, []);
  return <dialog ref={ref} className="finance-dialog" aria-labelledby="dialog-title" onCancel={event => { event.preventDefault(); if (!busy) onDismiss(); }}>
    <header className="dialog-heading"><div><span className="eyebrow">Tu actividad financiera</span><h2 id="dialog-title">{title}</h2></div><button type="button" className="button quiet" aria-label="Cerrar ventana" disabled={busy} onClick={onDismiss}>Cerrar ×</button></header>
    {children}
  </dialog>;
}
