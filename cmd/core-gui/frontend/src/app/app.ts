import { Component, OnInit, Inject, PLATFORM_ID, CUSTOM_ELEMENTS_SCHEMA, ViewChild, ElementRef } from '@angular/core';
import { CommonModule,  DOCUMENT, isPlatformBrowser } from '@angular/common';
import {RouterLink, RouterOutlet} from '@angular/router';
import { FooterComponent } from './shared/components/footer/footer.component';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import {Subscription} from 'rxjs';

@Component({
  selector: 'app-root',
  imports: [
    CommonModule,
    RouterOutlet,
    FooterComponent,
    TranslateModule,
    RouterLink
  ],
  templateUrl: './app.html',
  styleUrl: './app.css',
  standalone: true,
  schemas: [CUSTOM_ELEMENTS_SCHEMA],
})
export class App {
  @ViewChild('sidebar', { read: ElementRef, static: false }) sidebar?: ElementRef<HTMLElement>;

  sidebarOpen = false;
  userMenuOpen = false;
  currentRole = 'Developer';

  time: string = '';

  constructor(
    @Inject(DOCUMENT) private document: Document,
    @Inject(PLATFORM_ID) private platformId: object,
    private translateService: TranslateService
  ) {
    // Set default language
    this.translateService.use('en');
  }

}
