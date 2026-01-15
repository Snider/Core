import { Component, CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';

import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { TranslateModule } from '@ngx-translate/core';

@Component({
    selector: 'app-domain-manager-page',
    imports: [FormsModule, TranslateModule],
    schemas: [CUSTOM_ELEMENTS_SCHEMA],
    template: `
    <div class="domain-manager">
      <div class="domain-manager__header">
        <wa-input [placeholder]="'domainManager.searchPlaceholder' | translate" style="flex: 1; max-width: 400px;">
          <input slot="input" [(ngModel)]="searchQuery" />
        </wa-input>

        <div class="domain-manager__actions">
          <wa-button size="small" variant="neutral">
            <wa-icon slot="prefix" name="fa-solid fa-download"></wa-icon>
            {{ 'domainManager.export' | translate }}
          </wa-button>
          <wa-button size="small" variant="neutral">
            <wa-icon slot="prefix" name="fa-solid fa-arrow-right-arrow-left"></wa-icon>
            {{ 'domainManager.bulkTransfer' | translate }}
          </wa-button>
          <wa-button size="small" variant="primary">
            <wa-icon slot="prefix" name="fa-solid fa-receipt"></wa-icon>
            {{ 'domainManager.claimNamePayment' | translate }}
          </wa-button>
        </div>
      </div>

      @if (domains.length === 0) {
        <div class="domain-manager__content">
          <wa-callout variant="neutral">
            <div class="domain-manager__empty">
              <wa-icon name="fa-solid fa-folder-open" style="font-size: 3rem; opacity: 0.3; margin-bottom: 1rem;"></wa-icon>
              <p>{{ 'domainManager.emptyState' | translate }}</p>
              <p style="margin-top: 0.5rem;">
                <a routerLink="/domains" style="color: var(--wa-color-primary-600); text-decoration: underline;">{{ 'domainManager.browseDomainsLink' | translate }}</a> {{ 'domainManager.toGetStarted' | translate }}
              </p>
            </div>
          </wa-callout>
        </div>
      }

      @if (domains.length > 0) {
        <div class="domain-manager__content">
          <div class="domain-manager__table">
            <div class="table-header">
              <div class="table-col">{{ 'domainManager.name' | translate }}</div>
              <div class="table-col">{{ 'domainManager.expires' | translate }}</div>
              <div class="table-col">{{ 'domainManager.highestBid' | translate }}</div>
            </div>
            @for (domain of domains; track domain) {
              <div class="table-row" (click)="viewDomain(domain.name)">
                <div class="table-col">{{domain.name}}/</div>
                <div class="table-col">{{domain.expires}}</div>
                <div class="table-col">{{domain.highestBid}} {{ 'common.hns' | translate }}</div>
              </div>
            }
          </div>
          <div class="domain-manager__pagination">
            <wa-select size="small" value="10" style="width: 80px;">
              <wa-option value="5">5</wa-option>
              <wa-option value="10">10</wa-option>
              <wa-option value="20">20</wa-option>
              <wa-option value="50">50</wa-option>
            </wa-select>
            <span class="pagination-info">{{ 'domainManager.showingDomains' | translate: {count: domains.length} }}</span>
          </div>
        </div>
      }
    </div>
    `,
    styles: [`
    .domain-manager { display: flex; flex-direction: column; gap: 1.5rem; }

    .domain-manager__header { display: flex; gap: 1rem; align-items: center; flex-wrap: wrap; }
    .domain-manager__actions { display: flex; gap: 0.5rem; }

    .domain-manager__content { }
    .domain-manager__empty { text-align: center; padding: 2rem; }
    .domain-manager__empty p { margin: 0; color: var(--wa-color-neutral-600, #4b5563); }

    .domain-manager__table { border: 1px solid var(--wa-color-neutral-200, #e5e7eb); border-radius: 0.5rem; overflow: hidden; }
    .table-header { display: grid; grid-template-columns: 2fr 1fr 1fr; background: var(--wa-color-neutral-50, #fafafa); font-weight: 600; font-size: 0.875rem; }
    .table-row { display: grid; grid-template-columns: 2fr 1fr 1fr; border-top: 1px solid var(--wa-color-neutral-200, #e5e7eb); cursor: pointer; transition: background 0.15s; }
    .table-row:hover { background: var(--wa-color-neutral-50, #fafafa); }
    .table-col { padding: 0.75rem 1rem; }

    .domain-manager__pagination { display: flex; align-items: center; gap: 1rem; margin-top: 1rem; }
    .pagination-info { font-size: 0.875rem; color: var(--wa-color-neutral-600, #4b5563); }
  `]
})
export class DomainManagerPage {
  searchQuery = '';
  domains: any[] = [];

  constructor(private router: Router) {}

  viewDomain(name: string) {
    // Navigate to individual domain view
    console.log('View domain:', name);
  }
}
