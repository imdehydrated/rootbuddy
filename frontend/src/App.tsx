import { useEffect, useRef, useState } from "react";
import { boardLayoutForState } from "./boardLayouts";
import { describeKnownCardID } from "./cardCatalog";
import { AssistWorkflowPanel } from "./components/AssistWorkflowPanel";
import { BattleFlowPanel } from "./components/BattleFlowPanel";
import { CardHandTray } from "./components/CardHandTray";
import { BoardPanel } from "./components/BoardPanel";
import { CardVisibilityPanel } from "./components/CardVisibilityPanel";
import { EndgamePanel } from "./components/EndgamePanel";
import { GameLogPanel } from "./components/GameLogPanel";
import { GuideHelpPanel } from "./components/GuideHelpPanel";
import { InspectorPanel } from "./components/InspectorPanel";
import { JoinScreen } from "./components/JoinScreen";
import { LobbyScreen } from "./components/LobbyScreen";
import { PhaseBar } from "./components/PhaseBar";
import { PlayerActionsPanel } from "./components/PlayerActionsPanel";
import { PlayerPresenceBar } from "./components/PlayerPresenceBar";
import { SettingsPanel } from "./components/SettingsPanel";
import { SessionStatusPanel } from "./components/SessionStatusPanel";
import { SetupWizard } from "./components/SetupWizard";
import { TurnFlowPanel } from "./components/TurnFlowPanel";
import { TurnStatePanel } from "./components/TurnStatePanel";
import { TurnSummaryPanel } from "./components/TurnSummaryPanel";
import { VPTracker } from "./components/VPTracker";
import { rulerOfClearing, usedBuildSlots } from "./gameHelpers";
import { ACTION_TYPE, factionLabels, phaseLabels, setupStageLabels, stepLabels, suitLabels } from "./labels";
import { clearSavedSession } from "./localSession";
import { gameOverHeadline, initialState, useGameState } from "./hooks/useGameState";
import { emptyBattleModifiers, useBattleFlow } from "./hooks/useBattleFlow";
import { setupBoardPrompt, useBoardInteraction } from "./hooks/useBoardInteraction";
import { useMultiplayer } from "./hooks/useMultiplayer";
import { useSettings } from "./hooks/useSettings";
import { useSessionPersistence } from "./hooks/useSessionPersistence";
import type { ActiveModal } from "./modalState";
import type { Action } from "./types";

