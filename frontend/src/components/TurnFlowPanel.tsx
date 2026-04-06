import { factionLabels } from "../labels";
import type { Action, GameState } from "../types";

type TurnFlowPanelProps = {
  state: GameState;
  onApply: (action: Action) => Promise<void>;
  onGenerateActions: () => Promise<void>;
  onOpenAdvanced: () => void;
  onUpdateState: (mutator: (draft: GameState) => void) => void;
};

export function TurnFlowPanel({
  state,
  onApply,
  onGenerateActions,
  onOpenAdvanced,
  onUpdateState
}: TurnFlowPanelProps) {
  if (state.gamePhase !== 1) {
    return null;
  }

  return (
    <section className="panel sidebar-panel">
      <p className="eyebrow">Turn Controls</p>

      <div className="flow-guide-hero">
        <span className="summary-label">Current Turn</span>
        <strong>{factionLabels[state.factionTurn] ?? "Unknown"}</strong>
        <span className="summary-line">Use these shortcuts only for turn-state corrections and recovery, not routine play.</span>
      </div>

      <div className="summary-stack">
        <span className="summary-label">Acting Faction</span>
        <select
          value={state.factionTurn}
          onChange={(event) =>
            onUpdateState((draft) => {
              draft.factionTurn = Number(event.target.value);
            })
          }
        >
          {state.turnOrder.map((faction) => (
            <option key={faction} value={faction}>
              {factionLabels[faction] ?? `Faction ${faction}`}
            </option>
          ))}
        </select>
      </div>

      <div className="flow-step-list" style={{ marginTop: "0.9rem" }}>
        <div className="flow-step-card note">
          <strong>Phase Jump</strong>
          <span className="summary-line">Snap the turn to a broad phase checkpoint when the flow is out of sync.</span>
          <div className="shortcut-grid" style={{ marginTop: "0.5rem" }}>
            <button
              type="button"
              className="secondary"
              onClick={() =>
                onUpdateState((draft) => {
                  draft.currentPhase = 0;
                  draft.currentStep = 1;
                })
              }
            >
              Birdsong
            </button>
            <button
              type="button"
              className="secondary"
              onClick={() =>
                onUpdateState((draft) => {
                  draft.currentPhase = 1;
                  draft.currentStep = 3;
                })
              }
            >
              Daylight
            </button>
            <button
              type="button"
              className="secondary"
              onClick={() =>
                onUpdateState((draft) => {
                  draft.currentPhase = 2;
                  draft.currentStep = 4;
                })
              }
            >
              Evening
            </button>
          </div>
        </div>

        <div className="flow-step-card waiting">
          <strong>Correction Tools</strong>
          <span className="summary-line">Use these after fixing the board or when you need to force the flow forward.</span>
          <div className="shortcut-grid" style={{ marginTop: "0.5rem" }}>
            <button type="button" className="secondary" onClick={() => void onGenerateActions()}>
              Generate
            </button>
            <button
              type="button"
              className="secondary"
              onClick={() =>
                void onApply({
                  type: 24,
                  passPhase: {
                    faction: state.factionTurn
                  }
                })
              }
            >
              Advance
            </button>
            <button type="button" className="secondary" onClick={onOpenAdvanced}>
              Advanced
            </button>
          </div>
        </div>
      </div>
    </section>
  );
}
