import { factionLabels } from "../labels";
import { visibleHand } from "../cardPresentation";
import type { GameState } from "../types";
import { CardComponent } from "./CardComponent";

type CardHandTrayProps = {
  state: GameState;
  compactCards?: boolean;
};

export function CardHandTray({ state, compactCards = false }: CardHandTrayProps) {
  if (state.gamePhase !== 1) {
    return null;
  }

  const hand = visibleHand(state);
  const perspectiveLabel = factionLabels[state.playerFaction] ?? "Current";

  return (
    <section className="card-hand-tray" aria-label="Current hand tray">
      <div className="card-hand-tray-header">
        <div className="summary-stack">
          <span className="summary-label">Hand Tray</span>
          <strong>{perspectiveLabel} Hand</strong>
          <span className="summary-line">{hand.length} visible card{hand.length === 1 ? "" : "s"}</span>
        </div>
      </div>
      {hand.length > 0 ? (
        <div className="card-hand-strip">
          {hand.map((card) => (
            <div key={card.id} className="card-hand-item">
              <CardComponent card={card} zoneLabel="Hand" compact={compactCards} />
            </div>
          ))}
        </div>
      ) : (
        <div className="card-hand-empty">
          <span className="summary-line">No visible cards in hand.</span>
        </div>
      )}
    </section>
  );
}
