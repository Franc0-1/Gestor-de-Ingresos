import assert from 'node:assert/strict';
import test from 'node:test';
import { createTransaction, getTransactions, deleteTransaction } from '../src/services/api.ts';

test('send exact int64 JSON numbers and decode money as bigint', async (t) => {
  const previous = globalThis.fetch;
  t.after(() => { globalThis.fetch = previous; });
  globalThis.fetch = async (url, options) => {
    assert.equal(url, '/api/transactions');
    assert.equal(options.method, 'POST');
    assert.ok(options.body.includes('"amount":9223372036854775807}'));
    assert.equal(JSON.parse(options.body).description, 'Texto "con comillas"');
    return new Response('{"id":1,"amount":9223372036854775807}', { status: 201 });
  };
  const result = await createTransaction({ type: 'income', amount: 9223372036854775807n, description: 'Texto "con comillas"', categoryId: 1, accountId: 1, date: '2026-10-06T00:00:00Z' });
  assert.equal(result.amount, 9223372036854775807n);
});
test('query server-side pagination and handle a bodyless deletion', async (t) => {
  const previous = globalThis.fetch;
  t.after(() => { globalThis.fetch = previous; });
  globalThis.fetch = async (url, options) => {
    if (options.method === 'DELETE') { assert.equal(url, '/api/transactions/4'); return new Response(null, { status: 204 }); }
    const params = new URL(url, 'http://localhost').searchParams;
    assert.equal(params.get('search'), 'super & sueldo');
    assert.equal(params.get('page'), '2');
    assert.equal(params.get('pageSize'), '10');
    assert.equal(params.has('accountId'), false);
    return new Response('{"items":[],"pagination":{"page":2,"pageSize":10,"totalItems":11,"totalPages":2}}');
  };
  const page = await getTransactions({ search: 'super & sueldo', accountId: '', page: 2, pageSize: 10 });
  assert.equal(page.pagination.totalItems, 11);
  assert.equal(await deleteTransaction(4), undefined);
});
test('do not display raw internal backend errors', async (t) => {
  const previous = globalThis.fetch;
  t.after(() => { globalThis.fetch = previous; });
  globalThis.fetch = async () => new Response('{"error":"sensitive workbook path"}', { status: 500 });
  await assert.rejects(getTransactions({}), error => !error.message.includes('sensitive'));
});
