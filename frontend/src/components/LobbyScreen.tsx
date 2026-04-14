import type { CSSProperties } from "react";
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

const factionAccentColors = ["#b14d36", "#4c7a45", "#496aa0", "#8a6842"];

function factionAccentColor(faction: number) {
  return factionAccentColors[faction] ?? "#7a6045";
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

function readyPlayerCount(lobby: Lobby): number {
  return lobby.players.filter((player) => player.isReady).length;
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
  const hostPlayer = lobby.players.find((player) => player.isHost) ?? null;
  const readyCount = readyPlayerCount(lobby);

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

        <div className="lobby-hero">
          <div className="lobby-hero-copy">
            <p className="eyebrow">Woodland Gathering</p>
            <h2 className="lobby-title">Players are assembling around table {lobby.joinCode}.</h2>
            <p className="lobby-copy">
              Claim a faction, confirm readiness, and let the host begin once every occupied seat is settled.
            </p>
          </div>
          <div className="lobby-overview-grid">
            <article className="lobby-overview-card">
              <span className="summary-label">Table Status</span>
              <strong>{lobby.players.length} / {lobby.factions.length} seats filled</strong>
              <span className="summary-line">{readyCount} ready, {Math.max(lobby.players.length - readyCount, 0)} still staging</span>
            </article>
            <article className="lobby-overview-card">
              <span className="summary-label">Table Host</span>
              <strong>{hostPlayer ? hostPlayer.displayName : "Waiting for host"}</strong>
              <span className="summary-line">{hostPlayer?.connected ? "Connected to the table" : "Host connection not confirmed"}</span>
            </article>
            <article className="lobby-overview-card">
              <span className="summary-label">Your Seat</span>
              <strong>{self ? lobbyPlayerLabel(self) : "Reconnecting..."}</strong>
              <span className="summary-line">{self?.hasFaction ? "Faction claimed and ready controls unlocked." : "Claim a faction seat to participate."}</span>
            </article>
          </div>
        </div>

        <div className="summary-stack lobby-section">
          <span className="summary-label">Claim A Seat</span>
          <div className="lobby-seat-grid">
            {lobby.factions.map((faction) => {
              const claimedBy = claimedByFaction(lobby, faction);
              const selected = currentFaction === faction;
              const style = {
                "--faction-color": factionAccentColor(faction)
              } as CSSProperties;
              return (
                <button
                  key={faction}
                  type="button"
                  className={`lobby-seat-card faction-seat-card ${selected ? "selected" : ""} ${claimedBy ? "claimed" : "open"}`}
                  disabled={submitting || (!!claimedBy && !selected)}
                  onClick={() => void onClaimFaction(selected ? null : faction)}
                  style={style}
                >
                  <div className="lobby-seat-card-top">
                    <span className="summary-label">Faction Seat</span>
                    <span className={`lobby-seat-status ${selected ? "selected" : claimedBy ? "claimed" : "open"}`}>
                      {selected ? "Your seat" : claimedBy ? "Claimed" : "Open"}
                    </span>
                  </div>
                  <strong>{factionLabels[faction]}</strong>
                  <span className="summary-line">{claimedBy ? claimedBy.displayName : "Open seat"}</span>
                  <span className="summary-line">
                    {selected ? "You are sitting here." : claimedBy ? "Already claimed." : "Available to claim."}
                  </span>
                  <span className="lobby-seat-note">
                    {selected ? "Click again to leave this faction seat." : claimedBy ? "Wait for this player to release the seat." : "Select this faction to take the seat."}
                  </span>
                </button>
              );
            })}
          </div>
        </div>

        <div className="summary-stack lobby-section">
          <span className="summary-label">Players At The Table</span>
          <div className="lobby-player-list">
            {lobby.players.map((player, index) => (
              <article
                key={`${index}-${player.displayName}-${player.isHost}`}
                className={`lobby-player-card ${player.isHost ? "host" : "guest"} ${player.isReady ? "ready" : "staging"} ${player.connected ? "connected" : "disconnected"}`}
                style={player.hasFaction ? ({ "--faction-color": factionAccentColor(player.faction) } as CSSProperties) : undefined}
              >
                <div className="lobby-player-card-top">
                  <strong>{lobbyPlayerLabel(player)}</strong>
                  <span className={`presence-pill ${player.connected ? "connected" : "disconnected"}`}>
                    {player.connected ? "Connected" : "Away"}
                  </span>
                </div>
                <div className="lobby-player-meta">
                  <span className={`lobby-role-pill ${player.isHost ? "host" : "guest"}`}>{player.isHost ? "Host" : "Guest"}</span>
                  <span className={`lobby-ready-pill ${player.isReady ? "ready" : "staging"}`}>{player.isReady ? "Ready" : "Not ready"}</span>
                </div>
                <span className="summary-line">
                  {player.hasFaction ? factionLabels[player.faction] ?? `Faction ${player.faction}` : "No faction claimed"}
                </span>
                <span className="summary-line">
                  {player.hasFaction ? "Faction seat locked in for this table." : "Seat claim still pending."}
                </span>
              </article>
            ))}
          </div>
        </div>

        <div className="sidebar-actions footer lobby-actions">
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
