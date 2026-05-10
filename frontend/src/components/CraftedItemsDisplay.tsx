import { factionLabels, itemTypeLabels } from "../labels";
import type { GameState } from "../types";

type CraftedItemsDisplayProps = {
  state: GameState;
};

const craftingFactions = [0, 1, 2];

export function CraftedItemsDisplay({ state }: CraftedItemsDisplayProps) {
  return (
    <div className="crafted-items-board" aria-label="Crafted items">
      {craftingFactions.map((faction) => {
        const items = state.craftedItems[String(faction)] ?? [];

        return (
          <div className="crafted-items-row" key={faction}>
            <div className="crafted-items-faction">
              <span>{factionLabels[faction] ?? `Faction ${faction}`}</span>
              <small>Crafted Items</small>
            </div>
            <div className="crafted-items-slots">
              {items.length > 0 ? (
                items.map((itemType, index) => (
                  <span className="crafted-item-token" key={`${faction}-${itemType}-${index}`} title={`${itemTypeLabels[itemType] ?? "Unknown"} item`}>
                    <span className="crafted-item-icon">{itemTypeLabels[itemType]?.slice(0, 1) ?? "?"}</span>
                    <span>{itemTypeLabels[itemType] ?? `Item ${itemType}`}</span>
                  </span>
                ))
              ) : (
                <span className="crafted-items-empty">No crafted items</span>
              )}
            </div>
          </div>
        );
      })}
    </div>
  );
}
