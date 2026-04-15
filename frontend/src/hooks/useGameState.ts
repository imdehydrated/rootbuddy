import { startTransition, useDeferredValue, useEffect, useRef, useState } from "react";
import { applyAction, fetchValidActions, loadGame } from "../api";
import { syncDerivedFactionStateFromBoard } from "../gameHelpers";
import { factionLabels, setupStageLabels } from "../labels";
import type { SavedSession } from "../localSession";
import { loadSavedSession } from "../localSession";
import { sampleState } from "../sampleState";
import type { Action, Clearing, GameState } from "../types";

type UseGameStateOptions = {
  getMultiplayerToken: () => string | null;
  jsonEditorOpen: boolean;
};

export function stringifyState(nextState: GameState): string {
  return JSON.stringify(nextState, null, 2);
}

function hasWarriors(clearing: Clearing): boolean {
  return Object.values(clearing.warriors).some((count) => count > 0);
}

export function isBoardEmpty(state: GameState): boolean {
  return state.map.clearings.every(
    (clearing) => clearing.wood === 0 && !hasWarriors(clearing) && clearing.buildings.length === 0
  );
}

export function normalizeState(nextState: GameState): GameState {
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

export const initialState = normalizeState(sampleState);
export const initialJSON = JSON.stringify(initialState, null, 2);

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

export function gameOverHeadline(state: GameState): string {
  if (state.winningCoalition.length > 0) {
    return `${state.winningCoalition.map((faction) => factionLabels[faction] ?? `Faction ${faction}`).join(" + ")} win`;
  }
  return `${factionLabels[state.winner] ?? "Unknown"} win`;
}

export function gameOverStatusMessage(state: GameState): string {
  if (state.winningCoalition.length > 0) {
    return `Game over. Reviewing the coalition victory for ${state.winningCoalition
      .map((faction) => factionLabels[faction] ?? `Faction ${faction}`)
      .join(" + ")}.`;
  }
  return `Game over. Reviewing the final result for ${factionLabels[state.winner] ?? "Unknown"}.`;
}

function factionHandSize(state: GameState, faction: number): number {
  switch (faction) {
    case 0:
      return state.marquise.cardsInHand.length > 0 ? state.marquise.cardsInHand.length : state.otherHandCounts[String(faction)] ?? 0;
    case 1:
      return state.alliance.cardsInHand.length > 0 ? state.alliance.cardsInHand.length : state.otherHandCounts[String(faction)] ?? 0;
    case 2:
      return state.eyrie.cardsInHand.length > 0 ? state.eyrie.cardsInHand.length : state.otherHandCounts[String(faction)] ?? 0;
    case 3:
      return state.vagabond.cardsInHand.length > 0 ? state.vagabond.cardsInHand.length : state.otherHandCounts[String(faction)] ?? 0;
    default:
      return 0;
  }
}

export function useGameState({ getMultiplayerToken, jsonEditorOpen }: UseGameStateOptions) {
  const initialSavedSession = loadSavedSession();
  const [stateText, setStateText] = useState(initialJSON);
  const deferredStateText = useDeferredValue(stateText);
  const [parsedState, setParsedState] = useState<GameState>(initialState);
  const currentStateRef = useRef<GameState>(initialState);
  const [actions, setActions] = useState<Action[]>([]);
  const [selectedBattleIndex, setSelectedBattleIndex] = useState<number | null>(null);
  const [hoveredActionIndex, setHoveredActionIndex] = useState<number | null>(null);
  const [error, setError] = useState("");
  const [status, setStatus] = useState("Click a clearing to start setting the board.");
  const [serverGameID, setServerGameID] = useState<string | null>(null);
  const [serverRevision, setServerRevision] = useState<number | null>(initialSavedSession?.revision ?? null);

  useEffect(() => {
    if (!jsonEditorOpen) {
      return;
    }
    try {
      const nextState = JSON.parse(deferredStateText) as GameState;
      currentStateRef.current = nextState;
      setParsedState(nextState);
      setError("");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Invalid JSON");
    }
  }, [deferredStateText, jsonEditorOpen]);

  function syncState(nextState: GameState) {
    const normalizedState = normalizeState(nextState);
    currentStateRef.current = normalizedState;
    startTransition(() => {
      setParsedState(normalizedState);
      setStateText(stringifyState(normalizedState));
      setActions([]);
      setSelectedBattleIndex(null);
      setHoveredActionIndex(null);
      setError("");
    });
  }

  async function loadActionsForState(baseState: GameState, options?: { successStatus?: string }) {
    const requestState = normalizeState(baseState);
    currentStateRef.current = requestState;
    const { actions: nextActions, revision } = await fetchValidActions(requestState, serverGameID, getMultiplayerToken());
    if (revision !== null) {
      setServerRevision(revision);
    }

    startTransition(() => {
      setParsedState(requestState);
      setStateText(stringifyState(requestState));
      setActions(nextActions);
      setSelectedBattleIndex(null);
      setHoveredActionIndex(null);
      setError("");
    });

    setStatus(
      options?.successStatus ??
        (nextActions.length > 0
          ? requestState.gamePhase === 1 && requestState.factionTurn === requestState.playerFaction
            ? "Choose your next move."
            : requestState.gamePhase === 1
              ? "Record what happened on the board."
              : "Choose a highlighted setup target."
          : zeroActionHint(requestState))
    );

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
    if (isBoardEmpty(parsedState) && parsedState.gamePhase !== 0) {
      setStatus("Enter the current board state first.");
      return;
    }
    try {
      setStatus("Fetching valid actions...");
      await loadActionsForState(currentStateRef.current);
    } catch (err) {
      setStatus(err instanceof Error ? err.message : "Failed to fetch actions");
    }
  }

  function needsStandAndDeliverObservation(action: Action): boolean {
    const currentState = currentStateRef.current;
    return (
      currentState.gameMode === 1 &&
      !getMultiplayerToken() &&
      action.usePersistentEffect?.effectID === "stand_and_deliver" &&
      action.usePersistentEffect.targetFaction !== action.usePersistentEffect.faction &&
      factionHandSize(currentState, action.usePersistentEffect.targetFaction) > 0 &&
      (action.usePersistentEffect.observedCardID ?? 0) <= 0
    );
  }

  function isImpossibleStandAndDeliver(action: Action): boolean {
    const currentState = currentStateRef.current;
    return (
      action.usePersistentEffect?.effectID === "stand_and_deliver" &&
      action.usePersistentEffect.targetFaction !== action.usePersistentEffect.faction &&
      factionHandSize(currentState, action.usePersistentEffect.targetFaction) === 0
    );
  }

  async function applyFinalizedAction(actionToApply: Action) {
    try {
      setStatus("Applying action...");
      const { state: nextState, effectResult, revision } = await applyAction(
        currentStateRef.current,
        actionToApply,
        serverGameID,
        serverRevision,
        getMultiplayerToken()
      );
      if (revision !== null) {
        setServerRevision(revision);
      }
      if (nextState.gamePhase === 0) {
        await loadActionsForState(nextState, { successStatus: "Setup step applied." });
        return;
      }
      if (!getMultiplayerToken() && nextState.gamePhase === 1) {
        await loadActionsForState(nextState, { successStatus: effectResult?.message ?? "Action applied." });
        return;
      }
      syncState(nextState);
      setStatus(nextState.gamePhase === 2 ? gameOverStatusMessage(nextState) : effectResult?.message ?? "Action applied.");
    } catch (err) {
      setStatus(err instanceof Error ? err.message : "Failed to apply action");
    }
  }

  async function loadSavedGame(savedSession: SavedSession) {
    if (!savedSession.gameID) {
      return {
        state: savedSession.state,
        gameID: null,
        revision: savedSession.revision
      };
    }
    return loadGame(savedSession.gameID, getMultiplayerToken());
  }

  return {
    stateText,
    setStateText,
    parsedState,
    currentStateRef,
    actions,
    selectedBattleIndex,
    setSelectedBattleIndex,
    hoveredActionIndex,
    setHoveredActionIndex,
    error,
    setError,
    status,
    setStatus,
    serverGameID,
    setServerGameID,
    serverRevision,
    setServerRevision,
    boardIsEmpty: isBoardEmpty(parsedState),
    syncState,
    loadActionsForState,
    updateState,
    updateClearing,
    refreshActions,
    needsStandAndDeliverObservation,
    isImpossibleStandAndDeliver,
    applyFinalizedAction,
    loadSavedGame
  };
}
