import {
  battleTargetKey,
  battleTargetLabel,
  cardEffectChoiceLabel,
  craftCardID,
  craftRouteLabel,
  decreeCardKey,
  decreeCardLabel,
  decreeColumnAssignmentLabel,
  drawAdvanceChoiceLabel,
  factionChoiceDetail,
  factionChoiceLabel,
  factionSpatialChoiceDetail,
  factionSpatialChoiceLabel,
  groupActionsByIntent,
  observedPromptTemplates,
  type AssistIntentKey
} from "../assistDirector";
import { ACTION_TYPE, factionLabels, phaseLabels, stepLabels } from "../labels";
import { describeAction } from "../actionPresentation";
import { describeKnownCardID } from "../cardCatalog";
import type { Action, GameState } from "../types";
import { ExactActionDrawer, IntentGrid } from "./ActionPromptUi";
import { ObservedActionPanel, type ObservedTemplateKey } from "./ObservedActionPanel";
import { useEffect, useRef, useState } from "react";

type AssistWorkflowPanelProps = {
  state: GameState;
  actions: Action[];
  onApply: (action: Action) => Promise<void>;
  onGenerateActions: () => Promise<void>;
  onOpenTurnState: () => void;
  onOpenBattle: (actionIndex: number) => void;
  onBattleCandidatesChange?: (actionIndices: number[]) => void;
  onMovementCandidatesChange?: (actionIndices: number[]) => void;
  onBuildRecruitCandidatesChange?: (actionIndices: number[]) => void;
  onFactionSpatialCandidatesChange?: (actionIndices: number[]) => void;
};

