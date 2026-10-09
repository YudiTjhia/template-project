import { readFileSync } from 'node:fs';
import assert from 'node:assert/strict';
import test from 'node:test';
import { LoginApiError, LoginService } from '../LoginService';

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

test('LoginService.login sends credentials directly to the configured backend', { timeout: 10_000 }, async () => {
  const service = new LoginService({
    baseUrl: getLocalBackendUrl(),
    loginEndpoint: '/auth/login',
  });

  await assert.rejects(
    service.login({ email: 'login-test@example.invalid', password: 'invalid-test-password' }),
    (error: unknown) => {
      assert.ok(error instanceof LoginApiError);
      assert.equal(error.status, 400);
      assert.equal(error.message, 'email/mobilephone atau password tidak valid');
      return true;
    },
  );
});
