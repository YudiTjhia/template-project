import { readFileSync } from 'node:fs';
import assert from 'node:assert/strict';
import test from 'node:test';
import { AccountApiError, AccountService } from '../AccountService';
import { LoginApiError, LoginService } from '../LoginService';

const registration = {
  account_email: 'account-register-test@example.invalid',
  address: 'Test registration address',
  mobilephone: '081234567891',
  password: 'safe-password',
  confirm_password: 'safe-password',
  pin: '482617',
  confirm_pin: '482617',
  terms_accepted: true,
} as unknown as Parameters<typeof AccountService.prototype.register>[0];

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

test('valid account registration persists and login is blocked until OTP verification', { timeout: 15_000 }, async () => {
  const uniqueSuffix = Date.now().toString();
  const email = `account-register-${uniqueSuffix}@example.invalid`;
  const password = 'Safe-test-password-2026';
  const accountService = new AccountService({
    baseUrl: getLocalBackendUrl(),
    registerEndpoint: '/auth/accounts/register',
  });
  const loginService = new LoginService({
    baseUrl: getLocalBackendUrl(),
    loginEndpoint: '/auth/login',
  });

  // The backend registration contract requires additional fields beyond the
  // current public form's name/email/password values.
  const registration = {
    account_name: 'Integration Test Account',
    account_email: email,
    address: 'Integration test address',
    mobilephone: `08${uniqueSuffix.slice(-10)}`,
    password,
    confirm_password: password,
    pin: '482617',
    confirm_pin: '482617',
    terms_accepted: true,
  } as unknown as Parameters<typeof accountService.register>[0];

  const response = await accountService.register<{
    id: number;
    is_active: boolean;
    message: string;
  }>(registration);

  assert.ok(response.data.id > 0);
  assert.equal(response.data.is_active, false);
  assert.match(response.data.message, /OTP/i);

  await assert.rejects(
    loginService.login({ email, password }),
    (error: unknown) => {
      assert.ok(error instanceof LoginApiError);
      assert.equal(error.status, 400);
      assert.equal(error.message, 'user tidak aktif');
      return true;
    },
  );
});
