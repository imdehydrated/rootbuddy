import { describeAction, actionHeadline } from "./actionPresentation";
import { describeKnownCardID } from "./cardCatalog";
import { ACTION_TYPE, eyrieLeaderLabels, factionLabels, itemTypeLabels, suitLabels } from "./labels";
import type { Action, GameState, HighlightedClearing } from "./types";
import type { ObservedTemplateKey } from "./components/ObservedActionPanel";

export type AssistIntentKey =
  | "movement"
  | "battle"
  | "craft"
  | "build_recruit"
  | "faction"
  | "draw_advance"
  | "card_effect"
  | "other";

export type AssistIntentGroup = {
  key: AssistIntentKey;
  label: string;
  detail: string;
  actions: Action[];
};

export type MovementEndpoint = {
  fromClearingID: number;
  toClearingID: number;
  fromForestID: number;
  toForestID: number;
};

export type AssistMovementSource = {
  kind: "clearing" | "forest";
  id: number;
};

export type AssistBattleCandidate = {
  actionIndex: number;
  action: Action;
};

export type AssistMovementCandidate = {
  actionIndex: number;
  action: Action;
  endpoints: MovementEndpoint;
};

export type AssistClearingCandidate = {
  actionIndex: number;
  action: Action;
  clearingIDs: number[];
};

export type AssistBoardCandidates = {
  battleCandidates: AssistBattleCandidate[];
  movementCandidates: AssistMovementCandidate[];
  buildRecruitCandidates: AssistClearingCandidate[];
  factionSpatialCandidates: AssistClearingCandidate[];
};

export function actionIntentKey(action: Action): AssistIntentKey {
  switch (action.type) {
    case ACTION_TYPE.MOVEMENT:
    case ACTION_TYPE.SLIP:
      return "movement";
    case ACTION_TYPE.BATTLE:
      return "battle";
    case ACTION_TYPE.CRAFT:
      return "craft";
    case ACTION_TYPE.BUILD:
    case ACTION_TYPE.RECRUIT:
    case ACTION_TYPE.OVERWORK:
      return "build_recruit";
    case ACTION_TYPE.ADD_TO_DECREE:
    case ACTION_TYPE.SPREAD_SYMPATHY:
    case ACTION_TYPE.REVOLT:
    case ACTION_TYPE.MOBILIZE:
    case ACTION_TYPE.TRAIN:
    case ACTION_TYPE.ORGANIZE:
    case ACTION_TYPE.EXPLORE:
    case ACTION_TYPE.QUEST:
    case ACTION_TYPE.AID:
    case ACTION_TYPE.STRIKE:
    case ACTION_TYPE.REPAIR:
    case ACTION_TYPE.TURMOIL:
    case ACTION_TYPE.DAYBREAK:
    case ACTION_TYPE.BIRDSONG_WOOD:
    case ACTION_TYPE.SCORE_ROOSTS:
      return "faction";
    case ACTION_TYPE.EVENING_DRAW:
    case ACTION_TYPE.OTHER_PLAYER_DRAW:
    case ACTION_TYPE.PASS_PHASE:
      return "draw_advance";
    case ACTION_TYPE.OTHER_PLAYER_PLAY:
    case ACTION_TYPE.DISCARD_EFFECT:
    case ACTION_TYPE.ACTIVATE_DOMINANCE:
    case ACTION_TYPE.TAKE_DOMINANCE:
    case ACTION_TYPE.USE_PERSISTENT_EFFECT:
      return "card_effect";
    default:
      return "other";
  }
}

function intentMetadata(key: AssistIntentKey): Pick<AssistIntentGroup, "label" | "detail"> {
  switch (key) {
    case "movement":
      return { label: "Move", detail: "Pieces changed clearings or the Vagabond slipped." };
    case "battle":
      return { label: "Battle", detail: "A battle started and needs resolution." };
    case "craft":
      return { label: "Craft", detail: "A known card was crafted." };
    case "build_recruit":
      return { label: "Build / Recruit", detail: "Pieces, wood, or buildings were added." };
    case "faction":
      return { label: "Faction Action", detail: "A faction-specific public step happened." };
    case "draw_advance":
      return { label: "Draw / Advance", detail: "Cards were drawn or the phase advanced." };
    case "card_effect":
      return { label: "Card Effect", detail: "A known card, dominance card, or persistent effect was used." };
    default:
      return { label: "Other", detail: "A less common public action candidate." };
  }
}

