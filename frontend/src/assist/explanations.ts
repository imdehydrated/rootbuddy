import { describeKnownCardID } from "../cardCatalog";
import { ACTION_TYPE, buildingLabels, factionLabels, itemTypeLabels } from "../labels";
import type { Action, GameState } from "../types";

function factionLabel(faction: number | undefined) {
  return factionLabels[faction ?? -1] ?? "the target faction";
}

export function actionExplanation(action: Action, state: GameState): string {
  switch (action.type) {
    case ACTION_TYPE.MOVEMENT:
      return `Reposition warriors from clearing ${action.movement?.from ?? "?"} to ${action.movement?.to ?? "?"}. This keeps rule, battles, and build lines open from the current board.`;
    case ACTION_TYPE.BATTLE:
      return `Fight ${factionLabel(action.battle?.targetFaction)} in clearing ${action.battle?.clearingID ?? "?"}. Use this when you want to clear rule, remove pieces, or force the battle flow to resolve the conflict now.`;
    case ACTION_TYPE.BUILD:
      return `Spend available wood to place a ${buildingLabels[action.build?.buildingType ?? 0] ?? "building"} in clearing ${action.build?.clearingID ?? "?"}. This advances your engine rather than just trading tempo for a one-off action.`;
    case ACTION_TYPE.RECRUIT:
      return `Add warriors to clearing ${action.recruit?.clearingIDs.join(", ") || "?"}. Recruiting improves rule and gives you more material for later battles, moves, or faction actions.`;
    case ACTION_TYPE.OVERWORK:
      return `Spend ${describeKnownCardID(action.overwork?.cardID ?? 0)} to add wood in clearing ${action.overwork?.clearingID ?? "?"}. This is strongest when one extra wood unlocks a build line you would otherwise miss.`;
    case ACTION_TYPE.CRAFT:
      return `Craft ${describeKnownCardID(action.craft?.cardID ?? 0)} using workshops in clearings ${(action.craft?.usedWorkshopClearings ?? []).join(", ") || "?"}. Take this before other actions if the craft is your best scoring or effect window.`;
    case ACTION_TYPE.ADD_TO_DECREE:
      return `Commit ${(action.addToDecree?.cardIDs ?? []).map((cardID) => describeKnownCardID(cardID)).join(", ")} to the decree. Only add this if the future recruit, move, battle, or build obligations still look safe on the board.`;
    case ACTION_TYPE.SPREAD_SYMPATHY:
      return `Place sympathy in clearing ${action.spreadSympathy?.clearingID ?? "?"}. This adds pressure to enemy movement and helps set up future Alliance scoring and revolts.`;
    case ACTION_TYPE.REVOLT:
      return `Revolt in clearing ${action.revolt?.clearingID ?? "?"}. This is the Alliance's biggest swing because it flips a clearing into a base and opens officer growth.`;
    case ACTION_TYPE.MOBILIZE:
      return `Move ${describeKnownCardID(action.mobilize?.cardID ?? 0)} into supporters. Do this when you need better suits for sympathy or revolt more than you need another card in hand.`;
    case ACTION_TYPE.TRAIN:
      return `Spend ${describeKnownCardID(action.train?.cardID ?? 0)} to gain an officer. Officers convert your board presence into extra daylight actions later.`;
    case ACTION_TYPE.ORGANIZE:
      return `Turn one Alliance warrior into sympathy in clearing ${action.organize?.clearingID ?? "?"}. This is a direct scoring and board-pressure conversion.`;
    case ACTION_TYPE.EXPLORE:
      return `Explore the ruin in clearing ${action.explore?.clearingID ?? "?"} to claim ${itemTypeLabels[action.explore?.itemType ?? -1] ?? "an item"}. Vagabond tempo usually improves immediately when a ruin becomes gear.`;
    case ACTION_TYPE.QUEST:
      return `Complete the available quest from the current board position. Quests are a low-conflict way for the Vagabond to convert ready items into score or cards.`;
    case ACTION_TYPE.AID:
      return `Aid ${factionLabel(action.aid?.targetFaction)} in clearing ${action.aid?.clearingID ?? "?"} with ${describeKnownCardID(action.aid?.cardID ?? 0)}. This grows Vagabond relationships without forcing a fight.`;
    case ACTION_TYPE.STRIKE:
      return `Spend a crossbow shot to remove a key enemy piece in clearing ${action.strike?.clearingID ?? "?"}. Use it when one token, building, or warrior is blocking your route or score line.`;
    case ACTION_TYPE.REPAIR:
      return `Repair damaged item slot ${action.repair?.itemIndex ?? "?"}. Repairs recover future flexibility, especially if a key boot, sword, or hammer is offline.`;
    case ACTION_TYPE.TURMOIL:
      return "Go into turmoil and reset the Eyrie decree. This is a fallback when the current decree is no longer satisfiable, not a routine tempo play.";
    case ACTION_TYPE.DAYBREAK:
      return `Refresh ${action.daybreak?.refreshedItemIndexes.length ?? 0} Vagabond item(s). This converts exhausted gear back into usable actions for the turn.`;
    case ACTION_TYPE.SLIP:
      return `Slip to ${action.slip?.toForestID ? `forest ${action.slip.toForestID}` : `clearing ${action.slip?.to ?? "?"}`}. This is the Vagabond's safe reposition before committing to aid, quest, battle, or explore lines.`;
    case ACTION_TYPE.BIRDSONG_WOOD:
      return `Place wood in clearings ${(action.birdsongWood?.clearingIDs ?? []).join(", ") || "?"}. This is board setup for the Marquise build engine rather than immediate score.`;
    case ACTION_TYPE.EVENING_DRAW:
      return `Draw ${action.eveningDraw?.count ?? 0} card(s). Take this when the turn is functionally complete and more board actions are no longer available.`;
    case ACTION_TYPE.SCORE_ROOSTS:
      return `Score ${action.scoreRoosts?.points ?? 0} point(s) from roost count. This is bookkeeping progress for the Eyrie end step.`;
    case ACTION_TYPE.PASS_PHASE:
      return `Advance the turn flow from ${factionLabels[state.factionTurn] ?? "the active faction"}'s current phase. Use this only when the meaningful actions for the phase are actually done.`;
    case ACTION_TYPE.ACTIVATE_DOMINANCE:
      return `Switch to dominance using ${describeKnownCardID(action.activateDominance?.cardID ?? 0)}. This replaces the normal VP race with a board-control win condition.`;
    case ACTION_TYPE.TAKE_DOMINANCE:
      return `Take the available dominance card and spend ${describeKnownCardID(action.takeDominance?.spentCardID ?? 0)}. This keeps the dominance option live for a future conversion.`;
    case ACTION_TYPE.USE_PERSISTENT_EFFECT:
      switch (action.usePersistentEffect?.effectID) {
        case "better_burrow_bank":
          return `Use Better Burrow Bank with ${factionLabel(action.usePersistentEffect.targetFaction)} to trade the effect for immediate shared draw value.`;
        case "codebreakers":
          return `Use Codebreakers on ${factionLabel(action.usePersistentEffect.targetFaction)} when hand visibility matters more than saving the effect for later.`;
        case "royal_claim":
          return "Use Royal Claim now to cash the persistent effect into a VP burst before the board position changes.";
        case "stand_and_deliver":
          return `Use Stand and Deliver! against ${factionLabel(action.usePersistentEffect.targetFaction)} to convert an exposed hand into tempo and VP.`;
        case "tax_collector":
          return `Use Tax Collector in clearing ${action.usePersistentEffect.clearingID ?? "?"} to turn one warrior into a card when hand flow matters more than board presence.`;
        default:
          return "Use the persistent effect now to convert a stored card effect into immediate tempo or information.";
      }
    default:
      return "Follow the exact legal action when it matches what happened on the board. The explanation is intentionally conservative because the server and rule engine remain authoritative.";
  }
}
