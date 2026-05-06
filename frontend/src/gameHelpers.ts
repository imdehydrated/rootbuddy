import { ACTION_TYPE } from "./labels";
import type { Action, Building, GameState, HighlightedClearing, Token } from "./types";

const MARQUISE_FACTION = 0;
const ALLIANCE_FACTION = 1;
const EYRIE_FACTION = 2;
const RULING_FACTIONS = [MARQUISE_FACTION, ALLIANCE_FACTION, EYRIE_FACTION] as const;
const SAWMILL = 0;
const WORKSHOP = 1;
const RECRUITER = 2;
const ROOST = 3;
const BASE = 4;
const TOKEN_SYMPATHY = 1;
const MARQUISE_WARRIOR_POOL = 25;
const MARQUISE_WOOD_POOL = 8;

export function countBuildings(buildings: Building[], faction: number, type?: number): number {
  return buildings.filter((building) => {
    if (building.faction !== faction) {
      return false;
    }
    return type === undefined ? true : building.type === type;
  }).length;
}

export function countTokens(tokens: Token[], faction?: number, type?: number): number {
  return tokens.filter((token) => {
    if (faction !== undefined && token.faction !== faction) {
      return false;
    }
    return type === undefined ? true : token.type === type;
  }).length;
}

export function suitClass(suit: number): string {
  switch (suit) {
    case 0:
      return "fox";
    case 1:
      return "rabbit";
    case 2:
      return "mouse";
    default:
      return "bird";
  }
}

export function clearingPosition(
  clearingID: number,
  index: number,
  positions: Record<number, { left: number; top: number }>
) {
  const known = positions[clearingID];
  if (known) {
    return {
      left: `${known.left}%`,
      top: `${known.top}%`
    };
  }

  const columns = 4;
  const row = Math.floor(index / columns);
  const column = index % columns;
  return {
    left: `${18 + column * 20}%`,
    top: `${20 + row * 22}%`
  };
}

export function hasAnyIndicators(values: number[], extraPresence: boolean): boolean {
  return extraPresence || values.some((value) => value > 0);
}

export function usedBuildSlots(clearing: GameState["map"]["clearings"][number]): number {
  return clearing.buildings.length + (clearing.ruins ? 1 : 0);
}

export function openBuildSlots(clearing: GameState["map"]["clearings"][number]): number {
  return Math.max(0, clearing.buildSlots - usedBuildSlots(clearing));
}

export function countWarriors(warriors: Record<string, number>, faction: number): number {
  return warriors[String(faction)] ?? 0;
}

export function rulingPresence(clearing: GameState["map"]["clearings"][number], faction: number): number {
  return countWarriors(clearing.warriors, faction) + countBuildings(clearing.buildings, faction);
}

export function rulerOfClearing(clearing: GameState["map"]["clearings"][number]): number | null {
  let leadingFaction: number | null = null;
  let leadingPresence = 0;
  let tied = false;

  for (const faction of RULING_FACTIONS) {
    const presence = rulingPresence(clearing, faction);
    if (presence === 0) {
      continue;
    }
    if (presence > leadingPresence) {
      leadingFaction = faction;
      leadingPresence = presence;
      tied = false;
      continue;
    }
    if (presence === leadingPresence) {
      tied = true;
    }
  }

  if (leadingFaction === null || tied) {
    return null;
  }

  return leadingFaction;
}

export function syncDerivedFactionStateFromBoard(state: GameState): void {
  const totalMarquiseWarriors = state.map.clearings.reduce(
    (sum, clearing) => sum + countWarriors(clearing.warriors, MARQUISE_FACTION),
    0
  );

  state.marquise.warriorSupply = Math.max(0, MARQUISE_WARRIOR_POOL - totalMarquiseWarriors);
  const totalMarquiseWood = state.map.clearings.reduce((sum, clearing) => sum + clearing.wood, 0);
  state.marquise.woodSupply = Math.max(0, MARQUISE_WOOD_POOL - totalMarquiseWood);
  state.marquise.sawmillsPlaced = state.map.clearings.reduce(
    (sum, clearing) => sum + countBuildings(clearing.buildings, MARQUISE_FACTION, SAWMILL),
    0
  );
  state.marquise.workshopsPlaced = state.map.clearings.reduce(
    (sum, clearing) => sum + countBuildings(clearing.buildings, MARQUISE_FACTION, WORKSHOP),
    0
  );
  state.marquise.recruitersPlaced = state.map.clearings.reduce(
    (sum, clearing) => sum + countBuildings(clearing.buildings, MARQUISE_FACTION, RECRUITER),
    0
  );

  if (!state.map.clearings.some((clearing) => clearing.id === state.marquise.keepClearingID)) {
    state.marquise.keepClearingID = 0;
  }

  state.eyrie.roostsPlaced = state.map.clearings.reduce(
    (sum, clearing) => sum + countBuildings(clearing.buildings, EYRIE_FACTION, ROOST),
    0
  );

  state.alliance.sympathyPlaced = state.map.clearings.reduce(
    (sum, clearing) => sum + countTokens(clearing.tokens, ALLIANCE_FACTION, TOKEN_SYMPATHY),
    0
  );
  state.alliance.foxBasePlaced = state.map.clearings.some(
    (clearing) => clearing.suit === 0 && countBuildings(clearing.buildings, ALLIANCE_FACTION, BASE) > 0
  );
  state.alliance.rabbitBasePlaced = state.map.clearings.some(
    (clearing) => clearing.suit === 1 && countBuildings(clearing.buildings, ALLIANCE_FACTION, BASE) > 0
  );
  state.alliance.mouseBasePlaced = state.map.clearings.some(
    (clearing) => clearing.suit === 2 && countBuildings(clearing.buildings, ALLIANCE_FACTION, BASE) > 0
  );
}

