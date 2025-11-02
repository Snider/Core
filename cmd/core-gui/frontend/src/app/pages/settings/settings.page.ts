import { Component, CUSTOM_ELEMENTS_SCHEMA, inject } from '@angular/core';

import { FormsModule } from '@angular/forms';
import { FileDialogService } from '../../services/file-dialog.service';
import { ClipboardService } from '../../services/clipboard.service';

@Component({
    selector: 'app-settings-page',
    imports: [FormsModule],
    schemas: [CUSTOM_ELEMENTS_SCHEMA],
    template: `
    <div class="settings">
      <wa-tab-group>
        <wa-tab slot="nav" panel="general">General</wa-tab>
        <wa-tab slot="nav" panel="wallet">Wallet</wa-tab>
        <wa-tab slot="nav" panel="connection">Connection</wa-tab>
        <wa-tab slot="nav" panel="advanced">Advanced</wa-tab>

        <wa-tab-panel name="general">
          <div class="settings__section">
            <h3 class="settings__section-title">Language</h3>
            <wa-select value="en-US" style="max-width: 300px;">
              <wa-option value="en-US">English (US)</wa-option>
              <wa-option value="zh-CN">中文 (简体)</wa-option>
              <wa-option value="es-ES">Español</wa-option>
            </wa-select>
          </div>

          <div class="settings__section">
            <h3 class="settings__section-title">Block Explorer</h3>
            <wa-select value="hnsnetwork" style="max-width: 300px;">
              <wa-option value="hnsnetwork">HNS Network</wa-option>
              <wa-option value="niami">Niami</wa-option>
              <wa-option value="hnscan">HNScan</wa-option>
            </wa-select>
          </div>

          <div class="settings__section">
            <h3 class="settings__section-title">Theme</h3>
            <wa-select value="light" style="max-width: 300px;">
              <wa-option value="light">Light</wa-option>
              <wa-option value="dark">Dark</wa-option>
              <wa-option value="system">System</wa-option>
            </wa-select>
          </div>
        </wa-tab-panel>

        <wa-tab-panel name="wallet">
          <div class="settings__section">
            <h3 class="settings__section-title">Wallet Directory</h3>
            <p class="settings__description">Location where wallet data is stored</p>
            <wa-input value="~/.bob-wallet" readonly style="max-width: 500px;">
              <input slot="input" />
            </wa-input>
            <div style="margin-top: 0.75rem;">
              <wa-button size="small" (click)="pickDir()">Change Directory</wa-button>
            </div>
          </div>

          <div class="settings__section">
            <h3 class="settings__section-title">Backup</h3>
            <p class="settings__description">Export wallet seed phrase and settings</p>
            <wa-button size="small" (click)="saveFile()">
              <wa-icon slot="prefix" name="fa-solid fa-download"></wa-icon>
              Export Backup
            </wa-button>
          </div>

          <div class="settings__section">
            <h3 class="settings__section-title">Rescan Blockchain</h3>
            <p class="settings__description">Re-scan the blockchain for transactions</p>
            <wa-button size="small" variant="neutral">Rescan</wa-button>
          </div>
        </wa-tab-panel>

        <wa-tab-panel name="connection">
          <div class="settings__section">
            <h3 class="settings__section-title">Connection Type</h3>
            <wa-select value="full-node" style="max-width: 300px;">
              <wa-option value="full-node">Full Node</wa-option>
              <wa-option value="spv">SPV (Light)</wa-option>
              <wa-option value="custom">Custom RPC</wa-option>
            </wa-select>
          </div>

          <div class="settings__section">
            <h3 class="settings__section-title">Network</h3>
            <wa-select value="main" style="max-width: 300px;">
              <wa-option value="main">Mainnet</wa-option>
              <wa-option value="testnet">Testnet</wa-option>
              <wa-option value="regtest">Regtest</wa-option>
              <wa-option value="simnet">Simnet</wa-option>
            </wa-select>
          </div>
        </wa-tab-panel>

        <wa-tab-panel name="advanced">
          <div class="settings__section">
            <h3 class="settings__section-title">API Key</h3>
            <p class="settings__description">Node API authentication key</p>
            <wa-input type="password" style="max-width: 400px;">
              <input slot="input" type="password" />
            </wa-input>
          </div>

          <div class="settings__section">
            <h3 class="settings__section-title">Analytics</h3>
            <wa-checkbox checked>Share anonymous usage data to improve Bob</wa-checkbox>
          </div>

          <div class="settings__section">
            <h3 class="settings__section-title">Developer Options</h3>
            <wa-button size="small" variant="neutral">
              <wa-icon slot="prefix" name="fa-solid fa-bug"></wa-icon>
              Open Debug Console
            </wa-button>
          </div>
        </wa-tab-panel>
      </wa-tab-group>
    </div>
  `,
    styles: [`
    .settings { }

    .settings__section { padding: 1.5rem 0; border-bottom: 1px solid var(--wa-color-neutral-200, #e5e7eb); }
    .settings__section:last-child { border-bottom: none; }

    .settings__section-title { margin: 0 0 0.5rem 0; font-size: 1rem; font-weight: 600; }

    .settings__description { margin: 0 0 0.75rem 0; font-size: 0.875rem; color: var(--wa-color-neutral-600, #4b5563); }
  `]
})
export class SettingsPage {
  private fileDialog = inject(FileDialogService);
  private clipboard = inject(ClipboardService);

  locale = 'en-US';
  msg = '';
  pickedPath = '';

  saveLocale() {
    // TODO: connect to Setting.setLocale via IPC when available
    this.msg = `Saved locale: ${this.locale}`;
    setTimeout(() => (this.msg = ''), 1500);
  }

  async copyLocale() {
    await this.clipboard.copyText(this.locale);
    this.msg = 'Locale copied to clipboard';
    setTimeout(() => (this.msg = ''), 1500);
  }

  async pickDir() {
    const res = await this.fileDialog.pickDirectory();
    this.pickedPath = res?.path || res?.name || '';
  }

  async pickFile() {
    const res = await this.fileDialog.openFile({ multiple: false, accept: ['application/json'] });
    this.pickedPath = res?.[0]?.name || '';
  }

  async saveFile() {
    const data = new Blob([JSON.stringify({ locale: this.locale }, null, 2)], { type: 'application/json' });
    const file = await this.fileDialog.saveFile({ suggestedName: 'settings.json', blob: data });
    this.pickedPath = file?.name || '';
  }
}
