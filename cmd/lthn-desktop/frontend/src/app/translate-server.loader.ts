import { join } from 'path';
import { Observable, of } from 'rxjs';
import { TranslateLoader } from '@ngx-translate/core';
import * as fs from 'fs';

export class TranslateServerLoader implements TranslateLoader {
  constructor(private prefix: string = 'i18n', private suffix: string = '.json') {}

  public getTranslation(lang: string): Observable<any> {
    const path = join(process.cwd(), 'i18n', this.prefix, `${lang}${this.suffix}`);
    const data = JSON.parse(fs.readFileSync(path, 'utf8'));
    return of(data);
  }
}
