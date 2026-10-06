import { useEffect, useState } from 'react';
import { EmptyState } from './components/EmptyState';
import { Navigation } from './components/Navigation';
import type { Section } from './components/Navigation';
import { Dashboard } from './pages/Dashboard';
import { Transactions } from './pages/Transactions';
import { Transfers } from './pages/Transfers';

const paths: Record<Section, string> = { Dashboard: '/', Movimientos: '/movimientos', Cuentas: '/cuentas', Categorías: '/categorias', Transferencias: '/transferencias' };
function currentSection(): Section { return (Object.keys(paths) as Section[]).find(section => paths[section] === (window.location.pathname.replace(/\/$/, '') || '/')) ?? 'Dashboard'; }

export default function App() {
  const [section, setSection] = useState<Section>(currentSection);
  useEffect(() => { const handleBack = () => setSection(currentSection()); window.addEventListener('popstate', handleBack); return () => window.removeEventListener('popstate', handleBack); }, []);
  function navigate(next: Section) { if (window.location.pathname !== paths[next]) window.history.pushState(null, '', paths[next]); setSection(next); requestAnimationFrame(() => document.getElementById('main')?.focus()); }
  return <><a className="skip-link" href="#main">Saltar al contenido</a><div className="app-shell"><Navigation active={section} onChange={navigate} /><main id="main" tabIndex={-1}>{section === 'Dashboard' ? <Dashboard /> : section === 'Movimientos' ? <Transactions /> : section === 'Transferencias' ? <Transfers /> : <div className="upcoming-page"><div className="page-heading"><div><span className="eyebrow">Tu espacio financiero</span><h1>{section}<span>.</span></h1><p>{section === 'Cuentas' ? 'Un lugar para cada parte de tu dinero.' : 'Más claridad en cada ingreso y gasto.'}</p></div><span className="status-badge">En preparación</span></div><div className="upcoming-content"><EmptyState title={section === 'Cuentas' ? 'Tus cuentas, en un solo lugar' : 'Dale contexto a tus movimientos'} message={section === 'Cuentas' ? 'La gestión visual de cuentas estará disponible en una próxima etapa. Podés consultar los saldos actuales desde el Dashboard.' : 'La gestión visual de categorías estará disponible en una próxima etapa. Podés ver la distribución de ingresos y gastos desde el Dashboard.'} /><button className="button primary" onClick={() => navigate('Dashboard')}>Ver mi panorama <span aria-hidden="true">↗</span></button></div></div>}</main><footer className="site-footer"><span>Mis Finanzas <span aria-hidden="true">/</span> Tu dinero, con perspectiva.</span><span>Importes en unidad monetaria de cada cuenta.</span></footer></div></>;
}
