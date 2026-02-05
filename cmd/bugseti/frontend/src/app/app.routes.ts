import { Routes } from '@angular/router';

export const routes: Routes = [
  {
    path: '',
    redirectTo: 'tray',
    pathMatch: 'full'
  },
  {
    path: 'tray',
    loadComponent: () => import('./tray/tray.component').then(m => m.TrayComponent)
  },
  {
    path: 'workbench',
    loadComponent: () => import('./workbench/workbench.component').then(m => m.WorkbenchComponent)
  },
  {
    path: 'settings',
    loadComponent: () => import('./settings/settings.component').then(m => m.SettingsComponent)
  },
  {
    path: 'onboarding',
    loadComponent: () => import('./onboarding/onboarding.component').then(m => m.OnboardingComponent)
  }
];
