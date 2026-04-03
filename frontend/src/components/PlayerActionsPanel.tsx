import { useEffect, useState } from "react";
import { ACTION_TYPE, describeAction, phaseLabels } from "../labels";
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
            <div
              key={`${action.type}-${index}`}
              className="embedded-action-card"
              onMouseEnter={() => onPreviewAction?.(index)}
              onMouseLeave={() => onPreviewAction?.(null)}
              onFocus={() => onPreviewAction?.(index)}
              onBlur={() => onPreviewAction?.(null)}
            >
              <span className="summary-line">{describeAction(action)}</span>
              {action.type === ACTION_TYPE.BATTLE ? (
                <button type="button" className="secondary" onClick={() => onOpenBattle(index)}>
                  Resolve
                </button>
              ) : (
                <button type="button" className="secondary" onClick={() => void onApply(action)}>
                  Apply
                </button>
              )}
            </div>
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
