import { readFileSync } from 'node:fs';
import assert from 'node:assert/strict';
import test from 'node:test';
import { AccountApiError, AccountService } from '../AccountService';

const registration = {
  name: 'Test Account',
  email: 'account-register-test@example.invalid',
  password: 'safe-password',
};

function getLocalBackendUrl() {
  const configuredPort = process.env.BACKEND_PORT;
  const backendEnv = readFileSync(
    new URL('../../../../../infra/dev/backend/.env', import.meta.url),
    'utf8',
  );
  const envPort = backendEnv.match(/^\s*BACKEND_PORT\s*=\s*(\d+)\s*$/m)?.[1];
  const port = configuredPort ?? envPort;

  if (!port) {
    throw new Error('BACKEND_PORT tidak ditemukan di infra/dev/backend/.env.');
  }

  return `http://localhost:${port}`;
}

test('AccountService.register sends the request directly to the configured backend', { timeout: 10_000 }, async () => {
  const service = new AccountService({
    baseUrl: getLocalBackendUrl(),
    registerEndpoint: '/auth/accounts/register',
  });

  await assert.rejects(service.register(registration), (error: unknown) => {
    assert.ok(error instanceof AccountApiError);
    assert.equal(error.status, 400);
    assert.equal(error.message, 'account_name wajib diisi');
    return true;
  });
});
