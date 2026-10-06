import type { AccountBalance } from '../types/dashboard';
import { formatMoney } from '../utils/format';
import { EmptyState } from './EmptyState';

export function AccountBalances({ accounts }: { accounts: AccountBalance[] }) {
  return <section className="accounts-panel" aria-labelledby="accounts-title"><div className="section-heading"><div><span className="eyebrow">Dónde está tu dinero</span><h2 id="accounts-title">Tus cuentas</h2></div><span className="count">{accounts.length.toString().padStart(2, '0')}</span></div>{!accounts.length ? <EmptyState title="Sin cuentas activas" message="Las cuentas activas aparecerán aquí con su saldo actual." /> : <ul className="account-list">{accounts.map(account => <li key={account.accountId}><span className="account-monogram" aria-hidden="true">{account.accountName.charAt(0).toUpperCase() || 'C'}</span><div><strong>{account.accountName || `Cuenta ${account.accountId}`}</strong><small>{account.currency || 'Moneda sin identificar'}</small></div><span className={account.balance < 0n ? 'account-amount negative' : 'account-amount'}>{formatMoney(account.balance, account.currency)}</span></li>)}</ul>}<p className="panel-footnote">Saldos de todo el historial · Solo cuentas activas</p></section>;
}
