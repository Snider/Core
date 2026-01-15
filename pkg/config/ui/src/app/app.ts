import { Component, signal } from '@angular/core';

@Component({
  selector: 'config-element',
  templateUrl: './app.html',
  standalone: true
})
export class App {
  public readonly title = signal('config');
}
