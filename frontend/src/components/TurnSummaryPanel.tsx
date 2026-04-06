import { describeKnownCardID } from "../cardCatalog";
import {
  eyrieLeaderLabels,
  factionLabels,
  itemTypeLabels,
  itemStatusLabels,
  phaseLabels,
  relationshipLabels,
  setupStageLabels,
  stepLabels,
  suitLabels,
  vagabondCharacterLabels
} from "../labels";
import type { GameState } from "../types";
import { KnownCardPillList } from "./CardUi";

type TurnSummaryPanelProps = {
  state: GameState;
};

function formatTurnOrder(turnOrder: number[]): string {
  return turnOrder.map((faction) => factionLabels[faction] ?? `Faction ${faction}`).join(" -> ");
}

function hiddenCount(state: GameState, faction: number, zone: string): number {
  return state.hiddenCards.filter((card) => card.ownerFaction === faction && card.zone === zone).length;
}

function formatQuestItems(itemTypes: number[]): string {
  if (itemTypes.length === 0) {
    return "None";
  }

  return itemTypes.map((itemType) => itemTypeLabels[itemType] ?? `Item ${itemType}`).join(", ");
}

function renderCurrentFactionState(state: GameState) {
  switch (state.factionTurn) {
    case 0:
      return (
        <div className="faction-state-grid">
          <div className="faction-state-card">
            <span className="summary-label">Supply</span>
            <strong>{state.marquise.warriorSupply}</strong>
            <span className="summary-line">Warriors in reserve</span>
          </div>
          <div className="faction-state-card">
            <span className="summary-label">Buildings</span>
            <strong>{state.marquise.sawmillsPlaced}/{state.marquise.workshopsPlaced}/{state.marquise.recruitersPlaced}</strong>
            <span className="summary-line">Sawmill / Workshop / Recruiter</span>
          </div>
          <div className="faction-state-card">
            <span className="summary-label">Keep</span>
            <strong>{state.marquise.keepClearingID > 0 ? state.marquise.keepClearingID : "Unset"}</strong>
            <span className="summary-line">Home clearing</span>
          </div>
        </div>
      );
    case 1: {
      const visibleSupporters = state.playerFaction === 1 ? state.alliance.supporters : [];
      const supporterCount = visibleSupporters.length || hiddenCount(state, 1, "supporters");
      const baseLabels = [
        state.alliance.foxBasePlaced ? "Fox" : "",
        state.alliance.rabbitBasePlaced ? "Rabbit" : "",
        state.alliance.mouseBasePlaced ? "Mouse" : ""
      ].filter(Boolean);

      return (
        <div className="summary-stack">
          <div className="faction-state-grid">
            <div className="faction-state-card">
              <span className="summary-label">Officers</span>
              <strong>{state.alliance.officers}</strong>
              <span className="summary-line">Command actions</span>
            </div>
            <div className="faction-state-card">
              <span className="summary-label">Supporters</span>
              <strong>{supporterCount}</strong>
              <span className="summary-line">{visibleSupporters.length > 0 ? "Visible to Alliance" : "Hidden count"}</span>
            </div>
            <div className="faction-state-card">
              <span className="summary-label">Sympathy</span>
              <strong>{state.alliance.sympathyPlaced}</strong>
              <span className="summary-line">Tokens on map</span>
            </div>
          </div>
          <span className="summary-line">Bases: {baseLabels.join(", ") || "None"}</span>
          {visibleSupporters.length > 0 ? (
            <KnownCardPillList
              items={visibleSupporters.map((card) => ({
                key: `supporter-${card.id}`,
                label: `Supporter: ${card.name} (${suitLabels[card.suit] ?? "Unknown"})`
              }))}
            />
          ) : null}
        </div>
      );
    }
    case 2: {
      const decreeColumns = [
        { label: "Recruit", cards: state.eyrie.decree.recruit },
        { label: "Move", cards: state.eyrie.decree.move },
        { label: "Battle", cards: state.eyrie.decree.battle },
        { label: "Build", cards: state.eyrie.decree.build }
      ];
      return (
        <div className="summary-stack">
          <div className="faction-state-grid">
            <div className="faction-state-card">
              <span className="summary-label">Leader</span>
              <strong>{eyrieLeaderLabels[state.eyrie.leader] ?? "Unknown"}</strong>
              <span className="summary-line">Current leader</span>
            </div>
            <div className="faction-state-card">
              <span className="summary-label">Roosts</span>
              <strong>{state.eyrie.roostsPlaced}</strong>
              <span className="summary-line">Placed on map</span>
            </div>
          </div>
          <div className="decree-grid">
            {decreeColumns.map((column) => (
              <div key={column.label} className="decree-column-card">
                <span className="summary-label">{column.label}</span>
                {column.cards.length > 0 ? (
                  <KnownCardPillList
                    items={column.cards.map((cardID, index) => ({
                      key: `${column.label}-${cardID}-${index}`,
                      label: describeKnownCardID(cardID)
                    }))}
                  />
                ) : (
                  <span className="summary-line">No cards</span>
                )}
              </div>
            ))}
          </div>
        </div>
      );
    }
    case 3: {
      const itemSummary = [0, 1, 2].map((status) => ({
        key: `item-status-${status}`,
        label: `${itemStatusLabels[status]} ${state.vagabond.items.filter((item) => item.status === status).length}`
      }));
      const relationshipSummary = Object.entries(state.vagabond.relationships).map(([faction, level]) => ({
        key: `relationship-${faction}`,
        label: `${factionLabels[Number(faction)]}: ${relationshipLabels[level] ?? "Unknown"}`
      }));

      return (
        <div className="summary-stack">
          <div className="faction-state-grid">
            <div className="faction-state-card">
              <span className="summary-label">Character</span>
              <strong>{vagabondCharacterLabels[state.vagabond.character] ?? "Unknown"}</strong>
              <span className="summary-line">Vagabond role</span>
            </div>
            <div className="faction-state-card">
              <span className="summary-label">Location</span>
              <strong>{state.vagabond.inForest ? `Forest ${state.vagabond.forestID || "?"}` : `Clearing ${state.vagabond.clearingID || "?"}`}</strong>
              <span className="summary-line">{state.vagabond.inForest ? "In forest" : "On map"}</span>
            </div>
          </div>
          <KnownCardPillList items={itemSummary} />
          <KnownCardPillList items={relationshipSummary} emptyLabel="Relationships unset" />
        </div>
      );
    }
    default:
      return null;
  }
}

