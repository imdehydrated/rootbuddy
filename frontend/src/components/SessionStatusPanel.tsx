import { factionLabels } from "../labels";
import type { SavedSession } from "../localSession";
import type { GameState } from "../types";

type SessionStatusPanelProps = {
  state: GameState;
  hasSavedSession: boolean;
  serverGameID: string | null;
  savedSessionInfo: SavedSession | null;
};

function formatSavedAt(savedAt: string | undefined): string {
  if (!savedAt) {
    return "";
  }

  const parsed = new Date(savedAt);
  if (Number.isNaN(parsed.getTime())) {
    return "";
  }

  return parsed.toLocaleString();
}

export function SessionStatusPanel({ state, hasSavedSession, serverGameID, savedSessionInfo }: SessionStatusPanelProps) {
  const modeLabel = state.gameMode === 0 ? "Online" : "Assist";
  const lifecycleLabel =
    state.gamePhase === 0 ? "Setup in progress" : state.gamePhase === 2 ? "Reviewing final result" : "Active game";
  const hiddenInfoLabel =
    state.gameMode === 0
      ? "Server-authoritative hidden info with player redaction."
      : "Observed hidden info tracked as placeholders and counts.";

  return (
    <section className="panel sidebar-panel">
      <p className="eyebrow">Session</p>
      <div className="summary-stack">
        <span className="summary-label">Mode</span>
        <span className="summary-line">{lifecycleLabel}</span>
        <span className="summary-line">{modeLabel}</span>
        <span className="summary-line">Perspective: {factionLabels[state.playerFaction] ?? "Unknown"}</span>
        {state.gamePhase === 2 ? <span className="summary-line">Winner: {factionLabels[state.winner] ?? "Unknown"}</span> : null}
        <span className="summary-line">{hiddenInfoLabel}</span>
      </div>

      <div className="summary-stack" style={{ marginTop: "0.9rem" }}>
        <span className="summary-label">Persistence</span>
        <span className="summary-line">{hasSavedSession ? "Local autosave available" : "No local autosave yet"}</span>
        {savedSessionInfo?.savedAt ? <span className="summary-line">Last saved: {formatSavedAt(savedSessionInfo.savedAt)}</span> : null}
        {serverGameID ? <span className="summary-line">Online game ID: {serverGameID.slice(0, 8)}...</span> : null}
      </div>
    </section>
  );
}
