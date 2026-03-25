import { startTransition, useDeferredValue, useEffect, useState } from "react";
import { applyAction, fetchValidActions, resolveBattle } from "./api";
import { boardLayoutForState } from "./boardLayouts";
import { BoardPanel } from "./components/BoardPanel";
import { InspectorPanel } from "./components/InspectorPanel";
import { SetupWizard } from "./components/SetupWizard";
import { TurnStatePanel } from "./components/TurnStatePanel";
import { TurnSummaryPanel } from "./components/TurnSummaryPanel";
import { affectedClearings, syncDerivedFactionStateFromBoard } from "./gameHelpers";
import { ACTION_TYPE, describeAction, factionLabels, phaseLabels, stepLabels } from "./labels";
import { sampleState } from "./sampleState";
import type { Action, Clearing, GameState } from "./types";

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
  normalized.map.forests ??= [];
  normalized.vagabond.forestID ??= 0;
  normalized.vagabond.questsAvailable ??= [];
  normalized.vagabond.questsCompleted ??= [];
  for (const clearing of normalized.map.clearings) {
    clearing.ruinItems ??= [];
  }
  syncDerivedFactionStateFromBoard(normalized);
  return normalized;
}

const initialState = normalizeState(sampleState);
const initialJSON = JSON.stringify(initialState, null, 2);

function zeroActionHint(state: GameState): string {
  if (state.currentPhase === 0) {
    return "No legal birdsong actions found for the current faction state.";
  }

  if (state.currentPhase === 1) {
    return "No legal daylight actions found. Check faction-specific requirements like ruling, decree state, supporters, items, wood, and cards in hand.";
  }

  if (state.currentPhase === 2) {
    return "No legal evening actions found for the current faction state.";
  }

  return "No legal actions found for this state. Check the selected faction, phase, step, and faction-specific resources.";
}

export default function App() {
  const [showSetup, setShowSetup] = useState(true);
  const [stateText, setStateText] = useState(initialJSON);
  const deferredStateText = useDeferredValue(stateText);
  const [parsedState, setParsedState] = useState<GameState>(initialState);
  const [selectedClearingID, setSelectedClearingID] = useState<number>(
    initialState.map.clearings[0]?.id ?? 0
  );
  const [actions, setActions] = useState<Action[]>([]);
  const [selectedBattleIndex, setSelectedBattleIndex] = useState<number | null>(null);
  const [hoveredActionIndex, setHoveredActionIndex] = useState<number | null>(null);
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
      setHoveredActionIndex(null);
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
    if (action.type !== ACTION_TYPE.BATTLE) {
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
  const previewedAction =
    actions[hoveredActionIndex ?? selectedBattleIndex ?? -1] ?? null;
  const highlightedClearings = previewedAction ? affectedClearings(previewedAction) : [];

  if (showSetup) {
    return (
      <SetupWizard
        onStart={(state) => {
          syncState(state);
          setShowSetup(false);
          setStatus("Game created.");
          setActiveModal(null);
        }}
        onUseSample={() => {
          syncState(initialState);
          setShowSetup(false);
          setStatus("Loaded sample state.");
          setActiveModal("help");
        }}
      />
    );
  }

  return (
    <main className="app-shell workspace-shell">
      <div className="board-stage">
        <BoardPanel
          clearings={parsedState.map.clearings}
          boardLayout={boardLayout}
          selectedClearingID={selectedClearingID}
          keepClearingID={parsedState.marquise.keepClearingID}
          vagabondClearingID={parsedState.vagabond.clearingID}
          vagabondInForest={parsedState.vagabond.inForest}
          highlightedClearings={highlightedClearings}
          onSelectClearing={(clearingID) => {
            setSelectedClearingID(clearingID);
            setActiveModal("inspector");
          }}
        />

        {boardIsEmpty ? <div className="board-hint">Click a clearing to edit the board.</div> : null}
      </div>

      <aside className="app-sidebar">
        <section className="panel sidebar-panel">
          <p className="eyebrow">RootBuddy</p>
          <div className="status-block">
            <strong>{factionLabels[parsedState.factionTurn] ?? "Unknown"}</strong>
            <span>
              {phaseLabels[parsedState.currentPhase] ?? "Unknown"} / {stepLabels[parsedState.currentStep] ?? "Unknown"}
            </span>
          </div>
          <span className={error ? "message error" : "message"}>{error || status}</span>
        </section>

        <TurnSummaryPanel state={parsedState} />

        <section className="panel sidebar-panel sidebar-actions-panel">
          <p className="eyebrow">Controls</p>
          <div className="sidebar-actions">
            <button type="button" className="secondary" onClick={() => setActiveModal("help")}>
              Help
            </button>
            <button type="button" className="secondary" onClick={() => setShowSetup(true)}>
              Setup
            </button>
            <button type="button" className="secondary" onClick={() => setActiveModal("turn")}>
              Turn State
            </button>
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
          <div className="sidebar-actions footer">
            <button
              type="button"
              className="secondary"
              onClick={() => {
                syncState(initialState);
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
        </section>
      </aside>

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
              <TurnStatePanel state={parsedState} onUpdateState={updateState} onClose={() => setActiveModal(null)} />
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
                      const isBattle = action.type === ACTION_TYPE.BATTLE;
                      return (
                        <li
                          key={`${action.type}-${index}`}
                          className={`action-card ${index === (hoveredActionIndex ?? selectedBattleIndex) ? "previewed" : ""}`}
                          onMouseEnter={() => setHoveredActionIndex(index)}
                          onMouseLeave={() => setHoveredActionIndex(null)}
                        >
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
                                  setHoveredActionIndex(index);
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
                  <p>2. Open Turn State and set the current faction, phase, and step.</p>
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
