import { useEffect, useState } from "react";
import { phaseLabels } from "../labels";
import {
  battleTargetKey,
  battleTargetLabel,
  cardEffectChoiceLabel,
  craftCardID,
  craftRouteLabel,
  drawAdvanceChoiceLabel,
  factionChoiceDetail,
  factionChoiceLabel,
  groupActionsByIntent,
  type AssistActionCandidateRef,
  type AssistIntentKey
} from "../assistDirector";
import { describeAction } from "../actionPresentation";
import { describeKnownCardID } from "../cardCatalog";
import type { Action, GameState } from "../types";
import { ActionOptionCard, ExactActionDrawer, IntentGrid } from "./ActionPromptUi";

type PlayerActionsPanelProps = {
  state: GameState;
  actions: Action[];
  isMultiplayer: boolean;
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
  onApply,
  onGenerateActions,
  onOpenBattle,
  onPreviewAction,
  onMovementCandidatesChange,
  onBuildRecruitCandidatesChange,
  onFactionSpatialCandidatesChange
}: PlayerActionsPanelProps) {
  const [showAllActions, setShowAllActions] = useState(false);
  const [selectedIntent, setSelectedIntent] = useState<AssistIntentKey | null>(null);
  const [choiceMessage, setChoiceMessage] = useState("");
  const [selectedCraftCardID, setSelectedCraftCardID] = useState<number | null>(null);

  useEffect(() => {
    setShowAllActions(false);
    setSelectedIntent(null);
    setChoiceMessage("");
    setSelectedCraftCardID(null);
  }, [state.factionTurn, state.currentPhase, state.currentStep, actions.length]);

  if (state.gamePhase !== 1 || state.factionTurn !== state.playerFaction) {
    return null;
  }

  const actionGroups = groupActionsByIntent(actions);
  const selectedGroup = actionGroups.find((group) => group.key === selectedIntent) ?? null;
  const candidateRefsForSelectedGroup = (enabled: boolean): AssistActionCandidateRef[] =>
    enabled && selectedGroup
      ? selectedGroup.actions
          .map((action) => ({ actionIndex: actions.indexOf(action), action }))
          .filter((candidate) => candidate.actionIndex >= 0)
      : [];
  const movementCandidates =
    selectedIntent === "movement" && selectedGroup
      ? candidateRefsForSelectedGroup(true)
      : [];
  const buildRecruitCandidates =
    selectedIntent === "build_recruit" && selectedGroup
      ? candidateRefsForSelectedGroup(true)
      : [];
  const factionSpatialCandidates =
    selectedIntent === "faction" && selectedGroup
      ? candidateRefsForSelectedGroup(true)
      : [];
  const candidateKey = (candidates: AssistActionCandidateRef[]) =>
    candidates.map((candidate) => `${candidate.actionIndex}:${candidate.action.type}`).join(",");
  const movementCandidateKey = candidateKey(movementCandidates);
  const buildRecruitCandidateKey = candidateKey(buildRecruitCandidates);
  const factionSpatialCandidateKey = candidateKey(factionSpatialCandidates);
  const visibleActions = showAllActions ? actions : actions.slice(0, 6);
  const phaseLabel = phaseLabels[state.currentPhase] ?? "Unknown";
  const battleTargetChoices =
    selectedGroup?.key === "battle"
      ? Array.from(new Set(selectedGroup.actions.map(battleTargetKey).filter((key) => key.length > 0))).map((key) => {
          const matchingActions = selectedGroup.actions.filter((action) => battleTargetKey(action) === key);
          return {
            key,
            label: battleTargetLabel(matchingActions[0]),
            actions: matchingActions
          };
        })
      : [];
  const craftChoices =
    selectedGroup?.key === "craft"
      ? Array.from(new Set(selectedGroup.actions.map(craftCardID).filter((cardID) => cardID > 0))).map((cardID) => ({
          cardID,
          actions: selectedGroup.actions.filter((action) => craftCardID(action) === cardID)
        }))
      : [];
  const selectedCraftChoice = craftChoices.find((choice) => choice.cardID === selectedCraftCardID) ?? null;
  const factionChoiceActions =
    selectedGroup?.key === "faction"
      ? selectedGroup.actions
          .map((action) => ({ action, label: factionChoiceLabel(action, state), detail: factionChoiceDetail(action, state) }))
          .filter((choice): choice is { action: Action; label: string; detail: string } => choice.label !== null)
      : [];
  const drawAdvanceChoices = selectedGroup?.key === "draw_advance" ? selectedGroup.actions : [];
  const cardEffectChoices = selectedGroup?.key === "card_effect" ? selectedGroup.actions : [];

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

  return (
    <section className="panel sidebar-panel">
      <p className="eyebrow">Player Turn</p>
      <div className="summary-stack">
        <span className="summary-label">{phaseLabel}</span>
        <span className="summary-line">
          {visibleActions.length > 0
            ? `Choose your intent, then choose the matching legal action.`
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
              <span className="summary-label">{selectedGroup.label} Options</span>
              <span className="summary-line">
                Choose the legal action that matches what you want to do. Hover or focus an option to preview its board footprint.
              </span>
              {selectedGroup.key === "battle" ? (
                <>
                  <span className="summary-line">Choose the battle target first. This opens Battle Flow directly when the defender choice maps to one legal battle.</span>
                  <div className="assist-choice-grid">
                    {battleTargetChoices.map((choice) => (
                      <button
                        key={choice.key}
                        type="button"
                        className="assist-choice-card player-action-choice-card"
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
                          setChoiceMessage(`${choice.label} still maps to ${choice.actions.length} legal battle options. Choose the exact battle from the fallback list.`);
                        }}
                      >
                        <strong>{choice.label}</strong>
                        <span>{choice.actions.length === 1 ? "Open Battle Flow" : `${choice.actions.length} battle options`}</span>
                      </button>
                    ))}
                  </div>
                </>
              ) : selectedGroup.key === "craft" ? (
                <>
                  <span className="summary-line">Choose the card to craft first. If multiple workshop routes are legal, choose the route next.</span>
                  <div className="assist-choice-grid">
                    {craftChoices.map((choice) => (
                      <button
                        key={choice.cardID}
                        type="button"
                        className={`assist-choice-card player-action-choice-card ${selectedCraftCardID === choice.cardID ? "selected" : ""}`}
                        onClick={() => {
                          if (choice.actions.length === 1) {
                            setChoiceMessage("");
                            setSelectedCraftCardID(null);
                            void onApply(choice.actions[0]);
                            return;
                          }
                          setSelectedCraftCardID(choice.cardID);
                          setChoiceMessage(`${describeKnownCardID(choice.cardID)} has ${choice.actions.length} legal craft routes. Choose the workshop route.`);
                        }}
                      >
                        <strong>{describeKnownCardID(choice.cardID)}</strong>
                        <span>{choice.actions.length === 1 ? "Apply exact craft" : `${choice.actions.length} craft routes`}</span>
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
                  <span className="summary-line">Choose the faction-specific option directly. Board-based faction actions still use the generic choices for now.</span>
                  <div className="assist-choice-grid">
                    {factionChoiceActions.map((choice, actionIndex) => (
                      <button
                        key={`${choice.action.type}-${choice.label}-${actionIndex}`}
                        type="button"
                        className="assist-choice-card player-action-choice-card"
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
                  <span className="summary-line">Choose the turn bookkeeping directly. These actions do not need the exact legal-action browser first.</span>
                  <div className="assist-choice-grid">
                    {drawAdvanceChoices.map((action, actionIndex) => (
                      <button
                        key={`${action.type}-${actionIndex}`}
                        type="button"
                        className="assist-choice-card player-action-choice-card"
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
              <span className="summary-line">The matching legal options will appear here.</span>
            </div>
          )}

          <ExactActionDrawer
            title="Exact Legal Actions"
            summary="Use this only when you need to audit or apply from the raw generated action list."
            open={showAllActions}
            actions={visibleActions}
            allActions={actions}
            state={state}
            onToggle={setShowAllActions}
            onApply={onApply}
            onOpenBattle={onOpenBattle}
            onPreviewAction={onPreviewAction}
          />
        </div>
      )}

      <div className="sidebar-actions footer" style={{ marginTop: "0.9rem" }}>
        <button type="button" className="secondary" onClick={() => void onGenerateActions()}>
          Refresh
        </button>
      </div>
    </section>
  );
}
