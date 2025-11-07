import { describe, it, expect, beforeEach, afterEach } from 'vitest';

import { SessionService } from './SessionService';

describe('SessionService', () => {
  beforeEach(() => {
    // Clear localStorage before each test
    localStorage.clear();
  });

  afterEach(() => {
    // Clear localStorage after each test
    localStorage.clear();
  });

  describe('TestSessionService_SetToken', () => {
    it('should store token in localStorage', () => {
      const token = 'test-token-123';
      SessionService.setToken(token);

      const stored = localStorage.getItem('poker_session_token');
      expect(stored).toBe(token);
    });

    it('should overwrite existing token', () => {
      SessionService.setToken('old-token');
      SessionService.setToken('new-token');

      const stored = localStorage.getItem('poker_session_token');
      expect(stored).toBe('new-token');
    });
  });

  describe('TestSessionService_GetToken', () => {
    it('should retrieve token from localStorage', () => {
      const token = 'test-token-456';
      localStorage.setItem('poker_session_token', token);

      const retrieved = SessionService.getToken();
      expect(retrieved).toBe(token);
    });

    it('should return null when no token stored', () => {
      const retrieved = SessionService.getToken();
      expect(retrieved).toBeNull();
    });

    it('should return stored token', () => {
      SessionService.setToken('my-token');
      const retrieved = SessionService.getToken();
      expect(retrieved).toBe('my-token');
    });
  });

  describe('TestSessionService_ClearToken', () => {
    it('should remove token from localStorage', () => {
      SessionService.setToken('test-token');
      SessionService.clearToken();

      const stored = localStorage.getItem('poker_session_token');
      expect(stored).toBeNull();
    });

    it('should not throw when clearing empty localStorage', () => {
      expect(() => SessionService.clearToken()).not.toThrow();
    });

    it('should result in getToken returning null', () => {
      SessionService.setToken('token');
      SessionService.clearToken();
      const retrieved = SessionService.getToken();
      expect(retrieved).toBeNull();
    });
  });
});