export function TurnSummaryPanel({ state }: TurnSummaryPanelProps) {
  const activeDominanceEntries = Object.entries(state.activeDominance).map(([faction, cardID]) => ({
    factionLabel: factionLabels[Number(faction)] ?? `Faction ${faction}`,
    cardLabel: describeKnownCardID(Number(cardID))
  }));
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

      <div className="summary-stack">
        <span className="summary-label">Current Faction State</span>
        {renderCurrentFactionState(state)}
      </div>

      {activeDominanceEntries.length > 0 || state.availableDominance.length > 0 || state.coalitionActive ? (
        <div className="summary-stack">
          <span className="summary-label">Dominance</span>
          {activeDominanceEntries.map(({ factionLabel, cardLabel }) => (
            <div key={`${factionLabel}-${cardLabel}`} className="card-zone-row">
              <span className="summary-line card-zone-row-label">Active {factionLabel}</span>
              <KnownCardPillList items={[{ key: `${factionLabel}-${cardLabel}`, label: cardLabel }]} />
            </div>
          ))}
          {state.availableDominance.length > 0 ? (
            <KnownCardPillList
              items={state.availableDominance.map((cardID) => ({
                key: `turn-dominance-${cardID}`,
                label: describeKnownCardID(cardID)
              }))}
            />
          ) : null}
          {state.coalitionActive ? <span className="summary-line">Coalition: {coalitionLabel}</span> : null}
        </div>
      ) : null}

      {state.factionTurn === 3 && state.vagabond.questsAvailable.length > 0 ? (
        <div className="summary-stack">
          <span className="summary-label">Available Quests</span>
          <div className="quest-summary-grid">
            {state.vagabond.questsAvailable.map((quest) => (
              <article key={quest.id} className={`quest-summary-card ${suitLabels[quest.suit]?.toLowerCase() ?? "bird"}`}>
                <span className="summary-label">{suitLabels[quest.suit] ?? "Unknown"}</span>
                <strong>{quest.name}</strong>
                <span className="summary-line">Items: {formatQuestItems(quest.requiredItems)}</span>
              </article>
            ))}
          </div>
        </div>
      ) : null}
    </section>
  );
}
