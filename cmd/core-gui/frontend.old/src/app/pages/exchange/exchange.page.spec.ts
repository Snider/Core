import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ExchangePage } from './exchange.page';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';

describe('ExchangePage', () => {
  let component: ExchangePage;
  let fixture: ComponentFixture<ExchangePage>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ExchangePage],
      schemas: [CUSTOM_ELEMENTS_SCHEMA]
    }).compileComponents();

    fixture = TestBed.createComponent(ExchangePage);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
    // NEW: Verify component is an instance of ExchangePage
    expect(component instanceof ExchangePage).toBe(true);
  });

  it('should render exchange container', () => {
    const compiled = fixture.nativeElement as HTMLElement;
    const exchange = compiled.querySelector('.exchange');
    expect(exchange).toBeTruthy();
    // NEW: Verify exchange container has tab group
    const tabGroup = exchange?.querySelector('wa-tab-group');
    expect(tabGroup).toBeTruthy();
  });

  it('should render tab group', () => {
    const compiled = fixture.nativeElement as HTMLElement;
    const tabGroup = compiled.querySelector('wa-tab-group');
    expect(tabGroup).toBeTruthy();
    // NEW: Verify tab group contains tabs
    const tabs = tabGroup?.querySelectorAll('wa-tab');
    expect(tabs?.length).toBeGreaterThan(0);
  });

  it('should render three tabs', () => {
    const compiled = fixture.nativeElement as HTMLElement;
    const tabs = compiled.querySelectorAll('wa-tab');
    expect(tabs.length).toBe(3);
    // NEW: Verify corresponding tab panels exist
    const tabPanels = compiled.querySelectorAll('wa-tab-panel');
    expect(tabPanels.length).toBe(3);
  });

  it('should have Listings tab', () => {
    const compiled = fixture.nativeElement as HTMLElement;
    const tabs = Array.from(compiled.querySelectorAll('wa-tab'));
    const listingsTab = tabs.find(tab => tab.textContent?.includes('Listings'));
    expect(listingsTab).toBeTruthy();
    // NEW: Verify listings tab has correct panel attribute
    expect(listingsTab?.getAttribute('panel')).toBe('listings');
  });

  it('should have Fills tab', () => {
    const compiled = fixture.nativeElement as HTMLElement;
    const tabs = Array.from(compiled.querySelectorAll('wa-tab'));
    const fillsTab = tabs.find(tab => tab.textContent?.includes('Fills'));
    expect(fillsTab).toBeTruthy();
    // NEW: Verify fills tab has correct panel attribute
    expect(fillsTab?.getAttribute('panel')).toBe('fills');
  });

  it('should have Auctions tab', () => {
    const compiled = fixture.nativeElement as HTMLElement;
    const tabs = Array.from(compiled.querySelectorAll('wa-tab'));
    const auctionsTab = tabs.find(tab => tab.textContent?.includes('Auctions'));
    expect(auctionsTab).toBeTruthy();
    // NEW: Verify auctions tab has correct panel attribute
    expect(auctionsTab?.getAttribute('panel')).toBe('auctions');
  });

  it('should render three tab panels', () => {
    const compiled = fixture.nativeElement as HTMLElement;
    const tabPanels = compiled.querySelectorAll('wa-tab-panel');
    expect(tabPanels.length).toBe(3);
    // NEW: Verify each panel has correct name attribute
    const names = Array.from(tabPanels).map(p => p.getAttribute('name'));
    expect(names).toContain('listings');
    expect(names).toContain('fills');
    expect(names).toContain('auctions');
  });

  it('should display empty state for listings', () => {
    const compiled = fixture.nativeElement as HTMLElement;
    const listingsPanel = compiled.querySelector('wa-tab-panel[name="listings"]');
    expect(listingsPanel?.textContent).toContain('You have no active listings');
    // NEW: Verify listings panel has callout
    const callout = listingsPanel?.querySelector('wa-callout');
    expect(callout).toBeTruthy();
  });

  it('should display empty state for fills', () => {
    const compiled = fixture.nativeElement as HTMLElement;
    const fillsPanel = compiled.querySelector('wa-tab-panel[name="fills"]');
    expect(fillsPanel?.textContent).toContain('You have no filled orders');
    // NEW: Verify fills panel has empty state icon
    const icon = fillsPanel?.querySelector('wa-icon');
    expect(icon).toBeTruthy();
  });

  it('should display empty state for auctions', () => {
    const compiled = fixture.nativeElement as HTMLElement;
    const auctionsPanel = compiled.querySelector('wa-tab-panel[name="auctions"]');
    expect(auctionsPanel?.textContent).toContain('No active auctions found');
    // NEW: Verify auctions panel has refresh button
    const button = auctionsPanel?.querySelector('wa-button');
    expect(button).toBeTruthy();
  });
});
