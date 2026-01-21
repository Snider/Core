import { NgModule, CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import {  } from "@awesome.me/webawesome/dist/webawesome.loader.js"
// This module enables Angular to accept unknown custom elements (Web Awesome components)
// without throwing template parse errors.
@NgModule({
  schemas: [CUSTOM_ELEMENTS_SCHEMA],
})
export class CustomElementsModule {}
