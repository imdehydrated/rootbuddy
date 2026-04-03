import type { MultiplayerSocketMessage } from "./serverContract";
import type { MultiplayerConnectionStatus } from "./multiplayer";

const API_BASE = "http://localhost:8080/api";

type WebSocketClientOptions = {
  token: string;
  onMessage: (message: MultiplayerSocketMessage) => void;
  onConnectionChange?: (status: MultiplayerConnectionStatus) => void;
};

export class RootBuddyWebSocketClient {
  private readonly token: string;
  private readonly onMessage: (message: MultiplayerSocketMessage) => void;
  private readonly onConnectionChange?: (status: MultiplayerConnectionStatus) => void;
  private socket: WebSocket | null = null;
  private reconnectTimer: number | null = null;
  private reconnectDelay = 1000;
  private closedManually = false;
  private hasConnected = false;
  private reconnectScheduled = false;

  constructor(options: WebSocketClientOptions) {
    this.token = options.token;
    this.onMessage = options.onMessage;
    this.onConnectionChange = options.onConnectionChange;
  }

  connect() {
    this.closedManually = false;
    this.openSocket();
  }

  disconnect() {
    this.closedManually = true;
    if (this.reconnectTimer !== null) {
      window.clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    this.reconnectScheduled = false;
    if (!this.socket) {
      this.onConnectionChange?.("disconnected");
      return;
    }
    this.socket.close();
    this.socket = null;
  }

  private openSocket() {
    const wsURL = `${API_BASE.replace("http://", "ws://").replace("https://", "wss://")}/ws?token=${encodeURIComponent(this.token)}`;
    const socket = new WebSocket(wsURL);
    this.socket = socket;
    this.onConnectionChange?.(this.hasConnected || this.reconnectScheduled ? "reconnecting" : "connecting");

    socket.onopen = () => {
      this.hasConnected = true;
      this.reconnectScheduled = false;
      this.reconnectDelay = 1000;
      this.onConnectionChange?.("connected");
    };

    socket.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data) as MultiplayerSocketMessage;
        this.onMessage(message);
      } catch {
      }
    };

    socket.onerror = () => {
      socket.close();
    };

    socket.onclose = () => {
      if (this.closedManually) {
        this.reconnectScheduled = false;
        this.onConnectionChange?.("disconnected");
        return;
      }
      this.reconnectScheduled = true;
      this.onConnectionChange?.("reconnecting");
      this.scheduleReconnect();
    };
  }

  private scheduleReconnect() {
    if (this.reconnectTimer !== null) {
      return;
    }

    const delay = this.reconnectDelay;
    this.reconnectDelay = Math.min(this.reconnectDelay * 2, 15000);
    this.reconnectTimer = window.setTimeout(() => {
      this.reconnectTimer = null;
      this.openSocket();
    }, delay);
  }
}
