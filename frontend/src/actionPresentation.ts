import { describeKnownCardID } from "./cardCatalog";
import { ACTION_TYPE, factionLabels, buildingLabels, suitLabels } from "./labels";
import type { Action, Card, GameState } from "./types";

export type ActionCardReference = {
  cardID: number;
  zoneLabel: string;
};

type VisibleCardRef = {
  card: Card;
  zoneLabel: string;
};

function findQuestByID(state: GameState | undefined, questID: number | undefined) {
  if (!state || !questID) {
    return null;
  }

  return [...state.vagabond.questsAvailable, ...state.vagabond.questsCompleted].find((quest) => quest.id === questID) ?? null;
}

function describeQuestLabel(questID: number | undefined, state?: GameState): string {
  if (!questID) {
    return "Quest ?";
  }

  const quest = findQuestByID(state, questID);
  if (!quest) {
    return `Quest ${questID}`;
  }

  return `${quest.name} (${suitLabels[quest.suit] ?? "Unknown"})`;
}

export function createVisibleCardLookup(state: GameState): Map<number, VisibleCardRef> {
  const lookup = new Map<number, VisibleCardRef>();

  const register = (cards: Card[], zoneLabel: string) => {
    cards.forEach((card) => {
      lookup.set(card.id, { card, zoneLabel });
    });
  };

  switch (state.playerFaction) {
    case 0:
      register(state.marquise.cardsInHand, "Hand");
      break;
    case 1:
      register(state.alliance.cardsInHand, "Hand");
      register(state.alliance.supporters, "Supporter");
      break;
    case 2:
      register(state.eyrie.cardsInHand, "Hand");
      break;
    case 3:
      register(state.vagabond.cardsInHand, "Hand");
      break;
    default:
      break;
  }

  return lookup;
}

export function formatCardReferenceLabel(reference: ActionCardReference, lookup: Map<number, VisibleCardRef>): string {
  const visible = lookup.get(reference.cardID);
  const zoneLabel = visible?.zoneLabel ?? reference.zoneLabel;
  const cardLabel = visible ? `${visible.card.name} (${suitLabels[visible.card.suit] ?? "Unknown"})` : describeKnownCardID(reference.cardID);
  return `${zoneLabel}: ${cardLabel}`;
}

export function actionHeadline(action: Action): string {
  switch (action.type) {
    case ACTION_TYPE.MOVEMENT:
      return "Move";
    case ACTION_TYPE.SLIP:
      return "Slip";
    case ACTION_TYPE.CRAFT:
      return "Craft";
    case ACTION_TYPE.BUILD:
      return "Build";
    case ACTION_TYPE.RECRUIT:
      return "Recruit";
    case ACTION_TYPE.ORGANIZE:
      return "Organize";
    case ACTION_TYPE.EXPLORE:
      return "Explore";
    case ACTION_TYPE.AID:
      return "Aid";
    case ACTION_TYPE.STRIKE:
      return "Strike";
    case ACTION_TYPE.ADD_TO_DECREE:
      return "Add To Decree";
    case ACTION_TYPE.SPREAD_SYMPATHY:
      return "Spread Sympathy";
    case ACTION_TYPE.REVOLT:
      return "Revolt";
    case ACTION_TYPE.TRAIN:
      return "Train";
    case ACTION_TYPE.MOBILIZE:
      return "Mobilize";
    case ACTION_TYPE.OVERWORK:
      return "Overwork";
    case ACTION_TYPE.ACTIVATE_DOMINANCE:
      return "Activate Dominance";
    case ACTION_TYPE.TAKE_DOMINANCE:
      return "Take Dominance";
    case ACTION_TYPE.BATTLE:
      return "Battle";
    case ACTION_TYPE.OTHER_PLAYER_DRAW:
      return "Draw";
    case ACTION_TYPE.OTHER_PLAYER_PLAY:
      return "Play";
    default:
      return "Action";
  }
}

