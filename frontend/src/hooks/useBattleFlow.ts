import { useEffect, useState } from "react";
import { applyAction, fetchBattleContext, openBattle, resolveBattle, submitBattleResponse } from "../api";
import { ACTION_TYPE, factionLabels } from "../labels";
import { gameOverStatusMessage } from "./useGameState";
import type { Action, BattleContext, BattleModifiers, BattlePrompt, GameState } from "../types";

export const emptyBattleModifiers: BattleModifiers = {
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

type UseBattleFlowOptions = {
  actions: Action[];
  getMultiplayerToken: () => string | null;
  multiplayerBattlePrompt: BattlePrompt | null;
  parsedState: GameState;
  perspectiveFaction: number;
  selectedBattleIndex: number | null;
  serverGameID: string | null;
  serverRevision: number | null;
  setHoveredActionIndex: (index: number | null) => void;
  setMultiplayerBattlePrompt: (prompt: BattlePrompt | null) => void;
  setMultiplayerSubmitting: (submitting: boolean) => void;
  setSelectedBattleIndex: (index: number | null) => void;
  setServerRevision: (revision: number | null) => void;
  setStatus: (status: string) => void;
  syncState: (state: GameState) => void;
  loadActionsForState: (state: GameState, options?: { successStatus?: string }) => Promise<Action[]>;
};

export function useBattleFlow({
  actions,
  getMultiplayerToken,
  multiplayerBattlePrompt,
  parsedState,
  perspectiveFaction,
  selectedBattleIndex,
  serverGameID,
  serverRevision,
  setHoveredActionIndex,
  setMultiplayerBattlePrompt,
  setMultiplayerSubmitting,
  setSelectedBattleIndex,
  setServerRevision,
  setStatus,
  syncState,
  loadActionsForState
}: UseBattleFlowOptions) {
  const [attackerRoll, setAttackerRoll] = useState("1");
  const [defenderRoll, setDefenderRoll] = useState("0");
  const [battleModifiers, setBattleModifiers] = useState<BattleModifiers>(emptyBattleModifiers);
  const [battleContext, setBattleContext] = useState<BattleContext | null>(null);
  const [assistDefenderAmbushChoice, setAssistDefenderAmbushChoice] = useState<boolean | null>(null);

  const selectedBattleAction =
    selectedBattleIndex !== null && actions[selectedBattleIndex]?.type === ACTION_TYPE.BATTLE
      ? actions[selectedBattleIndex]
      : null;
  const activeBattleAction = multiplayerBattlePrompt?.action ?? selectedBattleAction;
  const activeBattleContext = multiplayerBattlePrompt?.battleContext ?? battleContext;
  const attackerFaction = activeBattleAction?.battle?.faction ?? -1;
  const defenderFaction = activeBattleAction?.battle?.targetFaction ?? -1;
  const assistDefenderAmbushPromptRequired = activeBattleContext?.assistDefenderAmbushPromptRequired ?? false;

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
  }, [assistDefenderAmbushPromptRequired, activeBattleAction, multiplayerBattlePrompt]);

  useEffect(() => {
    if (!multiplayerBattlePrompt) {
      return;
    }

    setBattleModifiers({
      ...emptyBattleModifiers,
      defenderAmbush: multiplayerBattlePrompt.defenderAmbush ?? false,
      defenderUsesArmorers: multiplayerBattlePrompt.defenderUsedArmorers ?? false,
      defenderUsesSappers: multiplayerBattlePrompt.defenderUsedSappers ?? false,
      attackerCounterAmbush: multiplayerBattlePrompt.attackerCounterAmbush ?? false,
      attackerUsesArmorers: multiplayerBattlePrompt.attackerUsedArmorers ?? false,
      attackerUsesBrutalTactics: multiplayerBattlePrompt.attackerUsedBrutalTactics ?? false
    });
  }, [multiplayerBattlePrompt]);

  useEffect(() => {
    let cancelled = false;

    async function loadCurrentBattleContext() {
      if (multiplayerBattlePrompt?.battleContext) {
        setBattleContext(multiplayerBattlePrompt.battleContext);
        return;
      }
      if (!selectedBattleAction?.battle) {
        setBattleContext(null);
        return;
      }

      try {
        const nextContext = await fetchBattleContext(parsedState, selectedBattleAction, serverGameID, getMultiplayerToken());
        if (!cancelled) {
          if (nextContext.revision !== null) {
            setServerRevision(nextContext.revision);
          }
          setBattleContext(nextContext.battleContext);
        }
      } catch {
        if (!cancelled) {
          setBattleContext(null);
        }
      }
    }

    void loadCurrentBattleContext();

    return () => {
      cancelled = true;
    };
  }, [multiplayerBattlePrompt, parsedState, selectedBattleAction, serverGameID, setServerRevision]);

  async function handleResolveAndApply() {
    const battleAction = multiplayerBattlePrompt?.action ?? (selectedBattleIndex !== null ? actions[selectedBattleIndex] : null);
    if (!battleAction) {
      setStatus("Select a battle action first.");
      return;
    }
    if (battleAction.type !== ACTION_TYPE.BATTLE) {
      setStatus("Selected action is not a battle.");
      return;
    }

    const multiplayerToken = getMultiplayerToken();
    if (multiplayerToken && multiplayerBattlePrompt && multiplayerBattlePrompt.stage !== "ready_to_resolve") {
      if (multiplayerBattlePrompt.waitingOnFaction === perspectiveFaction) {
        setStatus("Respond to the current battle prompt before resolving.");
      } else {
        setStatus(`Waiting on ${factionLabels[multiplayerBattlePrompt.waitingOnFaction] ?? "another player"} before resolving.`);
      }
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
        battleAction,
        multiplayerToken ? 0 : Number(attackerRoll),
        multiplayerToken ? 0 : Number(defenderRoll),
        multiplayerToken
          ? undefined
          : {
              ...battleModifiers,
              defenderAmbush: assistDefenderAmbushPromptRequired
                ? assistDefenderAmbushChoice === true
                : battleModifiers.defenderAmbush
            },
        serverGameID,
        multiplayerToken
      );
      if (resolved.revision !== null) {
        setServerRevision(resolved.revision);
      }

      const { state: nextState, effectResult, revision } = await applyAction(
        parsedState,
        resolved.action,
        serverGameID,
        resolved.revision ?? serverRevision,
        multiplayerToken
      );
      if (revision !== null) {
        setServerRevision(revision);
      }
      setMultiplayerBattlePrompt(null);
      syncState(nextState);
      if (!multiplayerToken && nextState.gamePhase === 1) {
        await loadActionsForState(nextState, {
          successStatus: effectResult?.message ?? "Battle resolved. Your next options are ready."
        });
        return;
      }
      setStatus(nextState.gamePhase === 2 ? gameOverStatusMessage(nextState) : effectResult?.message ?? "Battle resolved and applied.");
    } catch (err) {
      setStatus(err instanceof Error ? err.message : "Failed to resolve battle");
    }
  }

  async function handleSubmitBattleResponse() {
    const multiplayerToken = getMultiplayerToken();
    if (!multiplayerToken || !multiplayerBattlePrompt?.gameID) {
      setStatus("No multiplayer battle prompt is active.");
      return;
    }

    try {
      setMultiplayerSubmitting(true);
      setStatus("Submitting battle response...");
      const response = await submitBattleResponse(
        {
          gameID: multiplayerBattlePrompt.gameID,
          useAmbush: multiplayerBattlePrompt.stage === "defender_response" ? battleModifiers.defenderAmbush : undefined,
          useDefenderArmorers:
            multiplayerBattlePrompt.stage === "defender_response" ? battleModifiers.defenderUsesArmorers : undefined,
          useSappers: multiplayerBattlePrompt.stage === "defender_response" ? battleModifiers.defenderUsesSappers : undefined,
          useCounterAmbush:
            multiplayerBattlePrompt.stage === "attacker_response" ? battleModifiers.attackerCounterAmbush : undefined,
          useAttackerArmorers:
            multiplayerBattlePrompt.stage === "attacker_response" ? battleModifiers.attackerUsesArmorers : undefined,
          useBrutalTactics:
            multiplayerBattlePrompt.stage === "attacker_response" ? battleModifiers.attackerUsesBrutalTactics : undefined
        },
        multiplayerToken
      );
      if (response.revision !== null) {
        setServerRevision(response.revision);
      }
      setMultiplayerBattlePrompt(response.prompt);
      setStatus(response.prompt?.stage === "ready_to_resolve" ? "Battle choices locked in. Resolve when ready." : "Battle response submitted.");
    } catch (err) {
      setStatus(err instanceof Error ? err.message : "Failed to submit battle response");
    } finally {
      setMultiplayerSubmitting(false);
    }
  }

  function openBattleForActionIndex(actionIndex: number) {
    setSelectedBattleIndex(actionIndex);
    setHoveredActionIndex(actionIndex);
    setBattleModifiers(emptyBattleModifiers);
    setAssistDefenderAmbushChoice(null);

    const action = actions[actionIndex];
    const multiplayerToken = getMultiplayerToken();
    if (multiplayerToken && serverGameID && action?.type === ACTION_TYPE.BATTLE) {
      void (async () => {
        try {
          setMultiplayerSubmitting(true);
          setStatus("Opening multiplayer battle flow...");
          const response = await openBattle(parsedState, action, serverGameID, multiplayerToken);
          if (response.revision !== null) {
            setServerRevision(response.revision);
          }
          setMultiplayerBattlePrompt(response.prompt);
          setStatus(response.prompt?.stage === "ready_to_resolve" ? "Battle is ready to resolve." : "Battle selected. Follow the multiplayer response prompt in Battle Flow.");
        } catch (err) {
          setStatus(err instanceof Error ? err.message : "Failed to open multiplayer battle flow");
        } finally {
          setMultiplayerSubmitting(false);
        }
      })();
      return;
    }

    setStatus("Battle selected. Resolve it from Battle Flow.");
  }

  function clearBattleSelection() {
    setSelectedBattleIndex(null);
    setHoveredActionIndex(null);
    setBattleContext(null);
    setMultiplayerBattlePrompt(null);
    setBattleModifiers(emptyBattleModifiers);
    setAssistDefenderAmbushChoice(null);
    setStatus("Cleared selected battle.");
  }

  return {
    attackerRoll,
    setAttackerRoll,
    defenderRoll,
    setDefenderRoll,
    battleModifiers,
    setBattleModifiers,
    battleContext,
    assistDefenderAmbushChoice,
    setAssistDefenderAmbushChoice,
    selectedBattleAction,
    activeBattleAction,
    activeBattleContext,
    attackerFaction,
    defenderFaction,
    assistDefenderAmbushPromptRequired,
    handleResolveAndApply,
    handleSubmitBattleResponse,
    openBattleForActionIndex,
    clearBattleSelection
  };
}
