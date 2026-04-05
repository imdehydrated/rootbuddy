import { useEffect, useState } from "react";
import { describeKnownCardID } from "../cardCatalog";
import { ACTION_TYPE, describeAction, phaseLabels, suitLabels } from "../labels";
import type { Action, Card, GameState } from "../types";

type PlayerActionsPanelProps = {
  state: GameState;
  actions: Action[];
  isMultiplayer: boolean;
  onApply: (action: Action) => Promise<void>;
  onGenerateActions: () => Promise<void>;
  onOpenBattle: (actionIndex: number) => void;
  onPreviewAction?: (actionIndex: number | null) => void;
};

function visibleCards(state: GameState): Card[] {
  switch (state.playerFaction) {
    case 0:
      return state.marquise.cardsInHand;
    case 1:
      return [...state.alliance.cardsInHand, ...state.alliance.supporters];
    case 2:
      return state.eyrie.cardsInHand;
    case 3:
      return state.vagabond.cardsInHand;
    default:
      return [];
  }
}

function actionHeadline(action: Action): string {
  switch (action.type) {
    case ACTION_TYPE.CRAFT:
      return "Craft";
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
    default:
      return "Action";
  }
}

function relatedCardIDs(action: Action): number[] {
  switch (action.type) {
    case ACTION_TYPE.CRAFT:
      return action.craft?.cardID ? [action.craft.cardID] : [];
    case ACTION_TYPE.OVERWORK:
      return action.overwork?.cardID ? [action.overwork.cardID] : [];
    case ACTION_TYPE.ADD_TO_DECREE:
      return action.addToDecree?.cardIDs ?? [];
    case ACTION_TYPE.SPREAD_SYMPATHY:
      return action.spreadSympathy?.supporterCardIDs ?? [];
    case ACTION_TYPE.REVOLT:
      return action.revolt?.supporterCardIDs ?? [];
    case ACTION_TYPE.MOBILIZE:
      return action.mobilize?.cardID ? [action.mobilize.cardID] : [];
    case ACTION_TYPE.TRAIN:
      return action.train?.cardID ? [action.train.cardID] : [];
    case ACTION_TYPE.ACTIVATE_DOMINANCE:
      return action.activateDominance?.cardID ? [action.activateDominance.cardID] : [];
    case ACTION_TYPE.TAKE_DOMINANCE:
      return [action.takeDominance?.dominanceCardID, action.takeDominance?.spentCardID].filter(
        (cardID): cardID is number => typeof cardID === "number" && cardID !== 0
      );
    default:
      return [];
  }
}

function actionContextTags(action: Action): string[] {
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
    default:
      return [];
  }
}

export function PlayerActionsPanel({
  state,
  actions,
  isMultiplayer,
  onApply,
  onGenerateActions,
  onOpenBattle,
  onPreviewAction
}: PlayerActionsPanelProps) {
  const [showAllActions, setShowAllActions] = useState(false);

  useEffect(() => {
    setShowAllActions(false);
  }, [state.factionTurn, state.currentPhase, state.currentStep, actions.length]);

  if (state.gamePhase !== 1 || state.factionTurn !== state.playerFaction) {
    return null;
  }

  const visibleActions = showAllActions ? actions : actions.slice(0, 6);
  const phaseLabel = phaseLabels[state.currentPhase] ?? "Unknown";
  const visibleCardMap = new Map(visibleCards(state).map((card) => [card.id, card]));

  return (
    <section className="panel sidebar-panel">
      <p className="eyebrow">Player Turn</p>
      <div className="summary-stack">
        <span className="summary-label">{phaseLabel}</span>
        <span className="summary-line">
          {visibleActions.length > 0
            ? `${visibleActions.length} loaded action(s) ready below.`
            : isMultiplayer
              ? "Actions refresh automatically when the server hands you priority."
              : "No loaded actions yet."}
        </span>
      </div>

      {visibleActions.length === 0 ? (
        <div className="summary-stack" style={{ marginTop: "0.9rem" }}>
          <button type="button" onClick={() => void onGenerateActions()}>
            Load Actions
          </button>
        </div>
      ) : (
        <div className="embedded-action-list" style={{ marginTop: "0.9rem" }}>
          {visibleActions.map((action, index) => (
            <article
              key={`${action.type}-${index}`}
              className="player-action-card"
              onMouseEnter={() => onPreviewAction?.(index)}
              onMouseLeave={() => onPreviewAction?.(null)}
              onFocus={() => onPreviewAction?.(index)}
              onBlur={() => onPreviewAction?.(null)}
            >
              <div className="player-action-card-header">
                <strong>{actionHeadline(action)}</strong>
                <span className="player-action-index">#{index + 1}</span>
              </div>
              <span className="summary-line">{describeAction(action)}</span>
              {actionContextTags(action).length > 0 ? (
                <div className="player-action-chip-row">
                  {actionContextTags(action).map((tag) => (
                    <span key={`${index}-${tag}`} className="player-action-chip">
                      {tag}
                    </span>
                  ))}
                </div>
              ) : null}
              {relatedCardIDs(action).length > 0 ? (
                <div className="player-action-card-links">
                  <span className="summary-label">Cards</span>
                  <div className="known-card-pill-list">
                    {relatedCardIDs(action).map((cardID) => {
                      const visibleCard = visibleCardMap.get(cardID);
                      const label = visibleCard ? `${visibleCard.name} (${suitLabels[visibleCard.suit] ?? "Unknown"})` : describeKnownCardID(cardID);
                      return (
                        <span key={`${index}-card-${cardID}`} className="known-card-pill">
                          {label}
                        </span>
                      );
                    })}
                  </div>
                </div>
              ) : null}
              {action.type === ACTION_TYPE.BATTLE ? (
                <button type="button" className="secondary" onClick={() => onOpenBattle(index)}>
                  Resolve
                </button>
              ) : (
                <button type="button" className="secondary" onClick={() => void onApply(action)}>
                  Apply
                </button>
              )}
            </article>
          ))}
        </div>
      )}

      <div className="sidebar-actions footer" style={{ marginTop: "0.9rem" }}>
        <button type="button" className="secondary" onClick={() => void onGenerateActions()}>
          Refresh
        </button>
        {actions.length > 6 ? (
          <button type="button" className="secondary" onClick={() => setShowAllActions((current) => !current)}>
            {showAllActions ? "Show Less" : `Show All (${actions.length})`}
          </button>
        ) : null}
      </div>
    </section>
  );
}
