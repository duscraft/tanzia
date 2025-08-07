import { Injectable } from '@angular/core';

const msPerDay = 86400000; // 24 * 60 * 60 * 1000

@Injectable({
  providedIn: 'root',
})
export class AuthService {
  private _cookieName = 'auth_token';

  isAuthenticated(): boolean {
    return !!this.getCookie(this._cookieName);
  }

  getCookie(name: string): string | null {
    const match = document.cookie.match(new RegExp('(^| )' + name + '=([^;]+)'));
    return match ? match[2] : null;
  }

  setCookie(name: string, value: string, days = 1): void {
    const expires = new Date(Date.now() + days * msPerDay).toUTCString();
    document.cookie = `${name}=${value}; expires=${expires}; path=/`;
  }

  deleteCookie(name: string): void {
    document.cookie = `${name}=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;`;
  }

  login(token: string): void {
    this.setCookie(this._cookieName, token);
  }

  logout(): void {
    this.deleteCookie(this._cookieName);
  }
}
