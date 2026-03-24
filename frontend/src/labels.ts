import type { Action } from "./types";

export const suitLabels = ["Fox", "Rabbit", "Mouse", "Bird"];
export const phaseLabels = ["Birdsong", "Daylight", "Evening"];
export const stepLabels = ["Unspecified", "Birdsong", "Daylight Craft", "Daylight Actions", "Evening"];
export const factionLabels = ["Marquise", "Woodland Alliance", "Eyrie", "Vagabond"];
export const buildingLabels = ["Sawmill", "Workshop", "Recruiter", "Roost", "Base"];
export const eyrieLeaderLabels = ["Builder", "Charismatic", "Commander", "Despot"];
export const vagabondCharacterLabels = ["Thief", "Tinker", "Ranger"];
export const itemTypeLabels = ["Tea", "Coin", "Crossbow", "Hammer", "Sword", "Torch", "Boots", "Bag"];
export const itemStatusLabels = ["Ready", "Exhausted", "Damaged"];
export const relationshipLabels = ["Hostile", "Indifferent", "Friendly", "Allied"];

export const ACTION_TYPE = {
  MOVEMENT: 0,
  BATTLE: 1,
  BATTLE_RESOLUTION: 2,
  BUILD: 3,
  RECRUIT: 4,
  OVERWORK: 5,
  CRAFT: 6,
  ADD_TO_DECREE: 7,
  SPREAD_SYMPATHY: 8,
  REVOLT: 9,
  MOBILIZE: 10,
  TRAIN: 11,
  ORGANIZE: 12,
  EXPLORE: 13,
  QUEST: 14,
  AID: 15,
  STRIKE: 16,
  REPAIR: 17,
  TURMOIL: 18,
  DAYBREAK: 19,
  SLIP: 20,
  BIRDSONG_WOOD: 21,
  EVENING_DRAW: 22,
  SCORE_ROOSTS: 23,
  PASS_PHASE: 24
} as const;

export function describeAction(action: Action): string {
  switch (action.type) {
    case ACTION_TYPE.MOVEMENT:
      return `Move up to ${action.movement?.maxCount ?? 0} from ${action.movement?.from ?? "?"} to ${action.movement?.to ?? "?"}`;
    case ACTION_TYPE.BATTLE:
      return `Battle ${factionLabels[action.battle?.targetFaction ?? 0] ?? "Unknown"} in clearing ${action.battle?.clearingID ?? "?"}`;
    case ACTION_TYPE.BATTLE_RESOLUTION:
      return `Resolved battle in clearing ${action.battleResolution?.clearingID ?? "?"}`;
    case ACTION_TYPE.BUILD:
      return `Build ${buildingLabels[action.build?.buildingType ?? 0] ?? "building"} in clearing ${action.build?.clearingID ?? "?"}`;
    case ACTION_TYPE.RECRUIT:
      return `Recruit in clearings ${(action.recruit?.clearingIDs ?? []).join(", ")}`;
    case ACTION_TYPE.OVERWORK:
      return `Overwork in clearing ${action.overwork?.clearingID ?? "?"}`;
    case ACTION_TYPE.CRAFT:
      return `Craft card ${action.craft?.cardID ?? "?"}`;
    case ACTION_TYPE.ADD_TO_DECREE:
      return `Add decree cards ${(action.addToDecree?.cardIDs ?? []).join(", ")}`;
    case ACTION_TYPE.SPREAD_SYMPATHY:
      return `Spread sympathy to clearing ${action.spreadSympathy?.clearingID ?? "?"}`;
    case ACTION_TYPE.REVOLT:
      return `Revolt in clearing ${action.revolt?.clearingID ?? "?"}`;
    case ACTION_TYPE.MOBILIZE:
      return `Mobilize card ${action.mobilize?.cardID ?? "?"}`;
    case ACTION_TYPE.TRAIN:
      return `Train with card ${action.train?.cardID ?? "?"}`;
    case ACTION_TYPE.ORGANIZE:
      return `Organize in clearing ${action.organize?.clearingID ?? "?"}`;
    case ACTION_TYPE.EXPLORE:
      return `Explore ruins in clearing ${action.explore?.clearingID ?? "?"}`;
    case ACTION_TYPE.QUEST:
      return `Complete quest ${action.quest?.questID ?? "?"}`;
    case ACTION_TYPE.AID:
      return `Aid ${factionLabels[action.aid?.targetFaction ?? 0] ?? "Unknown"} in clearing ${action.aid?.clearingID ?? "?"}`;
    case ACTION_TYPE.STRIKE:
      return `Strike ${factionLabels[action.strike?.targetFaction ?? 0] ?? "Unknown"} in clearing ${action.strike?.clearingID ?? "?"}`;
    case ACTION_TYPE.REPAIR:
      return `Repair item ${action.repair?.itemIndex ?? "?"}`;
    case ACTION_TYPE.TURMOIL:
      return "Go into turmoil";
    case ACTION_TYPE.DAYBREAK:
      return `Refresh ${action.daybreak?.refreshedItemIndexes?.length ?? 0} item(s)`;
    case ACTION_TYPE.SLIP:
      return `Slip to ${action.slip?.toForestID ? `forest ${action.slip.toForestID}` : action.slip?.to ?? "?"}`;
    case ACTION_TYPE.BIRDSONG_WOOD:
      return `Place wood in clearings ${(action.birdsongWood?.clearingIDs ?? []).join(", ")}`;
    case ACTION_TYPE.EVENING_DRAW:
      return `Draw ${action.eveningDraw?.count ?? 0} card(s)`;
    case ACTION_TYPE.SCORE_ROOSTS:
      return `Score ${action.scoreRoosts?.points ?? 0} roost point(s)`;
    case ACTION_TYPE.PASS_PHASE:
      return "Pass phase";
    default:
      return "Unknown action";
  }
}
