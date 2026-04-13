import { factionLabels } from "../labels";
import { lobbyPlayerLabel } from "../multiplayer";
import type { LobbyPlayer } from "../types";

type PlayerPresenceBarProps = {
  players: LobbyPlayer[];
  factionTurn: number;
  perspectiveFaction: number;
};

export function PlayerPresenceBar({ players, factionTurn, perspectiveFaction }: PlayerPresenceBarProps) {
  if (players.length === 0) {
    return null;
  }

  return (
    <section className="player-presence-bar panel" aria-label="Player presence bar">
      <div className="player-presence-header">
        <span className="summary-label">Players In Game</span>
        <span className="summary-line">{players.length} connected seat{players.length === 1 ? "" : "s"} tracked</span>
      </div>
      <div className="player-presence-list">
        {players.map((player, index) => {
          const isActiveTurn = player.hasFaction && player.faction === factionTurn;
          const isPerspective = player.hasFaction && player.faction === perspectiveFaction;

          return (
            <article
              key={`${index}-${player.displayName}-${player.faction}-${player.isHost}`}
              className={`player-presence-pill ${isActiveTurn ? "active-turn" : ""} ${player.connected ? "connected" : "disconnected"} ${isPerspective ? "perspective" : ""}`.trim()}
            >
              <div className="player-presence-pill-top">
                <strong>{lobbyPlayerLabel(player)}</strong>
                <span className={`presence-pill ${player.connected ? "connected" : "disconnected"}`}>
                  {player.connected ? "Connected" : "Away"}
                </span>
              </div>
              <span className="summary-line">
                {player.hasFaction ? factionLabels[player.faction] ?? `Faction ${player.faction}` : "No faction claimed"}
              </span>
              <span className="summary-line">
                {isActiveTurn ? "Active turn" : isPerspective ? "You" : player.isReady ? "Ready" : "Waiting"}
              </span>
            </article>
          );
        })}
      </div>
    </section>
  );
}
