import { describeKnownCardID } from "../cardCatalog";
import { factionLabels } from "../labels";
import type { GameState } from "../types";

type EndgamePanelProps = {
  state: GameState;
  hasSavedSession: boolean;
  serverGameID: string | null;
  onNewGame: () => void;
  onReturnToSetup: () => void;
  onClearSavedSession: () => void;
  onOpenDebug: () => void;
};

function sortedScoreLines(state: GameState): Array<{ faction: number; score: number }> {
  return factionLabels
    .map((_, index) => ({
      faction: index,
      score: state.victoryPoints[String(index)] ?? 0
    }))
    .sort((left, right) => right.score - left.score);
}

function winningCoalitionSet(state: GameState): Set<number> {
  return new Set(state.winningCoalition);
}

function activeDominanceCardID(state: GameState, faction: number): number | null {
  const cardID = state.activeDominance[String(faction)];
  return typeof cardID === "number" ? cardID : null;
}

function victorySummary(state: GameState): { title: string; detail: string } {
  if (state.coalitionActive && state.winningCoalition.length > 0) {
    const partnerFaction = state.winningCoalition.find((faction) => faction !== 3) ?? state.winner;
    return {
      title: "Coalition Victory",
      detail: `${factionLabels[partnerFaction] ?? "Coalition partner"} won while allied with the Vagabond.`
    };
  }

  const dominanceCardID = activeDominanceCardID(state, state.winner);
  if (dominanceCardID !== null) {
    const cardLabel = describeKnownCardID(dominanceCardID);
    if (cardLabel.includes("(Bird)")) {
      return {
        title: "Bird Dominance Victory",
        detail: `${factionLabels[state.winner] ?? "Winner"} ruled both opposite corners with ${cardLabel}.`
      };
    }

    return {
      title: "Suit Dominance Victory",
      detail: `${factionLabels[state.winner] ?? "Winner"} met the clearing-rule requirement with ${cardLabel}.`
    };
  }

  return {
    title: "Victory Point Victory",
    detail: `${factionLabels[state.winner] ?? "Winner"} reached ${state.victoryPoints[String(state.winner)] ?? 0} victory points.`
  };
}

export function EndgamePanel({
  state,
  hasSavedSession,
  serverGameID,
  onNewGame,
  onReturnToSetup,
  onClearSavedSession,
  onOpenDebug
}: EndgamePanelProps) {
  if (state.gamePhase !== 2) {
    return null;
  }

  const summary = victorySummary(state);
  const coalitionLabel =
    state.winningCoalition.length > 0
      ? state.winningCoalition.map((faction) => factionLabels[faction] ?? `Faction ${faction}`).join(" + ")
      : "";
  const coalitionWinners = winningCoalitionSet(state);

  return (
    <section className="panel sidebar-panel endgame-panel">
      <p className="eyebrow">Game Over</p>
      <div className="endgame-hero">
        <span className="summary-label">{summary.title}</span>
        <strong>{factionLabels[state.winner] ?? "Unknown"} Wins</strong>
        <span className="summary-line">{summary.detail}</span>
      </div>

      <div className="summary-stack">
        <span className="summary-label">Final State</span>
        <span className="summary-line">Winner score: {state.victoryPoints[String(state.winner)] ?? 0}</span>
        {coalitionLabel ? <span className="summary-line">Coalition: {coalitionLabel}</span> : null}
        <span className="summary-line">Round: {state.roundNumber}</span>
        <span className="summary-line">
          Resume: {hasSavedSession ? "Final result is saved locally." : "No local autosave recorded."}
        </span>
        {serverGameID ? <span className="summary-line">Online save: {serverGameID.slice(0, 8)}...</span> : null}
      </div>

      <div className="summary-stack" style={{ marginTop: "0.9rem" }}>
        <span className="summary-label">Final Scores</span>
        {sortedScoreLines(state).map(({ faction, score }) => (
          <span
            key={faction}
            className={`summary-line ${faction === state.winner || coalitionWinners.has(faction) ? "endgame-winner-line" : ""}`}
          >
            {factionLabels[faction] ?? `Faction ${faction}`}: {score}
            {faction === state.winner ? " - Winner" : coalitionWinners.has(faction) ? " - Coalition" : ""}
          </span>
        ))}
      </div>

      <div className="summary-stack" style={{ marginTop: "0.9rem" }}>
        <span className="summary-label">Next Steps</span>
        <span className="summary-line">Return to Setup keeps this result available for review.</span>
        {hasSavedSession ? <span className="summary-line">Clear Saved Result removes the resumable review copy.</span> : null}
        <span className="summary-line">New Game clears the result and starts a fresh setup.</span>
      </div>

      <div className="sidebar-actions footer">
        <button type="button" className="secondary" onClick={onReturnToSetup}>
          Return to Setup
        </button>
        {hasSavedSession ? (
          <button type="button" className="secondary" onClick={onClearSavedSession}>
            Clear Saved Result
          </button>
        ) : null}
        <button type="button" className="secondary" onClick={onOpenDebug}>
          Debug
        </button>
        <button type="button" onClick={onNewGame}>
          New Game
        </button>
      </div>
    </section>
  );
}
