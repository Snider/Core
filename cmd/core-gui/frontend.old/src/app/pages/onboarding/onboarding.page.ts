import { Component, CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';


@Component({
    selector: 'app-onboarding-page',
    imports: [],
    schemas: [CUSTOM_ELEMENTS_SCHEMA],
    template: `
    <div class="onboarding">
      <div class="onboarding__header">
        <h2 style="margin: 0;">Welcome to Bob Wallet</h2>
        <p style="margin: 0.5rem 0 0 0; color: var(--wa-color-neutral-600);">
          Set up your wallet to start managing Handshake names
        </p>
      </div>

      <wa-tab-group>
        <wa-tab slot="nav" panel="create">Create New Wallet</wa-tab>
        <wa-tab slot="nav" panel="import">Import Seed</wa-tab>
        <wa-tab slot="nav" panel="ledger">Connect Ledger</wa-tab>

        <wa-tab-panel name="create">
          <div class="onboarding__content">
            <wa-callout variant="info">
              <strong>Important:</strong> Write down your seed phrase and store it in a secure location.
              You will need it to recover your wallet.
            </wa-callout>

            <div class="seed-display">
              <div class="seed-words">
                <div class="seed-word">abandon</div>
                <div class="seed-word">ability</div>
                <div class="seed-word">able</div>
                <div class="seed-word">about</div>
                <div class="seed-word">above</div>
                <div class="seed-word">absent</div>
                <div class="seed-word">absorb</div>
                <div class="seed-word">abstract</div>
                <div class="seed-word">absurd</div>
                <div class="seed-word">abuse</div>
                <div class="seed-word">access</div>
                <div class="seed-word">accident</div>
              </div>
            </div>

            <div class="onboarding__actions">
              <wa-button variant="primary">
                <wa-icon slot="prefix" name="fa-solid fa-copy"></wa-icon>
                Copy Seed Phrase
              </wa-button>
              <wa-button variant="neutral">
                <wa-icon slot="prefix" name="fa-solid fa-check"></wa-icon>
                I've Saved My Seed
              </wa-button>
            </div>
          </div>
        </wa-tab-panel>

        <wa-tab-panel name="import">
          <div class="onboarding__content">
            <wa-callout variant="neutral">
              Enter your 12 or 24 word seed phrase to restore an existing wallet.
            </wa-callout>

            <div style="margin-top: 1.5rem;">
              <label style="display: block; font-weight: 600; margin-bottom: 0.5rem;">Seed Phrase</label>
              <wa-input placeholder="Enter your seed phrase" style="width: 100%;">
                <textarea slot="input" rows="4" style="resize: vertical; font-family: monospace;"></textarea>
              </wa-input>
            </div>

            <div class="onboarding__actions">
              <wa-button variant="primary">
                <wa-icon slot="prefix" name="fa-solid fa-download"></wa-icon>
                Import Wallet
              </wa-button>
            </div>
          </div>
        </wa-tab-panel>

        <wa-tab-panel name="ledger">
          <div class="onboarding__content">
            <wa-callout variant="info">
              Connect your Ledger hardware wallet to manage your Handshake names securely.
            </wa-callout>

            <div class="ledger-instructions">
              <h4 style="margin: 1.5rem 0 1rem 0;">Instructions:</h4>
              <ol style="margin: 0; padding-left: 1.5rem; color: var(--wa-color-neutral-700);">
                <li style="margin-bottom: 0.5rem;">Connect your Ledger device via USB</li>
                <li style="margin-bottom: 0.5rem;">Enter your PIN on the device</li>
                <li style="margin-bottom: 0.5rem;">Open the Handshake app on your Ledger</li>
                <li style="margin-bottom: 0.5rem;">Click "Connect" below</li>
              </ol>
            </div>

            <div class="onboarding__actions">
              <wa-button variant="primary">
                <wa-icon slot="prefix" name="fa-solid fa-usb"></wa-icon>
                Connect Ledger
              </wa-button>
            </div>
          </div>
        </wa-tab-panel>
      </wa-tab-group>
    </div>
  `,
    styles: [`
    .onboarding { }

    .onboarding__header { margin-bottom: 2rem; }

    .onboarding__content { padding: 1.5rem 0; }

    .seed-display { margin: 1.5rem 0; padding: 1.5rem; background: var(--wa-color-neutral-50, #fafafa); border: 2px solid var(--wa-color-neutral-200, #e5e7eb); border-radius: 0.5rem; }

    .seed-words { display: grid; grid-template-columns: repeat(3, 1fr); gap: 0.75rem; }

    .seed-word { padding: 0.75rem; background: white; border: 1px solid var(--wa-color-neutral-300, #d1d5db); border-radius: 0.375rem; font-family: monospace; font-size: 0.875rem; text-align: center; }

    .onboarding__actions { display: flex; gap: 0.75rem; margin-top: 1.5rem; }

    .ledger-instructions { margin-top: 1rem; }
  `]
})
export class OnboardingPage {}
