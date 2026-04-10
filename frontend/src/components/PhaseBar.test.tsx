import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { PhaseBar } from "./PhaseBar";

describe("PhaseBar", () => {
  it("highlights the active phase with the acting faction color", () => {
    render(
      <PhaseBar
        gamePhase={1}
        currentPhase={1}
        currentStep={3}
        setupStage={0}
        factionTurn={2}
        roundNumber={4}
      />
    );

    expect(screen.getByText("Eyrie Turn")).toBeInTheDocument();
    expect(screen.getByText("Daylight")).toHaveClass("active");
    expect(screen.getByText("Daylight")).toHaveAttribute("aria-current", "step");
    expect(screen.getByLabelText("Phase bar")).toHaveStyle("--active-faction-color: #496aa0");
  });

  it("switches to setup mode before live play starts", () => {
    render(
      <PhaseBar
        gamePhase={0}
        currentPhase={0}
        currentStep={0}
        setupStage={1}
        factionTurn={0}
        roundNumber={1}
      />
    );

    expect(screen.getByText("Marquise Setup")).toBeInTheDocument();
    expect(screen.getByText("Follow the highlighted board targets to finish setup.")).toBeInTheDocument();
    expect(screen.queryByText("Birdsong")).not.toBeInTheDocument();
  });

  it("switches to game-over mode when the match is finished", () => {
    render(
      <PhaseBar
        gamePhase={2}
        currentPhase={2}
        currentStep={4}
        setupStage={4}
        factionTurn={3}
        roundNumber={7}
      />
    );

    expect(screen.getByText("Game Over")).toBeInTheDocument();
    expect(screen.getByText("Review the final board state and endgame summary.")).toBeInTheDocument();
  });
});
