import { factionLabels } from "../labels";
import type { MultiplayerConnectionStatus } from "../multiplayer";
import { lobbyPlayerLabel } from "../multiplayer";
import type { Lobby, LobbyPlayer } from "../types";

type LobbyScreenProps = {
  lobby: Lobby;
  self: LobbyPlayer | null;
  connectionStatus: MultiplayerConnectionStatus;
  status: string;
  submitting: boolean;
  onClaimFaction: (faction: number | null) => Promise<void>;
  onReady: (isReady: boolean) => Promise<void>;
  onStart: () => Promise<void>;
  onLeave: () => Promise<void>;
};

function claimedByFaction(lobby: Lobby, faction: number): LobbyPlayer | null {
  return lobby.players.find((player) => player.hasFaction && player.faction === faction) ?? null;
}

function connectionLabel(status: MultiplayerConnectionStatus): string {
  if (status === "connected") {
    return "Live";
  }
  if (status === "reconnecting") {
    return "Reconnecting";
  }
  if (status === "connecting") {
    return "Connecting";
  }
  return "Offline";
}

function lobbyReadyToStart(lobby: Lobby): boolean {
  return lobby.players.length >= 2 && lobby.players.every((player) => player.hasFaction && player.isReady);
}

export function LobbyScreen({
  lobby,
  self,
  connectionStatus,
  status,
  submitting,
  onClaimFaction,
  onReady,
  onStart,
  onLeave
}: LobbyScreenProps) {
  const readyToStart = lobbyReadyToStart(lobby);
  const currentFaction = self?.hasFaction ? self.faction : null;

  return (
    <main className="app-shell entry-shell">
      <section className="panel modal-panel multiplayer-screen entry-panel">
        <div className="panel-header">
          <div>
            <p className="eyebrow">Online Table</p>
            <h2>Join Code {lobby.joinCode}</h2>
          </div>
          <span className={`connection-pill ${connectionStatus}`}>{connectionLabel(connectionStatus)}</span>
        </div>

        <div className="lobby-hero pregame-table-hero">
          <div className="summary-stack">
            <span className="summary-label">Table Status</span>
            <span className="summary-line">
              {lobby.players.length} / {lobby.factions.length} seats filled
            </span>
          </div>
          <div className="summary-stack">
            <span className="summary-label">Your Seat</span>
            <span className="summary-line">{self ? lobbyPlayerLabel(self) : "Reconnecting..."}</span>
          </div>
        </div>

        <div className="summary-stack">
          <span className="summary-label">Claim A Seat</span>
          <div className="pregame-seat-grid">
            {lobby.factions.map((faction) => {
              const claimedBy = claimedByFaction(lobby, faction);
              const selected = currentFaction === faction;
              return (
                <button
                  key={faction}
                  type="button"
                  className={`pregame-seat-card faction-seat-card ${selected ? "selected" : ""}`}
                  disabled={submitting || (!!claimedBy && !selected)}
                  onClick={() => void onClaimFaction(selected ? null : faction)}
                >
                  <span className="summary-label">Faction Seat</span>
                  <strong>{factionLabels[faction]}</strong>
                  <span className="summary-line">{claimedBy ? claimedBy.displayName : "Open seat"}</span>
                  <span className="summary-line">{selected ? "You are sitting here." : claimedBy ? "Already claimed." : "Available to claim."}</span>
                </button>
              );
            })}
          </div>
        </div>

        <div className="summary-stack">
          <span className="summary-label">Players At The Table</span>
          <div className="lobby-player-list">
            {lobby.players.map((player, index) => (
              <article key={`${index}-${player.displayName}-${player.isHost}`} className="lobby-player-card">
                <div className="panel-header">
                  <strong>{lobbyPlayerLabel(player)}</strong>
                  <span className={`presence-pill ${player.connected ? "connected" : "disconnected"}`}>
                    {player.connected ? "Connected" : "Away"}
                  </span>
                </div>
                <span className="summary-line">
                  {player.hasFaction ? factionLabels[player.faction] ?? `Faction ${player.faction}` : "No faction claimed"}
                </span>
                <span className="summary-line">{player.isReady ? "Ready" : "Not ready"}</span>
              </article>
            ))}
          </div>
        </div>

        <div className="sidebar-actions footer">
          <button
            type="button"
            className="secondary"
            onClick={() => void onReady(!(self?.isReady ?? false))}
            disabled={submitting || !self?.hasFaction}
          >
            {self?.isReady ? "Unready" : "Ready Up"}
          </button>
          {self?.isHost ? (
            <button type="button" onClick={() => void onStart()} disabled={submitting || !readyToStart}>
              Start Game
            </button>
          ) : null}
          <button type="button" className="secondary" onClick={() => void onLeave()} disabled={submitting}>
            Leave Lobby
          </button>
        </div>
        <span className="message">
          {status || (readyToStart ? "Everyone is ready. Host can start." : "Claim a faction and ready up to begin.")}
        </span>
      </section>
    </main>
  );
}
