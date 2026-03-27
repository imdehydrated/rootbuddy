import {
  eyrieLeaderLabels,
  factionLabels,
  itemStatusLabels,
  phaseLabels,
  relationshipLabels,
  setupStageLabels,
  stepLabels,
  suitLabels,
  vagabondCharacterLabels
} from "../labels";
import type { GameState } from "../types";

type TurnSummaryPanelProps = {
  state: GameState;
};

function formatTurnOrder(turnOrder: number[]): string {
  return turnOrder.map((faction) => factionLabels[faction] ?? `Faction ${faction}`).join(" -> ");
}

function currentFactionLines(state: GameState): string[] {
  const hiddenCount = (faction: number, zone: string) =>
    state.hiddenCards.filter((card) => card.ownerFaction === faction && card.zone === zone).length;

  switch (state.factionTurn) {
    case 0:
      return [
        `Supply ${state.marquise.warriorSupply}`,
        `Buildings ${state.marquise.sawmillsPlaced}/${state.marquise.workshopsPlaced}/${state.marquise.recruitersPlaced}`,
        state.marquise.keepClearingID > 0 ? `Keep ${state.marquise.keepClearingID}` : "Keep unset"
      ];
    case 1:
      return [
        `Officers ${state.alliance.officers}`,
        `Supporters ${state.alliance.supporters.length || hiddenCount(1, "supporters")}`,
        `Sympathy ${state.alliance.sympathyPlaced}`,
        `Bases ${[state.alliance.foxBasePlaced ? "Fox" : "", state.alliance.rabbitBasePlaced ? "Rabbit" : "", state.alliance.mouseBasePlaced ? "Mouse" : ""].filter(Boolean).join(", ") || "None"}`
      ];
    case 2:
      return [
        `Leader ${eyrieLeaderLabels[state.eyrie.leader] ?? "Unknown"}`,
        `Roosts ${state.eyrie.roostsPlaced}`,
        `Decree ${state.eyrie.decree.recruit.length}/${state.eyrie.decree.move.length}/${state.eyrie.decree.battle.length}/${state.eyrie.decree.build.length}`
      ];
    case 3: {
      const itemSummary = [0, 1, 2]
        .map((status) => `${itemStatusLabels[status]} ${state.vagabond.items.filter((item) => item.status === status).length}`)
        .join(", ");
      const relationshipSummary = Object.entries(state.vagabond.relationships)
        .map(([faction, level]) => `${factionLabels[Number(faction)]}: ${relationshipLabels[level] ?? "Unknown"}`)
        .join("; ");

      return [
        `Character ${vagabondCharacterLabels[state.vagabond.character] ?? "Unknown"}`,
        state.vagabond.inForest ? `Forest ${state.vagabond.forestID || "?"}` : `Clearing ${state.vagabond.clearingID || "?"}`,
        itemSummary,
        relationshipSummary || "Relationships unset"
      ];
    }
    default:
      return [];
  }
}

export function TurnSummaryPanel({ state }: TurnSummaryPanelProps) {
  const hiddenHandLines = state.hiddenCards.reduce<Record<string, number>>((counts, card) => {
    if (card.zone !== "hand") {
      return counts;
    }
    const key = String(card.ownerFaction);
    counts[key] = (counts[key] ?? 0) + 1;
    return counts;
  }, {});
  const hiddenSupporterLines = state.hiddenCards.reduce<Record<string, number>>((counts, card) => {
    if (card.zone !== "supporters") {
      return counts;
    }
    const key = String(card.ownerFaction);
    counts[key] = (counts[key] ?? 0) + 1;
    return counts;
  }, {});
  const activeDominanceLines = Object.entries(state.activeDominance).map(([faction, cardID]) => {
    const factionLabel = factionLabels[Number(faction)] ?? `Faction ${faction}`;
    return `${factionLabel}: Card ${cardID}`;
  });
  const coalitionLabel =
    state.coalitionActive && state.winningCoalition.length > 0
      ? state.winningCoalition.map((faction) => factionLabels[faction] ?? `Faction ${faction}`).join(" + ")
      : state.coalitionActive
        ? `${factionLabels[state.coalitionPartner] ?? "Unknown"} + Vagabond`
        : "";

  return (
    <section className="panel floating-turn-summary">
      <p className="eyebrow">Turn Summary</p>
      <div className="summary-grid">
        <div className="summary-card">
          <span className="summary-label">Active Turn</span>
          <strong>{factionLabels[state.factionTurn] ?? "Unknown"}</strong>
          <span>
            {state.gamePhase === 0
              ? setupStageLabels[state.setupStage] ?? "Setup"
              : `${phaseLabels[state.currentPhase] ?? "Unknown"} / ${stepLabels[state.currentStep] ?? "Unknown"}`}
          </span>
        </div>
        <div className="summary-card">
          <span className="summary-label">Turn Order</span>
          <strong>{formatTurnOrder(state.turnOrder)}</strong>
        </div>
      </div>

      <div className="score-strip">
        {factionLabels.map((label, index) => (
          <span key={label} className="score-pill">
            {label}: {state.victoryPoints[String(index)] ?? 0}
          </span>
        ))}
      </div>

      {state.gamePhase === 2 ? (
        <div className="summary-stack">
          <span className="summary-label">Game Over</span>
          <span className="summary-line">Winner: {factionLabels[state.winner] ?? "Unknown"}</span>
          {state.winningCoalition.length > 0 ? <span className="summary-line">Coalition: {coalitionLabel}</span> : null}
        </div>
      ) : null}

      <div className="summary-stack">
        <span className="summary-label">Current Faction State</span>
        {currentFactionLines(state).map((line) => (
          <span key={line} className="summary-line">
            {line}
          </span>
        ))}
      </div>

      {activeDominanceLines.length > 0 || state.availableDominance.length > 0 || state.coalitionActive ? (
        <div className="summary-stack">
          <span className="summary-label">Dominance</span>
          {activeDominanceLines.map((line) => (
            <span key={line} className="summary-line">
              Active {line}
            </span>
          ))}
          {state.availableDominance.length > 0 ? (
            <span className="summary-line">Available cards: {state.availableDominance.join(", ")}</span>
          ) : null}
          {state.coalitionActive ? <span className="summary-line">Coalition: {coalitionLabel}</span> : null}
        </div>
      ) : null}

      {state.factionTurn === 3 && state.vagabond.questsAvailable.length > 0 ? (
        <div className="summary-stack">
          <span className="summary-label">Available Quests</span>
          {state.vagabond.questsAvailable.map((quest) => (
            <span key={quest.id} className="summary-line">
              {quest.name} ({suitLabels[quest.suit] ?? "Unknown"})
            </span>
          ))}
        </div>
      ) : null}

      {Object.keys(hiddenHandLines).length > 0 || Object.keys(hiddenSupporterLines).length > 0 ? (
        <div className="summary-stack">
          <span className="summary-label">Hidden Placeholders</span>
          {Object.entries(hiddenHandLines).map(([faction, count]) => (
            <span key={faction} className="summary-line">
              {factionLabels[Number(faction)] ?? `Faction ${faction}`}: hand {count}
            </span>
          ))}
          {Object.entries(hiddenSupporterLines).map(([faction, count]) => (
            <span key={`supporters-${faction}`} className="summary-line">
              {factionLabels[Number(faction)] ?? `Faction ${faction}`}: supporters {count}
            </span>
          ))}
        </div>
      ) : null}
    </section>
  );
}
