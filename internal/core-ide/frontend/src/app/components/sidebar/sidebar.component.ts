import { Component, Input, Output, EventEmitter } from '@angular/core';
import { CommonModule } from '@angular/common';
import { DomSanitizer, SafeHtml } from '@angular/platform-browser';

interface NavItem {
  id: string;
  label: string;
  icon: SafeHtml;
}

@Component({
  selector: 'app-sidebar',
  standalone: true,
  imports: [CommonModule],
  template: `
    <nav class="sidebar">
      <div class="sidebar-header">
        <div class="logo">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
              d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"/>
          </svg>
        </div>
      </div>

      <div class="nav-items">
        @for (item of navItems; track item.id) {
          <button
            class="nav-item"
            [class.active]="currentRoute === item.id"
            (click)="routeChange.emit(item.id)"
            [title]="item.label">
            <div class="nav-icon" [innerHTML]="item.icon"></div>
          </button>
        }
      </div>

      <div class="sidebar-footer">
        <button class="nav-item" (click)="routeChange.emit('settings')" title="Settings">
          <div class="nav-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
            </svg>
          </div>
        </button>
      </div>
    </nav>
  `,
  styles: [`
    .sidebar {
      display: flex;
      flex-direction: column;
      width: 56px;
      background: #16161e;
      border-right: 1px solid #24283b;
    }

    .sidebar-header {
      display: flex;
      align-items: center;
      justify-content: center;
      height: 56px;
      border-bottom: 1px solid #24283b;
    }

    .logo {
      width: 28px;
      height: 28px;
      color: #7aa2f7;
    }

    .logo svg {
      width: 100%;
      height: 100%;
    }

    .nav-items {
      flex: 1;
      display: flex;
      flex-direction: column;
      padding: 0.5rem 0;
      gap: 0.25rem;
    }

    .nav-item {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 100%;
      height: 44px;
      background: transparent;
      border: none;
      color: #565f89;
      cursor: pointer;
      transition: all 0.15s ease;
      position: relative;
    }

    .nav-item:hover {
      color: #a9b1d6;
      background: rgba(122, 162, 247, 0.1);
    }

    .nav-item.active {
      color: #7aa2f7;
      background: rgba(122, 162, 247, 0.15);
    }

    .nav-item.active::before {
      content: '';
      position: absolute;
      left: 0;
      top: 8px;
      bottom: 8px;
      width: 2px;
      background: #7aa2f7;
      border-radius: 0 2px 2px 0;
    }

    .nav-icon {
      width: 22px;
      height: 22px;
    }

    .nav-icon svg {
      width: 100%;
      height: 100%;
    }

    .sidebar-footer {
      border-top: 1px solid #24283b;
      padding: 0.5rem 0;
    }
  `]
})
export class SidebarComponent {
  @Input() currentRoute = 'dashboard';
  @Output() routeChange = new EventEmitter<string>();

  constructor(private sanitizer: DomSanitizer) {
    this.navItems = this.createNavItems();
  }

  navItems: NavItem[];

  private createNavItems(): NavItem[] {
    return [
      {
        id: 'dashboard',
        label: 'Dashboard',
        icon: this.sanitizer.bypassSecurityTrustHtml(`<svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z"/>
        </svg>`)
      },
      {
        id: 'explorer',
        label: 'Explorer',
        icon: this.sanitizer.bypassSecurityTrustHtml(`<svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"/>
        </svg>`)
      },
      {
        id: 'search',
        label: 'Search',
        icon: this.sanitizer.bypassSecurityTrustHtml(`<svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
        </svg>`)
      },
      {
        id: 'git',
        label: 'Source Control',
        icon: this.sanitizer.bypassSecurityTrustHtml(`<svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M16 3h5v5M4 20L21 3M21 16v5h-5M15 15l6 6M4 4l5 5"/>
        </svg>`)
      },
      {
        id: 'debug',
        label: 'Debug',
        icon: this.sanitizer.bypassSecurityTrustHtml(`<svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"/>
        </svg>`)
      },
      {
        id: 'terminal',
        label: 'Terminal',
        icon: this.sanitizer.bypassSecurityTrustHtml(`<svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/>
        </svg>`)
      },
    ];
  }
}
