import 'zone.js/testing';
import { TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { TranslateService, TranslateLoader, TranslateModule } from '@ngx-translate/core';
import { Observable, of } from 'rxjs';

// Provide TranslateService mock globally for tests to avoid NG0201 in standalone components
(() => {
  class FakeTranslateLoader implements TranslateLoader {
    getTranslation(lang: string): Observable<any> { return of({}); }
  }

  const translateServiceMock: Partial<TranslateService> = {
    use: (() => ({ toPromise: async () => undefined })) as any,
    instant: ((key: string) => key) as any,
    get: (((key: any) => ({ subscribe: (fn: any) => fn(key) })) as any),
    onLangChange: { subscribe: () => ({ unsubscribe() {} }) } as any,
  } as Partial<TranslateService>;

  // Patch TestBed.configureTestingModule to always include Translate support
  const originalConfigure = TestBed.configureTestingModule.bind(TestBed);
  (TestBed as any).configureTestingModule = (meta: any = {}) => {
    // Ensure providers include TranslateService mock if not already provided
    const providers = meta.providers ?? [];
    const hasTranslateProvider = providers.some((p: any) => p && (p.provide === TranslateService));
    meta.providers = hasTranslateProvider ? providers : [...providers, { provide: TranslateService, useValue: translateServiceMock }];

    // Ensure imports include TranslateModule.forRoot with a fake loader (brings internal _TranslateService)
    const imports = meta.imports ?? [];
    const hasTranslateModule = imports.some((imp: any) => imp && (imp === TranslateModule || (imp.ngModule && imp.ngModule === TranslateModule)));
    if (!hasTranslateModule) {
      imports.push(TranslateModule.forRoot({ loader: { provide: TranslateLoader, useClass: FakeTranslateLoader } }));
    }
    meta.imports = imports;

    return originalConfigure(meta);
  };
})();
