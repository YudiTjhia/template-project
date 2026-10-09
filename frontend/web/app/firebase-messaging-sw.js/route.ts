import { NextResponse } from 'next/server';
import { firebaseWebConfig, isFirebaseWebConfigured } from '@/app/firebase/config';

export function GET() {
  if (!isFirebaseWebConfigured()) {
    return new NextResponse('// Firebase Cloud Messaging is not configured.', {
      headers: { 'Content-Type': 'application/javascript; charset=utf-8', 'Cache-Control': 'no-store' },
    });
  }

  const config = JSON.stringify(firebaseWebConfig);
  const notificationsPath = `/${process.env.NEXT_PUBLIC_SUB_APP_TYPE === 'customer' ? 'customer' : 'admin'}/notifications`;
  const worker = `
importScripts('https://www.gstatic.com/firebasejs/13.0.0/firebase-app-compat.js');
importScripts('https://www.gstatic.com/firebasejs/13.0.0/firebase-messaging-compat.js');
firebase.initializeApp(${config});
const messaging = firebase.messaging();
messaging.onBackgroundMessage((payload) => {
  const title = payload.notification?.title || 'Notifikasi';
  const options = {
    body: payload.notification?.body || '',
    data: payload.data || {},
  };
  self.registration.showNotification(title, options);
});
self.addEventListener('notificationclick', (event) => {
  event.notification.close();
  event.waitUntil(clients.openWindow(${JSON.stringify(notificationsPath)}));
});
`;

  return new NextResponse(worker, {
    headers: {
      'Content-Type': 'application/javascript; charset=utf-8',
      'Cache-Control': 'no-store',
      'Service-Worker-Allowed': '/',
    },
  });
}
