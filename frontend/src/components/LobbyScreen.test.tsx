import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { LobbyScreen } from "./LobbyScreen";
import type { Lobby, LobbyPlayer } from "../types";

function player(overrides: Partial<LobbyPlayer> = {}): LobbyPlayer {
  return {
    displayName: "Host",
    connected: true,
    hasFaction: true,
    faction: 0,
    isReady: false,
    isHost: true,
    ...overrides
  };
}

function lobby(overrides: Partial<Lobby> = {}): Lobby {
  return {
    joinCode: "ABCD",
    state: 0,
    players: [
      player(),
      player({
        displayName: "Guest",
        hasFaction: false,
        isHost: false
      })
    ],
    factions: [0, 2, 1, 3],
    mapID: "autumn",
    createdAt: "",
    ...overrides
  };
}

describe("LobbyScreen", () => {
  it("renders faction seats as table-style claim cards", () => {
    render(
      <LobbyScreen
        lobby={lobby()}
        self={player()}
        connectionStatus="connected"
        status=""
        submitting={false}
        onClaimFaction={vi.fn(async () => undefined)}
        onReady={vi.fn(async () => undefined)}
        onStart={vi.fn(async () => undefined)}
        onLeave={vi.fn(async () => undefined)}
      />
    );

    expect(screen.getByText("Claim A Seat")).toBeInTheDocument();
    expect(screen.getAllByText("Faction Seat").length).toBeGreaterThan(1);
    expect(screen.getAllByText("Open seat").length).toBeGreaterThan(1);
    expect(screen.getByText("Players At The Table")).toBeInTheDocument();
  });
});
