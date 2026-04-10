import { cardSuitClass, formatCardEffectLabel, formatCardReward, formatCraftCost } from "../cardPresentation";
import { suitLabels } from "../labels";
import type { Card } from "../types";

type CardComponentProps = {
  card: Card;
  compact?: boolean;
  selected?: boolean;
  zoneLabel?: string;
  onSelect?: (card: Card) => void;
};

function CardBody({ card, zoneLabel }: { card: Card; zoneLabel?: string }) {
  const itemReward = formatCardReward(card);

  return (
    <>
      <span className={`card-component-accent ${cardSuitClass(card.suit)}`} aria-hidden="true" />
      <div className="card-component-body">
        <div className="card-component-header">
          <span className={`card-suit-tag ${cardSuitClass(card.suit)}`}>{suitLabels[card.suit] ?? "Unknown"}</span>
          {zoneLabel ? <span className="card-zone-tag">{zoneLabel}</span> : null}
        </div>
        <strong className="card-title">{card.name}</strong>
        <div className="card-component-meta">
          {card.effectID ? <span>{formatCardEffectLabel(card.effectID)}</span> : null}
          {itemReward ? <span>{itemReward}</span> : null}
          {card.vp > 0 ? <span>{card.vp} VP</span> : null}
        </div>
        <span className="card-cost-line">{formatCraftCost(card)}</span>
      </div>
    </>
  );
}

export function CardComponent({
  card,
  compact = false,
  selected = false,
  zoneLabel,
  onSelect
}: CardComponentProps) {
  const className = `card-component ${cardSuitClass(card.suit)} ${compact ? "compact" : ""} ${selected ? "selected" : ""}`.trim();

  if (onSelect) {
    return (
      <button type="button" className={className} aria-pressed={selected} onClick={() => onSelect(card)}>
        <CardBody card={card} zoneLabel={zoneLabel} />
      </button>
    );
  }

  return (
    <article className={className}>
      <CardBody card={card} zoneLabel={zoneLabel} />
    </article>
  );
}
