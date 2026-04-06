import { useEffect, useState } from "react";
import { ACTION_TYPE, phaseLabels } from "../labels";
import { actionCardReferences, actionContextTags, actionHeadline, createVisibleCardLookup, describeAction, formatCardReferenceLabel } from "../actionPresentation";
import type { Action, GameState } from "../types";

type PlayerActionsPanelProps = {
  state: GameState;
  actions: Action[];
  isMultiplayer: boolean;
  onApply: (action: Action) => Promise<void>;
  onGenerateActions: () => Promise<void>;
  onOpenBattle: (actionIndex: number) => void;
  onPreviewAction?: (actionIndex: number | null) => void;
};

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
  const visibleCardLookup = createVisibleCardLookup(state);

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
              <span className="summary-line">{describeAction(action, state)}</span>
              {actionContextTags(action).length > 0 ? (
                <div className="player-action-chip-row">
                  {actionContextTags(action).map((tag, tagIndex) => (
                    <span key={`${index}-${tag}-${tagIndex}`} className="player-action-chip">
                      {tag}
                    </span>
                  ))}
                </div>
              ) : null}
              {actionCardReferences(action).length > 0 ? (
                <div className="player-action-card-links">
                  <span className="summary-label">Cards</span>
                  <div className="known-card-pill-list">
                    {actionCardReferences(action).map((reference, referenceIndex) => (
                      <span key={`${index}-card-${reference.cardID}-${reference.zoneLabel}-${referenceIndex}`} className="known-card-pill">
                        {formatCardReferenceLabel(reference, visibleCardLookup)}
                      </span>
                    ))}
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
