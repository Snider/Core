import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';

interface Config {
  watchedRepos: string[];
  labels: string[];
  fetchIntervalMinutes: number;
  notificationsEnabled: boolean;
  notificationSound: boolean;
  workspaceDir: string;
  theme: string;
  autoSeedContext: boolean;
  workHours?: {
    enabled: boolean;
    startHour: number;
    endHour: number;
    days: number[];
    timezone: string;
  };
}

@Component({
  selector: 'app-settings',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="settings">
      <header class="settings-header">
        <h1>Settings</h1>
        <button class="btn btn--primary" (click)="saveSettings()">Save</button>
      </header>

      <div class="settings-content">
        <section class="settings-section">
          <h2>Repositories</h2>
          <p class="section-description">Add GitHub repositories to watch for issues.</p>

          <div class="repo-list">
            <div class="repo-item" *ngFor="let repo of config.watchedRepos; let i = index">
              <span>{{ repo }}</span>
              <button class="btn btn--danger btn--sm" (click)="removeRepo(i)">Remove</button>
            </div>
          </div>

          <div class="add-repo">
            <input type="text" class="form-input" [(ngModel)]="newRepo"
                   placeholder="owner/repo (e.g., facebook/react)">
            <button class="btn btn--secondary" (click)="addRepo()" [disabled]="!newRepo">Add</button>
          </div>
        </section>

        <section class="settings-section">
          <h2>Issue Labels</h2>
          <p class="section-description">Filter issues by these labels.</p>

          <div class="label-list">
            <span class="label-chip" *ngFor="let label of config.labels; let i = index">
              {{ label }}
              <button class="label-remove" (click)="removeLabel(i)">x</button>
            </span>
          </div>

          <div class="add-label">
            <input type="text" class="form-input" [(ngModel)]="newLabel"
                   placeholder="Add label (e.g., good first issue)">
            <button class="btn btn--secondary" (click)="addLabel()" [disabled]="!newLabel">Add</button>
          </div>
        </section>

        <section class="settings-section">
          <h2>Fetch Settings</h2>

          <div class="form-group">
            <label class="form-label">Fetch Interval (minutes)</label>
            <input type="number" class="form-input" [(ngModel)]="config.fetchIntervalMinutes" min="5" max="120">
          </div>

          <div class="form-group">
            <label class="checkbox-label">
              <input type="checkbox" [(ngModel)]="config.autoSeedContext">
              <span>Auto-prepare AI context for issues</span>
            </label>
          </div>
        </section>

        <section class="settings-section">
          <h2>Work Hours</h2>
          <p class="section-description">Only fetch issues during these hours.</p>

          <div class="form-group">
            <label class="checkbox-label">
              <input type="checkbox" [(ngModel)]="config.workHours!.enabled">
              <span>Enable work hours</span>
            </label>
          </div>

          <div class="work-hours-config" *ngIf="config.workHours?.enabled">
            <div class="form-group">
              <label class="form-label">Start Hour</label>
              <select class="form-select" [(ngModel)]="config.workHours!.startHour">
                <option *ngFor="let h of hours" [value]="h">{{ h }}:00</option>
              </select>
            </div>

            <div class="form-group">
              <label class="form-label">End Hour</label>
              <select class="form-select" [(ngModel)]="config.workHours!.endHour">
                <option *ngFor="let h of hours" [value]="h">{{ h }}:00</option>
              </select>
            </div>

            <div class="form-group">
              <label class="form-label">Days</label>
              <div class="day-checkboxes">
                <label class="checkbox-label" *ngFor="let day of days; let i = index">
                  <input type="checkbox" [checked]="isDaySelected(i)" (change)="toggleDay(i)">
                  <span>{{ day }}</span>
                </label>
              </div>
            </div>
          </div>
        </section>

        <section class="settings-section">
          <h2>Notifications</h2>

          <div class="form-group">
            <label class="checkbox-label">
              <input type="checkbox" [(ngModel)]="config.notificationsEnabled">
              <span>Enable desktop notifications</span>
            </label>
          </div>

          <div class="form-group">
            <label class="checkbox-label">
              <input type="checkbox" [(ngModel)]="config.notificationSound">
              <span>Play notification sounds</span>
            </label>
          </div>
        </section>

        <section class="settings-section">
          <h2>Appearance</h2>

          <div class="form-group">
            <label class="form-label">Theme</label>
            <select class="form-select" [(ngModel)]="config.theme">
              <option value="dark">Dark</option>
              <option value="light">Light</option>
              <option value="system">System</option>
            </select>
          </div>
        </section>

        <section class="settings-section">
          <h2>Storage</h2>

          <div class="form-group">
            <label class="form-label">Workspace Directory</label>
            <input type="text" class="form-input" [(ngModel)]="config.workspaceDir"
                   placeholder="Leave empty for default">
          </div>
        </section>
      </div>
    </div>
  `,
  styles: [`
    .settings {
      display: flex;
      flex-direction: column;
      height: 100%;
      background-color: var(--bg-secondary);
    }

    .settings-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: var(--spacing-md) var(--spacing-lg);
      background-color: var(--bg-primary);
      border-bottom: 1px solid var(--border-color);
    }

    .settings-header h1 {
      font-size: 18px;
      margin: 0;
    }

    .settings-content {
      flex: 1;
      overflow-y: auto;
      padding: var(--spacing-lg);
    }

    .settings-section {
      background-color: var(--bg-primary);
      border: 1px solid var(--border-color);
      border-radius: var(--radius-lg);
      padding: var(--spacing-lg);
      margin-bottom: var(--spacing-lg);
    }

    .settings-section h2 {
      font-size: 16px;
      margin-bottom: var(--spacing-xs);
    }

    .section-description {
      color: var(--text-muted);
      font-size: 13px;
      margin-bottom: var(--spacing-md);
    }

    .repo-list, .label-list {
      margin-bottom: var(--spacing-md);
    }

    .repo-item {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: var(--spacing-sm);
      background-color: var(--bg-secondary);
      border-radius: var(--radius-md);
      margin-bottom: var(--spacing-xs);
    }

    .add-repo, .add-label {
      display: flex;
      gap: var(--spacing-sm);
    }

    .add-repo .form-input, .add-label .form-input {
      flex: 1;
    }

    .label-list {
      display: flex;
      flex-wrap: wrap;
      gap: var(--spacing-xs);
    }

    .label-chip {
      display: inline-flex;
      align-items: center;
      gap: var(--spacing-xs);
      padding: var(--spacing-xs) var(--spacing-sm);
      background-color: var(--bg-tertiary);
      border-radius: 999px;
      font-size: 13px;
    }

    .label-remove {
      background: none;
      border: none;
      color: var(--text-muted);
      cursor: pointer;
      padding: 0;
      font-size: 14px;
      line-height: 1;
    }

    .label-remove:hover {
      color: var(--accent-danger);
    }

    .checkbox-label {
      display: flex;
      align-items: center;
      gap: var(--spacing-sm);
      cursor: pointer;
    }

    .checkbox-label input[type="checkbox"] {
      width: 16px;
      height: 16px;
    }

    .work-hours-config {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: var(--spacing-md);
      margin-top: var(--spacing-md);
    }

    .day-checkboxes {
      display: flex;
      flex-wrap: wrap;
      gap: var(--spacing-sm);
    }

    .day-checkboxes .checkbox-label {
      width: auto;
    }

    .btn--sm {
      padding: var(--spacing-xs) var(--spacing-sm);
      font-size: 12px;
    }
  `]
})
export class SettingsComponent implements OnInit {
  config: Config = {
    watchedRepos: [],
    labels: ['good first issue', 'help wanted'],
    fetchIntervalMinutes: 15,
    notificationsEnabled: true,
    notificationSound: true,
    workspaceDir: '',
    theme: 'dark',
    autoSeedContext: true,
    workHours: {
      enabled: false,
      startHour: 9,
      endHour: 17,
      days: [1, 2, 3, 4, 5],
      timezone: ''
    }
  };

  newRepo = '';
  newLabel = '';
  hours = Array.from({ length: 24 }, (_, i) => i);
  days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

  ngOnInit() {
    this.loadConfig();
  }

  async loadConfig() {
    try {
      if ((window as any).go?.main?.ConfigService?.GetConfig) {
        this.config = await (window as any).go.main.ConfigService.GetConfig();
        if (!this.config.workHours) {
          this.config.workHours = {
            enabled: false,
            startHour: 9,
            endHour: 17,
            days: [1, 2, 3, 4, 5],
            timezone: ''
          };
        }
      }
    } catch (err) {
      console.error('Failed to load config:', err);
    }
  }

  async saveSettings() {
    try {
      if ((window as any).go?.main?.ConfigService?.SetConfig) {
        await (window as any).go.main.ConfigService.SetConfig(this.config);
        alert('Settings saved!');
      }
    } catch (err) {
      console.error('Failed to save config:', err);
      alert('Failed to save settings.');
    }
  }

  addRepo() {
    if (this.newRepo && !this.config.watchedRepos.includes(this.newRepo)) {
      this.config.watchedRepos.push(this.newRepo);
      this.newRepo = '';
    }
  }

  removeRepo(index: number) {
    this.config.watchedRepos.splice(index, 1);
  }

  addLabel() {
    if (this.newLabel && !this.config.labels.includes(this.newLabel)) {
      this.config.labels.push(this.newLabel);
      this.newLabel = '';
    }
  }

  removeLabel(index: number) {
    this.config.labels.splice(index, 1);
  }

  isDaySelected(day: number): boolean {
    return this.config.workHours?.days.includes(day) || false;
  }

  toggleDay(day: number) {
    if (!this.config.workHours) return;

    const index = this.config.workHours.days.indexOf(day);
    if (index === -1) {
      this.config.workHours.days.push(day);
    } else {
      this.config.workHours.days.splice(index, 1);
    }
  }
}
