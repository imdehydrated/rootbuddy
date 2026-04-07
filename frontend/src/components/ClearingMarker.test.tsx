import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { ClearingMarker } from "./ClearingMarker";
import type { Clearing } from "../types";

function testClearing(overrides: Partial<Clearing> = {}): Clearing {
  return {
    id: 1,
    suit: 0,
    buildSlots: 2,
    adj: [2, 3],
    ruins: false,
    ruinItems: [],
    wood: 0,
    warriors: {},
    buildings: [],
    tokens: [],
    ...overrides
  };
}

describe("ClearingMarker", () => {
  it("renders pending setup pieces without mutating the clearing", () => {
    render(
      <ClearingMarker
        clearing={testClearing()}
        position={{ left: "50%", top: "50%" }}
        isSelected={false}
        hasKeep={false}
        hasVagabond={false}
        previewPieces={[{ kind: "sawmill", label: "Pending sawmill", preview: true }]}
        onClick={vi.fn()}
      />
    );

    expect(screen.getByLabelText("Pending sawmill")).toBeInTheDocument();
    expect(screen.getByText("1/2 slots")).toBeInTheDocument();
  });
});