export function actionCardReferences(action: Action): ActionCardReference[] {
  switch (action.type) {
    case ACTION_TYPE.CRAFT:
      return action.craft?.cardID ? [{ cardID: action.craft.cardID, zoneLabel: "Hand" }] : [];
    case ACTION_TYPE.OVERWORK:
      return action.overwork?.cardID ? [{ cardID: action.overwork.cardID, zoneLabel: "Hand" }] : [];
    case ACTION_TYPE.ADD_TO_DECREE:
      return (action.addToDecree?.cardIDs ?? []).map((cardID) => ({ cardID, zoneLabel: "Hand" }));
    case ACTION_TYPE.SPREAD_SYMPATHY:
      return (action.spreadSympathy?.supporterCardIDs ?? []).map((cardID) => ({ cardID, zoneLabel: "Supporter" }));
    case ACTION_TYPE.REVOLT:
      return (action.revolt?.supporterCardIDs ?? []).map((cardID) => ({ cardID, zoneLabel: "Supporter" }));
    case ACTION_TYPE.MOBILIZE:
      return action.mobilize?.cardID ? [{ cardID: action.mobilize.cardID, zoneLabel: "Hand" }] : [];
    case ACTION_TYPE.TRAIN:
      return action.train?.cardID ? [{ cardID: action.train.cardID, zoneLabel: "Hand" }] : [];
    case ACTION_TYPE.ACTIVATE_DOMINANCE:
      return action.activateDominance?.cardID ? [{ cardID: action.activateDominance.cardID, zoneLabel: "Dominance" }] : [];
    case ACTION_TYPE.TAKE_DOMINANCE:
      return [
        action.takeDominance?.dominanceCardID ? { cardID: action.takeDominance.dominanceCardID, zoneLabel: "Dominance" } : null,
        action.takeDominance?.spentCardID ? { cardID: action.takeDominance.spentCardID, zoneLabel: "Hand" } : null
      ].filter((reference): reference is ActionCardReference => reference !== null);
    case ACTION_TYPE.OTHER_PLAYER_PLAY:
      return action.otherPlayerPlay?.cardID ? [{ cardID: action.otherPlayerPlay.cardID, zoneLabel: "Played" }] : [];
    case ACTION_TYPE.DISCARD_EFFECT:
      return action.discardEffect?.cardID ? [{ cardID: action.discardEffect.cardID, zoneLabel: "Effect" }] : [];
    case ACTION_TYPE.ADD_CARD_TO_HAND:
      return action.addCardToHand?.cardID ? [{ cardID: action.addCardToHand.cardID, zoneLabel: "Hand" }] : [];
    case ACTION_TYPE.REMOVE_CARD_FROM_HAND:
      return action.removeCardFromHand?.cardID ? [{ cardID: action.removeCardFromHand.cardID, zoneLabel: "Hand" }] : [];
    default:
      return [];
  }
}

export function actionContextTags(action: Action): string[] {
  switch (action.type) {
    case ACTION_TYPE.MOVEMENT:
      return [`${action.movement?.from ?? "?"} -> ${action.movement?.to ?? "?"}`];
    case ACTION_TYPE.BATTLE:
      return [`Clearing ${action.battle?.clearingID ?? "?"}`];
    case ACTION_TYPE.BUILD:
      return [`Clearing ${action.build?.clearingID ?? "?"}`];
    case ACTION_TYPE.CRAFT:
      return (action.craft?.usedWorkshopClearings ?? []).map((clearingID) => `Workshop ${clearingID}`);
    case ACTION_TYPE.RECRUIT:
      return (action.recruit?.clearingIDs ?? []).map((clearingID) => `Clearing ${clearingID}`);
    case ACTION_TYPE.SPREAD_SYMPATHY:
      return [`Clearing ${action.spreadSympathy?.clearingID ?? "?"}`];
    case ACTION_TYPE.REVOLT:
      return [`Clearing ${action.revolt?.clearingID ?? "?"}`];
    case ACTION_TYPE.ADD_TO_DECREE:
      return (action.addToDecree?.columns ?? []).map((column) => `Column ${column}`);
    case ACTION_TYPE.AID:
      return [`Clearing ${action.aid?.clearingID ?? "?"}`, `Target ${factionLabels[action.aid?.targetFaction ?? -1] ?? "?"}`];
    case ACTION_TYPE.OTHER_PLAYER_DRAW:
      return [`Count ${action.otherPlayerDraw?.count ?? 0}`];
    default:
      return [];
  }
}

