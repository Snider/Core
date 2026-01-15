import { Component, signal } from '@angular/core';

@Component({
  selector: 'i18n-element',
  templateUrl: './app.html',
  standalone: true
})
export class App {
  protected readonly title = signal('i18n-element');
}
