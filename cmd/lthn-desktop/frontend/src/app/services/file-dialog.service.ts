import { Injectable } from '@angular/core';

// WAILS3 INTEGRATION:
// This service currently uses web-standard File System Access API.
// For Wails3, replace with Go service methods calling:
// - application.OpenFileDialog().PromptForSingleSelection()
// - application.SaveFileDialog().SetFilename().PromptForSelection()
// See WAILS3_INTEGRATION.md for complete examples.

export interface OpenFileOptions {
  multiple?: boolean;
  accept?: string[]; // e.g., ["application/json", "text/plain"]
}

export interface SaveFileOptions {
  suggestedName?: string;
  types?: { description?: string; accept?: Record<string, string[]> }[];
  blob: Blob;
}

@Injectable({ providedIn: 'root' })
export class FileDialogService {
  // Directory picker using File System Access API when available
  async pickDirectory(): Promise<any | null> {
    const nav: any = window.navigator;
    if ((window as any).showDirectoryPicker) {
      try {
        // @ts-ignore
        const handle: any = await (window as any).showDirectoryPicker({ mode: 'readwrite' });
        return handle;
      } catch (e) {
        return null;
      }
    }
    // Fallback: not supported in all browsers; inform the user
    alert('Directory picker is not supported in this browser.');
    return null;
  }

  // Open file(s) with <input type="file"> fallback if FS Access API not used
  async openFile(opts: OpenFileOptions = {}): Promise<File[] | null> {
    // Always supported fallback
    return new Promise<File[] | null>((resolve) => {
      const input = document.createElement('input');
      input.type = 'file';
      input.multiple = !!opts.multiple;
      if (opts.accept && opts.accept.length) {
        input.accept = opts.accept.join(',');
      }
      input.onchange = () => {
        const files = input.files ? Array.from(input.files) : null;
        resolve(files);
      };
      input.click();
    });
  }

  // Save file using File System Access API if available, otherwise trigger a download
  async saveFile(opts: SaveFileOptions): Promise<File | null> {
    if ((window as any).showSaveFilePicker) {
      try {
        // @ts-ignore
        const handle = await (window as any).showSaveFilePicker({
          suggestedName: opts.suggestedName,
          types: opts.types
        });
        const writable = await handle.createWritable();
        await writable.write(opts.blob);
        await writable.close();
        return { name: handle.name } as any;
      } catch (e) {
        return null;
      }
    }

    // Fallback: download
    const url = URL.createObjectURL(opts.blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = opts.suggestedName || 'download';
    a.click();
    URL.revokeObjectURL(url);
    return { name: opts.suggestedName || 'download' } as any;
  }
}