function effectLabel(effectID: string): string {
  switch (effectID) {
    case "better_burrow_bank":
      return "Better Burrow Bank";
    case "codebreakers":
      return "Codebreakers";
    case "royal_claim":
      return "Royal Claim";
    case "stand_and_deliver":
      return "Stand and Deliver!";
    case "tax_collector":
      return "Tax Collector";
    default:
      return effectID || "?";
  }
}

function describeBoardLocation(clearingID: number | undefined, forestID: number | undefined): string {
  if (forestID && forestID > 0) {
    return `forest ${forestID}`;
  }
  if (clearingID && clearingID > 0) {
    return `clearing ${clearingID}`;
  }
  return "?";
}

export function describeAction(action: Action, state?: GameState): string {
  switch (action.type) {
    case ACTION_TYPE.MOVEMENT:
      return `Move up to ${action.movement?.maxCount ?? 0} from ${describeBoardLocation(action.movement?.from, action.movement?.fromForestID)} to ${describeBoardLocation(action.movement?.to, action.movement?.toForestID)}`;
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
      return `Craft ${describeKnownCardID(action.craft?.cardID ?? 0)}`;
    case ACTION_TYPE.ADD_TO_DECREE:
      return `Add decree cards ${(action.addToDecree?.cardIDs ?? []).map((cardID) => describeKnownCardID(cardID)).join(", ")}`;
    case ACTION_TYPE.SPREAD_SYMPATHY:
      return `Spread sympathy to clearing ${action.spreadSympathy?.clearingID ?? "?"}`;
    case ACTION_TYPE.REVOLT:
      return `Revolt in clearing ${action.revolt?.clearingID ?? "?"}`;
    case ACTION_TYPE.MOBILIZE:
      return `Mobilize ${describeKnownCardID(action.mobilize?.cardID ?? 0)}`;
    case ACTION_TYPE.TRAIN:
      return `Train with ${describeKnownCardID(action.train?.cardID ?? 0)}`;
    case ACTION_TYPE.ORGANIZE:
      return `Organize in clearing ${action.organize?.clearingID ?? "?"}`;
    case ACTION_TYPE.EXPLORE:
      return `Explore ruins in clearing ${action.explore?.clearingID ?? "?"}`;
    case ACTION_TYPE.QUEST:
      return `Complete quest ${describeQuestLabel(action.quest?.questID, state)}`;
    case ACTION_TYPE.AID:
      return `Aid ${factionLabels[action.aid?.targetFaction ?? 0] ?? "Unknown"} in clearing ${action.aid?.clearingID ?? "?"}`;
    case ACTION_TYPE.STRIKE:
      return `Strike ${factionLabels[action.strike?.targetFaction ?? 0] ?? "Unknown"} in clearing ${action.strike?.clearingID ?? "?"}`;
    case ACTION_TYPE.REPAIR:
      return `Repair item slot ${action.repair?.itemIndex ?? "?"}`;
    case ACTION_TYPE.TURMOIL:
      return "Go into turmoil";
    case ACTION_TYPE.DAYBREAK:
      return `Refresh ${action.daybreak?.refreshedItemIndexes?.length ?? 0} item(s)`;
    case ACTION_TYPE.SLIP:
      return `Slip from ${describeBoardLocation(action.slip?.from, action.slip?.fromForestID)} to ${describeBoardLocation(action.slip?.to, action.slip?.toForestID)}`;
    case ACTION_TYPE.BIRDSONG_WOOD:
      return `Place wood in clearings ${(action.birdsongWood?.clearingIDs ?? []).join(", ")}`;
    case ACTION_TYPE.EVENING_DRAW:
      return `Draw ${action.eveningDraw?.count ?? 0} card(s)`;
    case ACTION_TYPE.SCORE_ROOSTS:
      return `Score ${action.scoreRoosts?.points ?? 0} roost point(s)`;
    case ACTION_TYPE.PASS_PHASE:
      return "Pass phase";
    case ACTION_TYPE.ADD_CARD_TO_HAND:
      return `Add ${describeKnownCardID(action.addCardToHand?.cardID ?? 0)} to hand`;
    case ACTION_TYPE.REMOVE_CARD_FROM_HAND:
      return `Remove ${describeKnownCardID(action.removeCardFromHand?.cardID ?? 0)} from hand`;
    case ACTION_TYPE.OTHER_PLAYER_DRAW:
      return `Record ${factionLabels[action.otherPlayerDraw?.faction ?? 0] ?? "Unknown"} drawing ${action.otherPlayerDraw?.count ?? 0} card(s)`;
    case ACTION_TYPE.OTHER_PLAYER_PLAY:
      return `Record ${factionLabels[action.otherPlayerPlay?.faction ?? 0] ?? "Unknown"} playing ${describeKnownCardID(action.otherPlayerPlay?.cardID ?? 0)}`;
    case ACTION_TYPE.DISCARD_EFFECT:
      return `Discard effect ${describeKnownCardID(action.discardEffect?.cardID ?? 0)}`;
    case ACTION_TYPE.ACTIVATE_DOMINANCE:
      return `Activate dominance ${describeKnownCardID(action.activateDominance?.cardID ?? 0)}`;
    case ACTION_TYPE.TAKE_DOMINANCE:
      return `Take dominance ${describeKnownCardID(action.takeDominance?.dominanceCardID ?? 0)}`;
    case ACTION_TYPE.MARQUISE_SETUP:
      return `Marquise setup: keep ${action.marquiseSetup?.keepClearingID ?? "?"}, sawmill ${action.marquiseSetup?.sawmillClearingID ?? "?"}, workshop ${action.marquiseSetup?.workshopClearingID ?? "?"}, recruiter ${action.marquiseSetup?.recruiterClearingID ?? "?"}`;
    case ACTION_TYPE.EYRIE_SETUP:
      return `Eyrie setup: start in clearing ${action.eyrieSetup?.clearingID ?? "?"}`;
    case ACTION_TYPE.VAGABOND_SETUP:
      return `Vagabond setup: start in forest ${action.vagabondSetup?.forestID ?? "?"}`;
    case ACTION_TYPE.USE_PERSISTENT_EFFECT: {
      const effectID = action.usePersistentEffect?.effectID ?? "";
      switch (effectID) {
        case "better_burrow_bank":
          return `Use Better Burrow Bank with ${factionLabels[action.usePersistentEffect?.targetFaction ?? 0] ?? "Unknown"}`;
        case "codebreakers":
          return `Use Codebreakers on ${factionLabels[action.usePersistentEffect?.targetFaction ?? 0] ?? "Unknown"}`;
        case "royal_claim":
          return "Use Royal Claim";
        case "stand_and_deliver":
          return `Use Stand and Deliver! on ${factionLabels[action.usePersistentEffect?.targetFaction ?? 0] ?? "Unknown"}`;
        case "tax_collector":
          return `Use Tax Collector in clearing ${action.usePersistentEffect?.clearingID ?? "?"}`;
        default:
          return `Use persistent effect ${effectLabel(effectID)}`;
      }
    }
    default:
      return "Unknown action";
  }
}
