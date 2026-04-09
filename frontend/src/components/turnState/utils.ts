import { describeKnownCardID } from "../../cardCatalog";
import { factionLabels, suitLabels } from "../../labels";
import type { Card } from "../../types";

export function parseNumberList(value: string): number[] {
  return value
    .split(",")
    .map((part) => part.trim())
    .filter((part) => part.length > 0)
    .map((part) => Number(part))
    .filter((entry) => Number.isInteger(entry));
}

export function formatNumberList(values: number[]): string {
  return values.join(", ");
}

export function isValidDecreeCardID(value: number): boolean {
  return value === -2 || value === -1 || value >= 1;
}

export function describeVisibleCard(card: Card): string {
  return `${card.name} (${suitLabels[card.suit] ?? "Unknown"})`;
}

export function duplicateValues(values: number[]): number[] {
  const seen = new Set<number>();
  const duplicates = new Set<number>();
  values.forEach((value) => {
    if (seen.has(value)) {
      duplicates.add(value);
      return;
    }
    seen.add(value);
  });
  return Array.from(duplicates);
}

export function formatFactionList(factions: number[]): string {
  return factions.map((faction) => factionLabels[faction] ?? `Faction ${faction}`).join(", ");
}

export function referenceItemsFromCardIDs(prefix: string, cardIDs: number[]) {
  return cardIDs.map((cardID, index) => ({
    key: `${prefix}-${cardID}-${index}`,
    label: describeKnownCardID(cardID)
  }));
}