export function groupActionsByIntent(actions: Action[]): AssistIntentGroup[] {
  const order: AssistIntentKey[] = ["movement", "battle", "craft", "build_recruit", "faction", "draw_advance", "card_effect", "other"];
  const buckets = new Map<AssistIntentKey, Action[]>();

  actions.forEach((action) => {
    const key = actionIntentKey(action);
    buckets.set(key, [...(buckets.get(key) ?? []), action]);
  });

  return order
    .map((key) => {
      const groupedActions = buckets.get(key) ?? [];
      if (groupedActions.length === 0) {
        return null;
      }
      return {
        key,
        ...intentMetadata(key),
        actions: groupedActions
      };
    })
    .filter((group): group is AssistIntentGroup => group !== null);
}

export function observedPromptTemplates(state: GameState): Array<{ label: string; template: ObservedTemplateKey }> {
  const prompts: Array<{ label: string; template: ObservedTemplateKey }> = [
    { label: "Hidden Draw", template: "other_player_draw" },
    { label: "Known Card Play", template: "other_player_play" }
  ];

  if (state.factionTurn === 2) {
    prompts.push({ label: "Decree Choice", template: "add_to_decree" });
  }
  if (state.factionTurn === 1) {
    prompts.push({ label: "Supporter Spend", template: "spread_sympathy" });
    prompts.push({ label: "Revolt", template: "revolt" });
  }
  if (state.factionTurn === 3) {
    prompts.push({ label: "Aid", template: "aid" });
  }

  prompts.push({ label: "Battle Result", template: "battle_resolution" });
  prompts.push({ label: "Dominance", template: "activate_dominance" });

  return prompts;
}

export function craftCardID(action: Action): number {
  return action.type === ACTION_TYPE.CRAFT ? action.craft?.cardID ?? 0 : 0;
}

export function craftRouteLabel(action: Action): string {
  const clearings = action.craft?.usedWorkshopClearings ?? [];
  if (clearings.length === 0) {
    return "No workshop cost";
  }
  return `Workshops ${clearings.join(" + ")}`;
}

function questLabel(action: Action, state: GameState): string {
  const questID = action.quest?.questID ?? 0;
  const quest = [...state.vagabond.questsAvailable, ...state.vagabond.questsCompleted].find((candidate) => candidate.id === questID);
  return quest ? quest.name : `Quest ${questID || "?"}`;
}

function questRewardLabel(reward: number | undefined): string {
  switch (reward) {
    case 0:
      return "Victory points";
    case 1:
      return "Draw 2 cards";
    default:
      return "Unknown reward";
  }
}

function itemIndexLabel(state: GameState, itemIndex: number | undefined): string {
  if (itemIndex === undefined || itemIndex < 0) {
    return "Item ?";
  }
  const item = state.vagabond.items[itemIndex];
  return `${itemTypeLabels[item?.type ?? -1] ?? "Item"} #${itemIndex + 1}`;
}

export function factionChoiceLabel(action: Action, state: GameState): string | null {
  switch (action.type) {
    case ACTION_TYPE.MOBILIZE:
      return `Mobilize ${describeKnownCardID(action.mobilize?.cardID ?? 0)}`;
    case ACTION_TYPE.TRAIN:
      return `Train with ${describeKnownCardID(action.train?.cardID ?? 0)}`;
    case ACTION_TYPE.QUEST:
      return `${questLabel(action, state)}: ${questRewardLabel(action.quest?.reward)}`;
    case ACTION_TYPE.REPAIR:
      return `Repair ${itemIndexLabel(state, action.repair?.itemIndex)}`;
    case ACTION_TYPE.TURMOIL:
      return `Choose ${eyrieLeaderLabels[action.turmoil?.newLeader ?? -1] ?? "new leader"}`;
    case ACTION_TYPE.DAYBREAK:
      return `Refresh ${(action.daybreak?.refreshedItemIndexes ?? []).map((index) => itemIndexLabel(state, index)).join(", ") || "no items"}`;
    case ACTION_TYPE.BIRDSONG_WOOD:
      return `Place wood in clearings ${(action.birdsongWood?.clearingIDs ?? []).join(", ")}`;
    case ACTION_TYPE.SCORE_ROOSTS:
      return `Score ${action.scoreRoosts?.points ?? 0} roost point(s)`;
    default:
      return null;
  }
}

