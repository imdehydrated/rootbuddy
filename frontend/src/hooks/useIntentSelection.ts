import { useEffect, useState } from "react";
import {
  battleTargetKey,
  battleTargetLabel,
  cardEffectChoiceLabel,
  craftCardID,
  factionChoiceDetail,
  factionChoiceLabel,
  groupActionsByIntent,
  type AssistActionCandidateRef,
  type AssistIntentGroup,
  type AssistIntentKey
} from "../assistDirector";
import type { Action, GameState } from "../types";

type UseIntentSelectionOptions = {
  actions: Action[];
  state: GameState;
  resetKey: string;
};

export function useIntentSelection({ actions, state, resetKey }: UseIntentSelectionOptions) {
  const [selectedIntent, setSelectedIntent] = useState<AssistIntentKey | null>(null);
  const [selectedCraftCardID, setSelectedCraftCardID] = useState<number | null>(null);

  useEffect(() => {
    setSelectedIntent(null);
    setSelectedCraftCardID(null);
  }, [resetKey]);

  const actionGroups = groupActionsByIntent(actions);
  const selectedGroup = actionGroups.find((group) => group.key === selectedIntent) ?? null;
  const candidateRefsForSelectedGroup = (enabled: boolean): AssistActionCandidateRef[] =>
    enabled && selectedGroup
      ? selectedGroup.actions
          .map((action) => ({ actionIndex: actions.indexOf(action), action }))
          .filter((candidate) => candidate.actionIndex >= 0)
      : [];

  const battleCandidates =
    selectedIntent === "battle" && selectedGroup
      ? candidateRefsForSelectedGroup(true)
      : [];
  const movementCandidates =
    selectedIntent === "movement" && selectedGroup
      ? candidateRefsForSelectedGroup(true)
      : [];
  const buildRecruitCandidates =
    selectedIntent === "build_recruit" && selectedGroup
      ? candidateRefsForSelectedGroup(true)
      : [];
  const factionSpatialCandidates =
    selectedIntent === "faction" && selectedGroup
      ? candidateRefsForSelectedGroup(true)
      : [];

  const battleTargetChoices =
    selectedGroup?.key === "battle"
      ? Array.from(new Set(selectedGroup.actions.map(battleTargetKey).filter((key) => key.length > 0))).map((key) => {
          const matchingActions = selectedGroup.actions.filter((action) => battleTargetKey(action) === key);
          return {
            key,
            label: battleTargetLabel(matchingActions[0]),
            actions: matchingActions
          };
        })
      : [];

  const craftChoices =
    selectedGroup?.key === "craft"
      ? Array.from(new Set(selectedGroup.actions.map(craftCardID).filter((cardID) => cardID > 0))).map((cardID) => ({
          cardID,
          actions: selectedGroup.actions.filter((action) => craftCardID(action) === cardID)
        }))
      : [];
  const selectedCraftChoice = craftChoices.find((choice) => choice.cardID === selectedCraftCardID) ?? null;

  const factionChoiceActions =
    selectedGroup?.key === "faction"
      ? selectedGroup.actions
          .map((action) => ({ action, label: factionChoiceLabel(action, state), detail: factionChoiceDetail(action, state) }))
          .filter((choice): choice is { action: Action; label: string; detail: string } => choice.label !== null)
      : [];

  return {
    actionGroups,
    selectedIntent,
    setSelectedIntent,
    selectedGroup,
    selectedCraftCardID,
    setSelectedCraftCardID,
    battleCandidates,
    movementCandidates,
    buildRecruitCandidates,
    factionSpatialCandidates,
    battleTargetChoices,
    craftChoices,
    selectedCraftChoice,
    factionChoiceActions,
    drawAdvanceChoices: selectedGroup?.key === "draw_advance" ? selectedGroup.actions : [],
    cardEffectChoices: selectedGroup?.key === "card_effect" ? selectedGroup.actions : []
  };
}
