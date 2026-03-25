import type { GameState } from "./types";

export const sampleState: GameState = {
  map: {
    id: "autumn",
    clearings: [
      { id: 1, suit: 0, buildSlots: 1, adj: [5, 10, 9], ruins: false, ruinItems: [], wood: 1, warriors: { "0": 3 }, buildings: [{ faction: 0, type: 0 }], tokens: [{ faction: 0, type: 0 }] },
      { id: 2, suit: 2, buildSlots: 2, adj: [5, 10, 6], ruins: false, ruinItems: [], wood: 0, warriors: { "0": 2 }, buildings: [{ faction: 0, type: 2 }], tokens: [] },
      { id: 3, suit: 1, buildSlots: 1, adj: [7, 11, 6], ruins: false, ruinItems: [], wood: 0, warriors: { "2": 3 }, buildings: [{ faction: 2, type: 3 }], tokens: [] },
      { id: 4, suit: 1, buildSlots: 1, adj: [9, 12, 8], ruins: false, ruinItems: [], wood: 0, warriors: {}, buildings: [], tokens: [{ faction: 1, type: 1 }] },
      { id: 5, suit: 1, buildSlots: 2, adj: [1, 2], ruins: false, ruinItems: [], wood: 0, warriors: {}, buildings: [], tokens: [] },
      { id: 6, suit: 0, buildSlots: 2, adj: [2, 11, 3], ruins: true, ruinItems: [1], wood: 0, warriors: { "0": 1 }, buildings: [], tokens: [] },
      { id: 7, suit: 2, buildSlots: 2, adj: [3, 12, 8], ruins: false, ruinItems: [], wood: 0, warriors: {}, buildings: [], tokens: [] },
      { id: 8, suit: 0, buildSlots: 2, adj: [7, 4], ruins: false, ruinItems: [], wood: 0, warriors: {}, buildings: [], tokens: [] },
      { id: 9, suit: 2, buildSlots: 2, adj: [1, 12, 4], ruins: false, ruinItems: [], wood: 1, warriors: { "0": 1 }, buildings: [{ faction: 0, type: 1 }], tokens: [] },
      { id: 10, suit: 1, buildSlots: 2, adj: [1, 2, 12], ruins: true, ruinItems: [3], wood: 0, warriors: {}, buildings: [], tokens: [] },
      { id: 11, suit: 2, buildSlots: 3, adj: [6, 3, 12], ruins: true, ruinItems: [7], wood: 0, warriors: { "1": 2 }, buildings: [{ faction: 1, type: 4 }], tokens: [] },
      { id: 12, suit: 0, buildSlots: 2, adj: [4, 9, 10, 11, 7], ruins: true, ruinItems: [3], wood: 0, warriors: {}, buildings: [], tokens: [] }
    ],
    forests: [
      { id: 1, adjacentClearings: [1, 5] },
      { id: 2, adjacentClearings: [5, 2, 10] },
      { id: 3, adjacentClearings: [2, 6, 11, 3] },
      { id: 4, adjacentClearings: [3, 7, 8] },
      { id: 5, adjacentClearings: [8, 4, 12, 7] },
      { id: 6, adjacentClearings: [4, 9, 1] },
      { id: 7, adjacentClearings: [1, 10, 12, 11, 6, 2] }
    ]
  },
  gameMode: 0,
  gamePhase: 1,
  playerFaction: 3,
  winner: 0,
  roundNumber: 3,
  factionTurn: 3,
  currentPhase: 1,
  currentStep: 3,
  turnOrder: [0, 2, 1, 3],
  deck: [12, 13, 14, 26, 27],
  discardPile: [7, 25, 36],
  itemSupply: {
    "0": 1,
    "1": 2,
    "2": 0,
    "3": 0,
    "4": 1,
    "5": 0,
    "6": 0,
    "7": 1
  },
  persistentEffects: {
    "0": [15],
    "2": [28]
  },
  questDeck: [2, 3, 4, 5],
  questDiscard: [6],
  otherHandCounts: {
    "0": 2,
    "1": 4,
    "2": 3
  },
  victoryPoints: {
    "0": 7,
    "1": 4,
    "2": 5,
    "3": 3
  },
  marquise: {
    cardsInHand: [],
    warriorSupply: 18,
    sawmillsPlaced: 1,
    workshopsPlaced: 1,
    recruitersPlaced: 1,
    keepClearingID: 1
  },
  eyrie: {
    cardsInHand: [],
    warriorSupply: 17,
    roostsPlaced: 1,
    leader: 2,
    availableLeaders: [0, 1, 2, 3],
    decree: {
      recruit: [-1],
      move: [-2],
      battle: [],
      build: []
    },
    craftedThisTurn: false
  },
  alliance: {
    cardsInHand: [],
    warriorSupply: 10,
    supporters: [],
    officers: 1,
    foxBasePlaced: false,
    rabbitBasePlaced: false,
    mouseBasePlaced: true,
    sympathyPlaced: 1
  },
  vagabond: {
    cardsInHand: [],
    character: 2,
    clearingID: 6,
    forestID: 0,
    inForest: false,
    items: [
      { type: 6, status: 0 },
      { type: 5, status: 0 },
      { type: 4, status: 0 },
      { type: 3, status: 0 },
      { type: 0, status: 1 }
    ],
    relationships: { "0": 1, "1": 2, "2": 1 },
    questsCompleted: [],
    questsAvailable: [
      { id: 1, name: "Expel Bandits", suit: 0, requiredItems: [5, 6] }
    ]
  },
  turnProgress: {
    actionsUsed: 0,
    bonusActions: 0,
    marchesUsed: 0,
    recruitUsed: false,
    usedWorkshopClearings: [],
    hasCrafted: false,
    decreeColumnsResolved: 0,
    decreeCardsResolved: 0,
    resolvedDecreeCardIDs: [],
    cardsAddedToDecree: 0,
    officerActionsUsed: 0,
    hasOrganized: false,
    hasSlipped: false
  }
};
