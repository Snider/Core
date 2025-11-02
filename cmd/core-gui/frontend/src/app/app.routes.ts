import { Routes } from '@angular/router';
import { HomePage } from './pages/home/home.page';
import { SearchTldPage } from './pages/search-tld/search-tld.page';
import { OnboardingPage } from './pages/onboarding/onboarding.page';
import { SettingsPage } from './pages/settings/settings.page';
import { DomainManagerPage } from './pages/domain-manager/domain-manager.page';
import { ExchangePage } from './pages/exchange/exchange.page';

export const routes: Routes = [
  { path: '', redirectTo: '/account', pathMatch: 'full' },
  { path: 'account', component: HomePage, title: 'Portfolio • Bob Wallet' },
  { path: 'send', component: HomePage, title: 'Send • Bob Wallet' },
  { path: 'receive', component: HomePage, title: 'Receive • Bob Wallet' },
  { path: 'domain-manager', component: DomainManagerPage, title: 'Domain Manager • Bob Wallet' },
  { path: 'domains', component: SearchTldPage, title: 'Browse Domains • Bob Wallet' },
  { path: 'bids', component: HomePage, title: 'Your Bids • Bob Wallet' },
  { path: 'watching', component: HomePage, title: 'Watching • Bob Wallet' },
  { path: 'exchange', component: ExchangePage, title: 'Exchange • Bob Wallet' },
  { path: 'get-coins', component: HomePage, title: 'Claim Airdrop • Bob Wallet' },
  { path: 'sign-message', component: HomePage, title: 'Sign Message • Bob Wallet' },
  { path: 'verify-message', component: HomePage, title: 'Verify Message • Bob Wallet' },
  { path: 'settings', component: SettingsPage, title: 'Settings • Bob Wallet' },
  { path: 'onboarding', component: OnboardingPage, title: 'Onboarding • Bob Wallet' },
  { path: '**', redirectTo: '/account' }
];
