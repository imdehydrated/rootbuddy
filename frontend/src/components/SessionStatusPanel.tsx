import { factionLabels } from "../labels";
import type { SavedSession } from "../localSession";
import type { MultiplayerConnectionStatus, MultiplayerSession } from "../multiplayer";
import type { BattlePrompt, GameState } from "../types";

type SessionStatusPanelProps = {
  state: GameState;
  hasSavedSession: boolean;
  serverGameID: string | null;
  savedSessionInfo: SavedSession | null;
  multiplayerSession: MultiplayerSession | null;
  multiplayerConnectionStatus: MultiplayerConnectionStatus;
  multiplayerBattlePrompt: BattlePrompt | null;
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

function promptStatusLine(prompt: BattlePrompt, perspective: number): string {
  if (prompt.stage === "defender_response") {
    return "Your defender response is required.";
  }
  if (prompt.stage === "attacker_response") {
    return "Your attacker response is required.";
  }
  if (prompt.stage === "ready_to_resolve") {
    return prompt.action.battle?.faction === perspective
      ? "Battle choices are locked in. You can resolve now."
      : `Battle choices are locked in. Waiting on ${factionLabels[prompt.action.battle?.faction ?? -1] ?? "the attacker"} to resolve.`;
  }
  return `Waiting on ${factionLabels[prompt.waitingOnFaction] ?? "another player"} for battle response.`;
}

export function SessionStatusPanel({
  state,
  hasSavedSession,
  serverGameID,
  savedSessionInfo,
  multiplayerSession,
  multiplayerConnectionStatus,
  multiplayerBattlePrompt
}: SessionStatusPanelProps) {
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

      {multiplayerSession ? (
        <div className="summary-stack" style={{ marginTop: "0.9rem" }}>
          <span className="summary-label">Realtime</span>
          <span className={`connection-pill ${multiplayerConnectionStatus}`}>{connectionLabel(multiplayerConnectionStatus)}</span>
          <span className="summary-line">Lobby: {multiplayerSession.joinCode}</span>
          <span className="summary-line">Player: {multiplayerSession.displayName}</span>
          <span className="summary-line">
            {multiplayerBattlePrompt ? promptStatusLine(multiplayerBattlePrompt, state.playerFaction) : `Turn: ${factionLabels[state.factionTurn] ?? "Unknown"}`}
          </span>
        </div>
      ) : null}

      <div className="summary-stack" style={{ marginTop: "0.9rem" }}>
        <span className="summary-label">Persistence</span>
        <span className="summary-line">
          {multiplayerSession ? "Browser reconnect token saved locally" : hasSavedSession ? "Local autosave available" : "No local autosave yet"}
        </span>
        {savedSessionInfo?.savedAt ? <span className="summary-line">Last saved: {formatSavedAt(savedSessionInfo.savedAt)}</span> : null}
        {serverGameID ? <span className="summary-line">Online game ID: {serverGameID.slice(0, 8)}...</span> : null}
      </div>
    </section>
  );
}
