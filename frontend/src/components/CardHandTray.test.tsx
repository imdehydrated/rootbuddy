import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { sampleState } from "../sampleState";
import type { Card, GameState } from "../types";
import { CardHandTray } from "./CardHandTray";

function sampleCard(overrides: Partial<Card> = {}): Card {
  return {
    id: 21,
    deck: 0,
    name: "Travel Gear",
    suit: 2,
    kind: 0,
    craftingCost: {
      fox: 0,
      rabbit: 1,
      mouse: 1,
      any: 0
    },
    craftedItem: null,
    effectID: "",
    vp: 0,
    ...overrides
  };
}

function trayState(overrides: Partial<GameState> = {}): GameState {
  return {
    ...structuredClone(sampleState),
    gamePhase: 1,
    ...overrides
  };
}

describe("CardHandTray", () => {
  it("shows the visible hand for the perspective faction", () => {
    const state = trayState({
      playerFaction: 1,
      alliance: {
        ...structuredClone(sampleState.alliance),
        cardsInHand: [sampleCard({ id: 31, name: "Alliance Plot" })]
      },
      marquise: {
        ...structuredClone(sampleState.marquise),
        cardsInHand: [sampleCard({ id: 32, name: "Marquise Card", suit: 0 })]
      }
    });

    render(<CardHandTray state={state} />);

    expect(screen.getByText("Woodland Alliance Hand")).toBeInTheDocument();
    expect(screen.getByText("Alliance Plot")).toBeInTheDocument();
    expect(screen.queryByText("Marquise Card")).not.toBeInTheDocument();
  });

  it("shows an empty state when the current hand has no visible cards", () => {
    render(<CardHandTray state={trayState({ playerFaction: 2 })} />);

    expect(screen.getByText("No visible cards in hand.")).toBeInTheDocument();
  });

  it("stays hidden outside the live game phase", () => {
    render(<CardHandTray state={trayState({ gamePhase: 0 })} />);

    expect(screen.queryByLabelText("Current hand tray")).not.toBeInTheDocument();
  });
});
