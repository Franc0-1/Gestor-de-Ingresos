import type { CurrencyBalance } from '../types/dashboard';
import { formatMoney } from '../utils/format';
import { EmptyState } from './EmptyState';

export function BalanceSummary({ balances, accountCount }: { balances: CurrencyBalance[]; accountCount: number }) {
  return <section className="balance-summary" aria-labelledby="balance-title"><div className="balance-heading"><span className="eyebrow" id="balance-title">Tu balance actual</span><span className="live-label"><i aria-hidden="true" />{accountCount} {accountCount === 1 ? 'cuenta activa' : 'cuentas activas'}</span></div>{balances.length ? <div className="balance-values">{balances.map(balance => <div className="balance-value" key={balance.currency}><span className="currency-tag">{balance.currency || 'Sin moneda'}</span><strong>{formatMoney(balance.amount, balance.currency)}</strong></div>)}</div> : <EmptyState title="Tu punto de partida" message="Los balances aparecerán cuando tengas cuentas activas." />}<div className="balance-foot"><span>El panorama completo de tu dinero.</span><small>{balances.length > 1 ? 'Cada moneda se muestra por separado, sin conversión.' : 'Saldo inicial + movimientos + transferencias.'}</small></div></section>;
}
