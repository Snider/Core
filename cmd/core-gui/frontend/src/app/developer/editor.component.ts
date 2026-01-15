import { Component, OnInit, OnDestroy } from '@angular/core';
import { MonacoEditorModule } from 'ngx-monaco-editor-v2';
import { FormsModule } from '@angular/forms';
import { CommonModule } from '@angular/common';
import { ActivatedRoute } from '@angular/router';
import { Events } from '@wailsio/runtime';
import * as IDE from '@lthn/ide/service';
import { DirectoryEntry } from '@lthn/ide/models';
import { SelectDirectory } from '@lthn/core/display/service';

interface TreeNode extends DirectoryEntry {
  expanded?: boolean;
  children?: TreeNode[];
  level?: number;
}

@Component({
  selector: 'dev-edit',
  standalone: true,
  imports: [MonacoEditorModule, FormsModule, CommonModule],
  template: `
    <div class="ide-container">
      <!-- Project Panel -->
      <div class="project-panel" [style.width.px]="panelWidth">
        <div class="panel-header">
          <span class="panel-title">PROJECT</span>
          <button class="panel-btn" (click)="openWorkspace()" title="Open Folder">
            <i class="fa-regular fa-folder-open"></i>
          </button>
          <button class="panel-btn" (click)="refreshTree()" title="Refresh">
            <i class="fa-regular fa-arrows-rotate"></i>
          </button>
        </div>
        @if (workspaceRoot) {
          <div class="workspace-name">{{ workspaceName }}</div>
          <div class="file-tree">
            @for (node of fileTree; track node.path) {
              <ng-container *ngTemplateOutlet="treeNodeTpl; context: { node: node }"></ng-container>
            }
          </div>
        } @else {
          <div class="no-workspace">
            <p>No folder open</p>
            <button class="open-folder-btn" (click)="openWorkspace()">
              <i class="fa-regular fa-folder-open"></i> Open Folder
            </button>
          </div>
        }
      </div>

      <!-- Resizer -->
      <div class="resizer" (mousedown)="startResize($event)"></div>

      <!-- Editor Area -->
      <div class="editor-area">
        @if (filePath) {
          <div class="editor-toolbar">
            <span class="file-path">{{ filePath }}</span>
            @if (isModified) {
              <span class="modified-indicator">●</span>
            }
          </div>
        }
        <ngx-monaco-editor
          [style.height]="filePath ? 'calc(100% - 30px)' : '100%'"
          [options]="editorOptions"
          [(ngModel)]="code"
          (ngModelChange)="onCodeChange()">
        </ngx-monaco-editor>
      </div>
    </div>

    <!-- Tree Node Template -->
    <ng-template #treeNodeTpl let-node="node">
      <div class="tree-item"
           [style.paddingLeft.px]="(node.level || 0) * 16 + 8"
           [class.is-dir]="node.isDir"
           [class.selected]="node.path === filePath"
           (click)="onNodeClick(node)">
        @if (node.isDir) {
          <i class="fa-regular" [class.fa-chevron-right]="!node.expanded" [class.fa-chevron-down]="node.expanded"></i>
          <i class="fa-regular fa-folder" [class.fa-folder-open]="node.expanded"></i>
        } @else {
          <i class="file-icon fa-regular" [class]="getFileIcon(node.name)"></i>
        }
        <span class="node-name">{{ node.name }}</span>
      </div>
      @if (node.isDir && node.expanded && node.children) {
        @for (child of node.children; track child.path) {
          <ng-container *ngTemplateOutlet="treeNodeTpl; context: { node: child }"></ng-container>
        }
      }
    </ng-template>
  `,
  styles: [`
    .ide-container {
      display: flex;
      height: 100vh;
      background: #1e1e1e;
    }
    .project-panel {
      background: #252526;
      display: flex;
      flex-direction: column;
      min-width: 150px;
      max-width: 500px;
    }
    .panel-header {
      display: flex;
      align-items: center;
      padding: 8px 12px;
      background: #333;
      border-bottom: 1px solid #444;
    }
    .panel-title {
      flex: 1;
      font-size: 11px;
      font-weight: 600;
      color: #999;
      letter-spacing: 0.5px;
    }
    .panel-btn {
      background: none;
      border: none;
      color: #888;
      cursor: pointer;
      padding: 4px 6px;
      margin-left: 4px;
      border-radius: 3px;
    }
    .panel-btn:hover {
      background: #444;
      color: #fff;
    }
    .workspace-name {
      padding: 8px 12px;
      font-size: 13px;
      font-weight: 500;
      color: #ccc;
      border-bottom: 1px solid #333;
      background: #2d2d2d;
    }
    .file-tree {
      flex: 1;
      overflow-y: auto;
      padding: 4px 0;
    }
    .tree-item {
      display: flex;
      align-items: center;
      padding: 4px 8px;
      cursor: pointer;
      color: #ccc;
      font-size: 13px;
      gap: 6px;
    }
    .tree-item:hover {
      background: #2a2d2e;
    }
    .tree-item.selected {
      background: #094771;
    }
    .tree-item i {
      font-size: 12px;
      width: 14px;
      text-align: center;
    }
    .tree-item .fa-chevron-right, .tree-item .fa-chevron-down {
      font-size: 10px;
      color: #888;
    }
    .tree-item .fa-folder { color: #dcb67a; }
    .tree-item .fa-folder-open { color: #dcb67a; }
    .node-name {
      flex: 1;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
    .no-workspace {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      height: 200px;
      color: #888;
    }
    .open-folder-btn {
      margin-top: 12px;
      padding: 8px 16px;
      background: #0e639c;
      border: none;
      color: #fff;
      border-radius: 4px;
      cursor: pointer;
      font-size: 13px;
    }
    .open-folder-btn:hover {
      background: #1177bb;
    }
    .resizer {
      width: 4px;
      cursor: col-resize;
      background: #333;
    }
    .resizer:hover {
      background: #007acc;
    }
    .editor-area {
      flex: 1;
      display: flex;
      flex-direction: column;
      min-width: 200px;
    }
    .editor-toolbar {
      height: 30px;
      background: #1e1e1e;
      color: #ccc;
      display: flex;
      align-items: center;
      padding: 0 10px;
      font-size: 12px;
      border-bottom: 1px solid #333;
    }
    .file-path {
      flex: 1;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
    .modified-indicator {
      color: #e0a000;
      margin-left: 8px;
      font-size: 16px;
    }
    /* File type icons */
    .file-icon { color: #888; }
    .file-icon.fa-file-code { color: #519aba; }
    .file-icon.fa-file-lines { color: #888; }
  `]
})
export class DeveloperEditorComponent implements OnInit, OnDestroy {
  editorOptions = { theme: 'vs-dark', language: 'typescript' };
  code: string = '';
  filePath: string = '';
  isModified: boolean = false;
  originalCode: string = '';

