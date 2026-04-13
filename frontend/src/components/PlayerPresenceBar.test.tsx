import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import type { LobbyPlayer } from "../types";
import { PlayerPresenceBar } from "./PlayerPresenceBar";

function player(overrides: Partial<LobbyPlayer> = {}): LobbyPlayer {
  return {
    displayName: "Alice",
    faction: 0,
    hasFaction: true,
    isHost: false,
    isReady: true,
    connected: true,
    ...overrides
  };
}

describe("PlayerPresenceBar", () => {
  it("renders players and highlights the active turn", () => {
    render(
      <PlayerPresenceBar
        players={[
          player({ displayName: "Alice", faction: 0 }),
          player({ displayName: "Bob", faction: 2 })
        ]}
        factionTurn={2}
        perspectiveFaction={0}
      />
    );

    expect(screen.getByText("Alice")).toBeInTheDocument();
    expect(screen.getByText("Bob").closest(".player-presence-pill")).toHaveClass("active-turn");
    expect(screen.getByText("Alice").closest(".player-presence-pill")).toHaveClass("perspective");
  });

  it("shows disconnected styling clearly", () => {
    render(
      <PlayerPresenceBar
        players={[player({ displayName: "Bob", faction: 2, connected: false, isReady: false })]}
        factionTurn={0}
        perspectiveFaction={0}
      />
    );

    expect(screen.getByText("Away")).toBeInTheDocument();
    expect(screen.getByText("Bob").closest(".player-presence-pill")).toHaveClass("disconnected");
  });
});
