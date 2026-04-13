import { useEffect, useState } from "react";
import { phaseLabels } from "../labels";
import {
  cardEffectChoiceLabel,
  craftRouteLabel,
  drawAdvanceChoiceLabel,
  type AssistActionCandidateRef
} from "../assistDirector";
import { actionExplanation } from "../assist/explanations";
import { describeAction } from "../actionPresentation";
import { describeKnownCardID } from "../cardCatalog";
import { useIntentSelection } from "../hooks/useIntentSelection";
import type { Action, GameState } from "../types";
import { ActionOptionCard, ExactActionDrawer, IntentGrid } from "./ActionPromptUi";

type PlayerActionsPanelProps = {
  state: GameState;
  actions: Action[];
  isMultiplayer: boolean;
  surface?: "sidebar" | "tray";
  showFallbackDrawer?: boolean;
  showRefreshButton?: boolean;
  onPromptChange?: (prompt: string | null) => void;
  onApply: (action: Action) => Promise<void>;
  onGenerateActions: () => Promise<void>;
  onOpenBattle: (actionIndex: number) => void;
  onPreviewAction?: (actionIndex: number | null) => void;
  onMovementCandidatesChange?: (candidates: AssistActionCandidateRef[]) => void;
  onBuildRecruitCandidatesChange?: (candidates: AssistActionCandidateRef[]) => void;
  onFactionSpatialCandidatesChange?: (candidates: AssistActionCandidateRef[]) => void;
};

