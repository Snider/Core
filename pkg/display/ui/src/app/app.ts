import { Component, signal } from '@angular/core';

@Component({
  selector: 'core-display',
  templateUrl: './app.html',
  standalone: true
})
export class App {
  protected readonly title = signal('display');
}
