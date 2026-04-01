import type { Lobby, LobbyPlayer } from "./types";

export type SetupScreen = "wizard" | "create-lobby" | "join-lobby";

export type MultiplayerConnectionStatus = "disconnected" | "connecting" | "connected" | "reconnecting";

export type MultiplayerSession = {
  playerToken: string;
  displayName: string;
  joinCode: string;
  gameID: string | null;
};

export function lobbyPlayerLabel(player: LobbyPlayer): string {
  return player.isHost ? `${player.displayName} (Host)` : player.displayName;
}

export function isLobbyPlayable(lobby: Lobby | null): boolean {
  return lobby?.state === 1 || lobby?.state === 2;
}
