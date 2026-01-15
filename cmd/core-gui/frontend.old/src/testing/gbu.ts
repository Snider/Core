// Good/Bad/Ugly test helpers for Jasmine
// Usage:
//   import { itGood, itBad, itUgly, trio } from 'src/testing/gbu';
//   itGood('does X', () => { /* ... */ });
//   trio('feature does Y', {
//     good: () => { /* ... */ },
//     bad:  () => { /* ... */ },
//     ugly: () => { /* ... */ },
//   });

export function suffix(base: string, tag: 'Good' | 'Bad' | 'Ugly'): string {
  return `${base}_${tag}`;
}

export function itGood(name: string, fn: jasmine.ImplementationCallback, timeout?: number): void {
  it(suffix(name, 'Good'), fn, timeout as any);
}

export function itBad(name: string, fn: jasmine.ImplementationCallback, timeout?: number): void {
  it(suffix(name, 'Bad'), fn, timeout as any);
}

export function itUgly(name: string, fn: jasmine.ImplementationCallback, timeout?: number): void {
  it(suffix(name, 'Ugly'), fn, timeout as any);
}

export function trio(name: string, impls: { good: () => void; bad: () => void; ugly: () => void; }): void {
  itGood(name, impls.good);
  itBad(name, impls.bad);
  itUgly(name, impls.ugly);
}
