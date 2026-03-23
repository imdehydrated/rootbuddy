export interface Building {
  faction: number;
  type: number;
}

export type HighlightedClearing = {
  clearingID: number;
  role: "source" | "target" | "affected";
};

export interface Card {
  id: number;
  deck: number;
  name: string;
  suit: number;
  kind: number;
  craftingCost: {
    fox: number;
    rabbit: number;
    mouse: number;
    any: number;
  };
  vp: number;
}

export interface Clearing {
  id: number;
  suit: number;
  buildSlots: number;
  adj: number[];
  ruins: boolean;
  wood: number;
  warriors: Record<string, number>;
  buildings: Building[];
}

export interface GameState {
  map: {
    id: string;
    clearings: Clearing[];
  };
  factionTurn: number;
  currentPhase: number;
  currentStep: number;
  marquise: {
    cardsInHand: Card[];
    warriorSupply: number;
    sawmillsPlaced: number;
    workshopsPlaced: number;
    recruitersPlaced: number;
    keepClearingID: number;
  };
  turnProgress: {
    recruitUsed: boolean;
    usedWorkshopClearings: number[];
  };
}

export interface Action {
  type: number;
  movement?: {
    faction: number;
    maxCount: number;
    from: number;
    to: number;
  } | null;
  battle?: {
    faction: number;
    clearingID: number;
    targetFaction: number;
  } | null;
  battleResolution?: {
    faction: number;
    clearingID: number;
    targetFaction: number;
    attackerRoll: number;
    defenderRoll: number;
    attackerHitModifier: number;
    defenderHitModifier: number;
    ignoreHitsToAttacker: boolean;
    ignoreHitsToDefender: boolean;
    attackerLosses: number;
    defenderLosses: number;
  } | null;
  build?: {
    faction: number;
    clearingID: number;
    buildingType: number;
  } | null;
  recruit?: {
    faction: number;
    clearingIDs: number[];
  } | null;
  overwork?: {
    faction: number;
    clearingID: number;
    cardID: number;
  } | null;
  craft?: {
    faction: number;
    cardID: number;
    usedWorkshopClearings: number[];
  } | null;
}
