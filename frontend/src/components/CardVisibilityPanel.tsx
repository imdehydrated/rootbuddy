import { describeKnownCardID } from "../cardCatalog";
import { factionLabels } from "../labels";
import { visibleHand } from "../cardPresentation";
import type { GameState } from "../types";
import { CardComponent } from "./CardComponent";
import { KnownCardPillList } from "./CardUi";

type CardVisibilityPanelProps = {
  state: GameState;
};

function hiddenSupporterCount(state: GameState, faction: number): number {
  return state.hiddenCards.filter((card) => card.ownerFaction === faction && card.zone === "supporters").length;
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
              <CardComponent key={card.id} card={card} zoneLabel="Hand" />
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
              <CardComponent key={card.id} card={card} zoneLabel="Supporter" />
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
            <KnownCardPillList
              items={state.discardPile.map((cardID) => ({
                key: `discard-${cardID}`,
                label: describeKnownCardID(cardID)
              }))}
            />
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
              <KnownCardPillList
                items={cardIDs.map((cardID) => ({
                  key: `effect-${faction}-${cardID}`,
                  label: describeKnownCardID(cardID)
                }))}
              />
            </div>
          ))}
          {state.availableDominance.length > 0 ? (
            <div className="card-zone-row">
              <span className="summary-line card-zone-row-label">Available dominance</span>
              <KnownCardPillList
                items={state.availableDominance.map((cardID) => ({
                  key: `dominance-${cardID}`,
                  label: describeKnownCardID(cardID)
                }))}
              />
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
