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

function phasePrompt(state: GameState): string {
  if (state.factionTurn === state.playerFaction) {
    return "Your turn is active. Use generated actions for your own play, and keep the observed tools nearby for public bookkeeping.";
  }

  switch (state.currentPhase) {
    case 0:
      return "Birdsong: record start-of-turn effects, setup-like public choices, and any hidden draws or revealed cards.";
    case 1:
      return "Daylight: record public builds, battles, decree additions, crafts, and card spends as they happen.";
    case 2:
      return "Evening: record end-of-turn draws, discards, and then advance to the next faction when the turn is done.";
    default:
      return "Record the current faction's observed actions, then advance the turn when the table state matches the board.";
  }
}

function phaseChecklist(state: GameState): string[] {
  if (state.factionTurn === state.playerFaction) {
    switch (state.currentPhase) {
      case 0:
        return ["Check persistent effects and faction start-of-turn tasks.", "Generate actions when the board state is current."];
      case 1:
        return ["Generate legal actions for your turn.", "Resolve battles from the battle tool when needed."];
      case 2:
        return ["Handle end-of-turn draws or scoring.", "Advance once your evening effects are complete."];
      default:
        return ["Keep the board state in sync, then generate actions."];
    }
  }

  switch (state.currentPhase) {
    case 0:
      return ["Record start-of-turn public effects.", "Use shortcuts for decree adds or early public plays.", "Advance when birdsong is complete."];
    case 1:
      return ["Record public battles, builds, crafts, and spends as they happen.", "Use the preset shortcut buttons before filling the form manually."];
    case 2:
      return ["Record end-of-turn draws with one click.", "Advance to the next faction when evening is done."];
    default:
      return ["Record the observed action, then advance the turn flow."];
  }
}

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

  useEffect(() => {
    setPreferredTemplate(null);
  }, [state.factionTurn, state.currentPhase, state.currentStep]);

  if (state.gameMode !== 1 || state.gamePhase !== 1) {
    return null;
  }

  const shortcuts = suggestedShortcuts(state);
  const observedGeneratedActions =
    state.factionTurn === state.playerFaction ? [] : actions.slice(0, 6);

  return (
    <section className="panel sidebar-panel assist-workflow-panel">
      <p className="eyebrow">Assist Workflow</p>
      <div className="summary-stack">
        <span className="summary-label">
          {factionLabels[state.factionTurn] ?? "Unknown"}: {phaseLabels[state.currentPhase] ?? "Unknown"} / {stepLabels[state.currentStep] ?? "Unknown"}
        </span>
        <span className="summary-line">{phasePrompt(state)}</span>
        {phaseChecklist(state).map((item) => (
          <span key={item} className="summary-line">
            {item}
          </span>
        ))}
      </div>

      <div className="sidebar-actions" style={{ marginTop: "0.9rem" }}>
        <button type="button" className="secondary" onClick={() => void onGenerateActions()}>
          Generate
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
          <span className="summary-label">Recommended Shortcuts</span>
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

      {state.factionTurn !== state.playerFaction ? (
        <div className="summary-stack" style={{ marginTop: "0.9rem" }}>
          <span className="summary-label">Generated Public Actions</span>
          {observedGeneratedActions.length === 0 ? (
            <>
              <span className="summary-line">Load generated actions when the current public board state is up to date.</span>
              <button type="button" className="secondary" onClick={() => void onGenerateActions()}>
                Load Public Actions
              </button>
            </>
          ) : (
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
          )}
        </div>
      ) : null}

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
