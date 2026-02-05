import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';

interface UpdateSettings {
  channel: string;
  autoUpdate: boolean;
  checkInterval: number;
  lastCheck: string;
}

interface VersionInfo {
  version: string;
  channel: string;
  commit: string;
  buildTime: string;
  goVersion: string;
  os: string;
  arch: string;
}

interface ChannelInfo {
  id: string;
  name: string;
  description: string;
}

interface UpdateCheckResult {
  available: boolean;
  currentVersion: string;
  latestVersion: string;
  release?: {
    version: string;
    channel: string;
    tag: string;
    name: string;
    body: string;
    publishedAt: string;
    htmlUrl: string;
  };
  error?: string;
  checkedAt: string;
}

@Component({
  selector: 'app-updates-settings',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="updates-settings">
      <div class="current-version">
        <div class="version-badge">
          <span class="version-number">{{ versionInfo?.version || 'Unknown' }}</span>
          <span class="channel-badge" [class]="'channel-' + (versionInfo?.channel || 'dev')">
            {{ versionInfo?.channel || 'dev' }}
          </span>
        </div>
        <p class="build-info" *ngIf="versionInfo">
          Built {{ versionInfo.buildTime | date:'medium' }} ({{ versionInfo.commit?.substring(0, 7) }})
        </p>
      </div>

      <div class="update-check" *ngIf="checkResult">
        <div class="update-available" *ngIf="checkResult.available">
          <div class="update-icon">!</div>
          <div class="update-info">
            <h4>Update Available</h4>
            <p>Version {{ checkResult.latestVersion }} is available</p>
            <a *ngIf="checkResult.release?.htmlUrl"
               [href]="checkResult.release.htmlUrl"
               target="_blank"
               class="release-link">
              View Release Notes
            </a>
          </div>
          <button class="btn btn--primary" (click)="installUpdate()" [disabled]="isInstalling">
            {{ isInstalling ? 'Installing...' : 'Install Update' }}
          </button>
        </div>

        <div class="up-to-date" *ngIf="!checkResult.available && !checkResult.error">
          <div class="check-icon">OK</div>
          <div class="check-info">
            <h4>Up to Date</h4>
            <p>You're running the latest version</p>
            <span class="last-check" *ngIf="checkResult.checkedAt">
              Last checked: {{ checkResult.checkedAt | date:'short' }}
            </span>
          </div>
        </div>

        <div class="check-error" *ngIf="checkResult.error">
          <div class="error-icon">X</div>
          <div class="error-info">
            <h4>Check Failed</h4>
            <p>{{ checkResult.error }}</p>
          </div>
        </div>
      </div>

      <div class="check-button-row">
        <button class="btn btn--secondary" (click)="checkForUpdates()" [disabled]="isChecking">
          {{ isChecking ? 'Checking...' : 'Check for Updates' }}
        </button>
      </div>

      <div class="settings-section">
        <h3>Update Channel</h3>
        <p class="section-description">Choose which release channel to follow for updates.</p>

        <div class="channel-options">
          <label class="channel-option" *ngFor="let channel of channels"
                 [class.selected]="settings.channel === channel.id">
            <input type="radio"
                   [name]="'channel'"
                   [value]="channel.id"
                   [(ngModel)]="settings.channel"
                   (change)="onSettingsChange()">
            <div class="channel-content">
              <span class="channel-name">{{ channel.name }}</span>
              <span class="channel-desc">{{ channel.description }}</span>
            </div>
          </label>
        </div>
      </div>

      <div class="settings-section">
        <h3>Automatic Updates</h3>

        <div class="form-group">
          <label class="checkbox-label">
            <input type="checkbox"
                   [(ngModel)]="settings.autoUpdate"
                   (change)="onSettingsChange()">
            <span>Automatically install updates</span>
          </label>
          <p class="setting-hint">When enabled, updates will be installed automatically on app restart.</p>
        </div>

        <div class="form-group">
          <label class="form-label">Check Interval</label>
          <select class="form-select"
                  [(ngModel)]="settings.checkInterval"
                  (change)="onSettingsChange()">
            <option [value]="0">Disabled</option>
            <option [value]="1">Every hour</option>
            <option [value]="6">Every 6 hours</option>
            <option [value]="12">Every 12 hours</option>
            <option [value]="24">Daily</option>
            <option [value]="168">Weekly</option>
          </select>
        </div>
      </div>

      <div class="save-status" *ngIf="saveMessage">
        <span [class.error]="saveError">{{ saveMessage }}</span>
      </div>
    </div>
  `,
  styles: [`
    .updates-settings {
      padding: var(--spacing-md);
    }

    .current-version {
      background: var(--bg-tertiary);
      border-radius: var(--radius-lg);
      padding: var(--spacing-lg);
      margin-bottom: var(--spacing-lg);
      text-align: center;
    }

    .version-badge {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: var(--spacing-sm);
      margin-bottom: var(--spacing-xs);
    }

    .version-number {
      font-size: 24px;
      font-weight: 600;
    }

    .channel-badge {
      padding: 2px 8px;
      border-radius: 999px;
      font-size: 11px;
      font-weight: 600;
      text-transform: uppercase;
    }

    .channel-stable { background: var(--accent-success); color: white; }
    .channel-beta { background: var(--accent-warning); color: black; }
    .channel-nightly { background: var(--accent-purple, #8b5cf6); color: white; }
    .channel-dev { background: var(--text-muted); color: var(--bg-primary); }

    .build-info {
      color: var(--text-muted);
      font-size: 12px;
      margin: 0;
    }

    .update-check {
      margin-bottom: var(--spacing-lg);
    }

    .update-available, .up-to-date, .check-error {
      display: flex;
      align-items: center;
      gap: var(--spacing-md);
      padding: var(--spacing-md);
      border-radius: var(--radius-md);
    }

    .update-available {
      background: var(--accent-warning-bg, rgba(245, 158, 11, 0.1));
      border: 1px solid var(--accent-warning);
    }

    .up-to-date {
      background: var(--accent-success-bg, rgba(34, 197, 94, 0.1));
      border: 1px solid var(--accent-success);
    }

    .check-error {
      background: var(--accent-danger-bg, rgba(239, 68, 68, 0.1));
      border: 1px solid var(--accent-danger);
    }

    .update-icon, .check-icon, .error-icon {
      width: 40px;
      height: 40px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      font-weight: bold;
      flex-shrink: 0;
    }

    .update-icon { background: var(--accent-warning); color: black; }
    .check-icon { background: var(--accent-success); color: white; }
    .error-icon { background: var(--accent-danger); color: white; }

    .update-info, .check-info, .error-info {
      flex: 1;
    }

    .update-info h4, .check-info h4, .error-info h4 {
      margin: 0 0 var(--spacing-xs) 0;
      font-size: 14px;
    }

    .update-info p, .check-info p, .error-info p {
      margin: 0;
      font-size: 13px;
      color: var(--text-muted);
    }

    .release-link {
      color: var(--accent-primary);
      font-size: 12px;
    }

    .last-check {
      font-size: 11px;
      color: var(--text-muted);
    }

    .check-button-row {
      margin-bottom: var(--spacing-lg);
    }

    .settings-section {
      background: var(--bg-primary);
      border: 1px solid var(--border-color);
      border-radius: var(--radius-lg);
      padding: var(--spacing-lg);
      margin-bottom: var(--spacing-lg);
    }

    .settings-section h3 {
      font-size: 14px;
      margin: 0 0 var(--spacing-xs) 0;
    }

    .section-description {
      color: var(--text-muted);
      font-size: 12px;
      margin-bottom: var(--spacing-md);
    }

    .channel-options {
      display: flex;
      flex-direction: column;
      gap: var(--spacing-sm);
    }

    .channel-option {
      display: flex;
      align-items: flex-start;
      gap: var(--spacing-sm);
      padding: var(--spacing-md);
      border: 1px solid var(--border-color);
      border-radius: var(--radius-md);
      cursor: pointer;
      transition: all 0.15s ease;
    }

    .channel-option:hover {
      border-color: var(--accent-primary);
    }

    .channel-option.selected {
      border-color: var(--accent-primary);
      background: var(--accent-primary-bg, rgba(59, 130, 246, 0.1));
    }

    .channel-option input[type="radio"] {
      margin-top: 2px;
    }

    .channel-content {
      display: flex;
      flex-direction: column;
      gap: 2px;
    }

    .channel-name {
      font-weight: 500;
      font-size: 14px;
    }

    .channel-desc {
      font-size: 12px;
      color: var(--text-muted);
    }

    .form-group {
      margin-bottom: var(--spacing-md);
    }

    .form-group:last-child {
      margin-bottom: 0;
    }

    .checkbox-label {
      display: flex;
      align-items: center;
      gap: var(--spacing-sm);
      cursor: pointer;
    }

    .setting-hint {
      color: var(--text-muted);
      font-size: 12px;
      margin: var(--spacing-xs) 0 0 24px;
    }

    .form-label {
      display: block;
      font-size: 13px;
      margin-bottom: var(--spacing-xs);
    }

    .form-select {
      width: 100%;
      padding: var(--spacing-sm);
      border: 1px solid var(--border-color);
      border-radius: var(--radius-md);
      background: var(--bg-secondary);
      color: var(--text-primary);
      font-size: 14px;
    }

    .save-status {
      text-align: center;
      font-size: 13px;
      color: var(--accent-success);
    }

    .save-status .error {
      color: var(--accent-danger);
    }

    .btn {
      padding: var(--spacing-sm) var(--spacing-md);
      border: none;
      border-radius: var(--radius-md);
      font-size: 14px;
      cursor: pointer;
      transition: all 0.15s ease;
    }

    .btn:disabled {
      opacity: 0.6;
      cursor: not-allowed;
    }

    .btn--primary {
      background: var(--accent-primary);
      color: white;
    }

    .btn--primary:hover:not(:disabled) {
      background: var(--accent-primary-hover, #2563eb);
    }

    .btn--secondary {
      background: var(--bg-tertiary);
      color: var(--text-primary);
      border: 1px solid var(--border-color);
    }

    .btn--secondary:hover:not(:disabled) {
      background: var(--bg-secondary);
    }
  `]
})
export class UpdatesComponent implements OnInit, OnDestroy {
  settings: UpdateSettings = {
    channel: 'stable',
    autoUpdate: false,
    checkInterval: 6,
    lastCheck: ''
  };

  versionInfo: VersionInfo | null = null;
  checkResult: UpdateCheckResult | null = null;

  channels: ChannelInfo[] = [
    { id: 'stable', name: 'Stable', description: 'Production releases - most stable, recommended for most users' },
    { id: 'beta', name: 'Beta', description: 'Pre-release builds - new features being tested before stable release' },
    { id: 'nightly', name: 'Nightly', description: 'Latest development builds - bleeding edge, may be unstable' }
  ];

  isChecking = false;
  isInstalling = false;
  saveMessage = '';
  saveError = false;

  private saveTimeout: ReturnType<typeof setTimeout> | null = null;

  ngOnInit() {
    this.loadSettings();
    this.loadVersionInfo();
  }

  ngOnDestroy() {
    if (this.saveTimeout) {
      clearTimeout(this.saveTimeout);
    }
  }

  async loadSettings() {
    try {
      const wails = (window as any).go?.main;
      if (wails?.UpdateService?.GetSettings) {
        this.settings = await wails.UpdateService.GetSettings();
      } else if (wails?.ConfigService?.GetUpdateSettings) {
        this.settings = await wails.ConfigService.GetUpdateSettings();
      }
    } catch (err) {
      console.error('Failed to load update settings:', err);
    }
  }

  async loadVersionInfo() {
    try {
      const wails = (window as any).go?.main;
      if (wails?.VersionService?.GetVersionInfo) {
        this.versionInfo = await wails.VersionService.GetVersionInfo();
      } else if (wails?.UpdateService?.GetVersionInfo) {
        this.versionInfo = await wails.UpdateService.GetVersionInfo();
      }
    } catch (err) {
      console.error('Failed to load version info:', err);
    }
  }

  async checkForUpdates() {
    this.isChecking = true;
    this.checkResult = null;

    try {
      const wails = (window as any).go?.main;
      if (wails?.UpdateService?.CheckForUpdate) {
        this.checkResult = await wails.UpdateService.CheckForUpdate();
      }
    } catch (err) {
      console.error('Failed to check for updates:', err);
      this.checkResult = {
        available: false,
        currentVersion: this.versionInfo?.version || 'unknown',
        latestVersion: '',
        error: 'Failed to check for updates',
        checkedAt: new Date().toISOString()
      };
    } finally {
      this.isChecking = false;
    }
  }

  async installUpdate() {
    if (!this.checkResult?.available || !this.checkResult.release) {
      return;
    }

    this.isInstalling = true;

    try {
      const wails = (window as any).go?.main;
      if (wails?.UpdateService?.InstallUpdate) {
        await wails.UpdateService.InstallUpdate();
      }
    } catch (err) {
      console.error('Failed to install update:', err);
      alert('Failed to install update. Please try again or download manually.');
    } finally {
      this.isInstalling = false;
    }
  }

  async onSettingsChange() {
    // Debounce save
    if (this.saveTimeout) {
      clearTimeout(this.saveTimeout);
    }

    this.saveTimeout = setTimeout(() => this.saveSettings(), 500);
  }

  async saveSettings() {
    try {
      const wails = (window as any).go?.main;
      if (wails?.UpdateService?.SetSettings) {
        await wails.UpdateService.SetSettings(this.settings);
      } else if (wails?.ConfigService?.SetUpdateSettings) {
        await wails.ConfigService.SetUpdateSettings(this.settings);
      }
      this.saveMessage = 'Settings saved';
      this.saveError = false;
    } catch (err) {
      console.error('Failed to save update settings:', err);
      this.saveMessage = 'Failed to save settings';
      this.saveError = true;
    }

    // Clear message after 2 seconds
    setTimeout(() => {
      this.saveMessage = '';
    }, 2000);
  }
}
