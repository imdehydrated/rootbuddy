export interface Building {
  faction: number;
  type: number;
}

export interface Token {
  faction: number;
  type: number;
}

export interface Decree {
  recruit: number[];
  move: number[];
  battle: number[];
  build: number[];
}

export interface Forest {
  id: number;
  adjacentClearings: number[];
}

export interface Item {
  type: number;
  status: number;
}

export interface Quest {
  id: number;
  name: string;
  suit: number;
  requiredItems: number[];
}

export interface HiddenCard {
  id: number;
  ownerFaction: number;
  zone: string;
  knownCardID: number;
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
  craftedItem: number | null;
  effectID: string;
  vp: number;
}

export interface EffectResult {
  effectID: string;
  message: string;
  cards: Card[];
}

export interface ActionLogEntry {
  roundNumber: number;
  faction: number;
  actionType: number;
  summary: string;
  timestamp: number;
}

export interface Clearing {
  id: number;
  suit: number;
  buildSlots: number;
  adj: number[];
  ruins: boolean;
  ruinItems: number[];
  wood: number;
  warriors: Record<string, number>;
  buildings: Building[];
  tokens: Token[];
}

export interface GameState {
  map: {
    id: string;
    clearings: Clearing[];
    forests: Forest[];
  };
  gameMode: number;
  randomSeed: number;
  shuffleCount: number;
  gamePhase: number;
  setupStage: number;
  playerFaction: number;
  winner: number;
  winningCoalition: number[];
  roundNumber: number;
  factionTurn: number;
  currentPhase: number;
  currentStep: number;
  turnOrder: number[];
  victoryPoints: Record<string, number>;
  deck: number[];
  discardPile: number[];
  availableDominance: number[];
  activeDominance: Record<string, number>;
  coalitionActive: boolean;
  coalitionPartner: number;
  itemSupply: Record<string, number>;
  persistentEffects: Record<string, number[]>;
  questDeck: number[];
  questDiscard: number[];
  otherHandCounts: Record<string, number>;
  hiddenCards: HiddenCard[];
  nextHiddenCardID: number;
  marquise: {
    cardsInHand: Card[];
    warriorSupply: number;
    sawmillsPlaced: number;
    workshopsPlaced: number;
    recruitersPlaced: number;
    keepClearingID: number;
  };
  eyrie: {
    cardsInHand: Card[];
    warriorSupply: number;
    roostsPlaced: number;
    leader: number;
    availableLeaders: number[];
    decree: Decree;
    craftedThisTurn: boolean;
  };
  alliance: {
    cardsInHand: Card[];
    warriorSupply: number;
    supporters: Card[];
    officers: number;
    foxBasePlaced: boolean;
    rabbitBasePlaced: boolean;
    mouseBasePlaced: boolean;
    sympathyPlaced: number;
  };
  vagabond: {
    cardsInHand: Card[];
    character: number;
    clearingID: number;
    forestID: number;
    inForest: boolean;
    items: Item[];
    relationships: Record<string, number>;
    questsCompleted: Quest[];
    questsAvailable: Quest[];
  };
  turnProgress: {
    actionsUsed: number;
    bonusActions: number;
    marchesUsed: number;
    recruitUsed: boolean;
    usedWorkshopClearings: number[];
    hasCrafted: boolean;
    decreeColumnsResolved: number;
    decreeCardsResolved: number;
    resolvedDecreeCardIDs: number[];
    cardsAddedToDecree: number;
    officerActionsUsed: number;
    hasOrganized: boolean;
    hasSlipped: boolean;
    usedPersistentEffectIDs: string[];
    birdsongMainActionTaken: boolean;
    daylightMainActionTaken: boolean;
    eveningMainActionTaken: boolean;
  };
}

export interface SetupRequest {
  gameMode: number;
  playerFaction: number;
  factions: number[];
  mapID: string;
  vagabondCharacter: number;
  eyrieLeader: number;
  randomSeed?: number;
}

