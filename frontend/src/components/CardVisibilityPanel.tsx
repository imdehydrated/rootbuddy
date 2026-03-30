import { describeKnownCardIDs } from "../cardCatalog";
import { factionLabels, suitLabels } from "../labels";
import type { Card, GameState } from "../types";

type CardVisibilityPanelProps = {
  state: GameState;
};

function visibleHand(state: GameState): Card[] {
  switch (state.playerFaction) {
    case 0:
      return state.marquise.cardsInHand;
    case 1:
      return state.alliance.cardsInHand;
    case 2:
      return state.eyrie.cardsInHand;
    case 3:
      return state.vagabond.cardsInHand;
    default:
      return [];
  }
}

function hiddenSupporterCount(state: GameState, faction: number): number {
  return state.hiddenCards.filter((card) => card.ownerFaction === faction && card.zone === "supporters").length;
}

function formatCard(card: Card): string {
  return `${card.name} (${suitLabels[card.suit] ?? "Unknown"})`;
}

export function CardVisibilityPanel({ state }: CardVisibilityPanelProps) {
  const hand = visibleHand(state);
  const visibleSupporters = state.playerFaction === 1 ? state.alliance.supporters : [];
  const persistentEffectLines = Object.entries(state.persistentEffects).flatMap(([faction, cardIDs]) =>
    cardIDs.length === 0 ? [] : [`${factionLabels[Number(faction)] ?? `Faction ${faction}`}: ${describeKnownCardIDs(cardIDs)}`]
  );
  const hiddenHandLines = factionLabels
    .map((label, index) => ({ label, index, count: state.otherHandCounts[String(index)] ?? 0 }))
    .filter(({ index, count }) => index !== state.playerFaction && count > 0);
  const hiddenSupporterLines =
    state.playerFaction === 1
      ? []
      : factionLabels
          .map((label, index) => ({ label, index, count: hiddenSupporterCount(state, index) }))
          .filter(({ count }) => count > 0);

  return (
    <section className="panel floating-turn-summary">
      <p className="eyebrow">{state.gamePhase === 2 ? "Final Card State" : "Card Visibility"}</p>

      <div className="summary-stack">
        <span className="summary-label">Your Hand</span>
        {hand.length > 0 ? (
          <div className="card-pill-list">
            {hand.map((card) => (
              <span key={card.id} className="card-pill">
                {formatCard(card)}
              </span>
            ))}
          </div>
        ) : (
          <span className="summary-line">No visible cards in hand.</span>
        )}
      </div>

      {visibleSupporters.length > 0 ? (
        <div className="summary-stack">
          <span className="summary-label">Visible Supporters</span>
          <div className="card-pill-list">
            {visibleSupporters.map((card) => (
              <span key={card.id} className="card-pill">
                {formatCard(card)}
              </span>
            ))}
          </div>
        </div>
      ) : null}

      <div className="summary-grid">
        <div className="summary-card">
          <span className="summary-label">Deck</span>
          <strong>{state.deck.length}</strong>
          <span>{state.gameMode === 0 ? "Hidden cards remaining" : "Tracked card count"}</span>
        </div>
        <div className="summary-card">
          <span className="summary-label">Discard Pile</span>
          <strong>{state.discardPile.length}</strong>
          <span>{state.discardPile.length > 0 ? describeKnownCardIDs(state.discardPile) : "Empty"}</span>
        </div>
      </div>

      {persistentEffectLines.length > 0 || state.availableDominance.length > 0 ? (
        <div className="summary-stack">
          <span className="summary-label">Public Card Zones</span>
          {persistentEffectLines.map((line) => (
            <span key={line} className="summary-line">
              Effects {line}
            </span>
          ))}
          {state.availableDominance.length > 0 ? (
            <span className="summary-line">Available dominance: {describeKnownCardIDs(state.availableDominance)}</span>
          ) : null}
        </div>
      ) : null}

      {hiddenHandLines.length > 0 || hiddenSupporterLines.length > 0 || state.questDeck.length > 0 ? (
        <div className="summary-stack">
          <span className="summary-label">Hidden Counts</span>
          {hiddenHandLines.map(({ label, count }) => (
            <span key={`hand-${label}`} className="summary-line">
              {label}: hand {count}
            </span>
          ))}
          {hiddenSupporterLines.map(({ label, count }) => (
            <span key={`supporters-${label}`} className="summary-line">
              {label}: supporters {count}
            </span>
          ))}
          {state.questDeck.length > 0 ? <span className="summary-line">Quest deck: {state.questDeck.length} hidden</span> : null}
        </div>
      ) : null}
    </section>
  );
}
