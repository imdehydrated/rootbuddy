import { startTransition, useDeferredValue, useEffect, useState } from "react";
import { applyAction, fetchBattleContext, fetchValidActions, resolveBattle } from "./api";
import { boardLayoutForState } from "./boardLayouts";
import { BoardPanel } from "./components/BoardPanel";
import { InspectorPanel } from "./components/InspectorPanel";
import { ObservedActionPanel } from "./components/ObservedActionPanel";
import { SetupFlowPanel } from "./components/SetupFlowPanel";
import { SetupWizard } from "./components/SetupWizard";
import { TurnStatePanel } from "./components/TurnStatePanel";
import { TurnSummaryPanel } from "./components/TurnSummaryPanel";
import { affectedClearings, syncDerivedFactionStateFromBoard } from "./gameHelpers";
import { ACTION_TYPE, describeAction, factionLabels, phaseLabels, setupStageLabels, stepLabels } from "./labels";
import { sampleState } from "./sampleState";
import type { Action, BattleContext, BattleModifiers, Clearing, GameState } from "./types";

type ActiveModal = "inspector" | "turn" | "actions" | "battle" | "observed" | "json" | "help" | null;

type MarquiseSetupDraft = {
  keepClearingID: number | null;
  sawmillClearingID: number | null;
  workshopClearingID: number | null;
};

const emptyMarquiseSetupDraft: MarquiseSetupDraft = {
  keepClearingID: null,
  sawmillClearingID: null,
  workshopClearingID: null
};

const emptyBattleModifiers: BattleModifiers = {
  attackerHitModifier: 0,
  defenderHitModifier: 0,
  ignoreHitsToAttacker: false,
  ignoreHitsToDefender: false,
  defenderAmbush: false,
  attackerCounterAmbush: false,
  attackerUsesArmorers: false,
  defenderUsesArmorers: false,
  attackerUsesBrutalTactics: false,
  defenderUsesSappers: false
};

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
  normalized.randomSeed ??= 0;
  normalized.shuffleCount ??= 0;
  normalized.setupStage ??= 0;
  normalized.map.clearings ??= [];
  normalized.map.forests ??= [];
  normalized.winningCoalition ??= [];
  normalized.turnOrder ??= [];
  normalized.victoryPoints ??= {};
  normalized.deck ??= [];
  normalized.discardPile ??= [];
  normalized.availableDominance ??= [];
  normalized.activeDominance ??= {};
  normalized.coalitionActive ??= false;
  normalized.coalitionPartner ??= 0;
  normalized.itemSupply ??= {};
  normalized.persistentEffects ??= {};
  normalized.questDeck ??= [];
  normalized.questDiscard ??= [];
  normalized.otherHandCounts ??= {};
  normalized.hiddenCards ??= [];
  normalized.nextHiddenCardID ??= 1;
  normalized.marquise.cardsInHand ??= [];
  normalized.eyrie.cardsInHand ??= [];
  normalized.eyrie.availableLeaders ??= [];
  normalized.eyrie.decree.recruit ??= [];
  normalized.eyrie.decree.move ??= [];
  normalized.eyrie.decree.battle ??= [];
  normalized.eyrie.decree.build ??= [];
  normalized.alliance.cardsInHand ??= [];
  normalized.alliance.supporters ??= [];
  normalized.vagabond.cardsInHand ??= [];
  normalized.vagabond.items ??= [];
  normalized.vagabond.relationships ??= {};
  normalized.vagabond.forestID ??= 0;
  normalized.vagabond.questsAvailable ??= [];
  normalized.vagabond.questsCompleted ??= [];
  normalized.turnProgress.usedWorkshopClearings ??= [];
  normalized.turnProgress.resolvedDecreeCardIDs ??= [];
  normalized.turnProgress.usedPersistentEffectIDs ??= [];
  normalized.turnProgress.birdsongMainActionTaken ??= false;
  normalized.turnProgress.daylightMainActionTaken ??= false;
  normalized.turnProgress.eveningMainActionTaken ??= false;
  for (const clearing of normalized.map.clearings) {
    clearing.adj ??= [];
    clearing.buildings ??= [];
    clearing.tokens ??= [];
    clearing.warriors ??= {};
    clearing.ruinItems ??= [];
  }
  syncDerivedFactionStateFromBoard(normalized);
  return normalized;
}

