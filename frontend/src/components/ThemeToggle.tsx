import { useState } from 'react';

export function ThemeToggle() {
  const [dark, setDark] = useState(() => document.documentElement.dataset.theme === 'dark');
  function toggle() {
    const next = !dark;
    setDark(next);
    document.documentElement.dataset.theme = next ? 'dark' : 'light';
    try { localStorage.setItem('finance-theme', next ? 'dark' : 'light'); } catch { /* Keep the toggle usable without storage. */ }
    document.querySelector('meta[name="theme-color"]')?.setAttribute('content', next ? '#151f1d' : '#f5f5f0');
  }
  return <button className="theme-toggle" type="button" onClick={toggle} aria-pressed={dark} aria-label="Modo oscuro" title={dark ? 'Cambiar a modo claro' : 'Cambiar a modo oscuro'}><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.6" aria-hidden="true">{dark ? <><circle cx="12" cy="12" r="4" /><path d="M12 2v2m0 16v2M2 12h2m16 0h2M5 5l1.5 1.5m11 11L19 19M5 19l1.5-1.5m11-11L19 5" /></> : <path d="M20.8 14.2A9 9 0 0 1 9.8 3.3a9 9 0 1 0 11 10.9Z" />}</svg><span>{dark ? 'Modo claro' : 'Modo oscuro'}</span></button>;
}
