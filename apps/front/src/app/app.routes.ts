import { Route } from '@angular/router';
import { Signup } from '../libs/auth/pages/signup';
import { Login } from '../libs/auth/pages/login';
import { Dashboard } from '../libs/dashboard/pages/dashboard';
import { AuthGuard } from '../libs/auth/guard/auth.guard';

export const appRoutes: Route[] = [
  { path: 'signup', component: Signup },
  { path: 'login', component: Login },
  { path: 'dashboard', component: Dashboard, canActivate: [AuthGuard] },
  { path: '', redirectTo: '/dashboard', pathMatch: 'full' },
];
