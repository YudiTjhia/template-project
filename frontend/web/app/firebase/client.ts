'use client';

import { getApp, getApps, initializeApp } from '@firebase/app';
import { getMessaging, isSupported, type Messaging } from '@firebase/messaging';
import { firebaseWebConfig, isFirebaseWebConfigured } from './config';

export function getFirebaseApp() {
  if (!isFirebaseWebConfigured()) return null;
  const appName = 'template-project-web';
  return getApps().some((app) => app.name === appName)
    ? getApp(appName)
    : initializeApp(firebaseWebConfig, appName);
}

export async function getFirebaseMessaging(): Promise<Messaging | null> {
  if (typeof window === 'undefined' || !(await isSupported())) return null;
  const app = getFirebaseApp();
  return app ? getMessaging(app) : null;
}
