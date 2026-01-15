import { Component, CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';

import { TranslateModule } from '@ngx-translate/core';

@Component({
    selector: 'app-exchange-page',
    imports: [TranslateModule],
    schemas: [CUSTOM_ELEMENTS_SCHEMA],
    template: `
    <div class="exchange">
      <wa-tab-group>
        <wa-tab slot="nav" panel="listings">{{ 'exchange.listings' | translate }}</wa-tab>
        <wa-tab slot="nav" panel="fills">{{ 'exchange.fills' | translate }}</wa-tab>
        <wa-tab slot="nav" panel="auctions">{{ 'exchange.auctions' | translate }}</wa-tab>

        <wa-tab-panel name="listings">
          <div class="exchange__content">
            <div class="exchange__header">
              <h3 style="margin: 0;">{{ 'exchange.yourListings' | translate }}</h3>
              <wa-button size="small" variant="primary">
                <wa-icon slot="prefix" name="fa-solid fa-plus"></wa-icon>
                {{ 'exchange.createListing' | translate }}
              </wa-button>
            </div>

            <wa-callout variant="neutral">
              <div class="exchange__empty">
                <wa-icon name="fa-solid fa-store" style="font-size: 2.5rem; opacity: 0.3; margin-bottom: 1rem;"></wa-icon>
                <p>{{ 'exchange.noActiveListings' | translate }}</p>
                <p style="margin-top: 0.5rem; font-size: 0.875rem; color: var(--wa-color-neutral-600);">
                  {{ 'exchange.listDomainsInfo' | translate }}
                </p>
              </div>
            </wa-callout>
          </div>
        </wa-tab-panel>

        <wa-tab-panel name="fills">
          <div class="exchange__content">
            <div class="exchange__header">
              <h3 style="margin: 0;">{{ 'exchange.yourFills' | translate }}</h3>
            </div>

            <wa-callout variant="neutral">
              <div class="exchange__empty">
                <wa-icon name="fa-solid fa-handshake" style="font-size: 2.5rem; opacity: 0.3; margin-bottom: 1rem;"></wa-icon>
                <p>{{ 'exchange.noFilledOrders' | translate }}</p>
                <p style="margin-top: 0.5rem; font-size: 0.875rem; color: var(--wa-color-neutral-600);">
                  {{ 'exchange.completedPurchasesInfo' | translate }}
                </p>
              </div>
            </wa-callout>
          </div>
        </wa-tab-panel>

        <wa-tab-panel name="auctions">
          <div class="exchange__content">
            <div class="exchange__header">
              <h3 style="margin: 0;">{{ 'exchange.marketplaceAuctions' | translate }}</h3>
              <wa-button size="small" variant="neutral">
                <wa-icon slot="prefix" name="fa-solid fa-rotate"></wa-icon>
                {{ 'exchange.refresh' | translate }}
              </wa-button>
            </div>

            <wa-callout variant="neutral">
              <div class="exchange__empty">
                <wa-icon name="fa-solid fa-gavel" style="font-size: 2.5rem; opacity: 0.3; margin-bottom: 1rem;"></wa-icon>
                <p>{{ 'exchange.noActiveAuctions' | translate }}</p>
                <p style="margin-top: 0.5rem; font-size: 0.875rem; color: var(--wa-color-neutral-600);">
                  {{ 'exchange.browseAuctionsInfo' | translate }}
                </p>
              </div>
            </wa-callout>
          </div>
        </wa-tab-panel>
      </wa-tab-group>
    </div>
  `,
    styles: [`
    .exchange { }

    .exchange__content { padding: 1.5rem 0; }

    .exchange__header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem; }

    .exchange__empty { text-align: center; padding: 2rem 1rem; }
    .exchange__empty p { margin: 0; color: var(--wa-color-neutral-700, #374151); }
  `]
})
export class ExchangePage {}
