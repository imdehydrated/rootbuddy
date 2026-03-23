import { startTransition, useDeferredValue, useEffect, useState } from "react";
import { applyAction, fetchValidActions, resolveBattle } from "./api";
import { boardLayoutForState } from "./boardLayouts";
import { BoardPanel } from "./components/BoardPanel";
import { InspectorPanel } from "./components/InspectorPanel";
import { syncMarquiseStateFromBoard } from "./gameHelpers";
import { describeAction, factionLabels, phaseLabels, stepLabels } from "./labels";
import { sampleState } from "./sampleState";
import type { Action, Clearing, GameState } from "./types";

const initialJSON = JSON.stringify(sampleState, null, 2);

type ActiveModal = "inspector" | "turn" | "actions" | "battle" | "json" | "help" | null;

function stringifyState(nextState: GameState): string {
  return JSON.stringify(nextState, null, 2);
}

function hasWarriors(clearing: Clearing): boolean {
  return Object.values(clearing.warriors).some((count) => count > 0);
}

function isBoardEmpty(state: GameState): boolean {
  return state.map.clearings.every(
    (clearing) => clearing.wood === 0 && !hasWarriors(clearing) && clearing.buildings.length === 0
  );
}

function normalizeState(nextState: GameState): GameState {
  const normalized = structuredClone(nextState);
  syncMarquiseStateFromBoard(normalized);
  return normalized;
}

function zeroActionHint(state: GameState): string {
  if (state.factionTurn !== 0) {
    return "No actions: the current turn is not set to Marquise.";
  }

  const recruitStep =
    state.currentStep === 1 || (state.currentStep === 0 && state.currentPhase === 0);
  const daylightStep =
    state.currentStep === 2 || (state.currentStep === 0 && state.currentPhase === 1);
  const eveningStep =
    state.currentStep === 3 || (state.currentStep === 0 && state.currentPhase === 2);

  if (recruitStep) {
    if (state.turnProgress.recruitUsed) {
      return "No actions: recruit is already marked used this turn.";
    }
    if (state.marquise.recruitersPlaced === 0) {
      return "No actions: the recruit step needs at least one recruiter on the board.";
    }
    if (state.marquise.warriorSupply <= 0) {
      return "No actions: the Marquise has no warriors left in supply to recruit.";
    }
  }

  if (daylightStep) {
    return "No legal daylight actions found. Check ruling, adjacency, wood, and cards in hand.";
  }

  if (eveningStep) {
    return "No actions: evening actions are not implemented yet.";
  }

  return "No legal actions found for this state. Check the selected faction, phase, and step.";
}

