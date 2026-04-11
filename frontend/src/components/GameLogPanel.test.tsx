import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import type { ActionLogEntry } from "../types";
import { GameLogPanel } from "./GameLogPanel";

function logEntry(overrides: Partial<ActionLogEntry> = {}): ActionLogEntry {
  return {
    roundNumber: 2,
    faction: 2,
    actionType: 0,
    summary: "Eyrie moved from clearing 3 to clearing 7.",
    timestamp: 1_700_000_000_000,
    ...overrides
  };
}

describe("GameLogPanel", () => {
  it("renders log entries in reverse chronological order", () => {
    render(
      <GameLogPanel
        factionTurn={2}
        entries={[
          logEntry({ timestamp: 10, summary: "Older action" }),
          logEntry({ timestamp: 20, faction: 0, summary: "Newest action" })
        ]}
      />
    );

    const summaries = screen.getAllByText(/action$/i);
    expect(summaries[0]).toHaveTextContent("Newest action");
    expect(summaries[1]).toHaveTextContent("Older action");
    expect(screen.getByText("Marquise")).toBeInTheDocument();
  });

  it("shows an empty state when no entries are present", () => {
    render(<GameLogPanel factionTurn={1} entries={[]} />);

    expect(screen.getByText("No multiplayer actions have been logged yet.")).toBeInTheDocument();
  });

  it("can collapse and reopen the log body", () => {
    render(<GameLogPanel factionTurn={3} entries={[logEntry()]} />);

    fireEvent.click(screen.getByRole("button", { name: "Hide Log" }));
    expect(screen.queryByText("Eyrie moved from clearing 3 to clearing 7.")).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Show Log" }));
    expect(screen.getByText("Eyrie moved from clearing 3 to clearing 7.")).toBeInTheDocument();
  });
});
