import { useEffect, useMemo, useRef, useState } from "react";
import { factionLabels } from "../labels";
import type { ActionLogEntry } from "../types";

type GameLogPanelProps = {
  entries: ActionLogEntry[];
  factionTurn: number;
};

const factionLogClasses = ["marquise", "alliance", "eyrie", "vagabond"];

function factionLogClass(faction: number) {
  return factionLogClasses[faction] ?? "unknown";
}

function formatTimestamp(timestamp: number) {
  if (!Number.isFinite(timestamp) || timestamp <= 0) {
    return null;
  }

  return new Intl.DateTimeFormat(undefined, {
    hour: "numeric",
    minute: "2-digit"
  }).format(new Date(timestamp));
}

export function GameLogPanel({ entries, factionTurn }: GameLogPanelProps) {
  const [collapsed, setCollapsed] = useState(false);
  const listRef = useRef<HTMLDivElement | null>(null);
  const orderedEntries = useMemo(() => [...entries].reverse(), [entries]);

  useEffect(() => {
    if (!listRef.current || collapsed) {
      return;
    }
    listRef.current.scrollTop = 0;
  }, [collapsed, orderedEntries]);

  return (
    <section className={`game-log-panel ${collapsed ? "collapsed" : ""}`} aria-label="Game log panel">
      <div className="game-log-header">
        <div className="summary-stack">
          <span className="summary-label">Recent Actions</span>
          <strong>Game Log</strong>
          <span className="summary-line">
            {entries.length > 0
              ? `${entries.length} recorded action${entries.length === 1 ? "" : "s"}`
              : `Waiting for ${factionLabels[factionTurn] ?? "the active faction"} to generate the first logged action.`}
          </span>
        </div>
        <button type="button" className="secondary game-log-toggle" onClick={() => setCollapsed((current) => !current)}>
          {collapsed ? "Show Log" : "Hide Log"}
        </button>
      </div>

      {!collapsed ? (
        <div ref={listRef} className="game-log-list">
          {orderedEntries.length > 0 ? (
            orderedEntries.map((entry, index) => {
              const factionLabel = factionLabels[entry.faction] ?? `Faction ${entry.faction}`;
              const timestampLabel = formatTimestamp(entry.timestamp);

              return (
                <article key={`${entry.timestamp}-${entry.summary}-${index}`} className={`game-log-entry ${factionLogClass(entry.faction)}`}>
                  <div className="game-log-entry-header">
                    <span className={`game-log-faction-badge ${factionLogClass(entry.faction)}`}>{factionLabel}</span>
                    <span className="game-log-round">Round {entry.roundNumber}</span>
                    {timestampLabel ? <span className="game-log-time">{timestampLabel}</span> : null}
                  </div>
                  <span className="game-log-summary">{entry.summary}</span>
                </article>
              );
            })
          ) : (
            <div className="game-log-empty">
              <span className="summary-line">No multiplayer actions have been logged yet.</span>
            </div>
          )}
        </div>
      ) : null}
    </section>
  );
}
