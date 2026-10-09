import assert from 'node:assert/strict';
import test from 'node:test';
import { UserService, type UserRecord } from '../UserService';

type RecordedRequest = {
  url: string;
  method: string;
  authorization: string | null;
  body?: unknown;
};

test('UserService list and CRUD methods call the expected endpoints', async () => {
  const user: UserRecord = {
    id: 17,
    username: 'qa-user',
    user_email: 'qa-user@example.invalid',
    mobilephone: '081234567890',
    user_type: 'user',
    login_type: 'cashier',
    user_roles: ['QA'],
    is_active: true,
  };
  const responses: unknown[] = [
    { data: [user] },
    { data: user },
    { id: user.id },
    { message: 'updated' },
    { message: 'deleted' },
  ];
  const requests: RecordedRequest[] = [];
  const service = new UserService({
    baseUrl: 'http://backend.test',
    tokenProvider: () => 'test-session-token',
    fetcher: async (input, init) => {
      const body = typeof init?.body === 'string' ? JSON.parse(init.body) : undefined;
      requests.push({
        url: String(input),
        method: init?.method ?? 'GET',
        authorization: new Headers(init?.headers).get('Authorization'),
        body,
      });
      return new Response(JSON.stringify(responses.shift()), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      });
    },
  });
  const payload = {
    username: 'qa-user-updated',
    user_email: 'qa-user@example.invalid',
    mobilephone: '081234567890',
    user_type: 'user',
    login_type: 'cashier',
    user_roles: ['QA'],
    is_active: true,
    user_password: 'test-password',
  };

  assert.deepEqual(await service.list('qa-user'), [user]);
  assert.deepEqual(await service.get(user.id), user);
  assert.deepEqual(await service.create(payload), { id: user.id });
  await service.update(user.id, payload);
  await service.remove(user.id);

  assert.deepEqual(requests.map(({ method }) => method), ['GET', 'GET', 'POST', 'PUT', 'DELETE']);
  assert.deepEqual(requests.map(({ url }) => url), [
    'http://backend.test/api/users?search=qa-user',
    'http://backend.test/api/users/17',
    'http://backend.test/api/users',
    'http://backend.test/api/users/17',
    'http://backend.test/api/users/17',
  ]);
  assert.ok(requests.every(({ authorization }) => authorization === 'Bearer test-session-token'));
  assert.deepEqual(requests[2].body, payload);
  assert.deepEqual(requests[3].body, payload);
  assert.equal(requests[4].body, undefined);
});