  workspaceRoot: string = '';
  workspaceName: string = '';
  fileTree: TreeNode[] = [];
  panelWidth: number = 250;

  private unsubscribeSave: (() => void) | null = null;
  private isResizing = false;

  constructor(private route: ActivatedRoute) {}

  async ngOnInit(): Promise<void> {
    this.unsubscribeSave = Events.On('ide:save', () => {
      this.saveFile();
    });

    this.route.queryParams.subscribe(async params => {
      if (params['file']) {
        await this.loadFile(params['file']);
      } else if (params['new']) {
        await this.newFile();
      } else if (params['workspace']) {
        await this.setWorkspace(params['workspace']);
      } else {
        this.code = '// Welcome to Core IDE\n// Open a folder to browse your project files\n// Or use File → Open to open a file';
      }
    });

    // Set up resize handlers
    document.addEventListener('mousemove', this.onResize.bind(this));
    document.addEventListener('mouseup', this.stopResize.bind(this));
  }

  ngOnDestroy(): void {
    if (this.unsubscribeSave) {
      this.unsubscribeSave();
    }
    document.removeEventListener('mousemove', this.onResize.bind(this));
    document.removeEventListener('mouseup', this.stopResize.bind(this));
  }

  async openWorkspace(): Promise<void> {
    try {
      const path = await SelectDirectory();
      if (path) {
        await this.setWorkspace(path);
      }
    } catch (error) {
      console.error('Error selecting directory:', error);
    }
  }

