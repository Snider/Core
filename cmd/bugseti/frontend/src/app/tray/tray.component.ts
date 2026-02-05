import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';

interface TrayStatus {
  running: boolean;
  currentIssue: string;
  queueSize: number;
  issuesFixed: number;
  prsMerged: number;
}

@Component({
  selector: 'app-tray',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="tray-panel">
      <header class="tray-header">
        <div class="logo">
          <span class="logo-icon">B</span>
          <span class="logo-text">BugSETI</span>
        </div>
        <span class="badge" [class.badge--success]="status.running" [class.badge--warning]="!status.running">
          {{ status.running ? 'Running' : 'Paused' }}
        </span>
      </header>

      <section class="stats-grid">
        <div class="stat-card">
          <span class="stat-value">{{ status.queueSize }}</span>
          <span class="stat-label">In Queue</span>
        </div>
        <div class="stat-card">
          <span class="stat-value">{{ status.issuesFixed }}</span>
          <span class="stat-label">Fixed</span>
        </div>
        <div class="stat-card">
          <span class="stat-value">{{ status.prsMerged }}</span>
          <span class="stat-label">Merged</span>
        </div>
      </section>

      <section class="current-issue" *ngIf="status.currentIssue">
        <h3>Current Issue</h3>
        <div class="issue-card">
          <p class="issue-title">{{ status.currentIssue }}</p>
          <div class="issue-actions">
            <button class="btn btn--primary btn--sm" (click)="openWorkbench()">
              Open Workbench
            </button>
            <button class="btn btn--secondary btn--sm" (click)="skipIssue()">
              Skip
            </button>
          </div>
        </div>
      </section>

      <section class="current-issue" *ngIf="!status.currentIssue">
        <div class="empty-state">
          <span class="empty-icon">[ ]</span>
          <p>No issue in progress</p>
          <button class="btn btn--primary btn--sm" (click)="nextIssue()" [disabled]="status.queueSize === 0">
            Get Next Issue
          </button>
        </div>
      </section>

      <footer class="tray-footer">
        <button class="btn btn--secondary btn--sm" (click)="toggleRunning()">
          {{ status.running ? 'Pause' : 'Start' }}
        </button>
        <button class="btn btn--secondary btn--sm" (click)="openSettings()">
          Settings
        </button>
      </footer>
    </div>
  `,
  styles: [`
    .tray-panel {
      display: flex;
      flex-direction: column;
      height: 100%;
      padding: var(--spacing-md);
      background-color: var(--bg-primary);
    }

    .tray-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: var(--spacing-md);
    }

    .logo {
      display: flex;
      align-items: center;
      gap: var(--spacing-sm);
    }

    .logo-icon {
      width: 28px;
      height: 28px;
      display: flex;
      align-items: center;
      justify-content: center;
      background: linear-gradient(135deg, var(--accent-primary), var(--accent-success));
      border-radius: var(--radius-md);
      font-weight: bold;
      color: white;
    }

    .logo-text {
      font-weight: 600;
      font-size: 16px;
    }

    .stats-grid {
      display: grid;
      grid-template-columns: repeat(3, 1fr);
      gap: var(--spacing-sm);
      margin-bottom: var(--spacing-md);
    }

    .stat-card {
      display: flex;
      flex-direction: column;
      align-items: center;
      padding: var(--spacing-sm);
      background-color: var(--bg-secondary);
      border-radius: var(--radius-md);
    }

    .stat-value {
      font-size: 24px;
      font-weight: bold;
      color: var(--accent-primary);
    }

    .stat-label {
      font-size: 11px;
      color: var(--text-muted);
      text-transform: uppercase;
    }

    .current-issue {
      flex: 1;
      margin-bottom: var(--spacing-md);
    }

    .current-issue h3 {
      font-size: 12px;
      color: var(--text-muted);
      text-transform: uppercase;
      margin-bottom: var(--spacing-sm);
    }

    .issue-card {
      background-color: var(--bg-secondary);
      border: 1px solid var(--border-color);
      border-radius: var(--radius-md);
      padding: var(--spacing-md);
    }

    .issue-title {
      font-size: 13px;
      line-height: 1.4;
      margin-bottom: var(--spacing-sm);
    }

    .issue-actions {
      display: flex;
      gap: var(--spacing-sm);
    }

    .empty-state {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      padding: var(--spacing-xl);
      text-align: center;
    }

    .empty-icon {
      font-size: 32px;
      color: var(--text-muted);
      margin-bottom: var(--spacing-sm);
    }

    .empty-state p {
      color: var(--text-muted);
      margin-bottom: var(--spacing-md);
    }

    .tray-footer {
      display: flex;
      gap: var(--spacing-sm);
      justify-content: center;
    }

    .btn--sm {
      padding: var(--spacing-xs) var(--spacing-sm);
      font-size: 12px;
    }
  `]
})
export class TrayComponent implements OnInit, OnDestroy {
  status: TrayStatus = {
    running: false,
    currentIssue: '',
    queueSize: 0,
    issuesFixed: 0,
    prsMerged: 0
  };

  private refreshInterval?: ReturnType<typeof setInterval>;

  ngOnInit() {
    this.loadStatus();
    this.refreshInterval = setInterval(() => this.loadStatus(), 5000);
  }

  ngOnDestroy() {
    if (this.refreshInterval) {
      clearInterval(this.refreshInterval);
    }
  }

  async loadStatus() {
    try {
      // Call Wails binding when available
      if ((window as any).go?.main?.TrayService?.GetStatus) {
        this.status = await (window as any).go.main.TrayService.GetStatus();
      }
    } catch (err) {
      console.error('Failed to load status:', err);
    }
  }

  async toggleRunning() {
    try {
      if (this.status.running) {
        if ((window as any).go?.main?.TrayService?.PauseFetching) {
          await (window as any).go.main.TrayService.PauseFetching();
        }
      } else {
        if ((window as any).go?.main?.TrayService?.StartFetching) {
          await (window as any).go.main.TrayService.StartFetching();
        }
      }
      this.loadStatus();
    } catch (err) {
      console.error('Failed to toggle running:', err);
    }
  }

  async nextIssue() {
    try {
      if ((window as any).go?.main?.TrayService?.NextIssue) {
        await (window as any).go.main.TrayService.NextIssue();
      }
      this.loadStatus();
    } catch (err) {
      console.error('Failed to get next issue:', err);
    }
  }

  async skipIssue() {
    try {
      if ((window as any).go?.main?.TrayService?.SkipIssue) {
        await (window as any).go.main.TrayService.SkipIssue();
      }
      this.loadStatus();
    } catch (err) {
      console.error('Failed to skip issue:', err);
    }
  }

  openWorkbench() {
    if ((window as any).wails?.Window) {
      (window as any).wails.Window.GetByName('workbench').then((w: any) => {
        w.Show();
        w.Focus();
      });
    }
  }

  openSettings() {
    if ((window as any).wails?.Window) {
      (window as any).wails.Window.GetByName('settings').then((w: any) => {
        w.Show();
        w.Focus();
      });
    }
  }
}