export function factionChoiceDetail(action: Action, state: GameState): string {
  if (action.type === ACTION_TYPE.QUEST) {
    const itemIndexes = action.quest?.itemIndexes ?? [];
    return `Items: ${itemIndexes.map((index) => itemIndexLabel(state, index)).join(", ") || "none"}`;
  }
  return describeAction(action, state);
}

function knownCardListLabel(cardIDs: number[]): string {
  return cardIDs.length > 0 ? cardIDs.map(describeKnownCardID).join(", ") : "No cards";
}

function knownCardLabel(cardID: number | undefined): string {
  return cardID && cardID > 0 ? describeKnownCardID(cardID) : "Unknown card";
}

export function factionSpatialChoiceLabel(action: Action): string | null {
  switch (action.type) {
    case ACTION_TYPE.SPREAD_SYMPATHY:
      return `Spread Sympathy to clearing ${action.spreadSympathy?.clearingID ?? "?"}`;
    case ACTION_TYPE.REVOLT:
      return `Revolt in clearing ${action.revolt?.clearingID ?? "?"}`;
    case ACTION_TYPE.ORGANIZE:
      return `Organize clearing ${action.organize?.clearingID ?? "?"}`;
    case ACTION_TYPE.EXPLORE:
      return `Explore clearing ${action.explore?.clearingID ?? "?"}`;
    case ACTION_TYPE.AID:
      return `Aid ${factionLabels[action.aid?.targetFaction ?? -1] ?? "Unknown faction"} in clearing ${action.aid?.clearingID ?? "?"}`;
    case ACTION_TYPE.STRIKE:
      return `Strike ${factionLabels[action.strike?.targetFaction ?? -1] ?? "Unknown faction"} in clearing ${action.strike?.clearingID ?? "?"}`;
    default:
      return null;
  }
}

export function factionSpatialChoiceDetail(action: Action): string {
  switch (action.type) {
    case ACTION_TYPE.SPREAD_SYMPATHY:
      return `Supporters: ${knownCardListLabel(action.spreadSympathy?.supporterCardIDs ?? [])}`;
    case ACTION_TYPE.REVOLT:
      return `Base: ${suitLabels[action.revolt?.baseSuit ?? -1] ?? "Unknown"}; supporters: ${knownCardListLabel(action.revolt?.supporterCardIDs ?? [])}`;
    case ACTION_TYPE.EXPLORE:
      return `Item: ${itemTypeLabels[action.explore?.itemType ?? -1] ?? "Unknown"}`;
    case ACTION_TYPE.AID:
      return `Card: ${knownCardLabel(action.aid?.cardID)}`;
    case ACTION_TYPE.STRIKE:
      return `Target: ${factionLabels[action.strike?.targetFaction ?? -1] ?? "Unknown faction"}`;
    case ACTION_TYPE.ORGANIZE:
      return "Remove one Alliance warrior there and place sympathy.";
    default:
      return describeAction(action);
  }
}

export function decreeCardKey(action: Action): string {
  return action.type === ACTION_TYPE.ADD_TO_DECREE ? (action.addToDecree?.cardIDs ?? []).join("|") : "";
}

export function decreeCardLabel(cardIDs: number[]): string {
  return cardIDs.map(describeKnownCardID).join(" + ");
}

function decreeColumnLabel(column: number): string {
  const labels = ["Recruit", "Move", "Battle", "Build"];
  return labels[column] ?? `Column ${column}`;
}

