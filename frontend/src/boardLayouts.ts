import type { GameState } from "./types";

export interface BoardPosition {
  left: number;
  top: number;
}

export interface BoardLayout {
  id: string;
  label: string;
  imagePath: string;
  clearingPositions: Record<number, BoardPosition>;
  forestPositions: Record<number, BoardPosition>;
}

const autumnBoardLayout: BoardLayout = {
  id: "autumn",
  label: "Autumn",
  imagePath: "/autumn-board-ai.jpg",
  clearingPositions: {
    1: { left: 8.4, top: 18.2 },
    5: { left: 50.2, top: 15.8 },
    2: { left: 86.8, top: 28.2 },
    9: { left: 9, top: 43 },
    10: { left: 42.8, top: 32 },
    11: { left: 64.5, top: 50 },
    6: { left: 82.8, top: 42.8 },
    4: { left: 12, top: 73 },
    12: { left: 33.8, top: 60 },
    7: { left: 61, top: 79 },
    8: { left: 23.5, top: 90 },
    3: { left: 85, top: 86 }
  },
  forestPositions: {
    1: { left: 29.2, top: 13.2 },
    2: { left: 71.6, top: 21.7 },
    3: { left: 78.7, top: 56.7 },
    4: { left: 53.5, top: 82.9 },
    5: { left: 25.6, top: 73.8 },
    6: { left: 7.4, top: 57.8 },
    7: { left: 50.6, top: 46.7 }
  }
};

export function boardLayoutForState(state: GameState): BoardLayout {
  if (state.map.id === autumnBoardLayout.id) {
    return autumnBoardLayout;
  }

  if (state.map.clearings.length === 12) {
    return autumnBoardLayout;
  }

  return autumnBoardLayout;
}
