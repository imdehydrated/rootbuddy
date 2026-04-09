import {
  eyrieLeaderLabels,
  factionLabels,
  phaseLabels,
  stepLabels,
  vagabondCharacterLabels
} from "../labels";
import type { GameState } from "../types";
import { AllianceEditor } from "./turnState/AllianceEditor";
import { EyrieEditor } from "./turnState/EyrieEditor";
import { MarquiseEditor } from "./turnState/MarquiseEditor";
import { VagabondEditor } from "./turnState/VagabondEditor";
import { duplicateValues, formatFactionList, formatNumberList, parseNumberList } from "./turnState/utils";

type TurnStatePanelProps = {
  state: GameState;
  onUpdateState: (mutator: (draft: GameState) => void) => void;
  title?: string;
  showCloseButton?: boolean;
  onClose: () => void;
};

export function TurnStatePanel({
  state,
  onUpdateState,
  title = "Advanced Turn State",
  showCloseButton = true,
  onClose
}: TurnStatePanelProps) {
  const missingTurnOrderFactions = factionLabels.map((_, index) => index).filter((faction) => !state.turnOrder.includes(faction));
  const duplicateTurnOrderFactions = duplicateValues(state.turnOrder);
  const currentPhaseLabel = phaseLabels[state.currentPhase] ?? `Phase ${state.currentPhase}`;
  const currentStepLabel = stepLabels[state.currentStep] ?? `Step ${state.currentStep}`;

  return (
    <section className="panel modal-panel">
      <div className="panel-header">
        <h2>{title}</h2>
        {showCloseButton ? (
          <button type="button" className="secondary" onClick={onClose}>
            Close
          </button>
        ) : null}
      </div>

      <div className="flow-guide-hero turn-state-hero">
        <span className="summary-label">Assist Recovery Surface</span>
        <strong>{factionLabels[state.factionTurn] ?? "Unknown"} - {currentPhaseLabel}</strong>
        <span className="summary-line">
          Use this when the observed table state and the app's turn flow have drifted. Prefer checkpoint shortcuts before editing faction internals.
        </span>
      </div>

      <div className="flow-step-list turn-state-guide">
        <div className="flow-step-card active">
          <strong>1. Confirm the turn owner</strong>
          <span className="summary-line">
            Current target is {factionLabels[state.factionTurn] ?? "Unknown"} at {currentPhaseLabel} / {currentStepLabel}.
          </span>
        </div>
        <div className="flow-step-card note">
          <strong>2. Use checkpoint shortcuts first</strong>
          <span className="summary-line">
            Birdsong, Daylight, and Evening only reset broad phase/step position. They are safer than editing counters directly.
          </span>
        </div>
        <div className="flow-step-card waiting">
          <strong>3. Edit internals only for real drift</strong>
          <span className="summary-line">
            Decree cards, supporter visibility, Vagabond position, and turn-progress counters are direct recovery controls.
          </span>
        </div>
      </div>

      <div className="turn-state-actions">
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
          Set Birdsong
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
          Set Daylight Actions
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
          Set Evening
        </button>
      </div>

      <div className="control-grid">
        <label>
          <span>Faction Turn</span>
          <select
            value={state.factionTurn}
            onChange={(event) =>
              onUpdateState((draft) => {
                draft.factionTurn = Number(event.target.value);
              })
            }
          >
            {factionLabels.map((label, index) => (
              <option key={label} value={index}>
                {label}
              </option>
            ))}
          </select>
        </label>

        <label>
          <span>Phase</span>
          <select
            value={state.currentPhase}
            onChange={(event) =>
              onUpdateState((draft) => {
                draft.currentPhase = Number(event.target.value);
              })
            }
          >
            {phaseLabels.map((label, index) => (
              <option key={label} value={index}>
                {label}
              </option>
            ))}
          </select>
        </label>

        <label>
          <span>Step</span>
          <select
            value={state.currentStep}
            onChange={(event) =>
              onUpdateState((draft) => {
                draft.currentStep = Number(event.target.value);
              })
            }
          >
            {stepLabels.map((label, index) => (
              <option key={label} value={index}>
                {label}
              </option>
            ))}
          </select>
        </label>

        <label>
          <span>Turn Order</span>
          <input
            type="text"
            value={formatNumberList(state.turnOrder)}
            onChange={(event) =>
              onUpdateState((draft) => {
                draft.turnOrder = parseNumberList(event.target.value);
              })
            }
          />
        </label>
      </div>

      {missingTurnOrderFactions.length > 0 || duplicateTurnOrderFactions.length > 0 ? (
        <div className="observed-issue-list" aria-live="polite">
          {missingTurnOrderFactions.length > 0 ? (
            <span className="message observed-issue error">
              Turn order is missing: {formatFactionList(missingTurnOrderFactions)}.
            </span>
          ) : null}
          {duplicateTurnOrderFactions.length > 0 ? (
            <span className="message observed-issue error">
              Turn order has duplicates: {formatFactionList(duplicateTurnOrderFactions)}.
            </span>
          ) : null}
        </div>
      ) : null}

      <div className="summary-section">
        <h3>Victory Points</h3>
        <div className="control-grid">
          {factionLabels.map((label, index) => (
            <label key={label}>
              <span>{label}</span>
              <input
                type="number"
                value={state.victoryPoints[String(index)] ?? 0}
                onChange={(event) =>
                  onUpdateState((draft) => {
                    draft.victoryPoints[String(index)] = Number(event.target.value);
                  })
                }
              />
            </label>
          ))}
        </div>
      </div>

      <div className="summary-section">
        <h3>Turn Progress</h3>
        <div className="control-grid">
          <label>
            <span>Actions Used</span>
            <input
              type="number"
              value={state.turnProgress.actionsUsed}
              onChange={(event) =>
                onUpdateState((draft) => {
                  draft.turnProgress.actionsUsed = Number(event.target.value);
                })
              }
            />
          </label>
          <label>
            <span>Bonus Actions</span>
            <input
              type="number"
              value={state.turnProgress.bonusActions}
              onChange={(event) =>
                onUpdateState((draft) => {
                  draft.turnProgress.bonusActions = Number(event.target.value);
                })
              }
            />
          </label>
          <label>
            <span>Marches Used</span>
            <input
              type="number"
              value={state.turnProgress.marchesUsed}
              onChange={(event) =>
                onUpdateState((draft) => {
                  draft.turnProgress.marchesUsed = Number(event.target.value);
                })
              }
            />
          </label>
          <label>
            <span>Officer Actions</span>
            <input
              type="number"
              value={state.turnProgress.officerActionsUsed}
              onChange={(event) =>
                onUpdateState((draft) => {
                  draft.turnProgress.officerActionsUsed = Number(event.target.value);
                })
              }
            />
          </label>
          <label className="checkbox">
            <input
              type="checkbox"
              checked={state.turnProgress.recruitUsed}
              onChange={(event) =>
                onUpdateState((draft) => {
                  draft.turnProgress.recruitUsed = event.target.checked;
                })
              }
            />
            Recruit Used
          </label>
          <label className="checkbox">
            <input
              type="checkbox"
              checked={state.turnProgress.hasSlipped}
              onChange={(event) =>
                onUpdateState((draft) => {
                  draft.turnProgress.hasSlipped = event.target.checked;
                })
              }
            />
            Has Slipped
          </label>
        </div>
      </div>

      <MarquiseEditor state={state} />
      <EyrieEditor state={state} onUpdateState={onUpdateState} />
      <AllianceEditor state={state} onUpdateState={onUpdateState} />
      <VagabondEditor state={state} onUpdateState={onUpdateState} />
    </section>
  );
}
