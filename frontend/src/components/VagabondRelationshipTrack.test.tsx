import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { sampleState } from "../sampleState";
import type { GameState } from "../types";
import { VagabondRelationshipTrack } from "./VagabondRelationshipTrack";

function relationshipState(overrides: Partial<GameState> = {}): GameState {
  return {
    ...structuredClone(sampleState),
    ...overrides
  };
}

describe("VagabondRelationshipTrack", () => {
  it("shows current relationship and Aid progress toward the next space", () => {
    const state = relationshipState();
    state.vagabond.relationships = { "0": 2, "1": 0, "2": 4 };
    state.turnProgress.vagabondAidCounts = { "0": 1 };
    state.vagabond.clearingID = 6;
    state.vagabond.inForest = false;
    const clearing = state.map.clearings.find((candidate) => candidate.id === 6);
    if (clearing) {
      clearing.warriors = { ...clearing.warriors, "2": 2 };
    }

    render(<VagabondRelationshipTrack state={state} />);

    expect(screen.getByText("Marquise")).toBeInTheDocument();
    expect(screen.getByText("1/2 Aid this turn toward Friendly.")).toBeInTheDocument();
    expect(screen.getByText("Hostile Aid allowed; relationship stays Hostile.")).toBeInTheDocument();
    expect(screen.getByText("Allied Aid scores 2 VP.")).toBeInTheDocument();
    expect(screen.getByText("2")).toBeInTheDocument();
    expect(screen.getByText("Move or battle with this ally")).toBeInTheDocument();
  });

  it("allows relationship editing when enabled", () => {
    const state = relationshipState();
    state.vagabond.relationships = { "0": 1, "1": 1, "2": 1 };
    const onSetRelationship = vi.fn();

    render(<VagabondRelationshipTrack state={state} editable onSetRelationship={onSetRelationship} />);

    fireEvent.click(screen.getAllByRole("button", { name: /Amiable/i })[0]);

    expect(onSetRelationship).toHaveBeenCalledWith(0, 2);
  });
});
