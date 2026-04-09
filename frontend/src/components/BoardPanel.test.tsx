import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { boardLayoutForState } from "../boardLayouts";
import { sampleState } from "../sampleState";
import { BoardPanel } from "./BoardPanel";

function renderBoard() {
  const state = structuredClone(sampleState);
  return render(
    <BoardPanel
      state={state}
      clearings={state.map.clearings}
      forests={state.map.forests}
      boardLayout={boardLayoutForState(state)}
      selectedClearingID={state.map.clearings[0]?.id ?? 1}
      keepClearingID={state.marquise.keepClearingID}
      vagabondClearingID={state.vagabond.clearingID}
      vagabondInForest={state.vagabond.inForest}
      onSelectClearing={vi.fn()}
      onSelectForest={vi.fn()}
    />
  );
}

describe("BoardPanel", () => {
  it("zooms the board view with the mouse wheel", () => {
    renderBoard();

    const canvas = screen.getByTestId("board-canvas");
    const boardView = screen.getByTestId("board-view");

    fireEvent.wheel(canvas, { deltaY: -100 });

    expect(boardView).toHaveStyle({ transform: "translate(0px, 0px) scale(1.12)" });
  });

  it("pans the board view by dragging empty board space", () => {
    renderBoard();

    const canvas = screen.getByTestId("board-canvas");
    const boardView = screen.getByTestId("board-view");

    fireEvent.mouseDown(canvas, { button: 0, clientX: 100, clientY: 80 });
    fireEvent.mouseMove(canvas, { clientX: 150, clientY: 120 });
    fireEvent.mouseUp(canvas);

    expect(boardView).toHaveStyle({ transform: "translate(50px, 40px) scale(1)" });
  });
});
