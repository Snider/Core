import { Injectable } from '@angular/core';

@Injectable({ providedIn: 'root' })
export class HardwareWalletService {
  // Placeholder for WebHID/WebUSB detection
  get isWebHIDAvailable() {
    return 'hid' in navigator;
  }

  get isWebUSBAvailable() {
    return 'usb' in navigator;
  }

  async connectLedger(): Promise<void> {
    // In a real implementation, prompt for a specific HID/USB device
    // and establish transport (e.g., via @ledgerhq/hw-transport-webhid).
    // This is a stub to document the integration point.
    throw new Error('HardwareWalletService.connectLedger is not implemented in the web build.');
  }

  async getAppVersion(): Promise<string> {
    // Should query the connected device/app for version information
    throw new Error('HardwareWalletService.getAppVersion is not implemented in the web build.');
  }

  async disconnect(): Promise<void> {
    // Close transport/session to the device
    throw new Error('HardwareWalletService.disconnect is not implemented in the web build.');
  }
}
