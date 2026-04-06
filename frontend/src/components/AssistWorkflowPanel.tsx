import { ACTION_TYPE, factionLabels, phaseLabels, stepLabels } from "../labels";
import { actionCardReferences, actionContextTags, actionHeadline, createVisibleCardLookup, describeAction, formatCardReferenceLabel } from "../actionPresentation";
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
  const visibleCardLookup = createVisibleCardLookup(state);
  const actorLabel = factionLabels[state.factionTurn] ?? "Unknown";
  const phaseLabel = phaseLabels[state.currentPhase] ?? "Unknown";
  const nextAssistSummary =
    actions.length > 0
      ? `Review ${actions.length} generated public action(s) first, then record the hidden or table-only events that are still missing.`
      : "Load generated public actions first if the board is current, then use the observed form for anything hidden or not captured automatically.";

  return (
    <section className="panel sidebar-panel assist-workflow-panel">
      <p className="eyebrow">Observed Turn</p>
      <div className="flow-guide-hero">
        <span className="summary-label">
          {actorLabel}: {phaseLabel} / {stepLabels[state.currentStep] ?? "Unknown"}
        </span>
        <strong>Assist Workflow</strong>
        <span className="summary-line">{nextAssistSummary}</span>
      </div>

      <div className="flow-step-list" style={{ marginTop: "0.9rem" }}>
        <div className={`flow-step-card ${actions.length > 0 ? "done" : "active"}`}>
          <strong>1. Generate the public action candidates</strong>
          <span className="summary-line">
            {actions.length > 0
              ? `${actions.length} public action(s) are loaded for this observed turn.`
              : "Pull the public action set from the current board state before filling anything manually."}
          </span>
          <div className="sidebar-actions" style={{ marginTop: "0.5rem" }}>
            <button type="button" className="secondary" onClick={() => void onGenerateActions()}>
              {actions.length > 0 ? "Refresh Public Actions" : "Load Public Actions"}
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
        </div>

        <div className={`flow-step-card ${actions.length > 0 ? "active" : "note"}`}>
          <strong>2. Use shortcuts and generated actions</strong>
          <span className="summary-line">
            Apply the obvious public actions from shortcuts or the generated list before dropping into manual observation.
          </span>
          {shortcuts.length > 0 ? (
            <div className="shortcut-grid" style={{ marginTop: "0.5rem" }}>
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
          ) : null}
        </div>

        <div className="flow-step-card waiting">
          <strong>3. Record what the generated actions cannot know</strong>
          <span className="summary-line">
            Use the observed form for hidden draws, supporter spending, decree choices, or any public result that still needs manual capture.
          </span>
          <div className="sidebar-actions" style={{ marginTop: "0.5rem" }}>
            <button type="button" className="secondary" onClick={onOpenTurnState}>
              Advanced Turn
            </button>
          </div>
        </div>
      </div>

      <div className="summary-stack" style={{ marginTop: "0.9rem" }}>
        <span className="summary-label">Generated Public Actions</span>
        {actions.length === 0 ? (
          <span className="summary-line">No public actions loaded yet.</span>
        ) : (
          <>
            <div className="embedded-action-list">
              {observedGeneratedActions.map((action, index) => (
                <div key={`${action.type}-${index}`} className="embedded-action-card">
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
