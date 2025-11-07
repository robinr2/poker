export type ConnectionStatus =
  | 'connecting'
  | 'connected'
  | 'disconnected'
  | 'error';

type MessageCallback = (data: string) => void;
type StatusCallback = (status: ConnectionStatus) => void;

export class WebSocketService {
  private ws: WebSocket | null = null;
  private url: string;
  private status: ConnectionStatus = 'disconnected';
  private messageCallbacks: MessageCallback[] = [];
  private statusCallbacks: StatusCallback[] = [];
  private reconnectAttempt = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private shouldReconnect = true;
  private connectResolve: (() => void) | null = null;
  private connectReject: ((reason?: unknown) => void) | null = null;

  constructor(url: string, token?: string) {
    this.url = this.appendTokenToUrl(url, token);
  }

  private appendTokenToUrl(url: string, token?: string): string {
    if (!token) {
      return url;
    }

    // Check if URL already has query parameters
    const separator = url.includes('?') ? '&' : '?';
    return `${url}${separator}token=${token}`;
  }

  async connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      this.connectResolve = resolve;
      this.connectReject = reject;
      this.attemptConnection();
    });
  }

  private attemptConnection(): void {
    if (this.status === 'connected') {
      return;
    }

    this.setStatus('connecting');

    try {
      this.ws = new WebSocket(this.url);

      this.ws.onopen = () => {
        this.reconnectAttempt = 0;
        this.setStatus('connected');
        if (this.connectResolve) {
          this.connectResolve();
          this.connectResolve = null;
          this.connectReject = null;
        }
      };

      this.ws.onmessage = (event: MessageEvent) => {
        this.messageCallbacks.forEach((callback) => callback(event.data));
      };

      this.ws.onerror = () => {
        this.handleConnectionError();
      };

      this.ws.onclose = () => {
        if (this.status !== 'disconnected') {
          this.setStatus('disconnected');
          if (this.shouldReconnect) {
            this.scheduleReconnect();
          }
        }
      };
    } catch {
      this.handleConnectionError();
    }
  }

  private handleConnectionError(): void {
    this.setStatus('error');

    if (this.connectReject) {
      this.connectReject(new Error('WebSocket connection failed'));
      this.connectResolve = null;
      this.connectReject = null;
    }

    if (this.shouldReconnect) {
      this.scheduleReconnect();
    }
  }

  private scheduleReconnect(): void {
    // Calculate exponential backoff: 1s, 2s, 4s, 8s, 16s, capped at 30s
    const backoffMs = Math.min(
      1000 * Math.pow(2, this.reconnectAttempt),
      30000
    );
    this.reconnectAttempt++;

    this.reconnectTimer = setTimeout(() => {
      if (this.shouldReconnect) {
        this.attemptConnection();
      }
    }, backoffMs);
  }

  disconnect(): void {
    this.shouldReconnect = false;

    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }

    this.setStatus('disconnected');
  }

  send(data: string): void {
    if (this.status !== 'connected' || !this.ws) {
      throw new Error('WebSocket is not connected');
    }

    this.ws.send(data);
  }

  onMessage(callback: MessageCallback): () => void {
    this.messageCallbacks.push(callback);
    return () => {
      const index = this.messageCallbacks.indexOf(callback);
      if (index > -1) this.messageCallbacks.splice(index, 1);
    };
  }

  onStatusChange(callback: StatusCallback): () => void {
    this.statusCallbacks.push(callback);
    return () => {
      const index = this.statusCallbacks.indexOf(callback);
      if (index > -1) this.statusCallbacks.splice(index, 1);
    };
  }

  getStatus(): ConnectionStatus {
    return this.status;
  }

  private setStatus(newStatus: ConnectionStatus): void {
    if (this.status !== newStatus) {
      this.status = newStatus;
      this.statusCallbacks.forEach((callback) => callback(newStatus));
    }
  }
}
