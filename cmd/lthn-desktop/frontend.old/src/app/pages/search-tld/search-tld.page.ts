import { Component, CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { Router } from '@angular/router';

@Component({
    selector: 'app-search-tld-page',
    imports: [FormsModule],
    schemas: [CUSTOM_ELEMENTS_SCHEMA],
    template: `
    <div class="search-tld">
      <div class="search-tld__header">
        <wa-input placeholder="Search for a name..." style="flex: 1;">
          <input slot="input" [(ngModel)]="query" (keyup.enter)="onSearch()" />
        </wa-input>
        <wa-button variant="primary" (click)="onSearch()" [disabled]="loading || !query">
          <wa-icon slot="prefix" name="fa-solid fa-magnifying-glass"></wa-icon>
          Search
        </wa-button>
      </div>

      @if (loading) {
        <div class="search-tld__content">
          <wa-spinner></wa-spinner>
          <p style="text-align: center; margin-top: 1rem;">Searching...</p>
        </div>
      }

      @if (!loading && !result) {
        <div class="search-tld__content">
          <wa-callout variant="neutral">
            <div style="text-align: center; padding: 1rem;">
              <wa-icon name="fa-solid fa-search" style="font-size: 2.5rem; opacity: 0.3; margin-bottom: 1rem;"></wa-icon>
              <p style="margin: 0;">Enter a name to search for availability and auction status.</p>
            </div>
          </wa-callout>
        </div>
      }

      @if (!loading && result) {
        <div class="search-tld__content">
          <wa-card>
            <div slot="header" style="display: flex; justify-content: space-between; align-items: center;">
              <h3 style="margin: 0; font-size: 1.5rem;">{{result.name}}/</h3>
              <wa-badge [attr.variant]="result.available ? 'success' : 'warning'">
                {{result.available ? 'Available' : 'In Auction'}}
              </wa-badge>
            </div>
            <div class="search-result">
              <div class="search-result__section">
                <div class="search-result__label">Status</div>
                <div class="search-result__value">{{result.status}}</div>
              </div>
              @if (result.currentBid) {
                <div class="search-result__section">
                  <div class="search-result__label">Current Bid</div>
                  <div class="search-result__value">{{result.currentBid}} HNS</div>
                </div>
              }
              @if (result.blocksUntil) {
                <div class="search-result__section">
                  <div class="search-result__label">{{result.blocksUntilLabel}}</div>
                  <div class="search-result__value">~{{result.blocksUntil}} blocks</div>
                </div>
              }
              <div class="search-result__actions">
                <wa-button variant="primary" [disabled]="!result.available">
                  <wa-icon slot="prefix" name="fa-solid fa-gavel"></wa-icon>
                  Place Bid
                </wa-button>
                <wa-button variant="neutral">
                  <wa-icon slot="prefix" name="fa-solid fa-eye"></wa-icon>
                  Watch
                </wa-button>
              </div>
            </div>
          </wa-card>
        </div>
      }
    </div>
    `,
    styles: [`
    .search-tld { display: flex; flex-direction: column; gap: 1.5rem; }

    .search-tld__header { display: flex; gap: 1rem; align-items: center; }

    .search-tld__content { }

    .search-result { padding: 1rem 0; }
    .search-result__section { display: flex; justify-content: space-between; padding: 0.75rem 0; border-bottom: 1px solid var(--wa-color-neutral-200, #e5e7eb); }
    .search-result__section:last-of-type { border-bottom: none; }
    .search-result__label { font-size: 0.875rem; font-weight: 600; color: var(--wa-color-neutral-600, #4b5563); }
    .search-result__value { font-size: 0.875rem; color: var(--wa-color-neutral-900, #111827); font-weight: 500; }
    .search-result__actions { display: flex; gap: 0.75rem; margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid var(--wa-color-neutral-200, #e5e7eb); }
  `]
})
export class SearchTldPage {
  query = '';
  loading = false;
  result: any = null;

  constructor(private router: Router) {}

  async onSearch() {
    if (!this.query.trim()) return;

    this.loading = true;
    this.result = null;

    // Simulate API call
    await new Promise(r => setTimeout(r, 800));

    const name = this.query.trim().replace('/', '');
    const available = Math.random() > 0.3;

    this.result = {
      name,
      available,
      status: available ? 'Available for bidding' : 'Auction in progress',
      currentBid: available ? null : (Math.random() * 100).toFixed(2),
      blocksUntil: available ? null : Math.floor(Math.random() * 5000),
      blocksUntilLabel: available ? null : 'Blocks until reveal'
    };

    this.loading = false;
  }
}