export function affectedClearings(action: Action): HighlightedClearing[] {
  const highlighted = new Map<number, HighlightedClearing["role"]>();

  const add = (clearingID: number | undefined, role: HighlightedClearing["role"]) => {
    if (clearingID === undefined) {
      return;
    }

    const existing = highlighted.get(clearingID);
    if (existing === "source" || (existing === "target" && role === "affected")) {
      return;
    }

    highlighted.set(clearingID, role);
  };

  switch (action.type) {
    case ACTION_TYPE.MOVEMENT:
      add(action.movement?.from, "source");
      add(action.movement?.to, "target");
      break;
    case ACTION_TYPE.BATTLE:
      add(action.battle?.clearingID, "affected");
      break;
    case ACTION_TYPE.BATTLE_RESOLUTION:
      add(action.battleResolution?.clearingID, "affected");
      break;
    case ACTION_TYPE.BUILD:
      for (const source of action.build?.woodSources ?? []) {
        add(source.clearingID, "source");
      }
      add(action.build?.clearingID, "affected");
      break;
    case ACTION_TYPE.RECRUIT:
      for (const clearingID of action.recruit?.clearingIDs ?? []) {
        add(clearingID, "affected");
      }
      break;
    case ACTION_TYPE.OVERWORK:
      add(action.overwork?.clearingID, "affected");
      break;
    case ACTION_TYPE.CRAFT:
      for (const clearingID of action.craft?.usedWorkshopClearings ?? []) {
        add(clearingID, "affected");
      }
      break;
    case ACTION_TYPE.SPREAD_SYMPATHY:
      add(action.spreadSympathy?.clearingID, "affected");
      break;
    case ACTION_TYPE.REVOLT:
      add(action.revolt?.clearingID, "affected");
      break;
    case ACTION_TYPE.ORGANIZE:
      add(action.organize?.clearingID, "affected");
      break;
    case ACTION_TYPE.EXPLORE:
      add(action.explore?.clearingID, "affected");
      break;
    case ACTION_TYPE.AID:
      add(action.aid?.clearingID, "affected");
      break;
    case ACTION_TYPE.STRIKE:
      add(action.strike?.clearingID, "affected");
      break;
    case ACTION_TYPE.SLIP:
      add(action.slip?.from, "source");
      add(action.slip?.to, "target");
      break;
    case ACTION_TYPE.BIRDSONG_WOOD:
      for (const clearingID of action.birdsongWood?.clearingIDs ?? []) {
        add(clearingID, "affected");
      }
      break;
    case ACTION_TYPE.EYRIE_NEW_ROOST:
      add(action.eyrieNewRoost?.clearingID, "affected");
      break;
    case ACTION_TYPE.MARQUISE_SETUP:
      add(action.marquiseSetup?.keepClearingID, "affected");
      add(action.marquiseSetup?.sawmillClearingID, "affected");
      add(action.marquiseSetup?.workshopClearingID, "affected");
      add(action.marquiseSetup?.recruiterClearingID, "affected");
      break;
    case ACTION_TYPE.EYRIE_SETUP:
      add(action.eyrieSetup?.clearingID, "affected");
      break;
    case ACTION_TYPE.USE_PERSISTENT_EFFECT:
      add(action.usePersistentEffect?.clearingID, "affected");
      break;
    case ACTION_TYPE.FIELD_HOSPITALS:
      add(action.fieldHospitals?.clearingID, "source");
      break;
    default:
      break;
  }

  return Array.from(highlighted, ([clearingID, role]) => ({ clearingID, role }));
}

export function actionTouchesClearing(action: Action, clearingID: number): boolean {
  return affectedClearings(action).some((highlight) => highlight.clearingID === clearingID);
}
