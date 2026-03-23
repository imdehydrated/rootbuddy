import type { GameState } from "./types";

export const sampleState: GameState = {
  map: {
    id: "autumn",
    clearings: [
      { id: 1, suit: 0, buildSlots: 1, adj: [5, 10, 9], ruins: false, wood: 0, warriors: {}, buildings: [] },
      { id: 2, suit: 2, buildSlots: 2, adj: [5, 10, 6], ruins: false, wood: 0, warriors: {}, buildings: [] },
      { id: 3, suit: 1, buildSlots: 1, adj: [7, 11, 6], ruins: false, wood: 0, warriors: {}, buildings: [] },
      { id: 4, suit: 1, buildSlots: 1, adj: [9, 12, 8], ruins: false, wood: 0, warriors: {}, buildings: [] },
      { id: 5, suit: 1, buildSlots: 2, adj: [1, 2], ruins: false, wood: 0, warriors: {}, buildings: [] },
      { id: 6, suit: 0, buildSlots: 2, adj: [2, 11, 3], ruins: true, wood: 0, warriors: {}, buildings: [] },
      { id: 7, suit: 2, buildSlots: 2, adj: [3, 12, 8], ruins: false, wood: 0, warriors: {}, buildings: [] },
      { id: 8, suit: 0, buildSlots: 2, adj: [7, 4], ruins: false, wood: 0, warriors: {}, buildings: [] },
      { id: 9, suit: 2, buildSlots: 2, adj: [1, 12, 4], ruins: false, wood: 0, warriors: {}, buildings: [] },
      { id: 10, suit: 1, buildSlots: 2, adj: [1, 2, 12], ruins: true, wood: 0, warriors: {}, buildings: [] },
      { id: 11, suit: 2, buildSlots: 3, adj: [6, 3, 12], ruins: true, wood: 0, warriors: {}, buildings: [] },
      { id: 12, suit: 0, buildSlots: 2, adj: [4, 9, 10, 11, 7], ruins: true, wood: 0, warriors: {}, buildings: [] }
    ]
  },
  factionTurn: 0,
  currentPhase: 0,
  currentStep: 0,
  marquise: {
    cardsInHand: [],
    warriorSupply: 25,
    sawmillsPlaced: 0,
    workshopsPlaced: 0,
    recruitersPlaced: 0,
    keepClearingID: 0
  },
  turnProgress: {
    recruitUsed: false,
    usedWorkshopClearings: []
  }
};
