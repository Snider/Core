import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';

interface Issue {
  id: string;
  number: number;
  repo: string;
  title: string;
  body: string;
  url: string;
  labels: string[];
  author: string;
  context?: IssueContext;
}

interface IssueContext {
  summary: string;
  relevantFiles: string[];
  suggestedFix: string;
  complexity: string;
  estimatedTime: string;
}

@Component({
  selector: 'app-workbench',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="workbench">
      <header class="workbench-header">
        <h1>BugSETI Workbench</h1>
        <div class="header-actions">
          <button class="btn btn--secondary" (click)="skipIssue()">Skip</button>
          <button class="btn btn--success" (click)="submitPR()" [disabled]="!canSubmit">Submit PR</button>
        </div>
      </header>

      <div class="workbench-content" *ngIf="currentIssue">
        <aside class="issue-panel">
          <div class="card">
            <div class="card__header">
              <h2 class="card__title">Issue #{{ currentIssue.number }}</h2>
              <a [href]="currentIssue.url" target="_blank" class="btn btn--secondary btn--sm">View on GitHub</a>
            </div>

            <h3>{{ currentIssue.title }}</h3>

            <div class="labels">
              <span class="badge badge--primary" *ngFor="let label of currentIssue.labels">
                {{ label }}
              </span>
            </div>

            <div class="issue-meta">
              <span>{{ currentIssue.repo }}</span>
              <span>by {{ currentIssue.author }}</span>
            </div>

            <div class="issue-body">
              <pre>{{ currentIssue.body }}</pre>
            </div>
          </div>

          <div class="card" *ngIf="currentIssue.context">
            <div class="card__header">
              <h2 class="card__title">AI Context</h2>
              <span class="badge" [ngClass]="{
                'badge--success': currentIssue.context.complexity === 'easy',
                'badge--warning': currentIssue.context.complexity === 'medium',
                'badge--danger': currentIssue.context.complexity === 'hard'
              }">
                {{ currentIssue.context.complexity }}
              </span>
            </div>

            <p class="context-summary">{{ currentIssue.context.summary }}</p>

            <div class="context-section" *ngIf="currentIssue.context.relevantFiles?.length">
              <h4>Relevant Files</h4>
              <ul class="file-list">
                <li *ngFor="let file of currentIssue.context.relevantFiles">
                  <code>{{ file }}</code>
                </li>
              </ul>
            </div>

            <div class="context-section" *ngIf="currentIssue.context.suggestedFix">
              <h4>Suggested Approach</h4>
              <p>{{ currentIssue.context.suggestedFix }}</p>
            </div>

            <div class="context-meta">
              <span>Est. time: {{ currentIssue.context.estimatedTime || 'Unknown' }}</span>
            </div>
          </div>
        </aside>

        <main class="editor-panel">
          <div class="card">
            <div class="card__header">
              <h2 class="card__title">PR Details</h2>
            </div>

            <div class="form-group">
              <label class="form-label">PR Title</label>
              <input type="text" class="form-input" [(ngModel)]="prTitle"
                     [placeholder]="'Fix #' + currentIssue.number + ': ' + currentIssue.title">
            </div>

            <div class="form-group">
              <label class="form-label">PR Description</label>
              <textarea class="form-textarea" [(ngModel)]="prBody" rows="8"
                        placeholder="Describe your changes..."></textarea>
            </div>

            <div class="form-group">
              <label class="form-label">Branch Name</label>
              <input type="text" class="form-input" [(ngModel)]="branchName"
                     [placeholder]="'bugseti/issue-' + currentIssue.number">
            </div>

            <div class="form-group">
              <label class="form-label">Commit Message</label>
              <textarea class="form-textarea" [(ngModel)]="commitMessage" rows="3"
                        [placeholder]="'fix: resolve issue #' + currentIssue.number"></textarea>
            </div>
          </div>
        </main>
      </div>

      <div class="empty-state" *ngIf="!currentIssue">
        <h2>No Issue Selected</h2>
        <p>Get an issue from the queue to start working.</p>
        <button class="btn btn--primary" (click)="nextIssue()">Get Next Issue</button>
      </div>
    </div>
  `,
  styles: [`
    .workbench {
      display: flex;
      flex-direction: column;
      height: 100%;
      background-color: var(--bg-secondary);
    }

    .workbench-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: var(--spacing-md) var(--spacing-lg);
      background-color: var(--bg-primary);
      border-bottom: 1px solid var(--border-color);
    }

    .workbench-header h1 {
      font-size: 18px;
      margin: 0;
    }

    .header-actions {
      display: flex;
      gap: var(--spacing-sm);
    }

    .workbench-content {
      display: grid;
      grid-template-columns: 400px 1fr;
      flex: 1;
      overflow: hidden;
    }

    .issue-panel {
      display: flex;
      flex-direction: column;
      gap: var(--spacing-md);
      padding: var(--spacing-md);
      overflow-y: auto;
      border-right: 1px solid var(--border-color);
    }

    .editor-panel {
      padding: var(--spacing-md);
      overflow-y: auto;
    }

    .labels {
      display: flex;
      flex-wrap: wrap;
      gap: var(--spacing-xs);
      margin: var(--spacing-sm) 0;
    }

    .issue-meta {
      display: flex;
      gap: var(--spacing-md);
      font-size: 12px;
      color: var(--text-muted);
      margin-bottom: var(--spacing-md);
    }

    .issue-body {
      padding: var(--spacing-md);
      background-color: var(--bg-tertiary);
      border-radius: var(--radius-md);
      max-height: 200px;
      overflow-y: auto;
    }

    .issue-body pre {
      white-space: pre-wrap;
      word-wrap: break-word;
      font-size: 13px;
      line-height: 1.5;
      margin: 0;
    }

    .context-summary {
      color: var(--text-secondary);
      margin-bottom: var(--spacing-md);
    }

    .context-section {
      margin-bottom: var(--spacing-md);
    }

    .context-section h4 {
      font-size: 12px;
      text-transform: uppercase;
      color: var(--text-muted);
      margin-bottom: var(--spacing-xs);
    }

    .file-list {
      list-style: none;
      padding: 0;
      margin: 0;
    }

    .file-list li {
      padding: var(--spacing-xs) 0;
    }

    .context-meta {
      font-size: 12px;
      color: var(--text-muted);
    }

    .empty-state {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      flex: 1;
      text-align: center;
    }

    .empty-state h2 {
      color: var(--text-secondary);
    }

    .empty-state p {
      color: var(--text-muted);
    }
  `]
})
export class WorkbenchComponent implements OnInit {
  currentIssue: Issue | null = null;
  prTitle = '';
  prBody = '';
  branchName = '';
  commitMessage = '';

  get canSubmit(): boolean {
    return !!this.currentIssue && !!this.prTitle;
  }

  ngOnInit() {
    this.loadCurrentIssue();
  }

  async loadCurrentIssue() {
    try {
      if ((window as any).go?.main?.TrayService?.GetCurrentIssue) {
        this.currentIssue = await (window as any).go.main.TrayService.GetCurrentIssue();
        if (this.currentIssue) {
          this.initDefaults();
        }
      }
    } catch (err) {
      console.error('Failed to load current issue:', err);
    }
  }

  initDefaults() {
    if (!this.currentIssue) return;

    this.prTitle = `Fix #${this.currentIssue.number}: ${this.currentIssue.title}`;
    this.branchName = `bugseti/issue-${this.currentIssue.number}`;
    this.commitMessage = `fix: resolve issue #${this.currentIssue.number}\n\n${this.currentIssue.title}`;
  }

  async nextIssue() {
    try {
      if ((window as any).go?.main?.TrayService?.NextIssue) {
        this.currentIssue = await (window as any).go.main.TrayService.NextIssue();
        if (this.currentIssue) {
          this.initDefaults();
        }
      }
    } catch (err) {
      console.error('Failed to get next issue:', err);
    }
  }

  async skipIssue() {
    try {
      if ((window as any).go?.main?.TrayService?.SkipIssue) {
        await (window as any).go.main.TrayService.SkipIssue();
        this.currentIssue = null;
        this.prTitle = '';
        this.prBody = '';
        this.branchName = '';
        this.commitMessage = '';
      }
    } catch (err) {
      console.error('Failed to skip issue:', err);
    }
  }

  async submitPR() {
    if (!this.currentIssue || !this.canSubmit) return;

    try {
      if ((window as any).go?.main?.SubmitService?.Submit) {
        const result = await (window as any).go.main.SubmitService.Submit({
          issue: this.currentIssue,
          title: this.prTitle,
          body: this.prBody,
          branch: this.branchName,
          commitMsg: this.commitMessage
        });

        if (result.success) {
          alert(`PR submitted successfully!\n\n${result.prUrl}`);
          this.currentIssue = null;
        } else {
          alert(`Failed to submit PR: ${result.error}`);
        }
      }
    } catch (err) {
      console.error('Failed to submit PR:', err);
      alert('Failed to submit PR. Check console for details.');
    }
  }
}
