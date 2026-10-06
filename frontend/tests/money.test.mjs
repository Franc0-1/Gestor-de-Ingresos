import assert from 'node:assert/strict';
import test from 'node:test';
import { parseMoneyInput, moneyInputValue, MAX_AMOUNT } from '../src/utils/money.ts';

test('parse and edit amounts without floating point arithmetic', () => {
  for (const [input, amount] of [['18500,50', 1850050n], ['0,01', 1n], ['100', 10000n], [' 12.3 ', 1230n], ['92233720368547758,07', MAX_AMOUNT]]) {
    assert.equal(parseMoneyInput(input), amount);
    assert.equal(parseMoneyInput(moneyInputValue(amount)), amount);
  }
});
test('reject invalid, non-positive, overly precise and overflowing amounts', () => {
  for (const input of ['', '0', '-1', '1,234', '1e3', '18.500,50', 'abc', '92233720368547758,08']) assert.throws(() => parseMoneyInput(input));
});
