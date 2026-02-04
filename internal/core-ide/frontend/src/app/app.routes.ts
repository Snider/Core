import { Routes } from '@angular/router';
import { TrayComponent } from './pages/tray/tray.component';
import { IdeComponent } from './pages/ide/ide.component';

export const routes: Routes = [
  // System tray panel - standalone compact UI
  { path: 'tray', component: TrayComponent },

  // Full IDE interface
  { path: 'ide', component: IdeComponent },

  // Default to tray for the root (tray panel is the default view)
  { path: '', redirectTo: 'tray', pathMatch: 'full' },

  // Catch-all
  { path: '**', redirectTo: 'tray' },
];