  async setWorkspace(path: string): Promise<void> {
    this.workspaceRoot = path;
    this.workspaceName = path.split('/').pop() || path;
    await this.refreshTree();
  }

  async refreshTree(): Promise<void> {
    if (!this.workspaceRoot) return;
    try {
      const entries = await IDE.ListDirectory(this.workspaceRoot);
      this.fileTree = this.sortEntries(entries.map(e => ({ ...e, level: 0 })));
    } catch (error) {
      console.error('Error loading directory:', error);
    }
  }

  async onNodeClick(node: TreeNode): Promise<void> {
    if (node.isDir) {
      node.expanded = !node.expanded;
      if (node.expanded && !node.children) {
        try {
          const entries = await IDE.ListDirectory(node.path);
          node.children = this.sortEntries(entries.map(e => ({
            ...e,
            level: (node.level || 0) + 1
          })));
        } catch (error) {
          console.error('Error loading directory:', error);
        }
      }
    } else {
      await this.loadFile(node.path);
    }
  }

  sortEntries(entries: TreeNode[]): TreeNode[] {
    return entries
      .filter(e => !e.name.startsWith('.')) // Hide hidden files
      .sort((a, b) => {
        if (a.isDir && !b.isDir) return -1;
        if (!a.isDir && b.isDir) return 1;
        return a.name.localeCompare(b.name);
      });
  }

  getFileIcon(filename: string): string {
    const ext = filename.split('.').pop()?.toLowerCase();
    const codeExts = ['ts', 'tsx', 'js', 'jsx', 'go', 'py', 'rs', 'java', 'c', 'cpp', 'h', 'cs', 'rb', 'php'];
    if (codeExts.includes(ext || '')) return 'fa-file-code';
    if (['md', 'txt', 'json', 'yaml', 'yml', 'toml', 'xml'].includes(ext || '')) return 'fa-file-lines';
    return 'fa-file';
  }

  startResize(event: MouseEvent): void {
    this.isResizing = true;
    event.preventDefault();
  }

  onResize(event: MouseEvent): void {
    if (!this.isResizing) return;
    this.panelWidth = Math.max(150, Math.min(500, event.clientX));
  }

  stopResize(): void {
    this.isResizing = false;
  }

  async newFile(): Promise<void> {
    try {
      const fileInfo = await IDE.NewFile('typescript');
      this.code = fileInfo.content;
      this.originalCode = fileInfo.content;
      this.filePath = '';
      this.editorOptions = { ...this.editorOptions, language: fileInfo.language };
      this.isModified = false;
    } catch (error) {
      console.error('Error creating new file:', error);
    }
  }

  async loadFile(path: string): Promise<void> {
    try {
      const fileInfo = await IDE.OpenFile(path);
      this.code = fileInfo.content;
      this.originalCode = fileInfo.content;
      this.filePath = fileInfo.path;
      this.editorOptions = { ...this.editorOptions, language: fileInfo.language };
      this.isModified = false;
    } catch (error) {
      console.error('Error loading file:', error);
      this.code = `// Error loading file: ${path}\n// ${error}`;
    }
  }

  async saveFile(): Promise<void> {
    if (!this.filePath) {
      console.log('No file path - need to implement save as dialog');
      return;
    }
    try {
      await IDE.SaveFile(this.filePath, this.code);
      this.originalCode = this.code;
      this.isModified = false;
      console.log('File saved:', this.filePath);
    } catch (error) {
      console.error('Error saving file:', error);
    }
  }

  onCodeChange(): void {
    this.isModified = this.code !== this.originalCode;
  }
}