export default function App() {
  const [stateText, setStateText] = useState(initialJSON);
  const deferredStateText = useDeferredValue(stateText);
  const [parsedState, setParsedState] = useState<GameState>(sampleState);
  const [selectedClearingID, setSelectedClearingID] = useState<number>(
    sampleState.map.clearings[0]?.id ?? 0
  );
  const [actions, setActions] = useState<Action[]>([]);
  const [selectedBattleIndex, setSelectedBattleIndex] = useState<number | null>(null);
  const [attackerRoll, setAttackerRoll] = useState("1");
  const [defenderRoll, setDefenderRoll] = useState("0");
  const [error, setError] = useState<string>("");
  const [status, setStatus] = useState<string>("Click a clearing to start setting the board.");
  const [activeModal, setActiveModal] = useState<ActiveModal>("help");

  useEffect(() => {
    try {
      const nextState = JSON.parse(deferredStateText) as GameState;
      setParsedState(nextState);
      setError("");
    } catch (err) {
      const message = err instanceof Error ? err.message : "Invalid JSON";
      setError(message);
    }
  }, [deferredStateText]);

  useEffect(() => {
    if (parsedState.map.clearings.some((clearing) => clearing.id === selectedClearingID)) {
      return;
    }
    setSelectedClearingID(parsedState.map.clearings[0]?.id ?? 0);
  }, [parsedState, selectedClearingID]);

  function syncState(nextState: GameState) {
    const normalizedState = normalizeState(nextState);
    startTransition(() => {
      setParsedState(normalizedState);
      setStateText(stringifyState(normalizedState));
      setActions([]);
      setSelectedBattleIndex(null);
      setError("");
    });
  }

  function updateState(mutator: (draft: GameState) => void) {
    const nextState = structuredClone(parsedState);
    mutator(nextState);
    syncState(nextState);
  }

  function updateClearing(clearingID: number, mutator: (clearing: Clearing) => void) {
    updateState((draft) => {
      const clearing = draft.map.clearings.find((item) => item.id === clearingID);
      if (!clearing) {
        return;
      }
      mutator(clearing);
    });
  }

  async function refreshActions() {
    if (error) {
      setStatus("Fix the JSON before requesting actions.");
      return;
    }

    if (boardIsEmpty) {
      setStatus("Enter the current board state first.");
      setActiveModal("help");
      return;
    }

    try {
      setStatus("Fetching valid actions...");
      const requestState = normalizeState(parsedState);
      const nextActions = await fetchValidActions(requestState);
      startTransition(() => {
        setParsedState(requestState);
        setStateText(stringifyState(requestState));
        setActions(nextActions);
        setSelectedBattleIndex(null);
      });
      setStatus(
        nextActions.length > 0
          ? `Loaded ${nextActions.length} action(s).`
          : zeroActionHint(requestState)
      );
      setActiveModal("actions");
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to fetch actions";
      setStatus(message);
    }
  }

  async function handleApply(action: Action) {
    try {
      setStatus("Applying action...");
      const nextState = await applyAction(parsedState, action);
      syncState(nextState);
      setStatus("Action applied.");
      setActiveModal(null);
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to apply action";
      setStatus(message);
    }
  }

  async function handleResolveAndApply() {
    if (selectedBattleIndex === null) {
      setStatus("Select a battle action first.");
      return;
    }

    const action = actions[selectedBattleIndex];
    if (action.type !== 1) {
      setStatus("Selected action is not a battle.");
      return;
    }

    try {
      setStatus("Resolving battle...");
      const resolved = await resolveBattle(
        parsedState,
        action,
        Number(attackerRoll),
        Number(defenderRoll)
      );
      const nextState = await applyAction(parsedState, resolved);
      syncState(nextState);
      setStatus("Battle resolved and applied.");
      setActiveModal(null);
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to resolve battle";
      setStatus(message);
    }
  }

  const selectedClearing =
    parsedState.map.clearings.find((clearing) => clearing.id === selectedClearingID) ??
    parsedState.map.clearings[0];
  const boardLayout = boardLayoutForState(parsedState);
  const boardIsEmpty = isBoardEmpty(parsedState);

  return (
    <main className="app-shell board-only-shell">
      <div className="board-stage">
        <BoardPanel
          clearings={parsedState.map.clearings}
          boardLayout={boardLayout}
          selectedClearingID={selectedClearingID}
          keepClearingID={parsedState.marquise.keepClearingID}
          vagabondClearingID={parsedState.vagabond.clearingID}
          vagabondInForest={parsedState.vagabond.inForest}
          onSelectClearing={(clearingID) => {
            setSelectedClearingID(clearingID);
            setActiveModal("inspector");
          }}
        />

        <div className="floating-status">
          <p className="eyebrow">RootBuddy</p>
          <span className={error ? "message error" : "message"}>{error || status}</span>
        </div>

        <div className="floating-actions top-left">
          <button type="button" className="secondary" onClick={() => setActiveModal("help")}>
            Help
          </button>
          <button type="button" className="secondary" onClick={() => setActiveModal("turn")}>
            Turn State
          </button>
        </div>

        <div className="floating-actions top-right">
          <button type="button" className="secondary" onClick={() => setActiveModal("actions")}>
            Actions
          </button>
          <button type="button" className="secondary" onClick={() => setActiveModal("battle")}>
            Resolve
          </button>
          <button type="button" className="secondary" onClick={() => setActiveModal("json")}>
            JSON
          </button>
        </div>

        <div className="floating-actions bottom-bar">
          <button
            type="button"
            className="secondary"
            onClick={() => {
              syncState(sampleState);
              setStatus("Board reset. Click a clearing to start setting the board.");
              setActiveModal("help");
            }}
          >
            Reset
          </button>
          <button type="button" onClick={refreshActions} disabled={!!error || boardIsEmpty}>
            Generate Actions
          </button>
        </div>

        {boardIsEmpty ? <div className="board-hint">Click a clearing to edit the board.</div> : null}
      </div>

      {activeModal ? (
        <div className="modal-backdrop" onClick={() => setActiveModal(null)}>
          <div className="modal-shell" onClick={(event) => event.stopPropagation()}>
            {activeModal === "inspector" ? (
              <InspectorPanel
                clearing={selectedClearing}
                keepClearingID={parsedState.marquise.keepClearingID}
                vagabondClearingID={parsedState.vagabond.clearingID}
                vagabondInForest={parsedState.vagabond.inForest}
                onUpdateClearing={updateClearing}
                onSetKeepClearing={(clearingID) =>
                  updateState((draft) => {
                    draft.marquise.keepClearingID = clearingID;
                  })
                }
                onSetVagabondClearing={(clearingID, inForest) =>
                  updateState((draft) => {
                    draft.vagabond.clearingID = clearingID;
                    draft.vagabond.inForest = inForest;
                  })
                }
                onClose={() => setActiveModal(null)}
              />
            ) : null}

            {activeModal === "turn" ? (
              <section className="panel modal-panel">
                <div className="panel-header">
                  <h2>Turn State</h2>
                  <button type="button" className="secondary" onClick={() => setActiveModal(null)}>
                    Close
                  </button>
                </div>
                <div className="turn-state-actions">
                  <button
                    type="button"
                    className="secondary"
                    onClick={() =>
                      updateState((draft) => {
                        draft.factionTurn = 0;
                        draft.currentPhase = 1;
                        draft.currentStep = 2;
                      })
                    }
                  >
                    Set Marquise Daylight
                  </button>
                  <button
                    type="button"
                    className="secondary"
                    onClick={() =>
                      updateState((draft) => {
                        draft.factionTurn = 0;
                        draft.currentPhase = 0;
                        draft.currentStep = 1;
                      })
                    }
                  >
                    Set Recruit Step
                  </button>
                </div>
                <div className="control-grid">
                  <label>
                    <span>Faction Turn</span>
                    <select
                      value={parsedState.factionTurn}
                      onChange={(event) =>
                        updateState((draft) => {
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
                      value={parsedState.currentPhase}
                      onChange={(event) =>
                        updateState((draft) => {
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
                      value={parsedState.currentStep}
                      onChange={(event) =>
                        updateState((draft) => {
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
                    <span>Warrior Supply</span>
                    <input type="number" min="0" value={parsedState.marquise.warriorSupply} readOnly />
                  </label>
                  <label>
                    <span>Sawmills Placed</span>
                    <input type="number" min="0" value={parsedState.marquise.sawmillsPlaced} readOnly />
                  </label>
                  <label>
                    <span>Workshops Placed</span>
                    <input type="number" min="0" value={parsedState.marquise.workshopsPlaced} readOnly />
                  </label>
                  <label>
                    <span>Recruiters Placed</span>
                    <input type="number" min="0" value={parsedState.marquise.recruitersPlaced} readOnly />
                  </label>
                  <label>
                    <span>Keep Clearing</span>
                    <input
                      type="text"
                      value={parsedState.marquise.keepClearingID || "Unset"}
                      readOnly
                    />
                  </label>
                  <label className="checkbox">
                    <input
                      type="checkbox"
                      checked={parsedState.turnProgress.recruitUsed}
                      onChange={(event) =>
                        updateState((draft) => {
                          draft.turnProgress.recruitUsed = event.target.checked;
                        })
                      }
                    />
                    Recruit Used
                  </label>
                </div>
              </section>
            ) : null}

            {activeModal === "actions" ? (
              <section className="panel modal-panel">
                <div className="panel-header">
                  <h2>Actions</h2>
                  <div className="inspector-header-actions">
                    <span className="pill">{actions.length}</span>
                    <button type="button" className="secondary" onClick={() => setActiveModal(null)}>
                      Close
                    </button>
                  </div>
                </div>
                {actions.length === 0 ? (
                  <p className="empty-state">
                    {boardIsEmpty
                      ? "Enter the board state, then generate actions."
                      : "No actions loaded yet."}
                  </p>
                ) : (
                  <ul className="action-list">
                    {actions.map((action, index) => {
                      const isBattle = action.type === 1;
                      return (
                        <li key={`${action.type}-${index}`} className="action-card">
                          <strong>{describeAction(action)}</strong>
                          <div className="action-controls">
                            <button
                              type="button"
                              onClick={() => handleApply(action)}
                              disabled={isBattle}
                            >
                              Apply
                            </button>
                            {isBattle ? (
                              <button
                                type="button"
                                className="secondary"
                                onClick={() => {
                                  setSelectedBattleIndex(index);
                                  setActiveModal("battle");
                                }}
                              >
                                Use for Battle
                              </button>
                            ) : null}
                          </div>
                        </li>
                      );
                    })}
                  </ul>
                )}
              </section>
            ) : null}

            {activeModal === "battle" ? (
              <section className="panel modal-panel">
                <div className="panel-header">
                  <h2>Resolve Battle</h2>
                  <div className="inspector-header-actions">
                    <span className="pill">
                      {selectedBattleIndex === null ? "None Selected" : `Action ${selectedBattleIndex + 1}`}
                    </span>
                    <button type="button" className="secondary" onClick={() => setActiveModal(null)}>
                      Close
                    </button>
                  </div>
                </div>
                <div className="resolve-grid">
                  <label>
                    <span>Attacker Roll</span>
                    <input
                      type="number"
                      min="0"
                      max="3"
                      value={attackerRoll}
                      onChange={(event) => setAttackerRoll(event.target.value)}
                    />
                  </label>
                  <label>
                    <span>Defender Roll</span>
                    <input
                      type="number"
                      min="0"
                      max="3"
                      value={defenderRoll}
                      onChange={(event) => setDefenderRoll(event.target.value)}
                    />
                  </label>
                  <button type="button" onClick={handleResolveAndApply}>
                    Resolve and Apply
                  </button>
                </div>
              </section>
            ) : null}

            {activeModal === "json" ? (
              <section className="panel modal-panel">
                <div className="panel-header">
                  <h2>Advanced JSON</h2>
                  <button type="button" className="secondary" onClick={() => setActiveModal(null)}>
                    Close
                  </button>
                </div>
                <textarea
                  className="state-editor"
                  value={stateText}
                  onChange={(event) => setStateText(event.target.value)}
                  spellCheck={false}
                />
              </section>
            ) : null}

            {activeModal === "help" ? (
              <section className="panel modal-panel">
                <div className="panel-header">
                  <h2>Quick Start</h2>
                  <button type="button" className="secondary" onClick={() => setActiveModal(null)}>
                    Close
                  </button>
                </div>
                <div className="compact-help">
                  <p>1. Click a clearing to place faction-specific warriors, buildings, sympathy, wood, ruins, the Keep, and the Vagabond.</p>
                  <p>2. Open Turn State and set the current Marquise step.</p>
                  <p>3. Use Generate Actions, then review or apply them from the Actions popup.</p>
                </div>
              </section>
            ) : null}
          </div>
        </div>
      ) : null}
    </main>
  );
}
