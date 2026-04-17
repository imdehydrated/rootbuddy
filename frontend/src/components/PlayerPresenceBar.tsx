import { useState } from "react";
import { factionLabels } from "../labels";
import { lobbyPlayerLabel } from "../multiplayer";
import type { LobbyPlayer } from "../types";

type PlayerPresenceBarProps = {
  players: LobbyPlayer[];
  factionTurn: number;
  perspectiveFaction: number;
};

export function PlayerPresenceBar({ players, factionTurn, perspectiveFaction }: PlayerPresenceBarProps) {
  const [collapsed, setCollapsed] = useState(true);

  if (players.length === 0) {
    return null;
  }

  return (
    <section className={`player-presence-bar panel ${collapsed ? "collapsed" : ""}`} aria-label="Player presence bar">
      <div className="player-presence-header">
        <div className="summary-stack">
          <span className="summary-label">Players In Game</span>
          <span className="summary-line">{players.length} seat{players.length === 1 ? "" : "s"} tracked</span>
        </div>
        <button type="button" className="secondary player-presence-toggle" onClick={() => setCollapsed((current) => !current)}>
          {collapsed ? "Show Players" : "Hide Players"}
        </button>
      </div>
      <div className={collapsed ? "player-presence-compact-list" : "player-presence-list"}>
        {players.map((player, index) => {
          const isActiveTurn = player.hasFaction && player.faction === factionTurn;
          const isPerspective = player.hasFaction && player.faction === perspectiveFaction;
          const presenceClass = `player-presence-pill ${isActiveTurn ? "active-turn" : ""} ${player.connected ? "connected" : "disconnected"} ${isPerspective ? "perspective" : ""}`.trim();

          return collapsed ? (
            <article key={`${index}-${player.displayName}-${player.faction}-${player.isHost}`} className={`${presenceClass} compact`}>
              <strong>{lobbyPlayerLabel(player)}</strong>
              <span className={`presence-pill ${player.connected ? "connected" : "disconnected"}`}>
                {player.connected ? "Online" : "Away"}
              </span>
            </article>
          ) : (
            <article
              key={`${index}-${player.displayName}-${player.faction}-${player.isHost}`}
              className={presenceClass}
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
