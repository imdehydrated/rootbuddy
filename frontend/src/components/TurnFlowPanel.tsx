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
      <p className="eyebrow">Turn Flow</p>

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
        <span className="summary-line">Use these controls for normal flow corrections. Full turn editing stays under Advanced.</span>
      </div>

      <div className="shortcut-grid" style={{ marginTop: "0.9rem" }}>
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
    </section>
  );
}
