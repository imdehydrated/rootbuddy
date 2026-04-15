import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { SetupWizard } from "./SetupWizard";
import { sampleState } from "../sampleState";

function renderWizard() {
  return render(
    <SetupWizard
      onStart={vi.fn()}
      onUseSample={vi.fn()}
      onResume={vi.fn(async () => undefined)}
      onClearSavedSession={vi.fn()}
      onOpenCreateLobby={vi.fn()}
      onOpenJoinLobby={vi.fn()}
      canResume={false}
      savedSessionInfo={null}
    />
  );
}

describe("SetupWizard", () => {
  it("keeps online play focused on lobby actions only", async () => {
    renderWizard();

    fireEvent.click(screen.getByRole("button", { name: /Online Play/i }));

    expect(await screen.findByRole("button", { name: "Create Lobby" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Join Lobby" })).toBeInTheDocument();
    expect(screen.queryByText("Factions In Game")).not.toBeInTheDocument();
    expect(screen.queryByText("My Faction")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Start Online Game" })).not.toBeInTheDocument();
  });

  it("keeps assist mode direct-start controls", async () => {
    renderWizard();

    fireEvent.click(screen.getByRole("button", { name: /Assist Mode/i }));

    expect(await screen.findByText("Factions In Game")).toBeInTheDocument();
    expect(screen.getByText("My Faction")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Start Assist Game" })).toBeInTheDocument();
  });

  it("does not show local saved-game resume controls on the landing screen", () => {
    render(
      <SetupWizard
        onStart={vi.fn()}
        onUseSample={vi.fn()}
        onResume={vi.fn(async () => undefined)}
        onClearSavedSession={vi.fn()}
        onOpenCreateLobby={vi.fn()}
        onOpenJoinLobby={vi.fn()}
        canResume={true}
        savedSessionInfo={{
          state: sampleState,
          gameID: null,
          revision: null,
          savedAt: "2026-04-15T00:00:00Z"
        }}
      />
    );

    expect(screen.queryByText("Saved Chronicle")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Resume Saved Game/i })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Clear Saved Game/i })).not.toBeInTheDocument();
  });

  it("keeps the landing screen reduced to the two primary mode buttons", () => {
    renderWizard();

    expect(screen.getByRole("button", { name: "Online Play" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Assist Mode" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Use Sample State/i })).not.toBeInTheDocument();
  });
});
