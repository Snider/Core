import { Injectable } from '@angular/core';

@Injectable({ providedIn: 'root' })
export class NotificationsService {
  async requestPermission(): Promise<NotificationPermission> {
    if (!('Notification' in window)) return 'denied';
    if (Notification.permission === 'default') {
      try {
        return await Notification.requestPermission();
      } catch {
        return Notification.permission;
      }
    }
    return Notification.permission;
  }

  async show(title: string, options?: NotificationOptions): Promise<void> {
    if (!('Notification' in window)) return;
    const perm = await this.requestPermission();
    if (perm === 'granted') {
      new Notification(title, options);
    }
  }
}