export default function App() {
  const [activeModal, setActiveModal] = useState<ActiveModal>(null);
  const [pendingStandAndDeliverAction, setPendingStandAndDeliverAction] = useState<Action | null>(null);
  const [standAndDeliverCardID, setStandAndDeliverCardID] = useState("");
  const [showGuideHelp, setShowGuideHelp] = useState(true);
  const [showAdvancedTurnPanel, setShowAdvancedTurnPanel] = useState(false);
  const [showBoardEditor, setShowBoardEditor] = useState(false);
  const [showContextDrawer, setShowContextDrawer] = useState(false);
  const [showWorkspaceTools, setShowWorkspaceTools] = useState(false);
  const [showRecoveryTools, setShowRecoveryTools] = useState(false);
  const [playerTrayPrompt, setPlayerTrayPrompt] = useState<string | null>(null);
  const autoLoadedActionKey = useRef("");
  const previousGamePhase = useRef(initialState.gamePhase);
  const { settings, updateSetting, resetSettings } = useSettings();

  function getMultiplayerToken() {
    return multiplayer.multiplayerToken;
  }

  function getPlayerFaction() {
    return gameState.parsedState.playerFaction;
  }

  const gameState = useGameState({
    getMultiplayerToken,
    jsonEditorOpen: activeModal === "json"
  });

  const session = useSessionPersistence({
    getMultiplayerToken,
    loadSavedGame: gameState.loadSavedGame,
    loadActionsForState: gameState.loadActionsForState,
    parsedState: gameState.parsedState,
    serverGameID: gameState.serverGameID,
    serverRevision: gameState.serverRevision,
    setActiveModal,
    setServerGameID: gameState.setServerGameID,
    setServerRevision: gameState.setServerRevision,
    setShowAdvancedTurnPanel,
    setShowBoardEditor,
    setShowContextDrawer,
    setShowGuideHelp,
    setShowRecoveryTools,
    setShowWorkspaceTools,
    setStatus: gameState.setStatus,
    syncState: gameState.syncState
  });

  const multiplayer = useMultiplayer({
    enterLoadedGame: session.enterLoadedGame,
    loadActionsForState: gameState.loadActionsForState,
    resetToSetup: session.resetToSetup,
    setServerGameID: gameState.setServerGameID,
    setServerRevision: gameState.setServerRevision,
    setShowBoardEditor,
    setShowGuideHelp,
    setShowSetup: session.setShowSetup,
    setSetupScreen: session.setSetupScreen,
    setStatus: gameState.setStatus,
    syncState: gameState.syncState,
    getPlayerFaction,
    serverGameID: gameState.serverGameID,
    showSetup: session.showSetup
  });

  const previewedAction =
    gameState.actions[gameState.hoveredActionIndex ?? gameState.selectedBattleIndex ?? -1] ?? null;

  const battle = useBattleFlow({
    actions: gameState.actions,
    getMultiplayerToken,
    multiplayerBattlePrompt: multiplayer.multiplayerBattlePrompt,
    parsedState: gameState.parsedState,
    perspectiveFaction: multiplayer.perspectiveFaction,
    selectedBattleIndex: gameState.selectedBattleIndex,
    serverGameID: gameState.serverGameID,
    serverRevision: gameState.serverRevision,
    setHoveredActionIndex: gameState.setHoveredActionIndex,
    setMultiplayerBattlePrompt: multiplayer.setMultiplayerBattlePrompt,
    setMultiplayerSubmitting: multiplayer.setMultiplayerSubmitting,
    setSelectedBattleIndex: gameState.setSelectedBattleIndex,
    setServerRevision: gameState.setServerRevision,
    setStatus: gameState.setStatus,
    syncState: gameState.syncState,
    loadActionsForState: gameState.loadActionsForState
  });

  async function handleApply(action: Action) {
    if (gameState.isImpossibleStandAndDeliver(action)) {
      gameState.setStatus("Stand and Deliver cannot target a faction with no recorded cards.");
      return;
    }
    if (gameState.needsStandAndDeliverObservation(action)) {
      setPendingStandAndDeliverAction(action);
      setStandAndDeliverCardID("");
      setActiveModal("standAndDeliver");
      gameState.setStatus("Confirm the Stand and Deliver observation before applying.");
      return;
    }
    await gameState.applyFinalizedAction(action);
  }

  const board = useBoardInteraction({
    actions: gameState.actions,
    activeModal,
    multiplayerToken: multiplayer.multiplayerToken,
    parsedState: gameState.parsedState,
    previewedAction,
    setStatus: gameState.setStatus,
    setShowBoardEditor,
    onApply: handleApply,
    onOpenBattle: battle.openBattleForActionIndex
  });

  const multiplayerToken = multiplayer.multiplayerToken;
  const parsedState = gameState.parsedState;
  const selectedClearing = board.selectedClearing;
  const selectedClearingRuler = selectedClearing ? rulerOfClearing(selectedClearing) : null;
  const selectedClearingRulerLabel =
    typeof selectedClearingRuler === "number" ? factionLabels[selectedClearingRuler] ?? "Unknown" : "None";
  const boardLayout = boardLayoutForState(parsedState);
  const isMultiplayerGame = multiplayerToken !== null;
  const showPrimarySetupFlow = parsedState.gamePhase === 0;
  const hasPrimaryBattleFlow = Boolean(battle.activeBattleAction?.battle);
  const showPrimaryReviewFlow = parsedState.gamePhase === 2;
  const showPrimaryAssistFlow =
    parsedState.gamePhase === 1 &&
    parsedState.gameMode === 1 &&
    parsedState.factionTurn !== multiplayer.perspectiveFaction &&
    !hasPrimaryBattleFlow;
  const showPrimaryPlayerFlow =
    parsedState.gamePhase === 1 &&
    parsedState.factionTurn === multiplayer.perspectiveFaction &&
    !hasPrimaryBattleFlow;

  useEffect(() => {
    const previousPhase = previousGamePhase.current;
    previousGamePhase.current = parsedState.gamePhase;

    if (parsedState.gamePhase === 2) {
      setShowContextDrawer(true);
      setShowWorkspaceTools(true);
      setShowRecoveryTools(false);
      setShowBoardEditor(false);
      if (previousPhase !== 2) {
        setShowGuideHelp(true);
      }
      return;
    }

    if (parsedState.gamePhase === 0) {
      setShowContextDrawer(false);
      setShowWorkspaceTools(true);
      setShowRecoveryTools(false);
      setShowBoardEditor(false);
      return;
    }

    setShowContextDrawer(false);
    setShowWorkspaceTools(false);
    if (multiplayerToken) {
      setShowRecoveryTools(false);
    }
    if (previousPhase === 2) {
      setShowGuideHelp(false);
    }
  }, [multiplayerToken, parsedState.gamePhase]);

  useEffect(() => {
    if (!session.showSetup && parsedState.gamePhase === 1 && !multiplayer.multiplayerBattlePrompt && parsedState.factionTurn === multiplayer.perspectiveFaction) {
      const loadKey = [
        gameState.serverGameID ?? "",
        gameState.serverRevision ?? "na",
        parsedState.roundNumber,
        parsedState.currentPhase,
        parsedState.currentStep,
        parsedState.factionTurn
      ].join(":");
      if (autoLoadedActionKey.current === loadKey) {
        return;
      }

      autoLoadedActionKey.current = loadKey;
      let cancelled = false;

      void (async () => {
        try {
          gameState.setStatus("Your turn. Getting the board ready...");
          await gameState.loadActionsForState(gameState.currentStateRef.current);
        } catch (err) {
          if (!cancelled) {
            autoLoadedActionKey.current = "";
            gameState.setStatus(err instanceof Error ? err.message : "Failed to prepare turn options");
          }
        }
      })();

      return () => {
        cancelled = true;
      };
    }

    autoLoadedActionKey.current = "";
  }, [
    gameState,
    multiplayer.multiplayerBattlePrompt,
    multiplayer.perspectiveFaction,
    parsedState.currentPhase,
    parsedState.currentStep,
    parsedState.factionTurn,
    parsedState.gamePhase,
    parsedState.roundNumber,
    session.showSetup
  ]);

  useEffect(() => {
    if (!showPrimaryPlayerFlow) {
      setPlayerTrayPrompt(null);
    }
  }, [showPrimaryPlayerFlow]);

  const setupPrompt = showPrimarySetupFlow ? setupBoardPrompt(parsedState.setupStage, board.marquiseSetupDraft) : null;
  const phaseStatusLabel =
    parsedState.gamePhase === 0
      ? setupStageLabels[parsedState.setupStage] ?? "Setup"
      : parsedState.gamePhase === 2
        ? gameOverHeadline(parsedState)
        : `${factionLabels[parsedState.factionTurn] ?? "Unknown"} • ${phaseLabels[parsedState.currentPhase] ?? "Unknown"} / ${stepLabels[parsedState.currentStep] ?? "Unknown"}`;
  const connectionStatusLabel = multiplayerToken
    ? multiplayer.multiplayerConnectionStatus === "connected"
      ? "Live"
      : multiplayer.multiplayerConnectionStatus === "reconnecting"
        ? "Reconnecting"
        : multiplayer.multiplayerConnectionStatus === "connecting"
          ? "Connecting"
          : "Offline"
    : null;

  const primaryFlowLabel = showPrimaryReviewFlow
    ? "Review"
    : hasPrimaryBattleFlow
      ? "Battle"
      : showPrimarySetupFlow
        ? "Setup"
        : showPrimaryAssistFlow
          ? "Observed Turn"
          : showPrimaryPlayerFlow
            ? "Your Turn"
            : isMultiplayerGame
              ? "Waiting"
              : "Flow";

  const primaryFlowSummary = (() => {
    if (showPrimaryReviewFlow) {
      return "The match is complete. Review the result here before touching restart or recovery tools.";
    }
    if (hasPrimaryBattleFlow) {
      if (multiplayer.multiplayerBattlePrompt) {
        if (
          (multiplayer.multiplayerBattlePrompt.stage === "defender_response" || multiplayer.multiplayerBattlePrompt.stage === "attacker_response") &&
          multiplayer.multiplayerBattlePrompt.waitingOnFaction === multiplayer.perspectiveFaction
        ) {
          return "Your battle response is blocking the turn. Submit it here to continue.";
        }
        if (
          (multiplayer.multiplayerBattlePrompt.stage === "defender_response" || multiplayer.multiplayerBattlePrompt.stage === "attacker_response") &&
          multiplayer.multiplayerBattlePrompt.waitingOnFaction !== multiplayer.perspectiveFaction
        ) {
          return `Battle progress is waiting on ${factionLabels[multiplayer.multiplayerBattlePrompt.waitingOnFaction] ?? "another player"}.`;
        }
        if (multiplayer.multiplayerBattlePrompt.stage === "ready_to_resolve") {
          return multiplayer.multiplayerBattlePrompt.action.battle?.faction === multiplayer.perspectiveFaction
            ? "All battle responses are in. Resolve here to return to the turn flow."
            : `All battle responses are in. Waiting on ${factionLabels[multiplayer.multiplayerBattlePrompt.action.battle?.faction ?? -1] ?? "the attacker"} to resolve.`;
        }
      }
      return "Battle flow is blocking the turn. Finish it here before returning to other actions.";
    }
    if (showPrimarySetupFlow) {
      return gameState.actions.length > 0
        ? "Highlighted setup targets are ready on the board."
        : "Load setup choices, then complete the staged board selections here.";
    }
    if (showPrimaryAssistFlow) {
      return gameState.actions.length > 0
        ? `${gameState.actions.length} guided observed step(s) are ready for the current table state.`
        : "Refresh the table view or record the table event manually here.";
    }
    if (showPrimaryPlayerFlow) {
      if (playerTrayPrompt) {
        return playerTrayPrompt;
      }
      if (gameState.actions.length > 0) {
        return "Choose your next move on the board or in the action tray.";
      }
      return isMultiplayerGame
        ? "Your turn has priority. Board-ready options refresh automatically here."
        : "Preparing board-ready turn options automatically.";
    }
    if (isMultiplayerGame) {
      if (multiplayer.multiplayerConnectionStatus === "reconnecting" || multiplayer.multiplayerConnectionStatus === "connecting") {
        return "Realtime connection is recovering. Stay here while the server resynchronizes state.";
      }
      return `Waiting on ${factionLabels[parsedState.factionTurn] ?? "another player"} until the server hands your faction priority.`;
    }
    return "Use this area for the active workflow before opening any secondary tools.";
  })();
  if (session.showSetup) {
    if (multiplayer.multiplayerLobby) {
      return (
        <LobbyScreen
          lobby={multiplayer.multiplayerLobby}
          self={multiplayer.multiplayerSelf}
          connectionStatus={multiplayer.multiplayerConnectionStatus}
          status={gameState.status}
          submitting={multiplayer.multiplayerSubmitting}
          onClaimFaction={multiplayer.handleClaimLobby}
          onReady={multiplayer.handleSetLobbyReady}
          onStart={multiplayer.handleStartLobby}
          onLeave={multiplayer.handleLeaveLobby}
        />
      );
    }

    if (session.setupScreen === "create-lobby" || session.setupScreen === "join-lobby") {
      return (
        <JoinScreen
          mode={session.setupScreen === "create-lobby" ? "create" : "join"}
          submitting={multiplayer.multiplayerSubmitting}
          status={gameState.status}
          onBack={() => {
            session.setSetupScreen("wizard");
            gameState.setStatus("Choose how you want to play.");
          }}
          onCreateLobby={multiplayer.handleCreateLobby}
          onJoinLobby={multiplayer.handleJoinLobby}
        />
      );
    }

    return (
      <SetupWizard
        canResume={session.hasSavedSession}
        savedSessionInfo={session.savedSessionInfo}
        onStart={async (state, gameID, revision) => {
          multiplayer.clearMultiplayerState();
          session.enterLoadedGame(state, gameID, revision, state.gamePhase === 0 ? "Setup started." : "Game created.");
          if (state.gamePhase === 0) {
            try {
              await gameState.loadActionsForState(state, { successStatus: "Choose a highlighted setup target." });
            } catch (err) {
              gameState.setStatus(err instanceof Error ? err.message : "Failed to load setup actions");
            }
          }
        }}
        onUseSample={() => {
          multiplayer.clearMultiplayerState();
          gameState.syncState(initialState);
          gameState.setServerGameID(null);
          gameState.setServerRevision(null);
          session.setShowSetup(false);
          setShowBoardEditor(false);
          gameState.setStatus("Loaded sample state.");
          setShowGuideHelp(true);
        }}
        onOpenCreateLobby={() => {
          session.setSetupScreen("create-lobby");
          gameState.setStatus("Enter your display name to create a lobby.");
        }}
        onOpenJoinLobby={() => {
          session.setSetupScreen("join-lobby");
          gameState.setStatus("Enter your display name and join code.");
        }}
        onClearSavedSession={() => {
          clearSavedSession();
          session.setHasSavedSession(false);
          session.setSavedSessionInfo(null);
          gameState.setStatus("Cleared saved game.");
        }}
        onResume={session.handleResumeSavedGame}
      />
    );
  }

  const standAndDeliverTargetFaction = pendingStandAndDeliverAction?.usePersistentEffect?.targetFaction ?? -1;
  const standAndDeliverTargetLabel = factionLabels[standAndDeliverTargetFaction] ?? "the target faction";
  const standAndDeliverParsedCardID = Number(standAndDeliverCardID);
  const standAndDeliverCardEntryIsInvalid =
    standAndDeliverCardID.trim().length > 0 &&
    (!Number.isInteger(standAndDeliverParsedCardID) || standAndDeliverParsedCardID <= 0);
  const standAndDeliverCardLabel =
    standAndDeliverCardID.trim().length > 0 && !standAndDeliverCardEntryIsInvalid
      ? describeKnownCardID(standAndDeliverParsedCardID)
      : null;

  return (
    <main className="app-shell workspace-shell">
      <div className="board-stage">
        <div className="board-phase-stack">
          <PhaseBar
            gamePhase={parsedState.gamePhase}
            currentPhase={parsedState.currentPhase}
            currentStep={parsedState.currentStep}
            setupStage={parsedState.setupStage}
            factionTurn={parsedState.factionTurn}
            roundNumber={parsedState.roundNumber}
          />
          {settings.showVPTracker ? (
            <VPTracker
              victoryPoints={parsedState.victoryPoints}
              turnOrder={parsedState.turnOrder}
              dominance={parsedState.activeDominance}
              coalitionActive={parsedState.coalitionActive}
              coalitionPartner={parsedState.coalitionPartner}
            />
          ) : null}
        </div>
        {setupPrompt ? (
          <>
            <div className="board-setup-prompt">
              <p className="eyebrow">Setup</p>
              <strong>{setupPrompt.instruction}</strong>
              <span>{setupPrompt.detail}</span>
              <span>{board.setupLegalChoiceCount} legal choice{board.setupLegalChoiceCount === 1 ? "" : "s"}</span>
              {parsedState.setupStage === 1 && board.hasMarquiseDraftSelection ? (
                <button
                  type="button"
                  className="secondary"
                  onClick={() => {
                    board.setMarquiseSetupDraft({
                      keepClearingID: null,
                      sawmillClearingID: null,
                      workshopClearingID: null,
                      recruiterClearingID: null
                    });
                    gameState.setStatus("Marquise setup draft reset.");
                  }}
                >
                  Reset Marquise setup
                </button>
              ) : null}
            </div>
          </>
        ) : (
          <>
            <div className="board-top-hud">
              <div className="board-status-stack">
                <section className="panel board-status-hud">
                  <p className="eyebrow">RootBuddy</p>
                  <div className="status-block">
                    <div className="status-block-main">
                      <strong>{factionLabels[parsedState.factionTurn] ?? "Unknown"}</strong>
                      <span>{phaseLabels[parsedState.currentPhase] ?? "Unknown"} / {stepLabels[parsedState.currentStep] ?? "Unknown"}</span>
                    </div>
                    {connectionStatusLabel ? (
                      <span className={`connection-pill compact ${multiplayer.multiplayerConnectionStatus}`}>{connectionStatusLabel}</span>
                    ) : null}
                  </div>
                </section>
                {isMultiplayerGame && multiplayer.multiplayerLobby ? (
                  <PlayerPresenceBar
                    players={multiplayer.multiplayerLobby.players}
                    factionTurn={parsedState.factionTurn}
                    perspectiveFaction={multiplayer.perspectiveFaction}
                  />
                ) : null}
              </div>
              <div className="board-utility-hud">
                {multiplayer.multiplayerNotice ? (
                  <section className={`panel notice-panel board-notice-panel ${multiplayer.multiplayerNotice.level}`}>
                    <p className="eyebrow">{multiplayer.multiplayerNotice.title}</p>
                    <span className="summary-line">{multiplayer.multiplayerNotice.detail}</span>
                  </section>
                ) : null}
                <div className="board-utility-actions">
                  <button type="button" className="secondary correction-mode-toggle" onClick={() => setActiveModal("settings")}>
                    Settings
                  </button>
                  <button type="button" className="secondary correction-mode-toggle" onClick={() => setActiveModal("correction")}>
                    Correction Mode
                  </button>
                </div>
              </div>
            </div>
            <div className="board-prompt-strip">
              <p className="eyebrow">{primaryFlowLabel}</p>
              <strong>{gameState.status}</strong>
              <span>{gameState.error || primaryFlowSummary}</span>
            </div>
          </>
        )}
        {parsedState.gamePhase === 2 ? (
          <div className="board-hint endgame-banner">
            {gameOverHeadline(parsedState)}. Review the result from the board event panel.
          </div>
        ) : null}
        <BoardPanel
          state={parsedState}
          clearings={parsedState.map.clearings}
          forests={parsedState.map.forests}
          boardLayout={boardLayout}
          selectedClearingID={board.selectedClearingID}
          keepClearingID={parsedState.marquise.keepClearingID}
          vagabondClearingID={parsedState.vagabond.clearingID}
          vagabondInForest={parsedState.vagabond.inForest}
          highlightedClearings={board.highlightedClearings}
          previewedAction={previewedAction}
          setupLegalClearingIDs={board.legalSetupClearingIDs}
          setupSelectedClearingIDs={board.selectedSetupClearingIDs}
          setupPreviewPiecesByClearing={board.setupPreviewPiecesByClearing}
          forestTargets={board.forestTargets}
          onSelectClearing={board.handleSetupClearingClick}
          onSelectForest={board.handleSetupForestClick}
        />
        {multiplayer.multiplayerToken && settings.showGameLog ? (
          <div className="board-log-shell">
            <GameLogPanel entries={multiplayer.actionLog} factionTurn={parsedState.factionTurn} />
          </div>
        ) : null}
        {settings.showCardTray ? (
          <div className={`board-card-tray-shell ${showPrimaryPlayerFlow || showPrimaryAssistFlow ? "with-action-tray" : ""}`}>
            <CardHandTray state={parsedState} compactCards={settings.compactCards} />
          </div>
        ) : null}
        {showPrimaryPlayerFlow ? (
          <div className="board-action-tray" aria-label="Live action tray">
            <PlayerActionsPanel
              state={parsedState}
              actions={gameState.actions}
              isMultiplayer={isMultiplayerGame}
              onPromptChange={setPlayerTrayPrompt}
              onApply={handleApply}
              onGenerateActions={gameState.refreshActions}
              onOpenBattle={battle.openBattleForActionIndex}
              onPreviewAction={gameState.setHoveredActionIndex}
              onMovementCandidatesChange={board.handlePlayerMovementCandidatesChange}
              onBuildRecruitCandidatesChange={board.handlePlayerBuildRecruitCandidatesChange}
              onFactionSpatialCandidatesChange={board.handlePlayerFactionSpatialCandidatesChange}
              surface="tray"
              showFallbackDrawer={false}
              showRefreshButton={false}
            />
          </div>
        ) : null}
        {showPrimaryAssistFlow ? (
          <div className="board-action-tray" aria-label="Observed action tray">
            <AssistWorkflowPanel
              state={parsedState}
              actions={gameState.actions}
              onApply={handleApply}
              onGenerateActions={gameState.refreshActions}
              onOpenTurnState={() => setActiveModal("correction")}
              onOpenBattle={battle.openBattleForActionIndex}
              onBattleCandidatesChange={board.handleAssistBattleCandidatesChange}
              onMovementCandidatesChange={board.handleAssistMovementCandidatesChange}
              onBuildRecruitCandidatesChange={board.handleAssistBuildRecruitCandidatesChange}
              onFactionSpatialCandidatesChange={board.handleAssistFactionSpatialCandidatesChange}
              surface="tray"
              showFallbackDrawer={false}
              showCorrectionControls={false}
            />
          </div>
        ) : null}
        {hasPrimaryBattleFlow ? (
          <div className="board-event-shell" aria-label="Battle event">
            <BattleFlowPanel
              selectedBattleIndex={gameState.selectedBattleIndex}
              selectedBattleAction={battle.selectedBattleAction}
              multiplayerBattlePrompt={multiplayer.multiplayerBattlePrompt}
              multiplayerPerspectiveFaction={multiplayer.perspectiveFaction}
              multiplayerSubmitting={multiplayer.multiplayerSubmitting}
              attackerFaction={battle.attackerFaction}
              defenderFaction={battle.defenderFaction}
              attackerRoll={battle.attackerRoll}
              defenderRoll={battle.defenderRoll}
              battleModifiers={battle.battleModifiers}
              battleContext={battle.battleContext}
              assistDefenderAmbushChoice={battle.assistDefenderAmbushChoice}
              onSetAttackerRoll={battle.setAttackerRoll}
              onSetDefenderRoll={battle.setDefenderRoll}
              onSetBattleModifiers={(updater) => battle.setBattleModifiers((current) => updater(current))}
              onSetAssistDefenderAmbushChoice={battle.setAssistDefenderAmbushChoice}
              onSubmitMultiplayerResponse={battle.handleSubmitBattleResponse}
              onResolveAndApply={battle.handleResolveAndApply}
              onClearSelection={battle.clearBattleSelection}
              surface="modal"
            />
          </div>
        ) : null}
        {showPrimaryReviewFlow ? (
          <div className="board-event-shell review-event-shell" aria-label="Review event">
            <EndgamePanel
              state={parsedState}
              hasSavedSession={session.hasSavedSession}
              serverGameID={gameState.serverGameID}
              onNewGame={() => {
                if (multiplayerToken) {
                  gameState.setStatus("Starting a new game from the in-game multiplayer workspace is not supported.");
                  return;
                }
                session.resetToSetup({ clearSaved: true, status: "Start a new game." });
              }}
              onReturnToSetup={() => {
                if (multiplayerToken) {
                  gameState.setStatus("Return to setup is disabled while a multiplayer session is active.");
                  return;
                }
                session.resetToSetup({ status: "Returned to setup. Resume is still available until you clear it." });
              }}
              onClearSavedSession={() => {
                clearSavedSession();
                session.setHasSavedSession(false);
                session.setSavedSessionInfo(null);
                gameState.setStatus("Cleared the saved endgame result.");
              }}
              onOpenDebug={() => {
                if (multiplayerToken) {
                  gameState.setStatus("Debug JSON is disabled in multiplayer.");
                  return;
                }
                setActiveModal("json");
              }}
              surface="modal"
            />
          </div>
        ) : null}
      </div>

      {activeModal ? (
        <div className="modal-backdrop" onClick={() => setActiveModal(null)}>
          <div className="modal-shell" onClick={(event) => event.stopPropagation()}>
            {activeModal === "correction" ? (
              <section className="panel modal-panel correction-mode-panel">
                <div className="panel-header">
                  <h2>Correction Mode</h2>
                  <button type="button" className="secondary" onClick={() => setActiveModal(null)}>
                    Close
                  </button>
                </div>
                <p className="message">
                  Recovery, inspection, and manual tools live here. Routine play should stay on the board.
                </p>
                <div className="correction-mode-grid">
                  <details className="panel secondary-drawer" open={showContextDrawer} onToggle={(event) => setShowContextDrawer(event.currentTarget.open)}>
                    <summary className="panel-summary">
                      <span className="summary-label">Context & Reference</span>
                      <span className="summary-line">Turn summary, card visibility, and session details.</span>
                    </summary>
                    <div className="context-drawer-body">
                      <TurnSummaryPanel state={parsedState} />
                      <CardVisibilityPanel state={parsedState} />
                      <SessionStatusPanel
                        state={parsedState}
                        hasSavedSession={session.hasSavedSession}
                        serverGameID={gameState.serverGameID}
                        savedSessionInfo={session.savedSessionInfo}
                        multiplayerSession={multiplayer.multiplayerSession}
                        multiplayerConnectionStatus={multiplayer.multiplayerConnectionStatus}
                        multiplayerBattlePrompt={multiplayer.multiplayerBattlePrompt}
                      />
                    </div>
                  </details>

                  <details className="panel secondary-drawer" open={showWorkspaceTools} onToggle={(event) => setShowWorkspaceTools(event.currentTarget.open)}>
                    <summary className="panel-summary">
                      <span className="summary-label">{parsedState.gamePhase === 2 ? "Review Workspace" : "Workspace Tools"}</span>
                      <span className="summary-line">Board selection, restart, and review controls live here.</span>
                    </summary>
                    <div className="summary-stack" style={{ margin: "0.9rem 0" }}>
                      <span className="summary-line">Selected clearing: {selectedClearing?.id ?? "None"}</span>
                      {selectedClearing ? (
                        <>
                          <span className="summary-line">Suit: {suitLabels[selectedClearing.suit] ?? "Unknown"}</span>
                          <span className="summary-line">Ruler: {selectedClearingRulerLabel}</span>
                          <span className="summary-line">
                            Paths / slots: {selectedClearing.adj.length} / {usedBuildSlots(selectedClearing)}/{selectedClearing.buildSlots}
                          </span>
                        </>
                      ) : null}
                    </div>
                    <div className="sidebar-actions">
                      <button
                        type="button"
                        className="secondary"
                        onClick={() => {
                          if (multiplayerToken) {
                            gameState.setStatus("Return to setup is disabled while a multiplayer session is active.");
                            return;
                          }
                          session.resetToSetup({
                            clearSaved: parsedState.gamePhase !== 2,
                            status: parsedState.gamePhase === 2 ? "Returned to setup. Resume is still available until you clear it." : "Start a new game."
                          });
                        }}
                      >
                        {parsedState.gamePhase === 2 ? "Return to Setup" : "Setup"}
                      </button>
                      {parsedState.gamePhase === 2 ? (
                        <button type="button" onClick={() => session.resetToSetup({ clearSaved: true, status: "Start a new game." })}>
                          New Game
                        </button>
                      ) : null}
                    </div>
                    {!multiplayerToken && parsedState.gamePhase !== 2 ? (
                      <div className="sidebar-actions footer" style={{ marginTop: "0.9rem" }}>
                        <button
                          type="button"
                          className="secondary"
                          onClick={() => {
                            gameState.syncState(initialState);
                            gameState.setServerGameID(null);
                            gameState.setServerRevision(null);
                            setShowBoardEditor(false);
                            gameState.setStatus("Board reset.");
                            setShowGuideHelp(true);
                          }}
                        >
                          Reset
                        </button>
                      </div>
                    ) : null}
                  </details>

                  <details className="panel secondary-drawer" open={showGuideHelp} onToggle={(event) => setShowGuideHelp(event.currentTarget.open)}>
                    <summary className="panel-summary">
                      <span className="summary-label">Guide</span>
                      <span className="summary-line">Reference help for the current phase.</span>
                    </summary>
                    <div className="context-drawer-body">
                      <GuideHelpPanel gamePhase={parsedState.gamePhase} onClose={() => setShowGuideHelp(false)} />
                    </div>
                  </details>

                  {!multiplayerToken ? (
                    <details className="panel secondary-drawer" open={showRecoveryTools} onToggle={(event) => setShowRecoveryTools(event.currentTarget.open)}>
                      <summary className="panel-summary">
                        <span className="summary-label">Recovery Tools</span>
                        <span className="summary-line">Manual correction and recovery controls.</span>
                      </summary>
                      {parsedState.gamePhase !== 2 ? (
                        <div style={{ marginTop: "0.9rem" }}>
                          <TurnFlowPanel
                            state={parsedState}
                            onApply={handleApply}
                            onGenerateActions={gameState.refreshActions}
                            onOpenAdvanced={() => setShowAdvancedTurnPanel(true)}
                            onUpdateState={gameState.updateState}
                          />
                        </div>
                      ) : null}
                      <div className="sidebar-actions" style={{ marginTop: "0.9rem" }}>
                        {parsedState.gamePhase !== 2 ? (
                          <button type="button" className="secondary" onClick={() => setShowAdvancedTurnPanel((current) => !current)}>
                            {showAdvancedTurnPanel ? "Hide Advanced Turn" : "Advanced Turn"}
                          </button>
                        ) : null}
                        <button type="button" className="secondary" onClick={() => setActiveModal("json")}>
                          Debug JSON
                        </button>
                      </div>
                      {showAdvancedTurnPanel && parsedState.gamePhase !== 2 ? (
                        <div style={{ marginTop: "0.9rem" }}>
                          <TurnStatePanel
                            state={parsedState}
                            onUpdateState={gameState.updateState}
                            title="Advanced Turn"
                            showCloseButton={false}
                            onClose={() => setShowAdvancedTurnPanel(false)}
                          />
                        </div>
                      ) : null}
                    </details>
                  ) : null}

                  {parsedState.gamePhase === 1 && !multiplayerToken ? (
                    <details className="panel secondary-drawer" open={showBoardEditor} onToggle={(event) => setShowBoardEditor(event.currentTarget.open)}>
                      <summary className="panel-summary">
                        <span className="summary-label">Board Editor</span>
                        <span className="summary-line">
                          {showBoardEditor ? `Editing clearing ${selectedClearing?.id ?? "?"}.` : "Select a clearing on the board, then open its editor here."}
                        </span>
                      </summary>
                      {showBoardEditor ? (
                        <div className="context-drawer-body">
                          <InspectorPanel
                            title="Board Editor"
                            showCloseButton={false}
                            clearing={selectedClearing}
                            keepClearingID={parsedState.marquise.keepClearingID}
                            vagabondClearingID={parsedState.vagabond.clearingID}
                            vagabondInForest={parsedState.vagabond.inForest}
                            onUpdateClearing={gameState.updateClearing}
                            onSetKeepClearing={(clearingID) =>
                              gameState.updateState((draft) => {
                                draft.marquise.keepClearingID = clearingID;
                              })
                            }
                            onSetVagabondClearing={(clearingID, inForest) =>
                              gameState.updateState((draft) => {
                                draft.vagabond.clearingID = clearingID;
                                draft.vagabond.inForest = inForest;
                              })
                            }
                            onClose={() => setShowBoardEditor(false)}
                          />
                        </div>
                      ) : null}
                    </details>
                  ) : null}
                </div>
              </section>
            ) : null}

            {activeModal === "json" ? (
              <section className="panel modal-panel">
                <div className="panel-header">
                  <h2>Debug JSON</h2>
                  <button type="button" className="secondary" onClick={() => setActiveModal(null)}>
                    Close
                  </button>
                </div>
                <p className="message">Use this only for debugging or recovery. Normal play should go through the guided panels.</p>
                <textarea
                  className="state-editor"
                  value={gameState.stateText}
                  onChange={(event) => gameState.setStateText(event.target.value)}
                  spellCheck={false}
                />
              </section>
            ) : null}

            {activeModal === "settings" ? (
              <SettingsPanel
                settings={settings}
                onChange={updateSetting}
                onReset={resetSettings}
                onClose={() => setActiveModal(null)}
              />
            ) : null}

            {activeModal === "standAndDeliver" && pendingStandAndDeliverAction?.usePersistentEffect ? (
              <section className="panel modal-panel">
                <div className="panel-header">
                  <h2>Stand and Deliver</h2>
                  <button
                    type="button"
                    className="secondary"
                    onClick={() => {
                      setActiveModal(null);
                      setPendingStandAndDeliverAction(null);
                      setStandAndDeliverCardID("");
                      gameState.setStatus("Stand and Deliver cancelled.");
                    }}
                  >
                    Cancel
                  </button>
                </div>
                <div className="flow-guide-hero stand-deliver-hero">
                  <span className="summary-label">Assist Observation</span>
                  <strong>Record the stolen card if it was revealed.</strong>
                  <span className="summary-line">
                    Target: {standAndDeliverTargetLabel}. Leave the card blank if the stolen card stayed hidden and record it later through observed tools.
                  </span>
                </div>
                <label className="stand-deliver-field">
                  <span>Observed Card ID</span>
                  <input
                    value={standAndDeliverCardID}
                    placeholder="Optional known card ID"
                    onChange={(event) => setStandAndDeliverCardID(event.target.value)}
                  />
                </label>
                {standAndDeliverCardEntryIsInvalid ? (
                  <span className="message error">Enter a positive integer card ID, or leave the field blank if the stolen card is unknown.</span>
                ) : standAndDeliverCardLabel ? (
                  <span className="summary-line">Known card: {standAndDeliverCardLabel}</span>
                ) : (
                  <span className="summary-line">No card ID entered. The action will record the stolen card as unknown.</span>
                )}
                <div className="sidebar-actions footer">
                  <button
                    type="button"
                    className="secondary"
                    onClick={async () => {
                      const actionToApply: Action = {
                        ...pendingStandAndDeliverAction,
                        usePersistentEffect: {
                          ...pendingStandAndDeliverAction.usePersistentEffect!,
                          observedCardID:
                            standAndDeliverCardID.trim().length > 0 && !standAndDeliverCardEntryIsInvalid
                              ? standAndDeliverParsedCardID
                              : 0
                        }
                      };
                      setActiveModal(null);
                      setPendingStandAndDeliverAction(null);
                      setStandAndDeliverCardID("");
                      await gameState.applyFinalizedAction(actionToApply);
                    }}
                    disabled={standAndDeliverCardEntryIsInvalid}
                  >
                    Apply Stand and Deliver
                  </button>
                </div>
              </section>
            ) : null}
          </div>
        </div>
      ) : null}
    </main>
  );
}