export function decreeColumnAssignmentLabel(action: Action): string {
  const cardIDs = action.addToDecree?.cardIDs ?? [];
  const columns = action.addToDecree?.columns ?? [];
  return cardIDs
    .map((cardID, index) => `${describeKnownCardID(cardID)} -> ${decreeColumnLabel(columns[index] ?? -1)}`)
    .join(", ");
}

export function drawAdvanceChoiceLabel(action: Action): string {
  switch (action.type) {
    case ACTION_TYPE.EVENING_DRAW:
      return `Draw ${action.eveningDraw?.count ?? 0} card(s)`;
    case ACTION_TYPE.OTHER_PLAYER_DRAW:
      return `Record draw ${action.otherPlayerDraw?.count ?? 0}`;
    case ACTION_TYPE.PASS_PHASE:
      return "Advance phase";
    default:
      return actionHeadline(action);
  }
}

export function cardEffectChoiceLabel(action: Action): string {
  switch (action.type) {
    case ACTION_TYPE.OTHER_PLAYER_PLAY:
      return `Play ${describeKnownCardID(action.otherPlayerPlay?.cardID ?? 0)}`;
    case ACTION_TYPE.DISCARD_EFFECT:
      return `Discard ${describeKnownCardID(action.discardEffect?.cardID ?? 0)}`;
    case ACTION_TYPE.ACTIVATE_DOMINANCE:
      return `Activate ${describeKnownCardID(action.activateDominance?.cardID ?? 0)}`;
    case ACTION_TYPE.TAKE_DOMINANCE:
      return `Take ${describeKnownCardID(action.takeDominance?.dominanceCardID ?? 0)}`;
    case ACTION_TYPE.USE_PERSISTENT_EFFECT:
      return describeAction(action);
    default:
      return actionHeadline(action);
  }
}

export function battleTargetKey(action: Action): string {
  if (action.type !== ACTION_TYPE.BATTLE) {
    return "";
  }
  return `${action.battle?.clearingID ?? 0}:${action.battle?.targetFaction ?? -1}`;
}

export function battleTargetLabel(action: Action): string {
  const clearingID = action.battle?.clearingID ?? 0;
  const targetFaction = action.battle?.targetFaction ?? -1;
  return `Clearing ${clearingID}: ${factionLabels[targetFaction] ?? "Unknown defender"}`;
}

export function movementEndpoints(action: Action): MovementEndpoint | null {
  if (action.type === ACTION_TYPE.MOVEMENT && action.movement) {
    const hasSource = action.movement.from > 0 || action.movement.fromForestID > 0;
    const hasDestination = action.movement.to > 0 || action.movement.toForestID > 0;
    if (!hasSource || !hasDestination) {
      return null;
    }
    return {
      fromClearingID: action.movement.from,
      toClearingID: action.movement.to,
      fromForestID: action.movement.fromForestID,
      toForestID: action.movement.toForestID
    };
  }

  if (action.type === ACTION_TYPE.SLIP && action.slip) {
    const hasSource = action.slip.from > 0 || action.slip.fromForestID > 0;
    const hasDestination = action.slip.to > 0 || action.slip.toForestID > 0;
    if (!hasSource || !hasDestination) {
      return null;
    }
    return {
      fromClearingID: action.slip.from,
      toClearingID: action.slip.to,
      fromForestID: action.slip.fromForestID,
      toForestID: action.slip.toForestID
    };
  }

  return null;
}

export function movementSourceMatches(endpoint: MovementEndpoint, source: AssistMovementSource): boolean {
  return source.kind === "clearing" ? endpoint.fromClearingID === source.id : endpoint.fromForestID === source.id;
}

export function buildRecruitClearingIDs(action: Action): number[] {
  switch (action.type) {
    case ACTION_TYPE.BUILD:
      return action.build?.clearingID ? [action.build.clearingID] : [];
    case ACTION_TYPE.RECRUIT:
      return action.recruit?.clearingIDs ?? [];
    case ACTION_TYPE.OVERWORK:
      return action.overwork?.clearingID ? [action.overwork.clearingID] : [];
    default:
      return [];
  }
}

