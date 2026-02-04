import { Component, signal, OnInit, OnDestroy, PLATFORM_ID, Inject } from '@angular/core';
import { CommonModule, isPlatformBrowser } from '@angular/common';
import { SidebarComponent } from '../../components/sidebar/sidebar.component';

@Component({
  selector: 'app-ide',
  standalone: true,
  imports: [CommonModule, SidebarComponent],
  template: `
    <div class="ide-layout">
      <app-sidebar [currentRoute]="currentRoute()" (routeChange)="onRouteChange($event)"></app-sidebar>

      <div class="ide-main">
        <!-- Top Bar -->
        <div class="top-bar">
          <div class="breadcrumb">
            <span class="breadcrumb-item">Core IDE</span>
            <span class="breadcrumb-sep">/</span>
            <span class="breadcrumb-item active">{{ currentRoute() }}</span>
          </div>
          <div class="top-bar-actions">
            <span class="time">{{ currentTime() }}</span>
          </div>
        </div>

        <!-- Content Area -->
        <div class="ide-content">
          @switch (currentRoute()) {
            @case ('dashboard') {
              <div class="dashboard-view">
                <h1>Welcome to Core IDE</h1>
                <p class="subtitle">Your development environment is ready.</p>

                <div class="stats-grid">
                  <div class="stat-card">
                    <div class="stat-icon projects">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                          d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"/>
                      </svg>
                    </div>
                    <div class="stat-info">
                      <span class="stat-value">{{ projectCount() }}</span>
                      <span class="stat-label">Projects</span>
                    </div>
                  </div>

                  <div class="stat-card">
                    <div class="stat-icon tasks">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                          d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"/>
                      </svg>
                    </div>
                    <div class="stat-info">
                      <span class="stat-value">{{ taskCount() }}</span>
                      <span class="stat-label">Tasks</span>
                    </div>
                  </div>

                  <div class="stat-card">
                    <div class="stat-icon git">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                          d="M16 3h5v5M4 20L21 3M21 16v5h-5M15 15l6 6M4 4l5 5"/>
                      </svg>
                    </div>
                    <div class="stat-info">
                      <span class="stat-value">{{ gitChanges() }}</span>
                      <span class="stat-label">Changes</span>
                    </div>
                  </div>

                  <div class="stat-card">
                    <div class="stat-icon status">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                          d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
                      </svg>
                    </div>
                    <div class="stat-info">
                      <span class="stat-value status-ok">OK</span>
                      <span class="stat-label">Status</span>
                    </div>
                  </div>
                </div>

                <div class="quick-actions">
                  <h2>Quick Actions</h2>
                  <div class="actions-grid">
                    <button class="action-card" (click)="emitAction('new-project')">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                          d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
                      </svg>
                      <span>New Project</span>
                    </button>
                    <button class="action-card" (click)="emitAction('open-project')">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                          d="M5 19a2 2 0 01-2-2V7a2 2 0 012-2h4l2 2h4a2 2 0 012 2v1M5 19h14a2 2 0 002-2v-5a2 2 0 00-2-2H9a2 2 0 00-2 2v5a2 2 0 01-2 2z"/>
                      </svg>
                      <span>Open Project</span>
                    </button>
                    <button class="action-card" (click)="emitAction('run-dev')">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                          d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"/>
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                          d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                      </svg>
                      <span>Run Dev Server</span>
                    </button>
                    <button class="action-card" (click)="onRouteChange('terminal')">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                          d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/>
                      </svg>
                      <span>Open Terminal</span>
                    </button>
                  </div>
                </div>
              </div>
            }
            @case ('explorer') {
              <div class="panel-view">
                <h2>File Explorer</h2>
                <p>Browse and manage your project files.</p>
              </div>
            }
            @case ('search') {
              <div class="panel-view">
                <h2>Search</h2>
                <p>Search across all files in your workspace.</p>
              </div>
            }
            @case ('git') {
              <div class="panel-view">
                <h2>Source Control</h2>
                <p>Manage your Git repositories and commits.</p>
              </div>
            }
            @case ('debug') {
              <div class="panel-view">
                <h2>Debug</h2>
                <p>Debug your applications.</p>
              </div>
            }
            @case ('terminal') {
              <div class="panel-view terminal">
                <h2>Terminal</h2>
                <div class="terminal-output">
                  <pre>$ core dev health
18 repos | clean | synced

$ _</pre>
                </div>
              </div>
            }
            @case ('settings') {
              <div class="panel-view">
                <h2>Settings</h2>
                <p>Configure your IDE preferences.</p>
              </div>
            }
            @default {
              <div class="panel-view">
                <h2>{{ currentRoute() }}</h2>
              </div>
            }
          }
        </div>

        <!-- Status Bar -->
        <div class="status-bar">
          <div class="status-left">
            <span class="status-item branch">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M16 3h5v5M4 20L21 3M21 16v5h-5M15 15l6 6M4 4l5 5"/>
              </svg>
              main
            </span>
            <span class="status-item">UTF-8</span>
          </div>
          <div class="status-right">
            <span class="status-item">Core IDE v0.1.0</span>
          </div>
        </div>
      </div>
    </div>
  `,
  styles: [`
    :host {
      display: block;
      width: 100%;
      height: 100%;
    }

    .ide-layout {
      display: flex;
      height: 100%;
      background: #1a1b26;
      color: #a9b1d6;
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    }

    .ide-main {
      flex: 1;
      display: flex;
      flex-direction: column;
      min-width: 0;
    }

    .top-bar {
      display: flex;
      align-items: center;
      justify-content: space-between;
      height: 40px;
      padding: 0 1rem;
      background: #16161e;
      border-bottom: 1px solid #24283b;
    }

    .breadcrumb {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      font-size: 0.8125rem;
    }

    .breadcrumb-item {
      color: #565f89;
    }

    .breadcrumb-item.active {
      color: #c0caf5;
      text-transform: capitalize;
    }

    .breadcrumb-sep {
      color: #414868;
    }

    .top-bar-actions {
      display: flex;
      align-items: center;
      gap: 1rem;
    }

    .time {
      font-size: 0.75rem;
      color: #565f89;
      font-family: 'JetBrains Mono', monospace;
    }

    .ide-content {
      flex: 1;
      overflow-y: auto;
      padding: 1.5rem;
    }

    .dashboard-view h1 {
      font-size: 1.75rem;
      font-weight: 600;
      color: #c0caf5;
      margin: 0 0 0.5rem 0;
    }

    .subtitle {
      color: #565f89;
      margin: 0 0 2rem 0;
    }

    .stats-grid {
      display: grid;
      grid-template-columns: repeat(4, 1fr);
      gap: 1rem;
      margin-bottom: 2rem;
    }

    .stat-card {
      display: flex;
      align-items: center;
      gap: 1rem;
      padding: 1.25rem;
      background: #16161e;
      border: 1px solid #24283b;
      border-radius: 8px;
    }

    .stat-icon {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 48px;
      height: 48px;
      border-radius: 8px;
    }

    .stat-icon svg {
      width: 24px;
      height: 24px;
    }

    .stat-icon.projects {
      background: rgba(122, 162, 247, 0.15);
      color: #7aa2f7;
    }

    .stat-icon.tasks {
      background: rgba(158, 206, 106, 0.15);
      color: #9ece6a;
    }

    .stat-icon.git {
      background: rgba(247, 118, 142, 0.15);
      color: #f7768e;
    }

    .stat-icon.status {
      background: rgba(158, 206, 106, 0.15);
      color: #9ece6a;
    }

    .stat-info {
      display: flex;
      flex-direction: column;
    }

    .stat-value {
      font-size: 1.5rem;
      font-weight: 600;
      color: #c0caf5;
    }

    .stat-value.status-ok {
      color: #9ece6a;
    }

    .stat-label {
      font-size: 0.8125rem;
      color: #565f89;
    }

    .quick-actions h2 {
      font-size: 1.125rem;
      font-weight: 600;
      color: #c0caf5;
      margin: 0 0 1rem 0;
    }

    .actions-grid {
      display: grid;
      grid-template-columns: repeat(4, 1fr);
      gap: 1rem;
    }

    .action-card {
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 0.75rem;
      padding: 1.5rem;
      background: #16161e;
      border: 1px solid #24283b;
      border-radius: 8px;
      color: #a9b1d6;
      cursor: pointer;
      transition: all 0.15s ease;
    }

    .action-card:hover {
      background: #1f2335;
      border-color: #7aa2f7;
      color: #c0caf5;
    }

    .action-card svg {
      width: 32px;
      height: 32px;
      color: #7aa2f7;
    }

    .action-card span {
      font-size: 0.875rem;
      font-weight: 500;
    }

    .panel-view {
      padding: 1rem;
    }

    .panel-view h2 {
      font-size: 1.25rem;
      font-weight: 600;
      color: #c0caf5;
      margin: 0 0 0.5rem 0;
    }

    .panel-view p {
      color: #565f89;
    }

    .panel-view.terminal {
      display: flex;
      flex-direction: column;
      height: 100%;
    }

    .terminal-output {
      flex: 1;
      background: #16161e;
      border: 1px solid #24283b;
      border-radius: 8px;
      padding: 1rem;
      font-family: 'JetBrains Mono', 'Fira Code', monospace;
      font-size: 0.875rem;
      overflow: auto;
    }

    .terminal-output pre {
      margin: 0;
      color: #9ece6a;
    }

    .status-bar {
      display: flex;
      align-items: center;
      justify-content: space-between;
      height: 24px;
      padding: 0 0.75rem;
      background: #7aa2f7;
      color: #1a1b26;
      font-size: 0.6875rem;
      font-weight: 500;
    }

    .status-left, .status-right {
      display: flex;
      align-items: center;
      gap: 1rem;
    }

    .status-item {
      display: flex;
      align-items: center;
      gap: 0.25rem;
    }

    .status-item svg {
      width: 12px;
      height: 12px;
    }

    @media (max-width: 1024px) {
      .stats-grid, .actions-grid {
        grid-template-columns: repeat(2, 1fr);
      }
    }

    @media (max-width: 640px) {
      .stats-grid, .actions-grid {
        grid-template-columns: 1fr;
      }
    }
  `]
})
export class IdeComponent implements OnInit, OnDestroy {
  private isBrowser: boolean;
  private timeEventCleanup?: () => void;
  currentRoute = signal('dashboard');
  currentTime = signal('');
  projectCount = signal(18);
  taskCount = signal(5);
  gitChanges = signal(12);

  constructor(@Inject(PLATFORM_ID) platformId: Object) {
    this.isBrowser = isPlatformBrowser(platformId);
  }

  ngOnInit() {
    if (!this.isBrowser) return;

    import('@wailsio/runtime').then(({ Events }) => {
      this.timeEventCleanup = Events.On('time', (time: { data: string }) => {
        this.currentTime.set(time.data);
      });
    });
  }

  ngOnDestroy() {
    this.timeEventCleanup?.();
  }

  onRouteChange(route: string) {
    this.currentRoute.set(route);
  }

  emitAction(action: string) {
    if (!this.isBrowser) return;
    import('@wailsio/runtime').then(({ Events }) => {
      Events.Emit('action', action);
    });
  }
}
