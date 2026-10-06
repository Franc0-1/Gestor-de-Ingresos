import assert from 'node:assert/strict';
import test from 'node:test';
import { transferInput } from '../src/utils/transfer.ts';
import { createTransfer, updateTransfer, getTransfer, getTransfers, deleteTransfer } from '../src/services/api.ts';

const accounts = [{ id: 1, active: true, currency: 'ARS' }, { id: 2, active: true, currency: 'USD' }, { id: 3, active: false, currency: 'ARS' }];
const fields = { fromAccountId: '1', toAccountId: '2', amount: '50000,50', description: '  Mover dinero  ', date: '2026-10-06' };

test('transfer form keeps exact amounts and permits different currencies without conversion', () => {
  assert.deepEqual(transferInput(fields, accounts), { fromAccountId: 1, toAccountId: 2, amount: 5000050n, description: 'Mover dinero', date: '2026-10-06T00:00:00Z' });
});

test('transfer form rejects missing, inactive or identical accounts, invalid amounts and dates', () => {
  for (const changes of [
    { fromAccountId: '' }, { toAccountId: '' }, { fromAccountId: '3' }, { toAccountId: '3' }, { toAccountId: '99' }, { toAccountId: '1' },
    { amount: '' }, { amount: '0' }, { amount: '-1' }, { amount: '1,234' }, { amount: '92233720368547758,08' },
    { date: '' }, { date: '2026-02-30' },
  ]) assert.throws(() => transferInput({ ...fields, ...changes }, accounts));
});

test('transfer CRUD uses existing endpoints and serializes int64 exactly', async (t) => {
  const previous = globalThis.fetch;
  t.after(() => { globalThis.fetch = previous; });
  const calls = [];
  globalThis.fetch = async (url, options) => {
    calls.push([url, options.method ?? 'GET']);
    if (options.method === 'DELETE') return new Response(null, { status: 204 });
    if (options.body) {
      assert.ok(options.body.includes('"amount":9223372036854775807}'));
      assert.equal(JSON.parse(options.body).fromAccountId, 1);
      assert.equal(JSON.parse(options.body).toAccountId, 2);
      assert.equal(JSON.parse(options.body).id, undefined);
    }
    return new Response('{"id":4,"amount":9223372036854775807}', { status: options.method === 'POST' ? 201 : 200 });
  };
  const input = { ...transferInput(fields, accounts), amount: 9223372036854775807n };
  assert.equal((await createTransfer(input)).amount, input.amount);
  assert.equal((await getTransfer(4)).amount, input.amount);
  assert.equal((await updateTransfer(4, input)).amount, input.amount);
  assert.equal(await deleteTransfer(4), undefined);
  assert.deepEqual(calls, [['/api/transfers', 'POST'], ['/api/transfers/4', 'GET'], ['/api/transfers/4', 'PUT'], ['/api/transfers/4', 'DELETE']]);
});

test('transfer search, filters, order and pagination are sent to the server', async (t) => {
  const previous = globalThis.fetch;
  t.after(() => { globalThis.fetch = previous; });
  const query = { fromAccountId: '1', toAccountId: '2', from: '2026-10-01', to: '2026-10-31', search: 'Mover & ahorrar', sortBy: 'amount', sortOrder: 'asc', page: 2, pageSize: 10 };
  globalThis.fetch = async url => {
    const parsed = new URL(url, 'http://localhost');
    assert.equal(parsed.pathname, '/api/transfers');
    for (const [key, value] of Object.entries(query)) assert.equal(parsed.searchParams.get(key), String(value));
    return new Response('{"items":[{"id":4,"amount":5000050}],"pagination":{"page":2,"pageSize":10,"totalItems":11,"totalPages":2}}');
  };
  const page = await getTransfers(query);
  assert.equal(page.items[0].amount, 5000050n);
  assert.equal(page.pagination.totalItems, 11);
});

test('transfer errors are understandable and hide server details', async (t) => {
  const previous = globalThis.fetch;
  t.after(() => { globalThis.fetch = previous; });
  for (const status of [400, 404, 500]) {
    globalThis.fetch = async () => new Response('{"error":"private workbook path"}', { status });
    await assert.rejects(updateTransfer(4, transferInput(fields, accounts)), error => !error.message.includes('private') && (status !== 404 || error.message.includes('transferencia')));
  }
});
