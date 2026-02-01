import { DoBootstrap, Injector, NgModule, provideBrowserGlobalErrorListeners } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { createCustomElement } from '@angular/elements';

import { App } from './app';

@NgModule({
  imports: [
    BrowserModule,
    App
  ],
  providers: [
    provideBrowserGlobalErrorListeners()
  ]
})
export class AppModule implements DoBootstrap {
  constructor(private injector: Injector) {
    const el = createCustomElement(App, { injector });
    customElements.define('core-element-template', el);
  }

  ngDoBootstrap() {}
}
