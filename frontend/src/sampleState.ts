import type { GameState } from "./types";

export const sampleState: GameState = {
  map: {
    id: "autumn",
    clearings: [
      { id: 1, suit: 0, buildSlots: 1, adj: [5, 10, 9], ruins: false, wood: 0, warriors: {}, buildings: [], tokens: [] },
      { id: 2, suit: 2, buildSlots: 2, adj: [5, 10, 6], ruins: false, wood: 0, warriors: {}, buildings: [], tokens: [] },
      { id: 3, suit: 1, buildSlots: 1, adj: [7, 11, 6], ruins: false, wood: 0, warriors: {}, buildings: [], tokens: [] },
      { id: 4, suit: 1, buildSlots: 1, adj: [9, 12, 8], ruins: false, wood: 0, warriors: {}, buildings: [], tokens: [] },
      { id: 5, suit: 1, buildSlots: 2, adj: [1, 2], ruins: false, wood: 0, warriors: {}, buildings: [], tokens: [] },
      { id: 6, suit: 0, buildSlots: 2, adj: [2, 11, 3], ruins: true, wood: 0, warriors: {}, buildings: [], tokens: [] },
      { id: 7, suit: 2, buildSlots: 2, adj: [3, 12, 8], ruins: false, wood: 0, warriors: {}, buildings: [], tokens: [] },
      { id: 8, suit: 0, buildSlots: 2, adj: [7, 4], ruins: false, wood: 0, warriors: {}, buildings: [], tokens: [] },
      { id: 9, suit: 2, buildSlots: 2, adj: [1, 12, 4], ruins: false, wood: 0, warriors: {}, buildings: [], tokens: [] },
      { id: 10, suit: 1, buildSlots: 2, adj: [1, 2, 12], ruins: true, wood: 0, warriors: {}, buildings: [], tokens: [] },
      { id: 11, suit: 2, buildSlots: 3, adj: [6, 3, 12], ruins: true, wood: 0, warriors: {}, buildings: [], tokens: [] },
      { id: 12, suit: 0, buildSlots: 2, adj: [4, 9, 10, 11, 7], ruins: true, wood: 0, warriors: {}, buildings: [], tokens: [] }
    ]
  },
  factionTurn: 0,
  currentPhase: 0,
  currentStep: 0,
  turnOrder: [0, 2, 1, 3],
  victoryPoints: {
    "0": 0,
    "1": 0,
    "2": 0,
    "3": 0
  },
  marquise: {
    cardsInHand: [],
    warriorSupply: 25,
    sawmillsPlaced: 0,
    workshopsPlaced: 0,
    recruitersPlaced: 0,
    keepClearingID: 0
  },
  eyrie: {
    cardsInHand: [],
    warriorSupply: 20,
    roostsPlaced: 0,
    leader: 0,
    availableLeaders: [0, 1, 2, 3],
    decree: {
      recruit: [],
      move: [],
      battle: [],
      build: []
    },
    craftedThisTurn: false
  },
  alliance: {
    cardsInHand: [],
    warriorSupply: 10,
    supporters: [],
    officers: 0,
    foxBasePlaced: false,
    rabbitBasePlaced: false,
    mouseBasePlaced: false,
    sympathyPlaced: 0
  },
  vagabond: {
    cardsInHand: [],
    character: 0,
    clearingID: 0,
    inForest: false,
    items: [],
    relationships: {},
    questsCompleted: [],
    questsAvailable: []
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
