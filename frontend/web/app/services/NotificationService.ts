import { getToken, onMessage, type MessagePayload, type Unsubscribe } from '@firebase/messaging';
import api from './SubAppApi';
import loginService from './LoginService';
import { getFirebaseMessaging } from '@/app/firebase/client';
import { firebaseWebVapidKey, isFirebaseWebConfigured } from '@/app/firebase/config';

export type NotificationItem = {
  id: number;
  title: string;
  message: string;
  entityType?: string;
  entityId?: number;
  entityNumber?: string;
  action?: string;
  isRead: boolean;
  readAt?: string | null;
  createdAt?: string;
};

type NotificationListResponse = {
  data?: NotificationItem[];
};

type UnreadCountResponse = {
  unreadCount?: number;
};

export class NotificationService {
  async getNotifications(options: { search?: string; limit?: number } = {}): Promise<NotificationItem[]> {
    const response = await api.get<NotificationListResponse>(
      '/api/notifications',
      'Gagal mengambil daftar notifikasi.',
      {
        token: this.requireSessionToken(),
        query: { search: options.search, limit: options.limit ?? 50 },
      },
    );
    return Array.isArray(response?.data) ? response.data : [];
  }

  async getUnreadCount(): Promise<number> {
    const response = await api.get<UnreadCountResponse>(
      '/api/notifications',
      'Gagal mengambil jumlah notifikasi belum dibaca.',
      { token: this.requireSessionToken(), query: { count: 'unread' } },
    );
    return Number(response?.unreadCount ?? 0);
  }

  async markAsRead(id: number): Promise<void> {
    if (!Number.isInteger(id) || id <= 0) throw new Error('ID notifikasi tidak valid.');
    await api.postWithOptions(
      `/api/notifications/${id}/read`,
      {},
      'Gagal menandai notifikasi sebagai dibaca.',
      { token: this.requireSessionToken() },
    );
  }

  async requestPushToken(): Promise<string> {
    if (!isFirebaseWebConfigured()) {
      throw new Error('Konfigurasi Firebase Web belum lengkap.');
    }
    if (typeof window === 'undefined' || !('Notification' in window) || !('serviceWorker' in navigator)) {
      throw new Error('Browser ini tidak mendukung notifikasi push.');
    }
    if (!firebaseWebVapidKey) {
      throw new Error('NEXT_PUBLIC_FIREBASE_VAPID_KEY belum dikonfigurasi.');
    }

    const permission = await Notification.requestPermission();
    if (permission !== 'granted') throw new Error('Izin notifikasi tidak diberikan.');

    const messaging = await getFirebaseMessaging();
    if (!messaging) throw new Error('Firebase Messaging tidak didukung di browser ini.');
    const serviceWorkerRegistration = await navigator.serviceWorker.register('/firebase-messaging-sw.js');
    const token = await getToken(messaging, { vapidKey: firebaseWebVapidKey, serviceWorkerRegistration });
    if (!token) throw new Error('Token Firebase Messaging tidak berhasil dibuat.');
    return token;
  }

  async listenForForegroundMessages(
    callback: (payload: MessagePayload) => void,
  ): Promise<Unsubscribe> {
    const messaging = await getFirebaseMessaging();
    return messaging ? onMessage(messaging, callback) : () => undefined;
  }

  private requireSessionToken() {
    const token = loginService.getSessionToken();
    if (!token) throw new Error('Sesi login tidak ditemukan. Silakan login kembali.');
    return token;
  }
}

const notificationService = new NotificationService();
export default notificationService;
