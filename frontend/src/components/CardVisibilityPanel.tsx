import { describeKnownCardID } from "../cardCatalog";
import { factionLabels, itemTypeLabels, suitLabels } from "../labels";
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

function cardSuitClass(suit: number): string {
  switch (suit) {
    case 0:
      return "fox";
    case 1:
      return "rabbit";
    case 2:
      return "mouse";
    default:
      return "bird";
  }
}

function formatCraftCost(card: Card): string {
  const costs = [
    card.craftingCost.fox > 0 ? `${card.craftingCost.fox} fox` : null,
    card.craftingCost.rabbit > 0 ? `${card.craftingCost.rabbit} rabbit` : null,
    card.craftingCost.mouse > 0 ? `${card.craftingCost.mouse} mouse` : null,
    card.craftingCost.any > 0 ? `${card.craftingCost.any} any` : null
  ].filter((entry): entry is string => entry !== null);

  return costs.length > 0 ? costs.join(" / ") : "No craft cost";
}

function formatEffectLabel(effectID: string): string {
  switch (effectID) {
    case "better_burrow_bank":
      return "Better Burrow Bank";
    case "command_warren":
      return "Command Warren";
    case "codebreakers":
      return "Codebreakers";
    case "royal_claim":
      return "Royal Claim";
    case "scouting_party":
      return "Scouting Party";
    case "stand_and_deliver":
      return "Stand and Deliver!";
    case "tax_collector":
      return "Tax Collector";
    default:
      return effectID.replaceAll("_", " ");
  }
}

function CardTile({ card, zoneLabel }: { card: Card; zoneLabel?: string }) {
  return (
    <article className={`card-tile ${cardSuitClass(card.suit)}`}>
      <div className="card-tile-header">
        <span className={`card-suit-tag ${cardSuitClass(card.suit)}`}>{suitLabels[card.suit] ?? "Unknown"}</span>
        {zoneLabel ? <span className="card-zone-tag">{zoneLabel}</span> : null}
      </div>
      <strong className="card-title">{card.name}</strong>
      <div className="card-tile-meta">
        {card.effectID ? <span>Effect: {formatEffectLabel(card.effectID)}</span> : null}
        {card.craftedItem !== null ? <span>Item reward: {itemTypeLabels[card.craftedItem] ?? `Item ${card.craftedItem}`}</span> : null}
        {card.vp > 0 ? <span>{card.vp} VP</span> : null}
      </div>
      <span className="card-cost-line">{formatCraftCost(card)}</span>
    </article>
  );
}

export function CardVisibilityPanel({ state }: CardVisibilityPanelProps) {
  const hand = visibleHand(state);
  const visibleSupporters = state.playerFaction === 1 ? state.alliance.supporters : [];
  const persistentEffects = Object.entries(state.persistentEffects)
    .map(([faction, cardIDs]) => ({
      faction: Number(faction),
      cardIDs
    }))
    .filter(({ cardIDs }) => cardIDs.length > 0);
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
          <div className="card-grid">
            {hand.map((card) => (
              <CardTile key={card.id} card={card} zoneLabel="Hand" />
            ))}
          </div>
        ) : (
          <span className="summary-line">No visible cards in hand.</span>
        )}
      </div>

      {visibleSupporters.length > 0 ? (
        <div className="summary-stack">
          <span className="summary-label">Visible Supporters</span>
          <div className="card-grid">
            {visibleSupporters.map((card) => (
              <CardTile key={card.id} card={card} zoneLabel="Supporter" />
            ))}
          </div>
        </div>
      ) : null}

      <div className="summary-grid card-zone-summary-grid">
        <div className="summary-card card-zone-summary-card">
          <span className="summary-label">Deck</span>
          <strong>{state.deck.length}</strong>
          <span>{state.gameMode === 0 ? "Hidden cards remaining" : "Tracked card count"}</span>
        </div>
        <div className="summary-card card-zone-summary-card">
          <span className="summary-label">Discard Pile</span>
          <strong>{state.discardPile.length}</strong>
          {state.discardPile.length > 0 ? (
            <div className="known-card-pill-list">
              {state.discardPile.map((cardID) => (
                <span key={`discard-${cardID}`} className="known-card-pill">
                  {describeKnownCardID(cardID)}
                </span>
              ))}
            </div>
          ) : (
            <span>Empty</span>
          )}
        </div>
      </div>

      {persistentEffects.length > 0 || state.availableDominance.length > 0 ? (
        <div className="summary-stack">
          <span className="summary-label">Public Card Zones</span>
          {persistentEffects.map(({ faction, cardIDs }) => (
            <div key={`effects-${faction}`} className="card-zone-row">
              <span className="summary-line card-zone-row-label">
                {factionLabels[faction] ?? `Faction ${faction}`} effects
              </span>
              <div className="known-card-pill-list">
                {cardIDs.map((cardID) => (
                  <span key={`effect-${faction}-${cardID}`} className="known-card-pill">
                    {describeKnownCardID(cardID)}
                  </span>
                ))}
              </div>
            </div>
          ))}
          {state.availableDominance.length > 0 ? (
            <div className="card-zone-row">
              <span className="summary-line card-zone-row-label">Available dominance</span>
              <div className="known-card-pill-list">
                {state.availableDominance.map((cardID) => (
                  <span key={`dominance-${cardID}`} className="known-card-pill">
                    {describeKnownCardID(cardID)}
                  </span>
                ))}
              </div>
            </div>
          ) : null}
        </div>
      ) : null}

      {hiddenHandLines.length > 0 || hiddenSupporterLines.length > 0 || state.questDeck.length > 0 ? (
        <div className="summary-stack">
          <span className="summary-label">Hidden Counts</span>
          <div className="card-count-grid">
            {hiddenHandLines.map(({ label, count }) => (
              <div key={`hand-${label}`} className="card-count-card">
                <span className="summary-label">{label}</span>
                <strong>{count}</strong>
                <span>Hidden hand cards</span>
              </div>
            ))}
            {hiddenSupporterLines.map(({ label, count }) => (
              <div key={`supporters-${label}`} className="card-count-card">
                <span className="summary-label">{label}</span>
                <strong>{count}</strong>
                <span>Hidden supporters</span>
              </div>
            ))}
            {state.questDeck.length > 0 ? (
              <div className="card-count-card">
                <span className="summary-label">Quest Deck</span>
                <strong>{state.questDeck.length}</strong>
                <span>Hidden quests</span>
              </div>
            ) : null}
          </div>
        </div>
      ) : null}
    </section>
  );
}
