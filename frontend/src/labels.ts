import type { Action } from "./types";

export const suitLabels = ["Fox", "Rabbit", "Mouse", "Bird"];
export const phaseLabels = ["Birdsong", "Daylight", "Evening"];
export const stepLabels = ["Unspecified", "Recruit", "Daylight Actions", "Evening"];
export const factionLabels = ["Marquise", "Woodland Alliance", "Eyrie", "Vagabond"];
export const buildingLabels = ["Sawmill", "Workshop", "Recruiter"];

export const ACTION_TYPE = {
  MOVEMENT: 0,
  BATTLE: 1,
  BATTLE_RESOLUTION: 2,
  BUILD: 3,
  RECRUIT: 4,
  OVERWORK: 5,
  CRAFT: 6
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
    default:
      return "Unknown action";
  }
}