export function PlayerActionsPanel({
  state,
  actions,
  isMultiplayer,
  surface = "sidebar",
  showFallbackDrawer = true,
  showRefreshButton = true,
  onPromptChange,
  onApply,
  onGenerateActions,
  onOpenBattle,
  onPreviewAction,
  onMovementCandidatesChange,
  onBuildRecruitCandidatesChange,
  onFactionSpatialCandidatesChange
}: PlayerActionsPanelProps) {
  const [showAllActions, setShowAllActions] = useState(false);
  const [choiceMessage, setChoiceMessage] = useState("");

  useEffect(() => {
    setShowAllActions(false);
    setChoiceMessage("");
  }, [state.factionTurn, state.currentPhase, state.currentStep, actions.length]);

  if (state.gamePhase !== 1 || state.factionTurn !== state.playerFaction) {
    return null;
  }

  const {
    actionGroups,
    selectedIntent,
    setSelectedIntent,
    selectedGroup,
    selectedCraftCardID,
    setSelectedCraftCardID,
    movementCandidates,
    buildRecruitCandidates,
    factionSpatialCandidates,
    battleTargetChoices,
    craftChoices,
    selectedCraftChoice,
    factionChoiceActions,
    drawAdvanceChoices,
    cardEffectChoices
  } = useIntentSelection({
    actions,
    state,
    resetKey: `${state.factionTurn}:${state.currentPhase}:${state.currentStep}:${actions.length}`
  });
  const candidateKey = (candidates: AssistActionCandidateRef[]) =>
    candidates.map((candidate) => `${candidate.actionIndex}:${candidate.action.type}`).join(",");
  const movementCandidateKey = candidateKey(movementCandidates);
  const buildRecruitCandidateKey = candidateKey(buildRecruitCandidates);
  const factionSpatialCandidateKey = candidateKey(factionSpatialCandidates);
  const visibleActions = showAllActions ? actions : actions.slice(0, 6);
  const phaseLabel = phaseLabels[state.currentPhase] ?? "Unknown";
  const currentPrompt =
    visibleActions.length === 0
      ? "Preparing your turn options."
      : selectedGroup?.key === "movement"
        ? "Choose a highlighted source clearing, then choose the destination."
        : selectedGroup?.key === "battle"
          ? "Choose the clearing and defender you want to battle."
          : selectedGroup?.key === "build_recruit"
            ? "Choose the clearing where this step happens."
            : selectedGroup?.key === "craft"
              ? selectedCraftChoice
                ? "Choose which workshop route to use."
                : "Choose the card you want to craft."
              : selectedGroup?.key === "faction"
                ? "Choose the faction action or target that matches what you want to do."
                : selectedGroup?.key === "draw_advance"
                  ? "Choose how to advance or finish this part of the turn."
                  : selectedGroup?.key === "card_effect"
                    ? "Choose the card or persistent effect you want to use."
                    : selectedGroup
                      ? "Choose the exact move you want to make."
                      : "Choose what you want to do next.";

  useEffect(() => {
    onMovementCandidatesChange?.(movementCandidates);
    return () => onMovementCandidatesChange?.([]);
  }, [movementCandidateKey, onMovementCandidatesChange]);

  useEffect(() => {
    onBuildRecruitCandidatesChange?.(buildRecruitCandidates);
    return () => onBuildRecruitCandidatesChange?.([]);
  }, [buildRecruitCandidateKey, onBuildRecruitCandidatesChange]);

  useEffect(() => {
    onFactionSpatialCandidatesChange?.(factionSpatialCandidates);
    return () => onFactionSpatialCandidatesChange?.([]);
  }, [factionSpatialCandidateKey, onFactionSpatialCandidatesChange]);

  useEffect(() => {
    onPromptChange?.(currentPrompt);
    return () => onPromptChange?.(null);
  }, [currentPrompt, onPromptChange]);

  return (
    <section className={`panel ${surface === "tray" ? "board-action-panel board-action-panel-tray player-actions-tray" : "sidebar-panel"}`}>
      <p className="eyebrow">Your Turn</p>
      <div className="summary-stack">
        <span className="summary-label">{phaseLabel}</span>
        <span className="summary-line">
          {visibleActions.length > 0
            ? `Choose your intent, then follow the board or prompt to carry it out.`
            : isMultiplayer
              ? "Preparing board-ready options for your turn."
              : "Preparing board-ready options for this turn."}
        </span>
      </div>

      {visibleActions.length === 0 ? (
        <div className="summary-stack" style={{ marginTop: "0.9rem" }}>
          <span className="summary-line">The tray will populate as soon as the current turn options are ready.</span>
          {showRefreshButton ? (
            <button type="button" className="secondary" onClick={() => void onGenerateActions()}>
              Refresh Turn
            </button>
          ) : null}
        </div>
      ) : (
        <div className="summary-stack" style={{ marginTop: "0.9rem" }}>
          <IntentGrid
            groups={actionGroups}
            selectedIntent={selectedIntent}
            countLabel="option(s)"
              onSelect={(intent) => {
                setSelectedIntent(intent);
                setShowAllActions(false);
                setChoiceMessage("");
                setSelectedCraftCardID(null);
            }}
          />

          {selectedGroup ? (
            <div className="summary-stack player-action-choice-stack">
              <span className="summary-label">{selectedGroup.label}</span>
              <span className="summary-line">
                Choose the target, card, or route that matches the move you want to make. Hover or focus an entry to preview it on the board.
              </span>
              {selectedGroup.key === "battle" ? (
                <>
                  <span className="summary-line">Choose the battle target first. If only one exact fight matches, battle opens immediately.</span>
                  <div className="assist-choice-grid">
                    {battleTargetChoices.map((choice) => (
                      <button
                        key={choice.key}
                        type="button"
                        className="assist-choice-card player-action-choice-card"
                        title={
                          choice.actions.length === 1
                            ? actionExplanation(choice.actions[0], state)
                            : "Multiple legal battle records share this defender. Use the audit drawer if you need the exact rule-backed fight."
                        }
                        onClick={() => {
                          if (choice.actions.length === 1) {
                            const actionIndex = actions.indexOf(choice.actions[0]);
                            if (actionIndex >= 0) {
                              setChoiceMessage("");
                              onOpenBattle(actionIndex);
                          return;
                        }
                      }
                      setShowAllActions(true);
                          setChoiceMessage(`${choice.label} still has ${choice.actions.length} possible battle paths. Open the audit drawer and choose the exact one.`);
                        }}
                      >
                        <strong>{choice.label}</strong>
                        <span>{choice.actions.length === 1 ? "Open battle" : `${choice.actions.length} battle paths`}</span>
                      </button>
                    ))}
                  </div>
                </>
              ) : selectedGroup.key === "craft" ? (
                <>
                  <span className="summary-line">Choose the card to craft first. If more than one workshop route can make it, choose the route next.</span>
                  <div className="assist-choice-grid">
                    {craftChoices.map((choice) => (
                      <button
                        key={choice.cardID}
                        type="button"
                        className={`assist-choice-card player-action-choice-card ${selectedCraftCardID === choice.cardID ? "selected" : ""}`}
                        title={
                          choice.actions.length === 1
                            ? actionExplanation(choice.actions[0], state)
                            : "Multiple workshop routes can craft this card. Choose the route that matches the board before applying it."
                        }
                        onClick={() => {
                          if (choice.actions.length === 1) {
                            setChoiceMessage("");
                            setSelectedCraftCardID(null);
                            void onApply(choice.actions[0]);
                            return;
                          }
                          setSelectedCraftCardID(choice.cardID);
                          setChoiceMessage(`${describeKnownCardID(choice.cardID)} has ${choice.actions.length} possible workshop routes. Choose the one you want to use.`);
                        }}
                      >
                        <strong>{describeKnownCardID(choice.cardID)}</strong>
                        <span>{choice.actions.length === 1 ? "Craft this card" : `${choice.actions.length} workshop routes`}</span>
                      </button>
                    ))}
                  </div>
                  {selectedCraftChoice ? (
                    <div className="assist-choice-grid">
                      {selectedCraftChoice.actions.map((action, actionIndex) => (
                        <button
                          key={`${selectedCraftChoice.cardID}-${craftRouteLabel(action)}-${actionIndex}`}
                          type="button"
                          className="assist-choice-card player-action-choice-card"
                          title={actionExplanation(action, state)}
                          onClick={() => {
                            setChoiceMessage("");
                            setSelectedCraftCardID(null);
                            void onApply(action);
                          }}
                        >
                          <strong>{craftRouteLabel(action)}</strong>
                          <span>Apply this craft route</span>
                        </button>
                      ))}
                    </div>
                  ) : null}
                </>
              ) : selectedGroup.key === "faction" && factionChoiceActions.length > 0 ? (
                <>
                  <span className="summary-line">Choose the faction-specific move directly. Board-based faction actions still use the board first.</span>
                  <div className="assist-choice-grid">
                    {factionChoiceActions.map((choice, actionIndex) => (
                      <button
                        key={`${choice.action.type}-${choice.label}-${actionIndex}`}
                        type="button"
                        className="assist-choice-card player-action-choice-card"
                        title={actionExplanation(choice.action, state)}
                        onClick={() => {
                          setChoiceMessage("");
                          void onApply(choice.action);
                        }}
                      >
                        <strong>{choice.label}</strong>
                        <span>{choice.detail}</span>
                      </button>
                    ))}
                  </div>
                </>
              ) : selectedGroup.key === "draw_advance" ? (
                <>
                  <span className="summary-line">Choose the turn bookkeeping directly. These steps do not need the audit drawer first.</span>
                  <div className="assist-choice-grid">
                    {drawAdvanceChoices.map((action, actionIndex) => (
                      <button
                        key={`${action.type}-${actionIndex}`}
                        type="button"
                        className="assist-choice-card player-action-choice-card"
                        title={actionExplanation(action, state)}
                        onClick={() => {
                          setChoiceMessage("");
                          void onApply(action);
                        }}
                      >
                        <strong>{drawAdvanceChoiceLabel(action)}</strong>
                        <span>{describeAction(action, state)}</span>
                      </button>
                    ))}
                  </div>
                </>
              ) : selectedGroup.key === "card_effect" ? (
                <>
                  <span className="summary-line">Choose the card or persistent effect directly.</span>
                  <div className="assist-choice-grid">
                    {cardEffectChoices.map((action, actionIndex) => (
                      <button
                        key={`${action.type}-${actionIndex}`}
                        type="button"
                        className="assist-choice-card player-action-choice-card"
                        title={actionExplanation(action, state)}
                        onClick={() => {
                          setChoiceMessage("");
                          void onApply(action);
                        }}
                      >
                        <strong>{cardEffectChoiceLabel(action)}</strong>
                        <span>{describeAction(action, state)}</span>
                      </button>
                    ))}
                  </div>
                </>
              ) : (
                <div className="assist-choice-grid">
                  {selectedGroup.actions.map((action) => {
                    const actionIndex = actions.indexOf(action);
                    return (
                      <ActionOptionCard
                        key={`choice-${action.type}-${actionIndex}`}
                        state={state}
                        action={action}
                        actionIndex={actionIndex}
                        variant="choice"
                        onApply={onApply}
                        onOpenBattle={onOpenBattle}
                        onPreviewAction={onPreviewAction}
                        choiceClassName="assist-choice-card player-action-choice-card"
                      />
                    );
                  })}
                </div>
              )}
              {choiceMessage ? <span className="message">{choiceMessage}</span> : null}
            </div>
          ) : (
            <div className="flow-step-card waiting assist-selection-empty">
              <strong>Select an intent above.</strong>
              <span className="summary-line">The matching turn prompt will appear here.</span>
            </div>
          )}

          {showFallbackDrawer ? (
            <ExactActionDrawer
              title="Action Audit"
              summary="Open this only when you need the exact rule-backed move list."
              open={showAllActions}
              actions={visibleActions}
              allActions={actions}
              state={state}
              onToggle={setShowAllActions}
              onApply={onApply}
              onOpenBattle={onOpenBattle}
              onPreviewAction={onPreviewAction}
            />
          ) : null}
        </div>
      )}

      {showRefreshButton ? (
        <div className="sidebar-actions footer" style={{ marginTop: "0.9rem" }}>
          <button type="button" className="secondary" onClick={() => void onGenerateActions()}>
            Refresh Turn
          </button>
        </div>
      ) : null}
    </section>
  );
}
