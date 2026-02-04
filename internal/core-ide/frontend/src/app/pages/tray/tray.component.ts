import { Component, signal, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Events } from '@wailsio/runtime';

@Component({
  selector: 'app-tray',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="tray-container">
      <!-- Header -->
      <div class="tray-header">
        <div class="tray-logo">
          <svg class="logo-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
              d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"/>
          </svg>
          <span>Core IDE</span>
        </div>
        <div class="tray-controls">
          <button class="control-btn" (click)="openIDE()" title="Open IDE">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"/>
            </svg>
          </button>
        </div>
      </div>

      <!-- Quick Stats -->
      <div class="tray-stats">
        <div class="stat-row">
          <span class="stat-label">Status</span>
          <span class="stat-value" [class.active]="isActive()">
            {{ isActive() ? 'Running' : 'Idle' }}
          </span>
        </div>
        <div class="stat-row">
          <span class="stat-label">Projects</span>
          <span class="stat-value">{{ projectCount() }}</span>
        </div>
        <div class="stat-row">
          <span class="stat-label">Active Tasks</span>
          <span class="stat-value">{{ taskCount() }}</span>
        </div>
        <div class="stat-row">
          <span class="stat-label">Time</span>
          <span class="stat-value mono">{{ currentTime() }}</span>
        </div>
      </div>

      <!-- Quick Actions -->
      <div class="actions-section">
        <div class="section-header">Quick Actions</div>
        <div class="actions-list">
          <button class="action-btn" (click)="emitAction('new-project')">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
            </svg>
            <span>New Project</span>
          </button>
          <button class="action-btn" (click)="emitAction('open-project')">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M5 19a2 2 0 01-2-2V7a2 2 0 012-2h4l2 2h4a2 2 0 012 2v1M5 19h14a2 2 0 002-2v-5a2 2 0 00-2-2H9a2 2 0 00-2 2v5a2 2 0 01-2 2z"/>
            </svg>
            <span>Open Project</span>
          </button>
          <button class="action-btn" (click)="emitAction('run-dev')">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"/>
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
            <span>Run Dev Server</span>
          </button>
          <button class="action-btn danger" (click)="emitAction('stop-all')">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M9 10a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z"/>
            </svg>
            <span>Stop All</span>
          </button>
        </div>
      </div>

      <!-- Recent Projects -->
      <div class="projects-section">
        <div class="section-header">Recent Projects</div>
        <div class="projects-list">
          @for (project of recentProjects(); track project.name) {
            <button class="project-item" (click)="openProject(project.path)">
              <div class="project-icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"/>
                </svg>
              </div>
              <div class="project-info">
                <span class="project-name">{{ project.name }}</span>
                <span class="project-path">{{ project.path }}</span>
              </div>
            </button>
          } @empty {
            <div class="no-projects">No recent projects</div>
          }
        </div>
      </div>

      <!-- Footer -->
      <div class="tray-footer">
        <div class="connection-status" [class.connected]="isActive()">
          <div class="status-dot"></div>
          <span>{{ isActive() ? 'Services Running' : 'Ready' }}</span>
        </div>
        <button class="footer-btn" (click)="openIDE()">Open IDE</button>
      </div>
    </div>
  `,
  styles: [`
    :host {
      display: block;
      width: 100%;
      height: 100%;
      overflow: hidden;
    }

    .tray-container {
      display: flex;
      flex-direction: column;
      height: 100%;
      background: #1a1b26;
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
      color: #a9b1d6;
    }

    .tray-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 0.75rem 1rem;
      background: #16161e;
      border-bottom: 1px solid #24283b;
    }

    .tray-logo {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      font-size: 0.9375rem;
      font-weight: 600;
      color: #c0caf5;
    }

    .logo-icon {
      width: 20px;
      height: 20px;
      color: #7aa2f7;
    }

    .tray-controls {
      display: flex;
      gap: 0.5rem;
    }

    .control-btn {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 32px;
      height: 32px;
      background: transparent;
      border: 1px solid #24283b;
      border-radius: 6px;
      color: #7aa2f7;
      cursor: pointer;
      transition: all 0.15s ease;
    }

    .control-btn:hover {
      background: #24283b;
      border-color: #7aa2f7;
    }

    .control-btn svg {
      width: 16px;
      height: 16px;
    }

    .tray-stats {
      display: flex;
      flex-direction: column;
      gap: 0.5rem;
      padding: 0.875rem 1rem;
      background: #16161e;
      border-bottom: 1px solid #24283b;
    }

    .stat-row {
      display: flex;
      align-items: center;
      justify-content: space-between;
    }

    .stat-label {
      font-size: 0.8125rem;
      color: #565f89;
    }

    .stat-value {
      font-size: 0.875rem;
      font-weight: 600;
      color: #c0caf5;
    }

    .stat-value.active {
      color: #9ece6a;
    }

    .stat-value.mono {
      font-family: 'JetBrains Mono', 'Fira Code', monospace;
      font-size: 0.8125rem;
    }

    .actions-section, .projects-section {
      display: flex;
      flex-direction: column;
    }

    .section-header {
      padding: 0.625rem 1rem;
      font-size: 0.6875rem;
      font-weight: 600;
      color: #565f89;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      background: #1a1b26;
      border-bottom: 1px solid #24283b;
    }

    .actions-list {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 0.5rem;
      padding: 0.75rem;
    }

    .action-btn {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      padding: 0.625rem 0.75rem;
      background: #16161e;
      border: 1px solid #24283b;
      border-radius: 6px;
      color: #a9b1d6;
      font-size: 0.8125rem;
      cursor: pointer;
      transition: all 0.15s ease;
    }

    .action-btn:hover {
      background: #24283b;
      border-color: #414868;
    }

    .action-btn.danger {
      color: #f7768e;
    }

    .action-btn.danger:hover {
      border-color: #f7768e;
      background: rgba(247, 118, 142, 0.1);
    }

    .action-btn svg {
      width: 16px;
      height: 16px;
      flex-shrink: 0;
    }

    .projects-section {
      flex: 1;
      min-height: 0;
    }

    .projects-list {
      flex: 1;
      overflow-y: auto;
      padding: 0.5rem;
    }

    .project-item {
      display: flex;
      align-items: center;
      gap: 0.75rem;
      width: 100%;
      padding: 0.625rem 0.75rem;
      background: transparent;
      border: none;
      border-radius: 6px;
      color: #a9b1d6;
      text-align: left;
      cursor: pointer;
      transition: background 0.15s ease;
    }

    .project-item:hover {
      background: #24283b;
    }

    .project-icon {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 32px;
      height: 32px;
      background: #24283b;
      border-radius: 6px;
      color: #7aa2f7;
    }

    .project-icon svg {
      width: 16px;
      height: 16px;
    }

    .project-info {
      display: flex;
      flex-direction: column;
      min-width: 0;
    }

    .project-name {
      font-size: 0.8125rem;
      font-weight: 500;
      color: #c0caf5;
    }

    .project-path {
      font-size: 0.6875rem;
      color: #565f89;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .no-projects {
      display: flex;
      align-items: center;
      justify-content: center;
      padding: 2rem;
      color: #565f89;
      font-size: 0.8125rem;
    }

    .tray-footer {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 0.625rem 1rem;
      background: #16161e;
      border-top: 1px solid #24283b;
    }

    .connection-status {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      font-size: 0.75rem;
      color: #565f89;
    }

    .status-dot {
      width: 6px;
      height: 6px;
      border-radius: 50%;
      background: #565f89;
    }

    .connection-status.connected .status-dot {
      background: #9ece6a;
      box-shadow: 0 0 4px #9ece6a;
    }

    .connection-status.connected {
      color: #9ece6a;
    }

    .footer-btn {
      padding: 0.375rem 0.75rem;
      background: #7aa2f7;
      border: none;
      border-radius: 4px;
      color: #1a1b26;
      font-size: 0.75rem;
      font-weight: 600;
      cursor: pointer;
      transition: background 0.15s ease;
    }

    .footer-btn:hover {
      background: #89b4fa;
    }
  `]
})
export class TrayComponent implements OnInit, OnDestroy {
  currentTime = signal('');
  isActive = signal(false);
  projectCount = signal(3);
  taskCount = signal(0);
  recentProjects = signal([
    { name: 'core', path: '~/Code/host-uk/core' },
    { name: 'core-gui', path: '~/Code/host-uk/core-gui' },
    { name: 'core-php', path: '~/Code/host-uk/core-php' },
  ]);
  private timeEventCleanup?: () => void;

  ngOnInit() {
    this.timeEventCleanup = Events.On('time', (time: { data: string }) => {
      this.currentTime.set(time.data);
    });
  }

  ngOnDestroy() {
    this.timeEventCleanup?.();
  }

  openIDE() {
    Events.Emit('action', 'open-ide');
  }

  emitAction(action: string) {
    Events.Emit('action', action);
  }

  openProject(path: string) {
    Events.Emit('open-project', path);
  }
}
