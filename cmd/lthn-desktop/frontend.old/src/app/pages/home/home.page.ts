import { Component, CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';

import { TranslateModule } from '@ngx-translate/core';

@Component({
    selector: 'app-home-page',
    imports: [TranslateModule],
    schemas: [CUSTOM_ELEMENTS_SCHEMA],
    template: `
    <div class="account">
      <!-- Balance Header -->
      <div class="account__header">
        <div class="account__header__section">
          <span class="label">{{ 'home.spendable' | translate }}</span>
          <p class="amount">0.00 {{ 'common.hns' | translate }}</p>
          <span class="subtext">~$0.00 {{ 'common.usd' | translate }}</span>
        </div>

        <div class="account__header__section">
          <span class="label">{{ 'home.locked' | translate }}</span>
          <p class="amount">0.00 {{ 'common.hns' | translate }}</p>
          <span class="subtext">{{ 'home.inBids' | translate }} (0 {{ 'home.bids' | translate }})</span>
        </div>
      </div>

      <!-- Actionable Items Cards -->
      <div class="account__cards">
        <wa-card class="account__card">
          <div slot="header" class="account__card__header">
            <wa-icon name="fa-solid fa-eye" class="account__card__icon"></wa-icon>
            <span>{{ 'home.revealable' | translate }}</span>
          </div>
          <div class="account__card__content">
            <div class="account__card__amount">0.00 {{ 'common.hns' | translate }}</div>
            <div class="account__card__detail">{{ 'home.bidsReadyToReveal' | translate: {count: 0} }}</div>
            <div class="account__card__action">
              <wa-button size="small" disabled>{{ 'home.revealAll' | translate }}</wa-button>
            </div>
          </div>
        </wa-card>

        <wa-card class="account__card">
          <div slot="header" class="account__card__header">
            <wa-icon name="fa-solid fa-gift" class="account__card__icon"></wa-icon>
            <span>{{ 'home.redeemable' | translate }}</span>
          </div>
          <div class="account__card__content">
            <div class="account__card__amount">0.00 {{ 'common.hns' | translate }}</div>
            <div class="account__card__detail">{{ 'home.bidsReadyToRedeem' | translate: {count: 0} }}</div>
            <div class="account__card__action">
              <wa-button size="small" disabled>{{ 'home.redeemAll' | translate }}</wa-button>
            </div>
          </div>
        </wa-card>

        <wa-card class="account__card">
          <div slot="header" class="account__card__header">
            <wa-icon name="fa-solid fa-pen-to-square" class="account__card__icon"></wa-icon>
            <span>{{ 'home.registerable' | translate }}</span>
          </div>
          <div class="account__card__content">
            <div class="account__card__amount">0.00 {{ 'common.hns' | translate }}</div>
            <div class="account__card__detail">{{ 'home.namesReadyToRegister' | translate: {count: 0} }}</div>
            <div class="account__card__action">
              <wa-button size="small" disabled>{{ 'home.registerAll' | translate }}</wa-button>
            </div>
          </div>
        </wa-card>

        <wa-card class="account__card">
          <div slot="header" class="account__card__header">
            <wa-icon name="fa-solid fa-clock-rotate-left" class="account__card__icon"></wa-icon>
            <span>{{ 'home.renewable' | translate }}</span>
          </div>
          <div class="account__card__content">
            <div class="account__card__detail">{{ 'home.domainsExpiringSoon' | translate: {count: 0} }}</div>
            <div class="account__card__action">
              <wa-button size="small" disabled>{{ 'home.renewAll' | translate }}</wa-button>
            </div>
          </div>
        </wa-card>

        <wa-card class="account__card">
          <div slot="header" class="account__card__header">
            <wa-icon name="fa-solid fa-arrow-right-arrow-left" class="account__card__icon"></wa-icon>
            <span>{{ 'home.transferring' | translate }}</span>
          </div>
          <div class="account__card__content">
            <div class="account__card__detail">{{ 'home.domainsInTransfer' | translate: {count: 0} }}</div>
          </div>
        </wa-card>

        <wa-card class="account__card">
          <div slot="header" class="account__card__header">
            <wa-icon name="fa-solid fa-check" class="account__card__icon"></wa-icon>
            <span>{{ 'home.finalizable' | translate }}</span>
          </div>
          <div class="account__card__content">
            <div class="account__card__detail">{{ 'home.transfersReadyToFinalize' | translate: {count: 0} }}</div>
            <div class="account__card__action">
              <wa-button size="small" disabled>{{ 'home.finalizeAll' | translate }}</wa-button>
            </div>
          </div>
        </wa-card>
      </div>

      <!-- Transaction History -->
      <div class="account__transactions">
        <div class="account__panel-title">{{ 'home.transactionHistory' | translate }}</div>
        <wa-callout variant="neutral">
          {{ 'home.noTransactions' | translate }}
        </wa-callout>
      </div>
    </div>
  `,
    styles: [`
    .account { display: flex; flex-direction: column; gap: 1.5rem; }

    .account__header { display: flex; gap: 1rem; flex-wrap: wrap; padding: 1.5rem; background: var(--wa-color-neutral-50, #fafafa); border-radius: 0.5rem; border: 1px solid var(--wa-color-neutral-200, #e5e7eb); }
    .account__header__section { display: flex; flex-direction: column; padding: 0 1rem; }
    .account__header__section:not(:last-child) { border-right: 1px solid var(--wa-color-neutral-300, #d1d5db); }
    .account__header__section .label { font-size: 0.75rem; font-weight: 600; text-transform: uppercase; color: var(--wa-color-neutral-600, #4b5563); margin-bottom: 0.25rem; }
    .account__header__section .amount { font-size: 1.875rem; font-weight: 700; color: var(--wa-color-neutral-900, #111827); margin: 0.25rem 0; }
    .account__header__section .subtext { font-size: 0.875rem; color: var(--wa-color-neutral-500, #6b7280); }

    .account__cards { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 1rem; }
    .account__card { height: 100%; }
    .account__card__header { display: flex; align-items: center; gap: 0.5rem; font-weight: 600; padding: 1rem; }
    .account__card__icon { font-size: 1.25rem; color: var(--wa-color-primary-600, #4f46e5); }
    .account__card__content { padding: 0 1rem 1rem; }
    .account__card__amount { font-size: 1.5rem; font-weight: 700; color: var(--wa-color-neutral-900, #111827); margin-bottom: 0.5rem; }
    .account__card__detail { font-size: 0.875rem; color: var(--wa-color-neutral-600, #4b5563); margin-bottom: 0.75rem; }
    .account__card__action { margin-top: 0.75rem; }

    .account__transactions { }
    .account__panel-title { font-size: 1.25rem; font-weight: 600; margin-bottom: 1rem; }
  `]
})
export class HomePage {}
