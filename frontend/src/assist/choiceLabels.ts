import { describeAction, actionHeadline } from "../actionPresentation";
import { describeKnownCardID } from "../cardCatalog";
import { ACTION_TYPE, eyrieLeaderLabels, factionLabels, itemTypeLabels, suitLabels } from "../labels";
import type { Action, GameState } from "../types";

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
    case ACTION_TYPE.EYRIE_EMERGENCY_ORDERS:
      return `Emergency Orders: draw ${action.eyrieEmergency?.count ?? 1}`;
    case ACTION_TYPE.EYRIE_NEW_ROOST:
      return `New Roost in clearing ${action.eyrieNewRoost?.clearingID ?? "?"}`;
    case ACTION_TYPE.VAGABOND_REST:
      return "Rest";
    case ACTION_TYPE.VAGABOND_DISCARD:
      return `Discard ${(action.vagabondDiscard?.cardIDs ?? []).map(describeKnownCardID).join(", ") || "no cards"}`;
    case ACTION_TYPE.VAGABOND_ITEM_CAPACITY:
      return `Capacity: remove ${(action.vagabondCapacity?.itemIndexes ?? []).map((index) => itemIndexLabel(state, index)).join(", ") || "no items"}`;
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
    case ACTION_TYPE.EYRIE_NEW_ROOST:
      return `New Roost in clearing ${action.eyrieNewRoost?.clearingID ?? "?"}`;
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
      return `Card: ${knownCardLabel(action.aid?.cardID)}; item slot: ${action.aid?.itemIndex ?? "?"}`;
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
