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
  imagePath: "/autumn-board.svg",
  clearingPositions: {
    1: { left: 14, top: 18 },
    5: { left: 56, top: 16 },
    2: { left: 85, top: 28 },
    9: { left: 15, top: 40 },
    10: { left: 40, top: 31 },
    11: { left: 63, top: 49 },
    6: { left: 85, top: 49 },
    4: { left: 14, top: 71 },
    12: { left: 39, top: 58 },
    7: { left: 60, top: 72 },
    8: { left: 39, top: 84 },
    3: { left: 84, top: 78 }
  },
  forestPositions: {
    1: { left: 35, top: 11 },
    2: { left: 71, top: 19 },
    3: { left: 75, top: 39 },
    4: { left: 72, top: 66 },
    5: { left: 34, top: 78 },
    6: { left: 11, top: 54 },
    7: { left: 48, top: 43 }
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
