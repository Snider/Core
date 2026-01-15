import { Component, signal } from '@angular/core';

@Component({
  selector: 'core-element-template',
  templateUrl: './app.html',
  standalone: true
})
export class App {
  protected readonly title = signal('core-element-template');
}
