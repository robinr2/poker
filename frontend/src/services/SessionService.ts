const STORAGE_KEY = 'poker_session_token';

export class SessionService {
  static setToken(token: string): void {
    localStorage.setItem(STORAGE_KEY, token);
  }

  static getToken(): string | null {
    return localStorage.getItem(STORAGE_KEY);
  }

  static clearToken(): void {
    localStorage.removeItem(STORAGE_KEY);
  }
}
