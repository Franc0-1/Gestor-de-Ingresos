import type { DashboardResponse, PeriodFilter } from '../types/dashboard';
import type { Account, Category, Transaction, TransactionInput, TransactionPage, TransactionQuery } from '../types/transactions';
import type { Transfer, TransferInput, TransferPage, TransferQuery } from '../types/transfers';

export class ApiError extends Error {}

const moneyKeys = new Set(['amount', 'initialBalance', 'balance', 'totalBalance', 'periodIncome', 'periodExpenses', 'netSavings']);

function parseResponse(text: string): unknown {
  return JSON.parse(text, (key: string, value: unknown, context?: { source?: string }): unknown => {
    if (typeof value !== 'number') return value;
    if (moneyKeys.has(key)) {
      if (context?.source && /^-?\d+$/.test(context.source)) return BigInt(context.source);
      if (Number.isSafeInteger(value)) return BigInt(value);
      throw new ApiError('Este navegador no puede representar este importe con precisión. Usá un navegador actualizado.');
    }
    if (!Number.isSafeInteger(value)) throw new ApiError('La respuesta contiene un identificador fuera del rango admitido.');
    return value;
  }) as unknown;
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  let response: Response;
  try {
    response = await fetch(`/api${path}`, { ...options, headers: { Accept: 'application/json', ...(options.body ? { 'Content-Type': 'application/json' } : {}), ...options.headers } });
  } catch (cause) {
    if (options.signal?.aborted) throw cause;
    throw new ApiError('No pudimos conectar con el servidor. Comprobá que esté iniciado e intentá de nuevo.');
  }
  if (!response.ok) {
    if (response.status === 404 && path.startsWith('/transfers/')) throw new ApiError('La transferencia ya no existe. Actualizá el listado.');
    if (response.status === 400 && path.startsWith('/transfers') && options.body) throw new ApiError('Revisá los datos: las cuentas deben existir, estar activas y ser diferentes; el monto y la fecha deben ser válidos.');
    if (response.status === 404) throw new ApiError(path.startsWith('/transactions/') ? 'El movimiento ya no existe. Actualizá el listado.' : 'No se encontró el recurso solicitado.');
    if (response.status === 400) throw new ApiError(path.startsWith('/transactions') && options.body ? 'Revisá los datos: la cuenta y categoría deben existir, y la categoría debe coincidir con el tipo de movimiento.' : 'Revisá los filtros y las fechas seleccionadas.');
    throw new ApiError('No se pudo completar la operación. Intentá de nuevo en unos instantes.');
  }
  if (response.status === 204) return undefined as T;
  try { return parseResponse(await response.text()) as T; }
  catch (cause) { if (cause instanceof ApiError) throw cause; throw new ApiError('El servidor devolvió una respuesta inesperada.'); }
}

export async function getDashboard(filter: PeriodFilter, signal: AbortSignal): Promise<DashboardResponse> {
  const params = new URLSearchParams();
  if (filter.from) params.set('from', filter.from);
  if (filter.to) params.set('to', filter.to);
  const result = await request<unknown>(`/dashboard${params.size ? `?${params}` : ''}`, { signal });
  if (!result || typeof result !== 'object' || !('period' in result) || !('accounts' in result) || !('summaryByCurrency' in result)) {
    throw new ApiError('El servidor devolvió una respuesta inesperada.');
  }
  return result as DashboardResponse;
}

export function getAccounts(signal?: AbortSignal): Promise<Account[]> { return request('/accounts', { signal }); }
export function getCategories(signal?: AbortSignal): Promise<Category[]> { return request('/categories', { signal }); }
export function getTransactions(query: TransactionQuery, signal?: AbortSignal): Promise<TransactionPage> {
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(query)) if (value !== undefined && value !== '') params.set(key, String(value));
  return request(`/transactions?${params}`, { signal });
}
export function getTransaction(id: number, signal?: AbortSignal): Promise<Transaction> { return request(`/transactions/${id}`, { signal }); }

// Serialize a bigint as an exact JSON number, which Go decodes directly as int64.
function moneyBody(input: TransactionInput | TransferInput): string {
  const { amount, ...fields } = input;
  return `${JSON.stringify(fields).slice(0, -1)},"amount":${amount.toString()}}`;
}
export function createTransaction(input: TransactionInput): Promise<Transaction> { return request('/transactions', { method: 'POST', body: moneyBody(input) }); }
export function updateTransaction(id: number, input: TransactionInput): Promise<Transaction> { return request(`/transactions/${id}`, { method: 'PUT', body: moneyBody(input) }); }
export function deleteTransaction(id: number): Promise<void> { return request(`/transactions/${id}`, { method: 'DELETE' }); }

export function getTransfers(query: TransferQuery, signal?: AbortSignal): Promise<TransferPage> {
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(query)) if (value !== undefined && value !== '') params.set(key, String(value));
  return request(`/transfers?${params}`, { signal });
}
export function getTransfer(id: number, signal?: AbortSignal): Promise<Transfer> { return request(`/transfers/${id}`, { signal }); }
export function createTransfer(input: TransferInput): Promise<Transfer> { return request('/transfers', { method: 'POST', body: moneyBody(input) }); }
export function updateTransfer(id: number, input: TransferInput): Promise<Transfer> { return request(`/transfers/${id}`, { method: 'PUT', body: moneyBody(input) }); }
export function deleteTransfer(id: number): Promise<void> { return request(`/transfers/${id}`, { method: 'DELETE' }); }
