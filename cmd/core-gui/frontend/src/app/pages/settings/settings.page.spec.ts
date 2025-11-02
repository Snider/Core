import { ComponentFixture, TestBed } from '@angular/core/testing';
import { SettingsPage } from './settings.page';
import { FileDialogService } from '../../services/file-dialog.service';
import { ClipboardService } from '../../services/clipboard.service';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';

describe('SettingsPage', () => {
  let component: SettingsPage;
  let fixture: ComponentFixture<SettingsPage>;
  let fileDialogService: jasmine.SpyObj<FileDialogService>;
  let clipboardService: jasmine.SpyObj<ClipboardService>;

  beforeEach(async () => {
    const fileDialogSpy = jasmine.createSpyObj('FileDialogService', ['pickDirectory', 'openFile', 'saveFile']);
    const clipboardSpy = jasmine.createSpyObj('ClipboardService', ['copyText']);

    await TestBed.configureTestingModule({
      imports: [SettingsPage],
      providers: [
        { provide: FileDialogService, useValue: fileDialogSpy },
        { provide: ClipboardService, useValue: clipboardSpy }
      ],
      schemas: [CUSTOM_ELEMENTS_SCHEMA]
    }).compileComponents();

    fileDialogService = TestBed.inject(FileDialogService) as jasmine.SpyObj<FileDialogService>;
    clipboardService = TestBed.inject(ClipboardService) as jasmine.SpyObj<ClipboardService>;

    fixture = TestBed.createComponent(SettingsPage);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
    // NEW: Verify component is an instance of SettingsPage
    expect(component instanceof SettingsPage).toBe(true);
  });

  it('should initialize with default locale', () => {
    expect(component.locale).toBe('en-US');
    // NEW: Verify msg is initialized as empty string
    expect(component.msg).toBe('');
  });

  it('should render settings tabs', () => {
    const compiled = fixture.nativeElement as HTMLElement;
    const tabGroup = compiled.querySelector('wa-tab-group');
    expect(tabGroup).toBeTruthy();
    // NEW: Verify tab group contains tab panels
    const tabPanels = compiled.querySelectorAll('wa-tab-panel');
    expect(tabPanels.length).toBeGreaterThan(0);
  });

  it('should render four tab panels', () => {
    const compiled = fixture.nativeElement as HTMLElement;
    const tabs = compiled.querySelectorAll('wa-tab');
    expect(tabs.length).toBe(4);
    // NEW: Verify corresponding tab panels exist
    const tabPanels = compiled.querySelectorAll('wa-tab-panel');
    expect(tabPanels.length).toBe(4);
  });

  it('should call pickDirectory when change directory button clicked', async () => {
    fileDialogService.pickDirectory.and.returnValue(Promise.resolve({ path: '/test/path' }));
    await component.pickDir();
    expect(fileDialogService.pickDirectory).toHaveBeenCalled();
    // NEW: Verify pickedPath is updated
    expect(component.pickedPath).toBe('/test/path');
  });

  it('should call saveFile when export backup button clicked', async () => {
    fileDialogService.saveFile.and.returnValue(Promise.resolve({ name: 'settings.json' } as any));
    await component.saveFile();
    expect(fileDialogService.saveFile).toHaveBeenCalled();
    // NEW: Verify pickedPath is updated with filename
    expect(component.pickedPath).toBe('settings.json');
  });

  it('should update message after saving locale', () => {
    component.saveLocale();
    expect(component.msg).toContain('Saved locale');
    // NEW: Verify message includes the locale value
    expect(component.msg).toContain('en-US');
  });

  it('should copy locale to clipboard', async () => {
    clipboardService.copyText.and.returnValue(Promise.resolve(true));
    await component.copyLocale();
    expect(clipboardService.copyText).toHaveBeenCalledWith(component.locale);
    expect(component.msg).toContain('copied to clipboard');
    // NEW: Verify clipboard was called exactly once
    expect(clipboardService.copyText).toHaveBeenCalledTimes(1);
  });
});
