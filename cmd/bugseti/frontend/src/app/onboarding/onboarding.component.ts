import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'app-onboarding',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="onboarding">
      <div class="onboarding-content">
        <!-- Step 1: Welcome -->
        <div class="step" *ngIf="step === 1">
          <div class="step-icon">B</div>
          <h1>Welcome to BugSETI</h1>
          <p class="subtitle">Distributed Bug Fixing - like SETI&#64;home but for code</p>

          <div class="feature-list">
            <div class="feature">
              <span class="feature-icon">[1]</span>
              <div>
                <strong>Find Issues</strong>
                <p>We pull beginner-friendly issues from OSS projects you care about.</p>
              </div>
            </div>
            <div class="feature">
              <span class="feature-icon">[2]</span>
              <div>
                <strong>Get Context</strong>
                <p>AI prepares relevant context to help you understand each issue.</p>
              </div>
            </div>
            <div class="feature">
              <span class="feature-icon">[3]</span>
              <div>
                <strong>Submit PRs</strong>
                <p>Fix bugs and submit PRs with minimal friction.</p>
              </div>
            </div>
          </div>

          <button class="btn btn--primary btn--lg" (click)="nextStep()">Get Started</button>
        </div>

        <!-- Step 2: GitHub Auth -->
        <div class="step" *ngIf="step === 2">
          <h2>Connect GitHub</h2>
          <p>BugSETI uses the GitHub CLI (gh) to interact with repositories.</p>

          <div class="auth-status" [class.auth-success]="ghAuthenticated">
            <span class="status-icon">{{ ghAuthenticated ? '[OK]' : '[!]' }}</span>
            <span>{{ ghAuthenticated ? 'GitHub CLI authenticated' : 'GitHub CLI not detected' }}</span>
          </div>

          <div class="auth-instructions" *ngIf="!ghAuthenticated">
            <p>To authenticate with GitHub CLI, run:</p>
            <code>gh auth login</code>
            <p class="note">After authenticating, click "Check Again".</p>
          </div>

          <div class="step-actions">
            <button class="btn btn--secondary" (click)="checkGhAuth()">Check Again</button>
            <button class="btn btn--primary" (click)="nextStep()" [disabled]="!ghAuthenticated">Continue</button>
          </div>
        </div>

        <!-- Step 3: Select Repos -->
        <div class="step" *ngIf="step === 3">
          <h2>Choose Repositories</h2>
          <p>Add repositories you want to contribute to.</p>

          <div class="repo-input">
            <input type="text" class="form-input" [(ngModel)]="newRepo"
                   placeholder="owner/repo (e.g., facebook/react)">
            <button class="btn btn--secondary" (click)="addRepo()" [disabled]="!newRepo">Add</button>
          </div>

          <div class="selected-repos" *ngIf="selectedRepos.length">
            <h3>Selected Repositories</h3>
            <div class="repo-chip" *ngFor="let repo of selectedRepos; let i = index">
              {{ repo }}
              <button class="repo-remove" (click)="removeRepo(i)">x</button>
            </div>
          </div>

          <div class="suggested-repos">
            <h3>Suggested Repositories</h3>
            <div class="suggested-list">
              <button class="suggestion" *ngFor="let repo of suggestedRepos" (click)="addSuggested(repo)">
                {{ repo }}
              </button>
            </div>
          </div>

          <div class="step-actions">
            <button class="btn btn--secondary" (click)="prevStep()">Back</button>
            <button class="btn btn--primary" (click)="nextStep()" [disabled]="selectedRepos.length === 0">Continue</button>
          </div>
        </div>

        <!-- Step 4: Complete -->
        <div class="step" *ngIf="step === 4">
          <div class="complete-icon">[OK]</div>
          <h2>You're All Set!</h2>
          <p>BugSETI is ready to help you contribute to open source.</p>

          <div class="summary">
            <p><strong>{{ selectedRepos.length }}</strong> repositories selected</p>
            <p>Looking for issues with these labels:</p>
            <div class="label-list">
              <span class="badge badge--primary">good first issue</span>
              <span class="badge badge--primary">help wanted</span>
              <span class="badge badge--primary">beginner-friendly</span>
            </div>
          </div>

          <button class="btn btn--success btn--lg" (click)="complete()">Start Finding Issues</button>
        </div>
      </div>

      <div class="step-indicators">
        <span class="indicator" [class.active]="step >= 1" [class.current]="step === 1"></span>
        <span class="indicator" [class.active]="step >= 2" [class.current]="step === 2"></span>
        <span class="indicator" [class.active]="step >= 3" [class.current]="step === 3"></span>
        <span class="indicator" [class.active]="step >= 4" [class.current]="step === 4"></span>
      </div>
    </div>
  `,
  styles: [`
    .onboarding {
      display: flex;
      flex-direction: column;
      height: 100%;
      background-color: var(--bg-primary);
    }

    .onboarding-content {
      flex: 1;
      display: flex;
      align-items: center;
      justify-content: center;
      padding: var(--spacing-xl);
    }

    .step {
      max-width: 500px;
      text-align: center;
    }

    .step-icon, .complete-icon {
      width: 80px;
      height: 80px;
      display: flex;
      align-items: center;
      justify-content: center;
      margin: 0 auto var(--spacing-lg);
      background: linear-gradient(135deg, var(--accent-primary), var(--accent-success));
      border-radius: var(--radius-lg);
      font-size: 32px;
      font-weight: bold;
      color: white;
    }

    .complete-icon {
      background: var(--accent-success);
    }

    h1 {
      font-size: 28px;
      margin-bottom: var(--spacing-sm);
    }

    h2 {
      font-size: 24px;
      margin-bottom: var(--spacing-sm);
    }

    .subtitle {
      color: var(--text-secondary);
      margin-bottom: var(--spacing-xl);
    }

    .feature-list {
      text-align: left;
      margin-bottom: var(--spacing-xl);
    }

    .feature {
      display: flex;
      gap: var(--spacing-md);
      margin-bottom: var(--spacing-md);
      padding: var(--spacing-md);
      background-color: var(--bg-secondary);
      border-radius: var(--radius-md);
    }

    .feature-icon {
      font-family: var(--font-mono);
      color: var(--accent-primary);
      font-weight: bold;
    }

    .feature strong {
      display: block;
      margin-bottom: var(--spacing-xs);
    }

    .feature p {
      color: var(--text-secondary);
      font-size: 13px;
      margin: 0;
    }

    .auth-status {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: var(--spacing-sm);
      padding: var(--spacing-md);
      background-color: var(--bg-tertiary);
      border-radius: var(--radius-md);
      margin: var(--spacing-lg) 0;
    }

    .auth-status.auth-success {
      background-color: rgba(63, 185, 80, 0.15);
      color: var(--accent-success);
    }

    .status-icon {
      font-family: var(--font-mono);
      font-weight: bold;
    }

    .auth-instructions {
      text-align: left;
      padding: var(--spacing-md);
      background-color: var(--bg-secondary);
      border-radius: var(--radius-md);
    }

    .auth-instructions code {
      display: block;
      margin: var(--spacing-md) 0;
      padding: var(--spacing-md);
      background-color: var(--bg-tertiary);
    }

    .auth-instructions .note {
      color: var(--text-muted);
      font-size: 13px;
      margin: 0;
    }

    .step-actions {
      display: flex;
      gap: var(--spacing-md);
      justify-content: center;
      margin-top: var(--spacing-xl);
    }

    .repo-input {
      display: flex;
      gap: var(--spacing-sm);
      margin-bottom: var(--spacing-lg);
    }

    .repo-input .form-input {
      flex: 1;
    }

    .selected-repos, .suggested-repos {
      text-align: left;
      margin-bottom: var(--spacing-lg);
    }

    .selected-repos h3, .suggested-repos h3 {
      font-size: 12px;
      text-transform: uppercase;
      color: var(--text-muted);
      margin-bottom: var(--spacing-sm);
    }

    .repo-chip {
      display: inline-flex;
      align-items: center;
      gap: var(--spacing-xs);
      padding: var(--spacing-xs) var(--spacing-sm);
      background-color: var(--bg-secondary);
      border-radius: var(--radius-md);
      margin-right: var(--spacing-xs);
      margin-bottom: var(--spacing-xs);
    }

    .repo-remove {
      background: none;
      border: none;
      color: var(--text-muted);
      cursor: pointer;
      padding: 0;
    }

    .suggested-list {
      display: flex;
      flex-wrap: wrap;
      gap: var(--spacing-xs);
    }

    .suggestion {
      padding: var(--spacing-xs) var(--spacing-sm);
      background-color: var(--bg-tertiary);
      border: 1px solid var(--border-color);
      border-radius: var(--radius-md);
      color: var(--text-secondary);
      cursor: pointer;
      font-size: 13px;
    }

    .suggestion:hover {
      background-color: var(--bg-secondary);
      border-color: var(--accent-primary);
    }

    .summary {
      padding: var(--spacing-lg);
      background-color: var(--bg-secondary);
      border-radius: var(--radius-md);
      margin-bottom: var(--spacing-xl);
    }

    .summary p {
      margin-bottom: var(--spacing-sm);
    }

    .label-list {
      display: flex;
      gap: var(--spacing-xs);
      justify-content: center;
      flex-wrap: wrap;
    }

    .step-indicators {
      display: flex;
      justify-content: center;
      gap: var(--spacing-sm);
      padding: var(--spacing-lg);
    }

    .indicator {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      background-color: var(--border-color);
    }

    .indicator.active {
      background-color: var(--accent-primary);
    }

    .indicator.current {
      width: 24px;
      border-radius: 4px;
    }

    .btn--lg {
      padding: var(--spacing-md) var(--spacing-xl);
      font-size: 16px;
    }
  `]
})
export class OnboardingComponent {
  step = 1;
  ghAuthenticated = false;
  newRepo = '';
  selectedRepos: string[] = [];
  suggestedRepos = [
    'facebook/react',
    'microsoft/vscode',
    'golang/go',
    'kubernetes/kubernetes',
    'rust-lang/rust',
    'angular/angular',
    'nodejs/node',
    'python/cpython'
  ];

  ngOnInit() {
    this.checkGhAuth();
  }

  nextStep() {
    if (this.step < 4) {
      this.step++;
    }
  }

  prevStep() {
    if (this.step > 1) {
      this.step--;
    }
  }

  async checkGhAuth() {
    try {
      // Check if gh CLI is authenticated
      // In a real implementation, this would call the backend
      this.ghAuthenticated = true; // Assume authenticated for demo
    } catch (err) {
      this.ghAuthenticated = false;
    }
  }

  addRepo() {
    if (this.newRepo && !this.selectedRepos.includes(this.newRepo)) {
      this.selectedRepos.push(this.newRepo);
      this.newRepo = '';
    }
  }

  removeRepo(index: number) {
    this.selectedRepos.splice(index, 1);
  }

  addSuggested(repo: string) {
    if (!this.selectedRepos.includes(repo)) {
      this.selectedRepos.push(repo);
    }
  }

  async complete() {
    try {
      // Save repos to config
      if ((window as any).go?.main?.ConfigService?.SetConfig) {
        const config = await (window as any).go.main.ConfigService.GetConfig() || {};
        config.watchedRepos = this.selectedRepos;
        await (window as any).go.main.ConfigService.SetConfig(config);
      }

      // Mark onboarding as complete
      if ((window as any).go?.main?.TrayService?.CompleteOnboarding) {
        await (window as any).go.main.TrayService.CompleteOnboarding();
      }

      // Close onboarding window and start fetching
      if ((window as any).wails?.Window) {
        (window as any).wails.Window.GetByName('onboarding').then((w: any) => w.Hide());
      }

      // Start fetching
      if ((window as any).go?.main?.TrayService?.StartFetching) {
        await (window as any).go.main.TrayService.StartFetching();
      }
    } catch (err) {
      console.error('Failed to complete onboarding:', err);
    }
  }
}
