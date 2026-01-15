import { Component, signal } from '@angular/core';

@Component({
  selector: 'help-element',
  templateUrl: './app.html',
  standalone: true
})
export class App {
  protected readonly title = signal('help');
}