export function factionSpatialClearingIDs(action: Action): number[] {
  switch (action.type) {
    case ACTION_TYPE.SPREAD_SYMPATHY:
      return action.spreadSympathy?.clearingID ? [action.spreadSympathy.clearingID] : [];
    case ACTION_TYPE.REVOLT:
      return action.revolt?.clearingID ? [action.revolt.clearingID] : [];
    case ACTION_TYPE.ORGANIZE:
      return action.organize?.clearingID ? [action.organize.clearingID] : [];
    case ACTION_TYPE.EXPLORE:
      return action.explore?.clearingID ? [action.explore.clearingID] : [];
    case ACTION_TYPE.AID:
      return action.aid?.clearingID ? [action.aid.clearingID] : [];
    case ACTION_TYPE.STRIKE:
      return action.strike?.clearingID ? [action.strike.clearingID] : [];
    default:
      return [];
  }
}

export function buildAssistBoardCandidates(params: {
  actions: Action[];
  battleCandidateIndices: number[];
  movementCandidateIndices: number[];
  buildRecruitCandidateIndices: number[];
  factionSpatialCandidateIndices: number[];
}): AssistBoardCandidates {
  return {
    battleCandidates: params.battleCandidateIndices
      .map((actionIndex) => ({ actionIndex, action: params.actions[actionIndex] }))
      .filter((candidate): candidate is AssistBattleCandidate => candidate.action?.type === ACTION_TYPE.BATTLE),
    movementCandidates: params.movementCandidateIndices
      .map((actionIndex) => ({ actionIndex, action: params.actions[actionIndex], endpoints: movementEndpoints(params.actions[actionIndex]) }))
      .filter((candidate): candidate is AssistMovementCandidate => candidate.endpoints !== null),
    buildRecruitCandidates: params.buildRecruitCandidateIndices
      .map((actionIndex) => ({ actionIndex, action: params.actions[actionIndex], clearingIDs: buildRecruitClearingIDs(params.actions[actionIndex]) }))
      .filter((candidate): candidate is AssistClearingCandidate => candidate.clearingIDs.length > 0),
    factionSpatialCandidates: params.factionSpatialCandidateIndices
      .map((actionIndex) => ({ actionIndex, action: params.actions[actionIndex], clearingIDs: factionSpatialClearingIDs(params.actions[actionIndex]) }))
      .filter((candidate): candidate is AssistClearingCandidate => candidate.clearingIDs.length > 0)
  };
}

export function assistBoardHighlights(params: AssistBoardCandidates & {
  movementSource: AssistMovementSource | null;
}): HighlightedClearing[] {
  const battleHighlights = Array.from(new Set(params.battleCandidates.map((candidate) => candidate.action.battle?.clearingID ?? 0)))
    .filter((clearingID) => clearingID > 0)
    .map((clearingID) => ({ clearingID, role: "affected" as const }));

  if (battleHighlights.length > 0) {
    return battleHighlights;
  }

  const movementHighlights =
    params.movementSource === null
      ? Array.from(new Set(params.movementCandidates.map((candidate) => candidate.endpoints.fromClearingID)))
          .filter((clearingID) => clearingID > 0)
          .map((clearingID) => ({ clearingID, role: "source" as const }))
      : Array.from(
          new Set(
            params.movementCandidates
              .filter((candidate) => movementSourceMatches(candidate.endpoints, params.movementSource as AssistMovementSource))
              .map((candidate) => candidate.endpoints.toClearingID)
          )
        )
          .filter((clearingID) => clearingID > 0)
          .map((clearingID) => ({ clearingID, role: "target" as const }));

  if (movementHighlights.length > 0) {
    return movementHighlights;
  }

  const buildRecruitHighlights = Array.from(new Set(params.buildRecruitCandidates.flatMap((candidate) => candidate.clearingIDs)))
    .filter((clearingID) => clearingID > 0)
    .map((clearingID) => ({ clearingID, role: "affected" as const }));

  if (buildRecruitHighlights.length > 0) {
    return buildRecruitHighlights;
  }

  return Array.from(new Set(params.factionSpatialCandidates.flatMap((candidate) => candidate.clearingIDs)))
    .filter((clearingID) => clearingID > 0)
    .map((clearingID) => ({ clearingID, role: "affected" as const }));
}
