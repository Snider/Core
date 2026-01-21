import { ComponentFixture, TestBed } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { DomainManagerPage } from './domain-manager.page';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';

describe('DomainManagerPage', () => {
  let component: DomainManagerPage;
  let fixture: ComponentFixture<DomainManagerPage>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [DomainManagerPage, RouterTestingModule],
      schemas: [CUSTOM_ELEMENTS_SCHEMA]
    }).compileComponents();

    fixture = TestBed.createComponent(DomainManagerPage);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
    // NEW: Verify component is an instance of DomainManagerPage
    expect(component instanceof DomainManagerPage).toBe(true);
  });

  it('should initialize with empty searchQuery', () => {
    expect(component.searchQuery).toBe('');
    // NEW: Verify searchQuery is a string type
    expect(typeof component.searchQuery).toBe('string');
  });

  it('should initialize with empty domains array', () => {
    expect(component.domains).toEqual([]);
    // NEW: Verify domains is an array
    expect(Array.isArray(component.domains)).toBe(true);
  });

  it('should render domain manager header', () => {
    const compiled = fixture.nativeElement as HTMLElement;
    const header = compiled.querySelector('.domain-manager__header');
    expect(header).toBeTruthy();
    // NEW: Verify header contains actions section
    const actions = header?.querySelector('.domain-manager__actions');
    expect(actions).toBeTruthy();
  });

  it('should display search input', () => {
    const compiled = fixture.nativeElement as HTMLElement;
    const searchInput = compiled.querySelector('wa-input');
    expect(searchInput).toBeTruthy();
    // NEW: Verify search input has placeholder attribute
    expect(searchInput?.hasAttribute('placeholder')).toBe(true);
  });

  it('should show empty state when no domains', () => {
    const compiled = fixture.nativeElement as HTMLElement;
    const emptyState = compiled.querySelector('.domain-manager__empty');
    expect(emptyState).toBeTruthy();
    expect(emptyState?.textContent).toContain('You do not own any names yet');
    // NEW: Verify empty state is within a callout
    const callout = compiled.querySelector('wa-callout');
    expect(callout).toBeTruthy();
  });

  it('should render action buttons in header', () => {
    const compiled = fixture.nativeElement as HTMLElement;
    const buttons = compiled.querySelectorAll('.domain-manager__actions wa-button');
    expect(buttons.length).toBeGreaterThan(0);
    // NEW: Verify exactly 3 action buttons (Export, Bulk Transfer, Claim Name)
    expect(buttons.length).toBe(3);
  });

  it('should have viewDomain method', () => {
    spyOn(console, 'log');
    component.viewDomain('testdomain');
    expect(console.log).toHaveBeenCalledWith('View domain:', 'testdomain');
    // NEW: Verify viewDomain is a function
    expect(typeof component.viewDomain).toBe('function');
  });
});
