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
  type AssistActionCandidateRef,
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
  surface?: "sidebar" | "tray";
  showFallbackDrawer?: boolean;
  showCorrectionControls?: boolean;
  onApply: (action: Action) => Promise<void>;
  onGenerateActions: () => Promise<void>;
  onOpenTurnState: () => void;
  onOpenBattle: (actionIndex: number) => void;
  onBattleCandidatesChange?: (candidates: AssistActionCandidateRef[]) => void;
  onMovementCandidatesChange?: (candidates: AssistActionCandidateRef[]) => void;
  onBuildRecruitCandidatesChange?: (candidates: AssistActionCandidateRef[]) => void;
  onFactionSpatialCandidatesChange?: (candidates: AssistActionCandidateRef[]) => void;
};

export function AssistWorkflowPanel({
  state,
  actions,
  surface = "sidebar",
  showFallbackDrawer = true,
  showCorrectionControls = true,
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
  const [observedNotesOpen, setObservedNotesOpen] = useState(false);
  const [choiceMessage, setChoiceMessage] = useState("");
  const [selectedCraftCardID, setSelectedCraftCardID] = useState<number | null>(null);
  const [selectedDecreeCardKey, setSelectedDecreeCardKey] = useState<string | null>(null);
  const autoLoadKey = useRef("");
  const traySurface = surface === "tray";

  useEffect(() => {
    setPreferredTemplate(null);
    setShowAllGeneratedActions(false);
    setSelectedIntent(null);
    setManualCaptureOpen(false);
    setObservedNotesOpen(false);
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
  const candidateRefsForSelectedGroup = (enabled: boolean): AssistActionCandidateRef[] =>
    enabled && selectedGroup
      ? selectedGroup.actions
          .map((action) => ({ actionIndex: actions.indexOf(action), action }))
          .filter((candidate) => candidate.actionIndex >= 0)
      : [];
  const battleCandidates =
    selectedIntent === "battle" && selectedGroup
      ? candidateRefsForSelectedGroup(true)
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
  const battleCandidateKey = candidateKey(battleCandidates);
  const movementCandidateKey = candidateKey(movementCandidates);
  const buildRecruitCandidateKey = candidateKey(buildRecruitCandidates);
  const factionSpatialCandidateKey = candidateKey(factionSpatialCandidates);
  const actorLabel = factionLabels[state.factionTurn] ?? "Unknown";
  const phaseLabel = phaseLabels[state.currentPhase] ?? "Unknown";
  const nextAssistSummary =
    actions.length > 0
      ? `Choose what ${actorLabel} just did on the physical board, then record the matching guided step.`
      : `Reading the public board state for ${actorLabel}. If hidden information mattered, use Other observed event.`;

  useEffect(() => {
    onBattleCandidatesChange?.(battleCandidates);
    return () => onBattleCandidatesChange?.([]);
  }, [battleCandidateKey, onBattleCandidatesChange]);

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

  if (state.gameMode !== 1 || state.gamePhase !== 1 || state.factionTurn === state.playerFaction) {
    return null;
  }

  const hiddenObservedControls = (
    <div className="shortcut-grid">
      {observedPromptTemplates(state).map((prompt) => (
        <button
          key={prompt.label}
          type="button"
          className="secondary"
          onClick={() => {
            setPreferredTemplate(prompt.template);
            setManualCaptureOpen(true);
            setObservedNotesOpen(true);
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
            setObservedNotesOpen(true);
            setSelectedIntent(null);
            setChoiceMessage("");
          setSelectedCraftCardID(null);
          setSelectedDecreeCardKey(null);
        }}
      >
        Other Observed Event
      </button>
    </div>
  );

  return (
    <section className={`panel assist-workflow-panel ${surface === "tray" ? "board-action-panel board-action-panel-tray assist-workflow-tray" : "sidebar-panel"}`}>
      {traySurface ? (
        <>
          <div className="summary-stack assist-tray-summary">
            <p className="eyebrow">Observed Turn</p>
            <span className="summary-label">
              {actorLabel}: {phaseLabel} / {stepLabels[state.currentStep] ?? "Unknown"}
            </span>
            <strong>Record what happened on the table.</strong>
            <span className="summary-line">{nextAssistSummary}</span>
          </div>
          {actions.length === 0 ? (
            <div className="summary-stack assist-loading-card">
              <span className="summary-line">Reading the public board state for this observed turn.</span>
              <button type="button" className="secondary" onClick={() => void onGenerateActions()}>
                Refresh Table View
              </button>
            </div>
          ) : (
            <IntentGrid
              groups={actionGroups}
              selectedIntent={selectedIntent}
              countLabel="option(s)"
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
          <details
            className="secondary-drawer assist-observed-notes-drawer"
            open={manualCaptureOpen || observedNotesOpen}
            onToggle={(event) => setObservedNotesOpen(event.currentTarget.open)}
          >
            <summary className="panel-summary">
              <span className="summary-label">Table Notes & Hidden Info</span>
              <span className="summary-line">Use this only when the board-first recorded steps do not fully match what happened at the physical table.</span>
            </summary>
            <div className="assist-observed-notes-body">{hiddenObservedControls}</div>
          </details>
        </>
      ) : (
        <>
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
                Pick the public intent first, then record what happened on the table one step at a time.
              </span>
              {actions.length === 0 ? (
                <div className="summary-stack assist-loading-card">
                  <span className="summary-line">Reading the public board state for this observed turn.</span>
                  <button type="button" className="secondary" onClick={() => void onGenerateActions()}>
                    Refresh Table View
                  </button>
                </div>
              ) : (
                <IntentGrid
                  groups={actionGroups}
                  selectedIntent={selectedIntent}
                  countLabel="option(s)"
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
                Use this only when the guided public options do not match what happened on the table or when unknown cards matter.
              </span>
              <div style={{ marginTop: "0.5rem" }}>{hiddenObservedControls}</div>
            </div>
          </div>
        </>
      )}

      {selectedGroup ? (
        <div className="summary-stack assist-candidate-list">
          <span className="summary-label">{selectedGroup.label}</span>
          <span className="summary-line">{selectedGroup.detail}</span>
          {selectedGroup.key === "movement" ? (
            <span className="summary-line">Use the board first: choose a highlighted source clearing, then a highlighted destination. Only use the record list if the same route could mean more than one move.</span>
          ) : null}
          {selectedGroup.key === "battle" ? (
            <span className="summary-line">Use the board first: choose a highlighted battle clearing. If that clearing could have more than one defender, choose the observed defender here.</span>
          ) : null}
          {selectedGroup.key === "build_recruit" ? (
            <span className="summary-line">Use the board first: choose a highlighted build, recruit, or overwork clearing. Use the record list only when one clearing could match more than one step.</span>
          ) : null}
          {selectedGroup.key === "faction" ? (
            <span className="summary-line">Use the board first for clearing-based faction actions like sympathy, revolt, organize, explore, aid, or strike. Use the record cards here for card, item, or leader choices.</span>
          ) : null}
          {selectedGroup.key === "faction" && factionChoiceActions.length > 0 ? (
            <>
              <span className="summary-line">For non-spatial faction choices, record the observed step directly instead of opening the audit drawer.</span>
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
                For clearing-based faction actions, use the board first or choose the matching observed card, item, or target here.
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
                    <span>{choice.actions.length === 1 ? "Record this decree add" : `${choice.actions.length} column choices`}</span>
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
                      <span>Record this decree choice</span>
                    </button>
                  ))}
                </div>
              ) : null}
            </>
          ) : null}
          {selectedGroup.key === "battle" ? (
            <>
              <span className="summary-line">Choose the observed battle target. This opens the battle event immediately when the defender choice maps to one matching fight.</span>
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
                      setChoiceMessage(`${choice.label} could still mean ${choice.actions.length} different battle records. Open the audit drawer and choose the matching one.`);
                    }}
                  >
                    <strong>{choice.label}</strong>
                    <span>{choice.actions.length === 1 ? "Open Battle" : `${choice.actions.length} battle records`}</span>
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
                      setChoiceMessage(`${describeKnownCardID(choice.cardID)} could be crafted through ${choice.actions.length} workshop paths. Choose the observed workshop route.`);
                    }}
                  >
                    <strong>{describeKnownCardID(choice.cardID)}</strong>
                    <span>{choice.actions.length === 1 ? "Record this craft" : `${choice.actions.length} workshop paths`}</span>
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
                      <span>Record this workshop route</span>
                    </button>
                  ))}
                </div>
              ) : null}
            </>
          ) : null}
          {selectedGroup.key === "draw_advance" ? (
            <>
              <span className="summary-line">Choose the observed turn bookkeeping directly. These are straightforward table-state steps, so they do not need the full record drawer first.</span>
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
          {showFallbackDrawer ? (
            <ExactActionDrawer
              title="Observed Record Audit"
              summary="Open only when the guided prompt is ambiguous or you need to inspect the exact rule-backed record choices."
              open={exactCandidateDrawerOpen}
              actions={observedGeneratedActions}
              allActions={actions}
              state={state}
              onToggle={setShowAllGeneratedActions}
              onApply={onApply}
              onOpenBattle={onOpenBattle}
            />
          ) : null}
        </div>
      ) : actions.length > 0 && !manualCaptureOpen ? (
        <div className="flow-step-card waiting assist-selection-empty">
          <strong>Select an intent above.</strong>
          <span className="summary-line">The matching guided record choices will appear here after you choose what happened on the physical board.</span>
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

      {showCorrectionControls ? (
        <details className="secondary-drawer assist-correction-drawer">
          <summary className="panel-summary">
            <span className="summary-label">Correction Mode</span>
            <span className="summary-line">Use only when the guided observed-turn flow cannot cleanly match the real table state.</span>
          </summary>
          <div className="sidebar-actions" style={{ marginTop: "0.8rem" }}>
            <button type="button" className="secondary" onClick={onOpenTurnState}>
              Advanced Turn
            </button>
            <button type="button" className="secondary" onClick={() => void onGenerateActions()}>
              Refresh Table View
            </button>
          </div>
        </details>
      ) : null}
    </section>
  );
}
