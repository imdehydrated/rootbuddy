export const suitLabels = ["Fox", "Rabbit", "Mouse", "Bird"];
export const phaseLabels = ["Birdsong", "Daylight", "Evening"];
export const stepLabels = ["Unspecified", "Birdsong", "Daylight Craft", "Daylight Actions", "Evening"];
export const setupStageLabels = ["Unspecified", "Marquise Setup", "Eyrie Setup", "Vagabond Setup", "Complete"];
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
  PASS_PHASE: 24,
  ADD_CARD_TO_HAND: 25,
  REMOVE_CARD_FROM_HAND: 26,
  OTHER_PLAYER_DRAW: 27,
  OTHER_PLAYER_PLAY: 28,
  DISCARD_EFFECT: 29,
  ACTIVATE_DOMINANCE: 30,
  TAKE_DOMINANCE: 31,
  MARQUISE_SETUP: 32,
  EYRIE_SETUP: 33,
  VAGABOND_SETUP: 34,
  USE_PERSISTENT_EFFECT: 35,
  EYRIE_EMERGENCY_ORDERS: 36,
  EYRIE_NEW_ROOST: 37,
  VAGABOND_REST: 38,
  VAGABOND_DISCARD: 39,
  VAGABOND_ITEM_CAPACITY: 40,
  EVENING_DISCARD: 41
} as const;
