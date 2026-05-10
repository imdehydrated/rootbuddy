import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { sampleState } from "../sampleState";
import type { GameState } from "../types";
import { CraftedItemsDisplay } from "./CraftedItemsDisplay";

function craftedItemsState(overrides: Partial<GameState> = {}): GameState {
  return {
    ...structuredClone(sampleState),
    ...overrides
  };
}

describe("CraftedItemsDisplay", () => {
  it("shows each non-Vagabond faction crafted item box", () => {
    const state = craftedItemsState({
      craftedItems: {
        "0": [6, 1],
        "1": [],
        "2": [4]
      }
    });

    render(<CraftedItemsDisplay state={state} />);

    expect(screen.getByText("Marquise")).toBeInTheDocument();
    expect(screen.getByText("Woodland Alliance")).toBeInTheDocument();
    expect(screen.getByText("Eyrie")).toBeInTheDocument();
    expect(screen.getByText("Boots")).toBeInTheDocument();
    expect(screen.getByText("Coin")).toBeInTheDocument();
    expect(screen.getByText("Sword")).toBeInTheDocument();
    expect(screen.getByText("No crafted items")).toBeInTheDocument();
  });
});