export interface Action {
  type: number;
  movement?: {
    faction: number;
    count: number;
    maxCount: number;
    from: number;
    to: number;
    fromForestID: number;
    toForestID: number;
    decreeCardID: number;
    sourceEffectID: string;
  } | null;
  battle?: {
    faction: number;
    clearingID: number;
    targetFaction: number;
    decreeCardID: number;
    sourceEffectID: string;
  } | null;
  battleResolution?: {
    faction: number;
    clearingID: number;
    targetFaction: number;
    decreeCardID: number;
    attackerRoll: number;
    defenderRoll: number;
    attackerHitModifier: number;
    defenderHitModifier: number;
    ignoreHitsToAttacker: boolean;
    ignoreHitsToDefender: boolean;
    defenderAmbushed: boolean;
    attackerCounterAmbush: boolean;
    attackerUsedArmorers: boolean;
    defenderUsedArmorers: boolean;
    attackerUsedBrutalTactics: boolean;
    defenderUsedSappers: boolean;
    ambushHitsToAttacker: number;
    attackerLosses: number;
    defenderLosses: number;
    sourceEffectID: string;
  } | null;
  build?: {
    faction: number;
    clearingID: number;
    buildingType: number;
    woodSources: Array<{
      clearingID: number;
      amount: number;
    }>;
    decreeCardID: number;
  } | null;
  recruit?: {
    faction: number;
    clearingIDs: number[];
    decreeCardID: number;
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
  addToDecree?: {
    faction: number;
    cardIDs: number[];
    columns: number[];
  } | null;
  spreadSympathy?: {
    faction: number;
    clearingID: number;
    supporterCardIDs: number[];
  } | null;
  revolt?: {
    faction: number;
    clearingID: number;
    baseSuit: number;
    supporterCardIDs: number[];
  } | null;
  mobilize?: {
    faction: number;
    cardID: number;
  } | null;
  train?: {
    faction: number;
    cardID: number;
  } | null;
  organize?: {
    faction: number;
    clearingID: number;
  } | null;
  explore?: {
    faction: number;
    clearingID: number;
    itemType: number;
  } | null;
  quest?: {
    faction: number;
    questID: number;
    itemIndexes: number[];
    reward: number;
  } | null;
  aid?: {
    faction: number;
    targetFaction: number;
    clearingID: number;
    cardID: number;
  } | null;
  strike?: {
    faction: number;
    clearingID: number;
    targetFaction: number;
  } | null;
  repair?: {
    faction: number;
    itemIndex: number;
  } | null;
  turmoil?: {
    faction: number;
    newLeader: number;
  } | null;
  daybreak?: {
    faction: number;
    refreshedItemIndexes: number[];
  } | null;
  slip?: {
    faction: number;
    from: number;
    to: number;
    fromForestID: number;
    toForestID: number;
  } | null;
  birdsongWood?: {
    faction: number;
    clearingIDs: number[];
    amount: number;
  } | null;
  eveningDraw?: {
    faction: number;
    count: number;
  } | null;
  scoreRoosts?: {
    faction: number;
    points: number;
  } | null;
  passPhase?: {
    faction: number;
  } | null;
  addCardToHand?: {
    faction: number;
    cardID: number;
  } | null;
  removeCardFromHand?: {
    faction: number;
    cardID: number;
  } | null;
  otherPlayerDraw?: {
    faction: number;
    count: number;
  } | null;
  otherPlayerPlay?: {
    faction: number;
    cardID: number;
  } | null;
  discardEffect?: {
    faction: number;
    cardID: number;
  } | null;
  activateDominance?: {
    faction: number;
    cardID: number;
    targetFaction: number;
  } | null;
  takeDominance?: {
    faction: number;
    dominanceCardID: number;
    spentCardID: number;
  } | null;
  marquiseSetup?: {
    faction: number;
    keepClearingID: number;
    sawmillClearingID: number;
    workshopClearingID: number;
    recruiterClearingID: number;
  } | null;
  eyrieSetup?: {
    faction: number;
    clearingID: number;
  } | null;
  vagabondSetup?: {
    faction: number;
    forestID: number;
  } | null;
  usePersistentEffect?: {
    faction: number;
    effectID: string;
    targetFaction: number;
    clearingID: number;
    observedCardID: number;
  } | null;
}

export interface BattleModifiers {
  attackerHitModifier: number;
  defenderHitModifier: number;
  ignoreHitsToAttacker: boolean;
  ignoreHitsToDefender: boolean;
  defenderAmbush: boolean;
  attackerCounterAmbush: boolean;
  attackerUsesArmorers: boolean;
  defenderUsesArmorers: boolean;
  attackerUsesBrutalTactics: boolean;
  defenderUsesSappers: boolean;
}

export interface BattleContext {
  action: Action;
  clearingSuit: number;
  timing: string[];
  attackerHasScoutingParty: boolean;
  canDefenderAmbush: boolean;
  canAttackerCounterAmbush: boolean;
  canAttackerArmorers: boolean;
  canDefenderArmorers: boolean;
  canAttackerBrutalTactics: boolean;
  canDefenderSappers: boolean;
  assistDefenderAmbushPromptRequired: boolean;
}

export interface BattlePrompt {
  gameID: string;
  revision: number;
  action: Action;
  stage: "waiting_defender" | "defender_response" | "ready_to_resolve";
  waitingOnFaction: number;
  battleContext: BattleContext;
  canUseAmbush?: boolean;
  canUseArmorers?: boolean;
  canUseSappers?: boolean;
  defenderAmbush?: boolean;
  defenderUsedArmorers?: boolean;
  defenderUsedSappers?: boolean;
}