const initialState = normalizeState(sampleState);
const initialJSON = JSON.stringify(initialState, null, 2);

function zeroActionHint(state: GameState): string {
  if (state.gamePhase === 0) {
    return `${setupStageLabels[state.setupStage] ?? "Setup"} has no legal actions.`;
  }

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

function marquiseSetupMatches(
  action: Action,
  draft: MarquiseSetupDraft & { recruiterClearingID?: number }
): boolean {
  const payload = action.marquiseSetup;
  if (!payload) {
    return false;
  }

  if (draft.keepClearingID !== null && payload.keepClearingID !== draft.keepClearingID) {
    return false;
  }
  if (draft.sawmillClearingID !== null && payload.sawmillClearingID !== draft.sawmillClearingID) {
    return false;
  }
  if (draft.workshopClearingID !== null && payload.workshopClearingID !== draft.workshopClearingID) {
    return false;
  }
  if (draft.recruiterClearingID !== undefined && payload.recruiterClearingID !== draft.recruiterClearingID) {
    return false;
  }

  return true;
}

export default function App() {
  const [showSetup, setShowSetup] = useState(true);
  const [serverGameID, setServerGameID] = useState<string | null>(null);
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
  const [battleModifiers, setBattleModifiers] = useState<BattleModifiers>(emptyBattleModifiers);
  const [battleContext, setBattleContext] = useState<BattleContext | null>(null);
  const [assistDefenderAmbushChoice, setAssistDefenderAmbushChoice] = useState<boolean | null>(null);
  const [error, setError] = useState<string>("");
  const [status, setStatus] = useState<string>("Click a clearing to start setting the board.");
  const [activeModal, setActiveModal] = useState<ActiveModal>("help");
  const [marquiseSetupDraft, setMarquiseSetupDraft] = useState<MarquiseSetupDraft>(emptyMarquiseSetupDraft);

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

  useEffect(() => {
    setMarquiseSetupDraft(emptyMarquiseSetupDraft);
  }, [parsedState.gamePhase, parsedState.setupStage]);

  function syncState(nextState: GameState) {
    const normalizedState = normalizeState(nextState);
    startTransition(() => {
      setParsedState(normalizedState);
      setStateText(stringifyState(normalizedState));
      setActions([]);
      setSelectedBattleIndex(null);
      setHoveredActionIndex(null);
      setBattleModifiers(emptyBattleModifiers);
      setBattleContext(null);
      setAssistDefenderAmbushChoice(null);
      setError("");
    });
  }

  async function loadActionsForState(
    baseState: GameState,
    options?: { openModal?: boolean; successStatus?: string }
  ) {
    const requestState = normalizeState(baseState);
    const nextActions = await fetchValidActions(requestState, serverGameID);

    startTransition(() => {
      setParsedState(requestState);
      setStateText(stringifyState(requestState));
      setActions(nextActions);
      setSelectedBattleIndex(null);
      setHoveredActionIndex(null);
      setBattleModifiers(emptyBattleModifiers);
      setBattleContext(null);
      setAssistDefenderAmbushChoice(null);
      setError("");
    });

    setStatus(
      options?.successStatus ??
        (nextActions.length > 0 ? `Loaded ${nextActions.length} action(s).` : zeroActionHint(requestState))
    );

    if (options?.openModal) {
      setActiveModal("actions");
    }

    return nextActions;
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

    if (boardIsEmpty && parsedState.gamePhase !== 0) {
      setStatus("Enter the current board state first.");
      setActiveModal("help");
      return;
    }

    try {
      setStatus("Fetching valid actions...");
      await loadActionsForState(parsedState, { openModal: parsedState.gamePhase !== 0 });
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to fetch actions";
      setStatus(message);
    }
  }

  async function handleApply(action: Action) {
    let actionToApply = action;
    if (
      parsedState.gameMode === 1 &&
      action.usePersistentEffect?.effectID === "stand_and_deliver" &&
      action.usePersistentEffect.faction === parsedState.playerFaction &&
      action.usePersistentEffect.targetFaction !== parsedState.playerFaction
    ) {
      const answer = window.prompt("Stand and Deliver!: if you know the stolen card ID, enter it now. Leave blank to record it manually later.");
      const observedCardID = answer ? Number(answer) : 0;
      actionToApply = {
        ...action,
        usePersistentEffect: {
          ...action.usePersistentEffect,
          observedCardID: Number.isFinite(observedCardID) && observedCardID > 0 ? observedCardID : 0
        }
      };
    }

    try {
      setStatus("Applying action...");
      const { state: nextState, effectResult } = await applyAction(parsedState, actionToApply, serverGameID);
      if (nextState.gamePhase === 0) {
        await loadActionsForState(nextState, { successStatus: "Setup step applied." });
        setActiveModal(null);
        return;
      }

      syncState(nextState);
      setStatus(effectResult?.message ?? "Action applied.");
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
    if (assistDefenderAmbushPromptRequired && assistDefenderAmbushChoice === null) {
      setStatus("Answer whether the defender used an ambush before resolving the battle.");
      return;
    }

    try {
      setStatus("Resolving battle...");
      const resolved = await resolveBattle(
        parsedState,
        action,
        Number(attackerRoll),
        Number(defenderRoll),
        {
          ...battleModifiers,
          defenderAmbush: assistDefenderAmbushPromptRequired
            ? assistDefenderAmbushChoice === true
            : battleModifiers.defenderAmbush
        },
        serverGameID
      );
      const { state: nextState, effectResult } = await applyAction(parsedState, resolved, serverGameID);
      syncState(nextState);
      setStatus(effectResult?.message ?? "Battle resolved and applied.");
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
  const selectedBattleAction =
    selectedBattleIndex !== null && actions[selectedBattleIndex]?.type === ACTION_TYPE.BATTLE
      ? actions[selectedBattleIndex]
      : null;
  const attackerFaction = selectedBattleAction?.battle?.faction ?? -1;
  const defenderFaction = selectedBattleAction?.battle?.targetFaction ?? -1;
  const attackerHasScoutingParty = battleContext?.attackerHasScoutingParty ?? false;
  const canDefenderAmbush = battleContext?.canDefenderAmbush ?? false;
  const assistDefenderAmbushPromptRequired = battleContext?.assistDefenderAmbushPromptRequired ?? false;
  const canAttackerCounterAmbush = battleContext?.canAttackerCounterAmbush ?? false;
  const canAttackerArmorers = battleContext?.canAttackerArmorers ?? false;
  const canDefenderArmorers = battleContext?.canDefenderArmorers ?? false;
  const canAttackerBrutalTactics = battleContext?.canAttackerBrutalTactics ?? false;
  const canDefenderSappers = battleContext?.canDefenderSappers ?? false;
  const marquiseSetupActions = actions.filter((action) => action.type === ACTION_TYPE.MARQUISE_SETUP);
  const eyrieSetupActions = actions.filter((action) => action.type === ACTION_TYPE.EYRIE_SETUP);
  const vagabondSetupActions = actions.filter((action) => action.type === ACTION_TYPE.VAGABOND_SETUP);

  const legalSetupClearingIDs =
    parsedState.gamePhase !== 0
      ? []
      : parsedState.setupStage === 1
        ? (() => {
            if (marquiseSetupDraft.keepClearingID === null) {
              return Array.from(new Set(marquiseSetupActions.map((action) => action.marquiseSetup?.keepClearingID ?? 0))).filter(
                (value) => value > 0
              );
            }

            const filteredByKeep = marquiseSetupActions.filter((action) =>
              marquiseSetupMatches(action, marquiseSetupDraft)
            );
            if (marquiseSetupDraft.sawmillClearingID === null) {
              return Array.from(new Set(filteredByKeep.map((action) => action.marquiseSetup?.sawmillClearingID ?? 0))).filter(
                (value) => value > 0
              );
            }

            const filteredBySawmill = filteredByKeep.filter((action) =>
              marquiseSetupMatches(action, marquiseSetupDraft)
            );
            if (marquiseSetupDraft.workshopClearingID === null) {
              return Array.from(new Set(filteredBySawmill.map((action) => action.marquiseSetup?.workshopClearingID ?? 0))).filter(
                (value) => value > 0
              );
            }

            return Array.from(
              new Set(
                filteredBySawmill
                  .filter((action) => marquiseSetupMatches(action, marquiseSetupDraft))
                  .map((action) => action.marquiseSetup?.recruiterClearingID ?? 0)
              )
            ).filter((value) => value > 0);
          })()
        : parsedState.setupStage === 2
          ? eyrieSetupActions.map((action) => action.eyrieSetup?.clearingID ?? 0).filter((value) => value > 0)
          : [];

  const selectedSetupClearingIDs =
    parsedState.gamePhase === 0 && parsedState.setupStage === 1
      ? [marquiseSetupDraft.keepClearingID, marquiseSetupDraft.sawmillClearingID, marquiseSetupDraft.workshopClearingID].filter(
          (value): value is number => value !== null
        )
      : [];

  const forestTargets =
    parsedState.gamePhase === 0 && parsedState.setupStage === 3
      ? parsedState.map.forests.map((forest) => ({
          forestID: forest.id,
          label: `Forest ${forest.id}`,
          legal: vagabondSetupActions.some((action) => action.vagabondSetup?.forestID === forest.id),
          selected: false
        }))
      : [];

  useEffect(() => {
    if (assistDefenderAmbushPromptRequired) {
      setAssistDefenderAmbushChoice(null);
      setBattleModifiers((current) => ({
        ...current,
        defenderAmbush: false,
        attackerCounterAmbush: false
      }));
      return;
    }

    setAssistDefenderAmbushChoice(false);
  }, [assistDefenderAmbushPromptRequired, selectedBattleIndex]);

  useEffect(() => {
    let cancelled = false;

    async function loadBattleContext() {
      if (!selectedBattleAction?.battle) {
        setBattleContext(null);
        return;
      }

      try {
        const nextContext = await fetchBattleContext(parsedState, selectedBattleAction, serverGameID);
        if (!cancelled) {
          setBattleContext(nextContext);
        }
      } catch {
        if (!cancelled) {
          setBattleContext(null);
        }
      }
    }

    void loadBattleContext();

    return () => {
      cancelled = true;
    };
  }, [parsedState, selectedBattleAction, serverGameID]);

  async function handleSetupClearingClick(clearingID: number) {
    if (parsedState.gamePhase !== 0) {
      setSelectedClearingID(clearingID);
      setActiveModal("inspector");
      return;
    }

    if (!legalSetupClearingIDs.includes(clearingID)) {
      setStatus("Choose one of the highlighted setup targets.");
      return;
    }

    if (parsedState.setupStage === 1) {
      if (marquiseSetupDraft.keepClearingID === null) {
        setMarquiseSetupDraft({ ...emptyMarquiseSetupDraft, keepClearingID: clearingID });
        setStatus("Choose the starting sawmill location.");
        return;
      }
      if (marquiseSetupDraft.sawmillClearingID === null) {
        setMarquiseSetupDraft({ ...marquiseSetupDraft, sawmillClearingID: clearingID });
        setStatus("Choose the starting workshop location.");
        return;
      }
      if (marquiseSetupDraft.workshopClearingID === null) {
        setMarquiseSetupDraft({ ...marquiseSetupDraft, workshopClearingID: clearingID });
        setStatus("Choose the starting recruiter location.");
        return;
      }

      const action = marquiseSetupActions.find((candidate) =>
        marquiseSetupMatches(candidate, { ...marquiseSetupDraft, recruiterClearingID: clearingID })
      );
      if (!action) {
        setStatus("That building placement is not legal.");
        return;
      }
      await handleApply(action);
      return;
    }

    if (parsedState.setupStage === 2) {
      const action = eyrieSetupActions.find((candidate) => candidate.eyrieSetup?.clearingID === clearingID);
      if (!action) {
        setStatus("That starting clearing is not legal for the Eyrie.");
        return;
      }
      await handleApply(action);
    }
  }

  async function handleSetupForestClick(forestID: number) {
    if (parsedState.gamePhase !== 0 || parsedState.setupStage !== 3) {
      return;
    }

    const action = vagabondSetupActions.find((candidate) => candidate.vagabondSetup?.forestID === forestID);
    if (!action) {
      setStatus("That forest is not a legal Vagabond starting forest.");
      return;
    }
    await handleApply(action);
  }

  if (showSetup) {
    return (
      <SetupWizard
        onStart={async (state, gameID) => {
          syncState(state);
          setServerGameID(gameID);
          setShowSetup(false);
          setStatus(state.gamePhase === 0 ? "Setup started." : "Game created.");
          setActiveModal(null);
          if (state.gamePhase === 0) {
            try {
              await loadActionsForState(state, { successStatus: "Choose a highlighted setup target." });
            } catch (err) {
              const message = err instanceof Error ? err.message : "Failed to load setup actions";
              setStatus(message);
            }
          }
        }}
        onUseSample={() => {
          syncState(initialState);
          setServerGameID(null);
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
          forests={parsedState.map.forests}
          boardLayout={boardLayout}
          selectedClearingID={selectedClearingID}
          keepClearingID={parsedState.marquise.keepClearingID}
          vagabondClearingID={parsedState.vagabond.clearingID}
          vagabondInForest={parsedState.vagabond.inForest}
          highlightedClearings={highlightedClearings}
          setupLegalClearingIDs={legalSetupClearingIDs}
          setupSelectedClearingIDs={selectedSetupClearingIDs}
          forestTargets={forestTargets}
          onSelectClearing={handleSetupClearingClick}
          onSelectForest={handleSetupForestClick}
        />

        {boardIsEmpty && parsedState.gamePhase !== 0 ? <div className="board-hint">Click a clearing to edit the board.</div> : null}
      </div>

      <aside className="app-sidebar">
        <section className="panel sidebar-panel">
          <p className="eyebrow">RootBuddy</p>
          <div className="status-block">
            <strong>{factionLabels[parsedState.factionTurn] ?? "Unknown"}</strong>
            <span>
              {parsedState.gamePhase === 0
                ? setupStageLabels[parsedState.setupStage] ?? "Setup"
                : `${phaseLabels[parsedState.currentPhase] ?? "Unknown"} / ${stepLabels[parsedState.currentStep] ?? "Unknown"}`}
            </span>
          </div>
          <span className={error ? "message error" : "message"}>{error || status}</span>
        </section>

        <TurnSummaryPanel state={parsedState} />
        {parsedState.gamePhase === 0 ? (
          <SetupFlowPanel
            stage={parsedState.setupStage}
            activeFaction={parsedState.factionTurn}
            legalChoiceCount={parsedState.setupStage === 3 ? forestTargets.filter((target) => target.legal).length : legalSetupClearingIDs.length}
            marquiseDraft={marquiseSetupDraft}
            onResetMarquiseDraft={() => {
              setMarquiseSetupDraft(emptyMarquiseSetupDraft);
              setStatus("Marquise setup draft reset.");
            }}
          />
        ) : null}

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
            {parsedState.gameMode === 1 ? (
              <button type="button" className="secondary" onClick={() => setActiveModal("observed")}>
                Observed
              </button>
            ) : null}
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
                setServerGameID(null);
                setStatus("Board reset. Click a clearing to start setting the board.");
                setActiveModal("help");
              }}
            >
              Reset
            </button>
            <button type="button" onClick={refreshActions} disabled={!!error || (boardIsEmpty && parsedState.gamePhase !== 0)}>
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
                    {boardIsEmpty && parsedState.gamePhase !== 0
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
                                  setBattleModifiers(emptyBattleModifiers);
                                  setAssistDefenderAmbushChoice(null);
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

            {activeModal === "observed" ? (
              <ObservedActionPanel
                state={parsedState}
                onApply={handleApply}
                onClose={() => setActiveModal(null)}
              />
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
                {assistDefenderAmbushPromptRequired ? (
                  <div className="summary-stack" style={{ marginTop: "1rem" }}>
                    <span className="summary-label">Assist Prompt</span>
                    <span className="summary-line">
                      Did {factionLabels[defenderFaction] ?? "the defender"} play an Ambush?
                    </span>
                    <div className="sidebar-actions">
                      <button
                        type="button"
                        className={assistDefenderAmbushChoice === true ? "" : "secondary"}
                        onClick={() => {
                          setAssistDefenderAmbushChoice(true);
                          setBattleModifiers((current) => ({
                            ...current,
                            defenderAmbush: true
                          }));
                        }}
                      >
                        Yes
                      </button>
                      <button
                        type="button"
                        className={assistDefenderAmbushChoice === false ? "" : "secondary"}
                        onClick={() => {
                          setAssistDefenderAmbushChoice(false);
                          setBattleModifiers((current) => ({
                            ...current,
                            defenderAmbush: false,
                            attackerCounterAmbush: false
                          }));
                        }}
                      >
                        No
                      </button>
                    </div>
                  </div>
                ) : null}
                <div className="control-grid" style={{ marginTop: "1rem" }}>
                  {!assistDefenderAmbushPromptRequired ? (
                    <label className="checkbox">
                      <input
                        type="checkbox"
                        checked={battleModifiers.defenderAmbush}
                        disabled={!canDefenderAmbush}
                        onChange={(event) =>
                          setBattleModifiers((current) => ({
                            ...current,
                            defenderAmbush: event.target.checked,
                            attackerCounterAmbush: event.target.checked ? current.attackerCounterAmbush : false
                          }))
                        }
                      />
                      Defender Ambush
                    </label>
                  ) : null}
                  <label className="checkbox">
                    <input
                      type="checkbox"
                      checked={battleModifiers.attackerCounterAmbush}
                      disabled={
                        !(assistDefenderAmbushPromptRequired ? assistDefenderAmbushChoice === true : battleModifiers.defenderAmbush) ||
                        !canAttackerCounterAmbush
                      }
                      onChange={(event) =>
                        setBattleModifiers((current) => ({
                          ...current,
                          attackerCounterAmbush: event.target.checked
                        }))
                      }
                    />
                    Attacker Counter-Ambush
                  </label>
                  <label className="checkbox">
                    <input
                      type="checkbox"
                      checked={battleModifiers.attackerUsesArmorers}
                      disabled={!canAttackerArmorers}
                      onChange={(event) =>
                        setBattleModifiers((current) => ({
                          ...current,
                          attackerUsesArmorers: event.target.checked
                        }))
                      }
                    />
                    Attacker Armorers
                  </label>
                  <label className="checkbox">
                    <input
                      type="checkbox"
                      checked={battleModifiers.defenderUsesArmorers}
                      disabled={!canDefenderArmorers}
                      onChange={(event) =>
                        setBattleModifiers((current) => ({
                          ...current,
                          defenderUsesArmorers: event.target.checked
                        }))
                      }
                    />
                    Defender Armorers
                  </label>
                  <label className="checkbox">
                    <input
                      type="checkbox"
                      checked={battleModifiers.attackerUsesBrutalTactics}
                      disabled={!canAttackerBrutalTactics}
                      onChange={(event) =>
                        setBattleModifiers((current) => ({
                          ...current,
                          attackerUsesBrutalTactics: event.target.checked
                        }))
                      }
                    />
                    Attacker Brutal Tactics
                  </label>
                  <label className="checkbox">
                    <input
                      type="checkbox"
                      checked={battleModifiers.defenderUsesSappers}
                      disabled={!canDefenderSappers}
                      onChange={(event) =>
                        setBattleModifiers((current) => ({
                          ...current,
                          defenderUsesSappers: event.target.checked
                        }))
                      }
                    />
                    Defender Sappers
                  </label>
                </div>
                {attackerHasScoutingParty ? (
                  <p className="message" style={{ marginTop: "0.8rem" }}>
                    Attacker has Scouting Party, so defender ambushes are ignored.
                  </p>
                ) : null}
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
