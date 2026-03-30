import { ACTION_TYPE, describeAction } from "../labels";
import type { Action, GameState } from "../types";

type PlayerActionsPanelProps = {
  state: GameState;
  actions: Action[];
  onApply: (action: Action) => Promise<void>;
  onGenerateActions: () => Promise<void>;
  onOpenBattle: (actionIndex: number) => void;
  onOpenAllActions: () => void;
};

export function PlayerActionsPanel({
  state,
  actions,
  onApply,
  onGenerateActions,
  onOpenBattle,
  onOpenAllActions
}: PlayerActionsPanelProps) {
  if (state.gamePhase !== 1 || state.factionTurn !== state.playerFaction) {
    return null;
  }

  const visibleActions = actions.slice(0, 6);

  return (
    <section className="panel sidebar-panel">
      <p className="eyebrow">Player Actions</p>

      {visibleActions.length === 0 ? (
        <div className="summary-stack">
          <span className="summary-line">Load legal actions for the current board and turn state.</span>
          <button type="button" onClick={() => void onGenerateActions()}>
            Load Actions
          </button>
        </div>
      ) : (
        <div className="embedded-action-list">
          {visibleActions.map((action, index) => (
            <div key={`${action.type}-${index}`} className="embedded-action-card">
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
        <button type="button" className="secondary" onClick={onOpenAllActions}>
          Full List
        </button>
      </div>
    </section>
  );
}
