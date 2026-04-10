import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { sampleState } from "../sampleState";
import type { GameState } from "../types";
import { VPTracker } from "./VPTracker";

function trackerState(overrides: Partial<GameState> = {}): GameState {
  return {
    ...structuredClone(sampleState),
    ...overrides
  };
}

describe("VPTracker", () => {
  it("renders faction scores in turn order", () => {
    const state = trackerState({
      turnOrder: [2, 0, 1, 3],
      victoryPoints: { "0": 7, "1": 4, "2": 12, "3": 3 }
    });

    render(
      <VPTracker
        victoryPoints={state.victoryPoints}
        turnOrder={state.turnOrder}
        dominance={state.activeDominance}
        coalitionActive={state.coalitionActive}
        coalitionPartner={state.coalitionPartner}
      />
    );

    expect(screen.getByText("Eyrie")).toBeInTheDocument();
    expect(screen.getByText("12 / 30 VP")).toBeInTheDocument();
    expect(screen.getByText("7 / 30 VP")).toBeInTheDocument();
  });

  it("shows dominance and coalition indicators", () => {
    const state = trackerState({
      activeDominance: { "2": 14 },
      coalitionActive: true,
      coalitionPartner: 0
    });

    render(
      <VPTracker
        victoryPoints={state.victoryPoints}
        turnOrder={state.turnOrder}
        dominance={state.activeDominance}
        coalitionActive={state.coalitionActive}
        coalitionPartner={state.coalitionPartner}
      />
    );

    expect(screen.getByText(/Dominance:/i)).toBeInTheDocument();
    expect(screen.getByText("Coalition partner")).toBeInTheDocument();
    expect(screen.getByText("Coalition active")).toBeInTheDocument();
  });
});
