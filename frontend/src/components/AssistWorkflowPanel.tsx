import { ACTION_TYPE, describeAction, factionLabels, phaseLabels, stepLabels } from "../labels";
import type { Action, GameState } from "../types";
import { ObservedActionPanel, type ObservedTemplateKey } from "./ObservedActionPanel";
import { useEffect, useState } from "react";

type AssistWorkflowPanelProps = {
  state: GameState;
  actions: Action[];
  onApply: (action: Action) => Promise<void>;
  onGenerateActions: () => Promise<void>;
  onOpenTurnState: () => void;
  onOpenBattle: (actionIndex: number) => void;
};

type Shortcut = {
  label: string;
  action?: Action;
  template?: ObservedTemplateKey;
};

function shortcutAction(state: GameState, count: number): Action {
  return {
    type: 27,
    otherPlayerDraw: {
      faction: state.factionTurn,
      count
    }
  };
}

function suggestedShortcuts(state: GameState): Shortcut[] {
  if (state.factionTurn === state.playerFaction) {
    return [];
  }

  switch (state.currentPhase) {
    case 0:
      if (state.factionTurn === 2) {
        return [{ label: "Add To Decree", template: "add_to_decree" }, { label: "Advance", action: { type: 24, passPhase: { faction: state.factionTurn } } }];
      }
      return [{ label: "Advance", action: { type: 24, passPhase: { faction: state.factionTurn } } }];
    case 1: {
      const phaseShortcuts: Shortcut[] = [{ label: "Battle", template: "battle_resolution" }, { label: "Craft", template: "craft" }];
      switch (state.factionTurn) {
        case 0:
          phaseShortcuts.push({ label: "Overwork", template: "overwork" });
          break;
        case 1:
          phaseShortcuts.push({ label: "Spread Sympathy", template: "spread_sympathy" });
          phaseShortcuts.push({ label: "Revolt", template: "revolt" });
          break;
        case 3:
          phaseShortcuts.push({ label: "Aid", template: "aid" });
          break;
      }
      phaseShortcuts.push({ label: "Advance", action: { type: 24, passPhase: { faction: state.factionTurn } } });
      return phaseShortcuts;
    }
    case 2:
      return [
        { label: "Draw 1", action: shortcutAction(state, 1) },
        { label: "Draw 2", action: shortcutAction(state, 2) },
        { label: "Advance", action: { type: 24, passPhase: { faction: state.factionTurn } } }
      ];
    default:
      return [{ label: "Advance", action: { type: 24, passPhase: { faction: state.factionTurn } } }];
  }
}

export function AssistWorkflowPanel({ state, actions, onApply, onGenerateActions, onOpenTurnState, onOpenBattle }: AssistWorkflowPanelProps) {
  const [preferredTemplate, setPreferredTemplate] = useState<ObservedTemplateKey | null>(null);
  const [showAllGeneratedActions, setShowAllGeneratedActions] = useState(false);

  useEffect(() => {
    setPreferredTemplate(null);
    setShowAllGeneratedActions(false);
  }, [state.factionTurn, state.currentPhase, state.currentStep]);

  if (state.gameMode !== 1 || state.gamePhase !== 1 || state.factionTurn === state.playerFaction) {
    return null;
  }

  const shortcuts = suggestedShortcuts(state);
  const observedGeneratedActions = showAllGeneratedActions ? actions : actions.slice(0, 6);

  return (
    <section className="panel sidebar-panel assist-workflow-panel">
      <p className="eyebrow">Observed Turn</p>
      <div className="summary-stack">
        <span className="summary-label">
          {factionLabels[state.factionTurn] ?? "Unknown"}: {phaseLabels[state.currentPhase] ?? "Unknown"} / {stepLabels[state.currentStep] ?? "Unknown"}
        </span>
        <span className="summary-line">Use this panel to record public actions and hidden bookkeeping for the current non-player turn.</span>
      </div>

      <div className="sidebar-actions" style={{ marginTop: "0.9rem" }}>
        <button type="button" className="secondary" onClick={() => void onGenerateActions()}>
          Load Public Actions
        </button>
        <button type="button" className="secondary" onClick={onOpenTurnState}>
          Advanced Turn
        </button>
        <button
          type="button"
          className="secondary"
          onClick={() =>
            void onApply({
              type: 24,
              passPhase: {
                faction: state.factionTurn
              }
            })
          }
        >
          Advance
        </button>
      </div>

      {shortcuts.length > 0 ? (
        <div className="summary-stack" style={{ marginTop: "0.9rem" }}>
          <span className="summary-label">Shortcuts</span>
          <div className="shortcut-grid">
            {shortcuts.map((shortcut) => (
              <button
                key={shortcut.label}
                type="button"
                className="secondary"
                onClick={() => {
                  if (shortcut.action) {
                    void onApply(shortcut.action);
                    return;
                  }
                  if (shortcut.template) {
                    setPreferredTemplate(shortcut.template);
                  }
                }}
              >
                {shortcut.label}
              </button>
            ))}
          </div>
        </div>
      ) : null}

      <div className="summary-stack" style={{ marginTop: "0.9rem" }}>
        <span className="summary-label">Generated Public Actions</span>
        {actions.length === 0 ? (
          <span className="summary-line">No public actions loaded yet.</span>
        ) : (
          <>
            <div className="embedded-action-list">
              {observedGeneratedActions.map((action, index) => (
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
            {actions.length > 6 ? (
              <div className="embedded-action-card">
                <button type="button" className="secondary" onClick={() => setShowAllGeneratedActions((current) => !current)}>
                  {showAllGeneratedActions ? "Show Fewer Public Actions" : `Show All Public Actions (${actions.length})`}
                </button>
              </div>
            ) : null}
          </>
        )}
      </div>

      <div style={{ marginTop: "1rem" }}>
        <ObservedActionPanel
          state={state}
          onApply={onApply}
          embedded
          preferredActorFaction={state.factionTurn}
          preferredTemplate={preferredTemplate}
        />
      </div>
    </section>
  );
}
