import { ACTION_TYPE } from "../labels";
import type { Action } from "../types";

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
    case ACTION_TYPE.EYRIE_EMERGENCY_ORDERS:
    case ACTION_TYPE.EYRIE_NEW_ROOST:
    case ACTION_TYPE.VAGABOND_REST:
    case ACTION_TYPE.VAGABOND_DISCARD:
    case ACTION_TYPE.VAGABOND_ITEM_CAPACITY:
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
