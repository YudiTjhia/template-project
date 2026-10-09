import assert from 'node:assert/strict';
import test from 'node:test';
import { RoleService, type RoleActionRecord, type RoleRecord } from '../RoleService';

type RecordedRequest = {
  url: string;
  method: string;
  authorization: string | null;
  body?: unknown;
};

test('RoleService list, action lookup, and CRUD methods call the expected endpoints', async () => {
  const role: RoleRecord = { id: 23, role_name: 'QA Role', actions: ['users.read'] };
  const action: RoleActionRecord = {
    id: 4,
    value: 'users.read',
    label: 'Read users',
    sort_order: 1,
    active: true,
  };
  const responses: unknown[] = [
    { data: [role] },
    { data: role },
    { data: [action] },
    { id: role.id },
    { message: 'updated' },
    { message: 'deleted' },
  ];
  const requests: RecordedRequest[] = [];
  const service = new RoleService({
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
  const payload = { role_name: 'QA Role Updated', actions: ['users.read', 'users.write'] };

  assert.deepEqual(await service.list('QA'), [role]);
  assert.deepEqual(await service.get(role.id), role);
  assert.deepEqual(await service.listActions('users'), [action]);
  assert.deepEqual(await service.create(payload), { id: role.id });
  await service.update(role.id, payload);
  await service.remove(role.id);

  assert.deepEqual(requests.map(({ method }) => method), ['GET', 'GET', 'GET', 'POST', 'PUT', 'DELETE']);
  assert.deepEqual(requests.map(({ url }) => url), [
    'http://backend.test/api/roles?search=QA',
    'http://backend.test/api/roles/23',
    'http://backend.test/api/roles/actions?search=users',
    'http://backend.test/api/roles',
    'http://backend.test/api/roles/23',
    'http://backend.test/api/roles/23',
  ]);
  assert.ok(requests.every(({ authorization }) => authorization === 'Bearer test-session-token'));
  assert.deepEqual(requests[3].body, payload);
  assert.deepEqual(requests[4].body, payload);
  assert.equal(requests[5].body, undefined);
});
