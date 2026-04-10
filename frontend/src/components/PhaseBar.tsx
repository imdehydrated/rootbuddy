import type { CSSProperties } from "react";
import { factionLabels, phaseLabels, setupStageLabels, stepLabels } from "../labels";

type PhaseBarProps = {
  gamePhase: number;
  currentPhase: number;
  currentStep: number;
  setupStage: number;
  factionTurn: number;
  roundNumber: number;
};

const factionAccentColors = ["#b14d36", "#4c7a45", "#496aa0", "#8a6842"];

function factionAccentColor(faction: number) {
  return factionAccentColors[faction] ?? "#7a6045";
}

export function PhaseBar({
  gamePhase,
  currentPhase,
  currentStep,
  setupStage,
  factionTurn,
  roundNumber
}: PhaseBarProps) {
  const activeFactionLabel = factionLabels[factionTurn] ?? "Unknown";
  const activePhaseLabel = phaseLabels[currentPhase] ?? "Unknown";
  const activeStepLabel = stepLabels[currentStep] ?? "Unknown";
  const mode = gamePhase === 0 ? "setup" : gamePhase === 2 ? "game-over" : "live";
  const style = {
    "--active-faction-color": factionAccentColor(factionTurn)
  } as CSSProperties;

  if (mode === "setup") {
    return (
      <section className="phase-bar phase-bar-setup" aria-label="Phase bar" data-mode={mode} style={style}>
        <div className="phase-bar-header">
          <span className="phase-bar-kicker">Setup</span>
          <strong>{setupStageLabels[setupStage] ?? "Setup"}</strong>
          <span>Round {roundNumber} staging</span>
        </div>
        <div className="phase-bar-status">Follow the highlighted board targets to finish setup.</div>
      </section>
    );
  }

  if (mode === "game-over") {
    return (
      <section className="phase-bar phase-bar-game-over" aria-label="Phase bar" data-mode={mode} style={style}>
        <div className="phase-bar-header">
          <span className="phase-bar-kicker">Final State</span>
          <strong>Game Over</strong>
          <span>Round {roundNumber} complete</span>
        </div>
        <div className="phase-bar-status">Review the final board state and endgame summary.</div>
      </section>
    );
  }

  return (
    <section className="phase-bar phase-bar-live" aria-label="Phase bar" data-mode={mode} style={style}>
      <div className="phase-bar-header">
        <span className="phase-bar-kicker">Round {roundNumber}</span>
        <strong>{activeFactionLabel} Turn</strong>
        <span>{activePhaseLabel} / {activeStepLabel}</span>
      </div>
      <div className="phase-bar-track">
        {phaseLabels.map((label, index) => (
          <span
            key={label}
            className={`phase-bar-segment ${index === currentPhase ? "active" : ""}`}
            aria-current={index === currentPhase ? "step" : undefined}
          >
            {label}
          </span>
        ))}
      </div>
    </section>
  );
}
