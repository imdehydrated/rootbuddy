import { itemTypeLabels } from "./labels";
import type { Card, GameState } from "./types";

export function visibleHand(state: GameState): Card[] {
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

export function cardSuitClass(suit: number): string {
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

export function formatCraftCost(card: Card): string {
  const costs = [
    card.craftingCost.fox > 0 ? `${card.craftingCost.fox} fox` : null,
    card.craftingCost.rabbit > 0 ? `${card.craftingCost.rabbit} rabbit` : null,
    card.craftingCost.mouse > 0 ? `${card.craftingCost.mouse} mouse` : null,
    card.craftingCost.any > 0 ? `${card.craftingCost.any} any` : null
  ].filter((entry): entry is string => entry !== null);

  return costs.length > 0 ? costs.join(" / ") : "No craft cost";
}

export function formatCardEffectLabel(effectID: string): string {
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

export function formatCardReward(card: Card): string | null {
  return card.craftedItem !== null ? `Item reward: ${itemTypeLabels[card.craftedItem] ?? `Item ${card.craftedItem}`}` : null;
}
