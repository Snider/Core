import { ComponentFixture, TestBed } from '@angular/core/testing';
import { HomePage } from './home.page';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';

describe('HomePage', () => {
  let component: HomePage;
  let fixture: ComponentFixture<HomePage>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [HomePage],
      schemas: [CUSTOM_ELEMENTS_SCHEMA]
    }).compileComponents();

    fixture = TestBed.createComponent(HomePage);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
    // NEW: Verify component is an instance of HomePage
    expect(component instanceof HomePage).toBe(true);
  });

  it('should render account header with balance sections', () => {
    const compiled = fixture.nativeElement as HTMLElement;
    const header = compiled.querySelector('.account__header');
    expect(header).toBeTruthy();
    // NEW: Verify header has at least 2 sections (spendable and locked)
    const sections = compiled.querySelectorAll('.account__header__section');
    expect(sections.length).toBeGreaterThanOrEqual(2);
  });

  it('should display spendable balance section', () => {
    const compiled = fixture.nativeElement as HTMLElement;
    const spendableLabel = compiled.querySelector('.label');
    expect(spendableLabel?.textContent).toContain('Spendable');
    // NEW: Verify there's an amount display
    const amount = compiled.querySelector('.amount');
    expect(amount).toBeTruthy();
  });

  it('should render actionable cards grid', () => {
    const compiled = fixture.nativeElement as HTMLElement;
    const cardsGrid = compiled.querySelector('.account__cards');
    expect(cardsGrid).toBeTruthy();
    // NEW: Verify grid contains wa-card elements
    const cards = cardsGrid?.querySelectorAll('wa-card');
    expect(cards?.length).toBeGreaterThan(0);
  });

  it('should render six action cards', () => {
    const compiled = fixture.nativeElement as HTMLElement;
    const cards = compiled.querySelectorAll('wa-card');
    expect(cards.length).toBe(6);
    // NEW: Verify each card has an icon
    const icons = compiled.querySelectorAll('.account__card__icon');
    expect(icons.length).toBe(6);
  });

  it('should render transaction history section', () => {
    const compiled = fixture.nativeElement as HTMLElement;
    const transactions = compiled.querySelector('.account__transactions');
    expect(transactions).toBeTruthy();
    // NEW: Verify transactions section has a title
    const title = transactions?.querySelector('.account__panel-title');
    expect(title).toBeTruthy();
  });

  it('should display transaction history title', () => {
    const compiled = fixture.nativeElement as HTMLElement;
    const title = compiled.querySelector('.account__panel-title');
    expect(title?.textContent).toContain('Transaction History');
    // NEW: Verify title is not empty
    expect(title?.textContent?.trim().length).toBeGreaterThan(0);
  });

  it('should show empty state for transactions', () => {
    const compiled = fixture.nativeElement as HTMLElement;
    const callout = compiled.querySelector('wa-callout');
    expect(callout?.textContent).toContain('No transactions yet');
    // NEW: Verify callout has neutral variant attribute
    expect(callout?.getAttribute('variant')).toBe('neutral');
  });
});
