import { ThemeToggle } from './ThemeToggle';

export const sections = ['Dashboard', 'Movimientos', 'Cuentas', 'Categorías', 'Transferencias'] as const;
export type Section = typeof sections[number];

export function Navigation({ active, onChange }: { active: Section; onChange: (section: Section) => void }) {
  return <header className="site-header"><a className="brand" href="/" onClick={event => { event.preventDefault(); onChange('Dashboard'); }}><span className="brand-mark" aria-hidden="true">f<span>.</span></span><span>Mis Finanzas<small>Tu dinero, con perspectiva.</small></span></a><nav aria-label="Navegación principal">{sections.map(section => <button key={section} className={section === active ? 'nav-link active' : 'nav-link'} aria-current={section === active ? 'page' : undefined} onClick={() => onChange(section)}>{section}</button>)}</nav><ThemeToggle /></header>;
}
