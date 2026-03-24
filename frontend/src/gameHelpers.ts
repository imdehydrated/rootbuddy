import { ACTION_TYPE } from "./labels";
import type { Action, Building, GameState, HighlightedClearing, Token } from "./types";

const MARQUISE_FACTION = 0;
const ALLIANCE_FACTION = 1;
const EYRIE_FACTION = 2;
const SAWMILL = 0;
const WORKSHOP = 1;
const RECRUITER = 2;
const ROOST = 3;
const BASE = 4;
const TOKEN_SYMPATHY = 1;
const MARQUISE_WARRIOR_POOL = 25;

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

export function countWarriors(warriors: Record<string, number>, faction: number): number {
  return warriors[String(faction)] ?? 0;
}

export function syncDerivedFactionStateFromBoard(state: GameState): void {
  const totalMarquiseWarriors = state.map.clearings.reduce(
    (sum, clearing) => sum + countWarriors(clearing.warriors, MARQUISE_FACTION),
    0
  );

  state.marquise.warriorSupply = Math.max(0, MARQUISE_WARRIOR_POOL - totalMarquiseWarriors);
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
    default:
      break;
  }

  return Array.from(highlighted, ([clearingID, role]) => ({ clearingID, role }));
}