export function AssistWorkflowPanel({
  state,
  actions,
  onApply,
  onGenerateActions,
  onOpenTurnState,
  onOpenBattle,
  onBattleCandidatesChange,
  onMovementCandidatesChange,
  onBuildRecruitCandidatesChange,
  onFactionSpatialCandidatesChange
}: AssistWorkflowPanelProps) {
  const [preferredTemplate, setPreferredTemplate] = useState<ObservedTemplateKey | null>(null);
  const [showAllGeneratedActions, setShowAllGeneratedActions] = useState(false);
  const [selectedIntent, setSelectedIntent] = useState<AssistIntentKey | null>(null);
  const [manualCaptureOpen, setManualCaptureOpen] = useState(false);
  const [choiceMessage, setChoiceMessage] = useState("");
  const [selectedCraftCardID, setSelectedCraftCardID] = useState<number | null>(null);
  const [selectedDecreeCardKey, setSelectedDecreeCardKey] = useState<string | null>(null);
  const autoLoadKey = useRef("");

  useEffect(() => {
    setPreferredTemplate(null);
    setShowAllGeneratedActions(false);
    setSelectedIntent(null);
    setManualCaptureOpen(false);
    setChoiceMessage("");
    setSelectedCraftCardID(null);
    setSelectedDecreeCardKey(null);
  }, [state.factionTurn, state.currentPhase, state.currentStep]);

  useEffect(() => {
    const nextKey = `${state.factionTurn}:${state.currentPhase}:${state.currentStep}:${state.roundNumber}`;
    if (state.gameMode !== 1 || state.gamePhase !== 1 || state.factionTurn === state.playerFaction) {
      return;
    }
    if (actions.length > 0 || autoLoadKey.current === nextKey) {
      return;
    }

    autoLoadKey.current = nextKey;
    void onGenerateActions();
  }, [actions.length, onGenerateActions, state.currentPhase, state.currentStep, state.factionTurn, state.gameMode, state.gamePhase, state.playerFaction, state.roundNumber]);

  const actionGroups = groupActionsByIntent(actions);
  const selectedGroup = actionGroups.find((group) => group.key === selectedIntent) ?? null;
  const exactCandidateDrawerOpen = Boolean(showAllGeneratedActions || selectedGroup?.key === "other");
  const observedGeneratedActions = selectedGroup ? (exactCandidateDrawerOpen ? selectedGroup.actions : selectedGroup.actions.slice(0, 6)) : [];
  const craftChoices =
    selectedGroup?.key === "craft"
      ? Array.from(new Set(selectedGroup.actions.map(craftCardID).filter((cardID) => cardID > 0))).map((cardID) => ({
          cardID,
          actions: selectedGroup.actions.filter((action) => craftCardID(action) === cardID)
        }))
      : [];
  const selectedCraftChoice = craftChoices.find((choice) => choice.cardID === selectedCraftCardID) ?? null;
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
  const decreeActions =
    selectedGroup?.key === "faction"
      ? selectedGroup.actions.filter((action) => action.type === ACTION_TYPE.ADD_TO_DECREE && (action.addToDecree?.cardIDs.length ?? 0) > 0)
      : [];
  const decreeCardChoices = Array.from(new Set(decreeActions.map(decreeCardKey).filter((key) => key.length > 0))).map((key) => {
    const matchingActions = decreeActions.filter((action) => decreeCardKey(action) === key);
    return {
      key,
      cardIDs: matchingActions[0]?.addToDecree?.cardIDs ?? [],
      actions: matchingActions
    };
  });
  const selectedDecreeCardChoice = decreeCardChoices.find((choice) => choice.key === selectedDecreeCardKey) ?? null;
  const factionChoiceActions =
    selectedGroup?.key === "faction"
      ? selectedGroup.actions
          .map((action) => ({ action, label: factionChoiceLabel(action, state), detail: factionChoiceDetail(action, state) }))
          .filter((choice): choice is { action: Action; label: string; detail: string } => choice.label !== null)
      : [];
  const factionSpatialChoiceActions =
    selectedGroup?.key === "faction"
      ? selectedGroup.actions
          .map((action) => ({ action, label: factionSpatialChoiceLabel(action), detail: factionSpatialChoiceDetail(action) }))
          .filter((choice): choice is { action: Action; label: string; detail: string } => choice.label !== null)
      : [];
  const drawAdvanceChoices = selectedGroup?.key === "draw_advance" ? selectedGroup.actions : [];
  const cardEffectChoices = selectedGroup?.key === "card_effect" ? selectedGroup.actions : [];
  const battleCandidateIndices =
    selectedIntent === "battle" && selectedGroup
      ? selectedGroup.actions.map((action) => actions.indexOf(action)).filter((index) => index >= 0)
      : [];
  const movementCandidateIndices =
    selectedIntent === "movement" && selectedGroup
      ? selectedGroup.actions.map((action) => actions.indexOf(action)).filter((index) => index >= 0)
      : [];
  const buildRecruitCandidateIndices =
    selectedIntent === "build_recruit" && selectedGroup
      ? selectedGroup.actions.map((action) => actions.indexOf(action)).filter((index) => index >= 0)
      : [];
  const factionSpatialCandidateIndices =
    selectedIntent === "faction" && selectedGroup
      ? selectedGroup.actions.map((action) => actions.indexOf(action)).filter((index) => index >= 0)
      : [];
  const battleCandidateKey = battleCandidateIndices.join(",");
  const movementCandidateKey = movementCandidateIndices.join(",");
  const buildRecruitCandidateKey = buildRecruitCandidateIndices.join(",");
  const factionSpatialCandidateKey = factionSpatialCandidateIndices.join(",");
  const actorLabel = factionLabels[state.factionTurn] ?? "Unknown";
  const phaseLabel = phaseLabels[state.currentPhase] ?? "Unknown";
  const nextAssistSummary =
    actions.length > 0
      ? `Choose what ${actorLabel} just did, then apply the matching public action candidate.`
      : `Loading public candidates for ${actorLabel}. If the action involved hidden information, use Other observed event.`;

  useEffect(() => {
    onBattleCandidatesChange?.(battleCandidateIndices);
    return () => onBattleCandidatesChange?.([]);
  }, [battleCandidateKey, onBattleCandidatesChange]);

  useEffect(() => {
    onMovementCandidatesChange?.(movementCandidateIndices);
    return () => onMovementCandidatesChange?.([]);
  }, [movementCandidateKey, onMovementCandidatesChange]);

  useEffect(() => {
    onBuildRecruitCandidatesChange?.(buildRecruitCandidateIndices);
    return () => onBuildRecruitCandidatesChange?.([]);
  }, [buildRecruitCandidateKey, onBuildRecruitCandidatesChange]);

  useEffect(() => {
    onFactionSpatialCandidatesChange?.(factionSpatialCandidateIndices);
    return () => onFactionSpatialCandidatesChange?.([]);
  }, [factionSpatialCandidateKey, onFactionSpatialCandidatesChange]);

  if (state.gameMode !== 1 || state.gamePhase !== 1 || state.factionTurn === state.playerFaction) {
    return null;
  }

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

      <div className="assist-director-stack">
        <div className="flow-step-card active">
          <strong>What did {actorLabel} do?</strong>
          <span className="summary-line">
            Pick the public intent first. This keeps assist mode close to Root Digital: one observed decision, then one matching action.
          </span>
          {actions.length === 0 ? (
            <div className="summary-stack assist-loading-card">
              <span className="summary-line">Public candidates are loading for this observed turn.</span>
              <button type="button" className="secondary" onClick={() => void onGenerateActions()}>
                Refresh Public Candidates
              </button>
            </div>
          ) : (
            <IntentGrid
              groups={actionGroups}
              selectedIntent={selectedIntent}
              countLabel="candidate(s)"
              onSelect={(intent) => {
                setSelectedIntent(intent);
                setShowAllGeneratedActions(false);
                setManualCaptureOpen(false);
                setChoiceMessage("");
                setSelectedCraftCardID(null);
                setSelectedDecreeCardKey(null);
              }}
            />
          )}
        </div>

        <div className="flow-step-card note">
          <strong>Hidden or uncaptured event?</strong>
          <span className="summary-line">
            Use this only when the public candidates do not represent the observed table event or the event involved unknown cards.
          </span>
          <div className="shortcut-grid" style={{ marginTop: "0.5rem" }}>
            {observedPromptTemplates(state).map((prompt) => (
              <button
                key={prompt.label}
                type="button"
                className="secondary"
                onClick={() => {
                  setPreferredTemplate(prompt.template);
                  setManualCaptureOpen(true);
                  setSelectedIntent(null);
                  setChoiceMessage("");
                  setSelectedCraftCardID(null);
                  setSelectedDecreeCardKey(null);
                }}
              >
                {prompt.label}
              </button>
            ))}
            <button
              type="button"
              className="secondary"
              onClick={() => {
                setPreferredTemplate(null);
                setManualCaptureOpen((current) => !current);
                setSelectedIntent(null);
                setChoiceMessage("");
                setSelectedCraftCardID(null);
                setSelectedDecreeCardKey(null);
              }}
            >
              Other Observed Event
            </button>
          </div>
        </div>
      </div>

      {selectedGroup ? (
        <div className="summary-stack assist-candidate-list">
          <span className="summary-label">{selectedGroup.label} Candidates</span>
          <span className="summary-line">{selectedGroup.detail}</span>
          {selectedGroup.key === "movement" ? (
            <span className="summary-line">Use the board first: choose a highlighted source clearing, then a highlighted destination. Use the list only if multiple candidates share the same route.</span>
          ) : null}
          {selectedGroup.key === "battle" ? (
            <span className="summary-line">Use the board first: choose a highlighted battle clearing. If multiple defenders are possible there, choose the observed defender here.</span>
          ) : null}
          {selectedGroup.key === "build_recruit" ? (
            <span className="summary-line">Use the board first: choose a highlighted build, recruit, or overwork clearing. Use the list when one click would match multiple candidates.</span>
          ) : null}
          {selectedGroup.key === "faction" ? (
            <span className="summary-line">Use the board first for clearing-based faction actions like sympathy, revolt, organize, explore, aid, or strike. Use the list for card, item, or leader choices.</span>
          ) : null}
          {selectedGroup.key === "faction" && factionChoiceActions.length > 0 ? (
            <>
              <span className="summary-line">For non-spatial faction choices, choose the observed option directly instead of opening exact generated candidates.</span>
              <div className="assist-choice-grid">
                {factionChoiceActions.map((choice, actionIndex) => (
                  <button
                    key={`${choice.action.type}-${choice.label}-${actionIndex}`}
                    type="button"
                    className="assist-choice-card"
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
          ) : null}
          {selectedGroup.key === "faction" && factionSpatialChoiceActions.length > 0 ? (
            <>
              <span className="summary-line">
                For clearing-based faction actions, use the board first or choose the exact observed card, item, or target option here.
              </span>
              <div className="assist-choice-grid">
                {factionSpatialChoiceActions.map((choice, actionIndex) => (
                  <button
                    key={`${choice.action.type}-${choice.label}-${choice.detail}-${actionIndex}`}
                    type="button"
                    className="assist-choice-card"
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
          ) : null}
          {selectedGroup.key === "faction" && decreeCardChoices.length > 0 ? (
            <>
              <span className="summary-line">For Eyrie decree additions, choose the observed card set first, then choose the decree column assignment.</span>
              <div className="assist-choice-grid">
                {decreeCardChoices.map((choice) => (
                  <button
                    key={choice.key}
                    type="button"
                    className={`assist-choice-card ${selectedDecreeCardKey === choice.key ? "selected" : ""}`}
                    onClick={() => {
                      if (choice.actions.length === 1) {
                        setChoiceMessage("");
                        setSelectedDecreeCardKey(null);
                        void onApply(choice.actions[0]);
                        return;
                      }
                      setSelectedDecreeCardKey(choice.key);
                      setShowAllGeneratedActions(true);
                      setChoiceMessage(`${decreeCardLabel(choice.cardIDs)} can be added to multiple decree columns. Choose the observed column assignment.`);
                    }}
                  >
                    <strong>{decreeCardLabel(choice.cardIDs)}</strong>
                    <span>{choice.actions.length === 1 ? "Apply exact decree add" : `${choice.actions.length} column assignments`}</span>
                  </button>
                ))}
              </div>
              {selectedDecreeCardChoice ? (
                <div className="assist-choice-grid">
                  {selectedDecreeCardChoice.actions.map((action, actionIndex) => (
                    <button
                      key={`${selectedDecreeCardChoice.key}-${actionIndex}`}
                      type="button"
                      className="assist-choice-card"
                      onClick={() => {
                        setChoiceMessage("");
                        setSelectedDecreeCardKey(null);
                        void onApply(action);
                      }}
                    >
                      <strong>{decreeColumnAssignmentLabel(action)}</strong>
                      <span>Apply this decree assignment</span>
                    </button>
                  ))}
                </div>
              ) : null}
            </>
          ) : null}
          {selectedGroup.key === "battle" ? (
            <>
              <span className="summary-line">Choose the observed battle target. This opens Battle Flow directly when the defender choice maps to one legal battle.</span>
              <div className="assist-choice-grid">
                {battleTargetChoices.map((choice) => (
                  <button
                    key={choice.key}
                    type="button"
                    className="assist-choice-card"
                    onClick={() => {
                      if (choice.actions.length === 1) {
                        const actionIndex = actions.indexOf(choice.actions[0]);
                        if (actionIndex >= 0) {
                          setChoiceMessage("");
                          onOpenBattle(actionIndex);
                          return;
                        }
                      }
                      setShowAllGeneratedActions(true);
                      setChoiceMessage(`${choice.label} still maps to ${choice.actions.length} generated battle candidates. Choose the exact battle from the fallback list.`);
                    }}
                  >
                    <strong>{choice.label}</strong>
                    <span>{choice.actions.length === 1 ? "Open Battle Flow" : `${choice.actions.length} battle candidates`}</span>
                  </button>
                ))}
              </div>
            </>
          ) : null}
          {selectedGroup.key === "craft" ? (
            <>
              <span className="summary-line">Choose the crafted card first. If multiple workshop routes can craft it, choose the observed workshop route next.</span>
              <div className="assist-choice-grid">
                {craftChoices.map((choice) => (
                  <button
                    key={choice.cardID}
                    type="button"
                    className={`assist-choice-card ${selectedCraftCardID === choice.cardID ? "selected" : ""}`}
                    onClick={() => {
                      if (choice.actions.length === 1) {
                        setChoiceMessage("");
                        setSelectedCraftCardID(null);
                        void onApply(choice.actions[0]);
                        return;
                      }
                      setSelectedCraftCardID(choice.cardID);
                      setChoiceMessage(`${describeKnownCardID(choice.cardID)} has ${choice.actions.length} legal craft routes. Choose the observed workshop route.`);
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
                      className="assist-choice-card"
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
          ) : null}
          {selectedGroup.key === "draw_advance" ? (
            <>
              <span className="summary-line">Choose the observed turn bookkeeping directly. These are low-ambiguity flow actions, so they do not need the full candidate browser first.</span>
              <div className="assist-choice-grid">
                {drawAdvanceChoices.map((action, actionIndex) => (
                  <button
                    key={`${action.type}-${actionIndex}`}
                    type="button"
                    className="assist-choice-card"
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
          ) : null}
          {selectedGroup.key === "card_effect" ? (
            <>
              <span className="summary-line">Choose the observed card or persistent effect directly. If hidden information is unknown, use the hidden/uncaptured event prompts instead.</span>
              <div className="assist-choice-grid">
                {cardEffectChoices.map((action, actionIndex) => (
                  <button
                    key={`${action.type}-${actionIndex}`}
                    type="button"
                    className="assist-choice-card"
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
          ) : null}
          {choiceMessage ? <span className="message">{choiceMessage}</span> : null}
          <ExactActionDrawer
            title="Exact Generated Candidates"
            summary="Open only when the guided prompt is ambiguous or you need to audit the rules-engine candidates."
            open={exactCandidateDrawerOpen}
            actions={observedGeneratedActions}
            allActions={actions}
            state={state}
            onToggle={setShowAllGeneratedActions}
            onApply={onApply}
            onOpenBattle={onOpenBattle}
          />
        </div>
      ) : actions.length > 0 && !manualCaptureOpen ? (
        <div className="flow-step-card waiting assist-selection-empty">
          <strong>Select an intent above.</strong>
          <span className="summary-line">The matching public candidates will appear here after you choose what happened on the physical board.</span>
        </div>
      ) : null}

      {manualCaptureOpen ? (
      <div className="assist-manual-capture">
        <ObservedActionPanel
          state={state}
          onApply={onApply}
          embedded
          preferredActorFaction={state.factionTurn}
          preferredTemplate={preferredTemplate}
        />
      </div>
      ) : null}

      <details className="secondary-drawer assist-correction-drawer">
        <summary className="panel-summary">
          <span className="summary-label">Correction Mode</span>
          <span className="summary-line">Use only when the guided observed-turn flow cannot recover the table state.</span>
        </summary>
        <div className="sidebar-actions" style={{ marginTop: "0.8rem" }}>
          <button type="button" className="secondary" onClick={onOpenTurnState}>
            Advanced Turn
          </button>
          <button type="button" className="secondary" onClick={() => void onGenerateActions()}>
            Refresh Candidates
          </button>
        </div>
      </details>
    </section>
  );
}
