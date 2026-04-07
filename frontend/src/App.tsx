import { startTransition, useCallback, useDeferredValue, useEffect, useRef, useState } from "react";
import {
  applyAction,
  claimLobbyFaction,
  createLobby,
  fetchBattleSession,
  fetchBattleContext,
  fetchLobbyState,
  fetchValidActions,
  joinLobby,
  leaveLobby,
  loadGame,
  openBattle,
  resolveBattle,
  setLobbyReady,
  startLobby,
  submitBattleResponse
} from "./api";
import {
  assistBoardHighlights,
  buildAssistBoardCandidates,
  movementSourceMatches,
  type AssistActionCandidateRef,
  type AssistMovementSource
} from "./assistDirector";
import { boardLayoutForState } from "./boardLayouts";
import { describeKnownCardID } from "./cardCatalog";
import { AssistWorkflowPanel } from "./components/AssistWorkflowPanel";
import { BoardPanel } from "./components/BoardPanel";
import { BattleFlowPanel } from "./components/BattleFlowPanel";
import { CardVisibilityPanel } from "./components/CardVisibilityPanel";
import { EndgamePanel } from "./components/EndgamePanel";
import { FlowGuidePanel } from "./components/FlowGuidePanel";
import { GuideHelpPanel } from "./components/GuideHelpPanel";
import { InspectorPanel } from "./components/InspectorPanel";
import { JoinScreen } from "./components/JoinScreen";
import { LobbyScreen } from "./components/LobbyScreen";
import { PlayerActionsPanel } from "./components/PlayerActionsPanel";
import { SessionStatusPanel } from "./components/SessionStatusPanel";
import { TurnFlowPanel } from "./components/TurnFlowPanel";
import { SetupWizard } from "./components/SetupWizard";
import { TurnStatePanel } from "./components/TurnStatePanel";
import { TurnSummaryPanel } from "./components/TurnSummaryPanel";
import { affectedClearings, rulerOfClearing, syncDerivedFactionStateFromBoard, usedBuildSlots } from "./gameHelpers";
import { ACTION_TYPE, factionLabels, phaseLabels, setupStageLabels, stepLabels, suitLabels } from "./labels";
import {
  clearSavedMultiplayerSession,
  clearSavedSession,
  loadSavedMultiplayerSession,
  loadSavedSession,
  saveSavedMultiplayerSession,
  saveSavedSession,
  type SavedSession
} from "./localSession";
import { actionHeadline } from "./actionPresentation";
import type { ClearingPreviewPiece } from "./components/ClearingMarker";
import type { MultiplayerConnectionStatus, MultiplayerSession, SetupScreen } from "./multiplayer";
import { sampleState } from "./sampleState";
import type { Action, BattleContext, BattleModifiers, BattlePrompt, Clearing, GameState, Lobby, LobbyPlayer } from "./types";
import { RootBuddyWebSocketClient } from "./wsClient";

type ActiveModal = "json" | "standAndDeliver" | null;
type MultiplayerNotice = {
  level: "warning" | "error";
  title: string;
  detail: string;
} | null;

type MarquiseSetupDraft = {
  keepClearingID: number | null;
  sawmillClearingID: number | null;
  workshopClearingID: number | null;
  recruiterClearingID: number | null;
};

const emptyMarquiseSetupDraft: MarquiseSetupDraft = {
  keepClearingID: null,
  sawmillClearingID: null,
  workshopClearingID: null,
  recruiterClearingID: null
};

const emptyBattleModifiers: BattleModifiers = {
  attackerHitModifier: 0,
  defenderHitModifier: 0,
  ignoreHitsToAttacker: false,
  ignoreHitsToDefender: false,
  defenderAmbush: false,
  attackerCounterAmbush: false,
  attackerUsesArmorers: false,
  defenderUsesArmorers: false,
  attackerUsesBrutalTactics: false,
  defenderUsesSappers: false
};

function stringifyState(nextState: GameState): string {
  return JSON.stringify(nextState, null, 2);
}

function hasWarriors(clearing: Clearing): boolean {
  return Object.values(clearing.warriors).some((count) => count > 0);
}

function isBoardEmpty(state: GameState): boolean {
  return state.map.clearings.every(
    (clearing) => clearing.wood === 0 && !hasWarriors(clearing) && clearing.buildings.length === 0
  );
}

function normalizeState(nextState: GameState): GameState {
  const normalized = structuredClone(nextState);
  normalized.randomSeed ??= 0;
  normalized.shuffleCount ??= 0;
  normalized.setupStage ??= 0;
  normalized.map.clearings ??= [];
  normalized.map.forests ??= [];
  normalized.winningCoalition ??= [];
  normalized.turnOrder ??= [];
  normalized.victoryPoints ??= {};
  normalized.deck ??= [];
  normalized.discardPile ??= [];
  normalized.availableDominance ??= [];
  normalized.activeDominance ??= {};
  normalized.coalitionActive ??= false;
  normalized.coalitionPartner ??= 0;
  normalized.itemSupply ??= {};
  normalized.persistentEffects ??= {};
  normalized.questDeck ??= [];
  normalized.questDiscard ??= [];
  normalized.otherHandCounts ??= {};
  normalized.hiddenCards ??= [];
  normalized.nextHiddenCardID ??= 1;
  normalized.marquise.cardsInHand ??= [];
  normalized.eyrie.cardsInHand ??= [];
  normalized.eyrie.availableLeaders ??= [];
  normalized.eyrie.decree.recruit ??= [];
  normalized.eyrie.decree.move ??= [];
  normalized.eyrie.decree.battle ??= [];
  normalized.eyrie.decree.build ??= [];
  normalized.alliance.cardsInHand ??= [];
  normalized.alliance.supporters ??= [];
  normalized.vagabond.cardsInHand ??= [];
  normalized.vagabond.items ??= [];
  normalized.vagabond.relationships ??= {};
  normalized.vagabond.forestID ??= 0;
  normalized.vagabond.questsAvailable ??= [];
  normalized.vagabond.questsCompleted ??= [];
  normalized.turnProgress.usedWorkshopClearings ??= [];
  normalized.turnProgress.resolvedDecreeCardIDs ??= [];
  normalized.turnProgress.usedPersistentEffectIDs ??= [];
  normalized.turnProgress.birdsongMainActionTaken ??= false;
  normalized.turnProgress.daylightMainActionTaken ??= false;
  normalized.turnProgress.eveningMainActionTaken ??= false;
  for (const clearing of normalized.map.clearings) {
    clearing.adj ??= [];
    clearing.buildings ??= [];
    clearing.tokens ??= [];
    clearing.warriors ??= {};
    clearing.ruinItems ??= [];
  }
  syncDerivedFactionStateFromBoard(normalized);
  return normalized;
}

const initialState = normalizeState(sampleState);
const initialJSON = JSON.stringify(initialState, null, 2);

function zeroActionHint(state: GameState): string {
  if (state.gamePhase === 0) {
    return `${setupStageLabels[state.setupStage] ?? "Setup"} has no legal actions.`;
  }

  if (state.currentPhase === 0) {
    return "No legal birdsong actions found for the current faction state.";
  }

  if (state.currentPhase === 1) {
    return "No legal daylight actions found. Check faction-specific requirements like ruling, decree state, supporters, items, wood, and cards in hand.";
  }

  if (state.currentPhase === 2) {
    return "No legal evening actions found for the current faction state.";
  }

  return "No legal actions found for this state. Check the selected faction, phase, step, and faction-specific resources.";
}

function marquiseSetupMatches(
  action: Action,
  draft: MarquiseSetupDraft
): boolean {
  const payload = action.marquiseSetup;
  if (!payload) {
    return false;
  }

  if (draft.keepClearingID !== null && payload.keepClearingID !== draft.keepClearingID) {
    return false;
  }
  if (draft.sawmillClearingID !== null && payload.sawmillClearingID !== draft.sawmillClearingID) {
    return false;
  }
  if (draft.workshopClearingID !== null && payload.workshopClearingID !== draft.workshopClearingID) {
    return false;
  }
  if (draft.recruiterClearingID !== undefined && draft.recruiterClearingID !== null && payload.recruiterClearingID !== draft.recruiterClearingID) {
    return false;
  }

  return true;
}

function gameOverHeadline(state: GameState): string {
  if (state.winningCoalition.length > 0) {
    return `${state.winningCoalition.map((faction) => factionLabels[faction] ?? `Faction ${faction}`).join(" + ")} win`;
  }

  return `${factionLabels[state.winner] ?? "Unknown"} win`;
}

function gameOverStatusMessage(state: GameState): string {
  if (state.winningCoalition.length > 0) {
    return `Game over. Reviewing the coalition victory for ${state.winningCoalition
      .map((faction) => factionLabels[faction] ?? `Faction ${faction}`)
      .join(" + ")}.`;
  }

  return `Game over. Reviewing the final result for ${factionLabels[state.winner] ?? "Unknown"}.`;
}

function setupBoardPrompt(stage: number, draft: MarquiseSetupDraft): { title: string; instruction: string; detail: string } {
  switch (stage) {
    case 1:
      if (draft.keepClearingID === null) {
        return {
          title: "Marquise Setup",
          instruction: "Choose the Keep corner",
          detail: "Click one of the highlighted corner clearings."
        };
      }
      if (draft.sawmillClearingID === null) {
        return {
          title: "Marquise Setup",
          instruction: "Place the sawmill",
          detail: "Click a highlighted legal clearing. The pending building appears immediately."
        };
      }
      if (draft.workshopClearingID === null) {
        return {
          title: "Marquise Setup",
          instruction: "Place the workshop",
          detail: "Click a highlighted legal clearing. The pending building appears immediately."
        };
      }
      return {
        title: "Marquise Setup",
        instruction: "Place the recruiter",
        detail: "Click a highlighted legal clearing to finish Marquise setup."
      };
    case 2:
      return {
        title: "Eyrie Setup",
        instruction: "Choose the starting roost",
        detail: "Click one of the highlighted corner clearings."
      };
    case 3:
      return {
        title: "Vagabond Setup",
        instruction: "Choose the starting forest",
        detail: "Click one of the highlighted forest markers."
      };
    default:
      return {
        title: "Setup",
        instruction: "Choose a highlighted setup target",
        detail: "The board highlights the legal setup choices."
      };
  }
}

function factionHandSize(state: GameState, faction: number): number {
  switch (faction) {
    case 0:
      return state.marquise.cardsInHand.length > 0 ? state.marquise.cardsInHand.length : state.otherHandCounts[String(faction)] ?? 0;
    case 1:
      return state.alliance.cardsInHand.length > 0 ? state.alliance.cardsInHand.length : state.otherHandCounts[String(faction)] ?? 0;
    case 2:
      return state.eyrie.cardsInHand.length > 0 ? state.eyrie.cardsInHand.length : state.otherHandCounts[String(faction)] ?? 0;
    case 3:
      return state.vagabond.cardsInHand.length > 0 ? state.vagabond.cardsInHand.length : state.otherHandCounts[String(faction)] ?? 0;
    default:
      return 0;
  }
}

function sameCandidateRefs(current: AssistActionCandidateRef[], next: AssistActionCandidateRef[]) {
  return (
    current.length === next.length &&
    current.every((candidate, index) => candidate.actionIndex === next[index].actionIndex && candidate.action === next[index].action)
  );
}

export default function App() {
  const initialSavedSession = loadSavedSession();
  const initialSavedMultiplayerSession = loadSavedMultiplayerSession();
  const [showSetup, setShowSetup] = useState(true);
  const [setupScreen, setSetupScreen] = useState<SetupScreen>("wizard");
  const [hasSavedSession, setHasSavedSession] = useState(() => initialSavedSession !== null);
  const [savedSessionInfo, setSavedSessionInfo] = useState<SavedSession | null>(initialSavedSession);
  const [multiplayerSession, setMultiplayerSession] = useState<MultiplayerSession | null>(
    initialSavedMultiplayerSession
      ? {
          playerToken: initialSavedMultiplayerSession.playerToken,
          displayName: initialSavedMultiplayerSession.displayName,
          joinCode: initialSavedMultiplayerSession.joinCode,
          gameID: initialSavedMultiplayerSession.gameID
        }
      : null
  );
  const [multiplayerLobby, setMultiplayerLobby] = useState<Lobby | null>(null);
  const [multiplayerSelf, setMultiplayerSelf] = useState<LobbyPlayer | null>(null);
  const [multiplayerConnectionStatus, setMultiplayerConnectionStatus] =
    useState<MultiplayerConnectionStatus>("disconnected");
  const [multiplayerSubmitting, setMultiplayerSubmitting] = useState(false);
  const [serverGameID, setServerGameID] = useState<string | null>(null);
  const [serverRevision, setServerRevision] = useState<number | null>(initialSavedSession?.revision ?? null);
  const [stateText, setStateText] = useState(initialJSON);
  const deferredStateText = useDeferredValue(stateText);
  const [parsedState, setParsedState] = useState<GameState>(initialState);
  const currentStateRef = useRef<GameState>(initialState);
  const [selectedClearingID, setSelectedClearingID] = useState<number>(
    initialState.map.clearings[0]?.id ?? 0
  );
  const [actions, setActions] = useState<Action[]>([]);
  const [selectedBattleIndex, setSelectedBattleIndex] = useState<number | null>(null);
  const [hoveredActionIndex, setHoveredActionIndex] = useState<number | null>(null);
  const [assistBattleCandidateRefs, setAssistBattleCandidateRefs] = useState<AssistActionCandidateRef[]>([]);
  const [assistMovementCandidateRefs, setAssistMovementCandidateRefs] = useState<AssistActionCandidateRef[]>([]);
  const [assistMovementSource, setAssistMovementSource] = useState<AssistMovementSource | null>(null);
  const [playerMovementCandidateRefs, setPlayerMovementCandidateRefs] = useState<AssistActionCandidateRef[]>([]);
  const [playerMovementSource, setPlayerMovementSource] = useState<AssistMovementSource | null>(null);
  const [assistBuildRecruitCandidateRefs, setAssistBuildRecruitCandidateRefs] = useState<AssistActionCandidateRef[]>([]);
  const [playerBuildRecruitCandidateRefs, setPlayerBuildRecruitCandidateRefs] = useState<AssistActionCandidateRef[]>([]);
  const [assistFactionSpatialCandidateRefs, setAssistFactionSpatialCandidateRefs] = useState<AssistActionCandidateRef[]>([]);
  const [playerFactionSpatialCandidateRefs, setPlayerFactionSpatialCandidateRefs] = useState<AssistActionCandidateRef[]>([]);
  const [attackerRoll, setAttackerRoll] = useState("1");
  const [defenderRoll, setDefenderRoll] = useState("0");
  const [battleModifiers, setBattleModifiers] = useState<BattleModifiers>(emptyBattleModifiers);
  const [battleContext, setBattleContext] = useState<BattleContext | null>(null);
  const [multiplayerBattlePrompt, setMultiplayerBattlePrompt] = useState<BattlePrompt | null>(null);
  const [multiplayerNotice, setMultiplayerNotice] = useState<MultiplayerNotice>(null);
  const [assistDefenderAmbushChoice, setAssistDefenderAmbushChoice] = useState<boolean | null>(null);
  const [error, setError] = useState<string>("");
  const [status, setStatus] = useState<string>("Click a clearing to start setting the board.");
  const [activeModal, setActiveModal] = useState<ActiveModal>(null);
  const [pendingStandAndDeliverAction, setPendingStandAndDeliverAction] = useState<Action | null>(null);
  const [standAndDeliverCardID, setStandAndDeliverCardID] = useState("");
  const [showGuideHelp, setShowGuideHelp] = useState(true);
  const [showAdvancedTurnPanel, setShowAdvancedTurnPanel] = useState(false);
  const [showBoardEditor, setShowBoardEditor] = useState(false);
  const [showContextDrawer, setShowContextDrawer] = useState(false);
  const [showWorkspaceTools, setShowWorkspaceTools] = useState(false);
  const [showRecoveryTools, setShowRecoveryTools] = useState(false);
  const [marquiseSetupDraft, setMarquiseSetupDraft] = useState<MarquiseSetupDraft>(emptyMarquiseSetupDraft);
  const [eyrieSetupDraftClearingID, setEyrieSetupDraftClearingID] = useState<number | null>(null);
  const [vagabondSetupDraftForestID, setVagabondSetupDraftForestID] = useState<number | null>(null);
  const multiplayerToken = multiplayerSession?.playerToken ?? null;
  const previousConnectionStatus = useRef<MultiplayerConnectionStatus>("disconnected");
  const autoLoadedActionKey = useRef<string>("");
  const previousGamePhase = useRef(initialState.gamePhase);

  useEffect(() => {
    try {
      const nextState = JSON.parse(deferredStateText) as GameState;
      currentStateRef.current = nextState;
      setParsedState(nextState);
      setError("");
    } catch (err) {
      const message = err instanceof Error ? err.message : "Invalid JSON";
      setError(message);
    }
  }, [deferredStateText]);

  useEffect(() => {
    if (parsedState.map.clearings.some((clearing) => clearing.id === selectedClearingID)) {
      return;
    }
    setSelectedClearingID(parsedState.map.clearings[0]?.id ?? 0);
  }, [parsedState, selectedClearingID]);

  useEffect(() => {
    setMarquiseSetupDraft(emptyMarquiseSetupDraft);
    setEyrieSetupDraftClearingID(null);
    setVagabondSetupDraftForestID(null);
  }, [parsedState.gamePhase, parsedState.setupStage]);

  useEffect(() => {
    if (showSetup) {
      return;
    }
    if (multiplayerToken) {
      return;
    }

    const session: SavedSession = {
      state: parsedState,
      gameID: serverGameID,
      revision: serverRevision,
      savedAt: new Date().toISOString()
    };
    const saved = saveSavedSession({
      state: session.state,
      gameID: session.gameID,
      revision: session.revision,
      savedAt: session.savedAt
    });
    if (saved) {
      setHasSavedSession(true);
      setSavedSessionInfo(session);
    }
  }, [multiplayerToken, parsedState, serverGameID, serverRevision, showSetup]);

  useEffect(() => {
    if (!multiplayerSession) {
      clearSavedMultiplayerSession();
      return;
    }

    saveSavedMultiplayerSession({
      ...multiplayerSession,
      savedAt: new Date().toISOString()
    });
  }, [multiplayerSession]);

  useEffect(() => {
    const savedMultiplayerSession = initialSavedMultiplayerSession;
    if (!savedMultiplayerSession) {
      return;
    }
    const savedPlayerToken = savedMultiplayerSession.playerToken;

    let cancelled = false;

    async function resumeMultiplayerSession() {
      setMultiplayerSubmitting(true);
      setStatus("Rejoining multiplayer session...");
      try {
        const { lobby, self } = await fetchLobbyState(savedPlayerToken);
        if (cancelled) {
          return;
        }

        setMultiplayerLobby(lobby);
        setMultiplayerSelf(self);
        setMultiplayerSession((current) =>
          current
            ? {
                ...current,
                joinCode: lobby.joinCode,
                gameID: lobby.gameID ?? current.gameID
              }
            : current
        );
        setSetupScreen("wizard");

        if (lobby.gameID) {
          const loaded = await loadGame(lobby.gameID, savedPlayerToken);
          if (cancelled) {
            return;
          }

          setMultiplayerSession((current) =>
            current
              ? {
                  ...current,
                  joinCode: lobby.joinCode,
                  gameID: loaded.gameID
                }
              : current
          );
          setServerGameID(loaded.gameID);
          setServerRevision(loaded.revision);
          syncState(loaded.state);
          setShowSetup(false);
          setShowBoardEditor(false);
          setShowGuideHelp(false);
          setStatus(loaded.state.gamePhase === 2 ? "Rejoined finished multiplayer game." : "Rejoined multiplayer game.");
          return;
        }

        setShowSetup(true);
        setStatus(`Rejoined lobby ${lobby.joinCode}.`);
      } catch (err) {
        if (cancelled) {
          return;
        }
        setMultiplayerSession(null);
        setMultiplayerLobby(null);
        setMultiplayerSelf(null);
        setStatus(err instanceof Error ? err.message : "Saved multiplayer session expired.");
      } finally {
        if (!cancelled) {
          setMultiplayerSubmitting(false);
        }
      }
    }

    void resumeMultiplayerSession();

    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    if (!multiplayerToken) {
      setMultiplayerConnectionStatus("disconnected");
      return;
    }

    const client = new RootBuddyWebSocketClient({
      token: multiplayerToken,
      onConnectionChange: (nextStatus) => {
        setMultiplayerConnectionStatus(nextStatus);
      },
      onMessage: (message) => {
        if (message.type === "lobby.update") {
          setMultiplayerNotice(null);
          setMultiplayerLobby(message.lobby);
          setMultiplayerSelf(message.self);
          setShowSetup(true);
          setMultiplayerSession((current) =>
            current
              ? {
                  ...current,
                  joinCode: message.lobby.joinCode,
                  gameID: message.lobby.gameID ?? current.gameID
                }
              : current
          );
          return;
        }

        if (message.type === "game.start" || message.type === "game.state" || message.type === "conflict") {
          setMultiplayerBattlePrompt(null);
          setMultiplayerLobby((current) =>
            current
              ? {
                  ...current,
                  gameID: message.gameID,
                  state: 1
                }
              : current
          );
          setMultiplayerSession((current) =>
            current
              ? {
                  ...current,
                  gameID: message.gameID
                }
              : current
          );
          setServerGameID(message.gameID);
          setServerRevision(message.revision);
          syncState(message.state);
          setShowSetup(false);
          setShowBoardEditor(false);
          setShowGuideHelp(false);
          if (message.type === "conflict") {
            setMultiplayerNotice({
              level: "warning",
              title: "Server State Updated",
              detail: message.error
            });
            setStatus(message.error);
          } else {
            setMultiplayerNotice(null);
            if (message.state.gamePhase === 2) {
              setStatus(gameOverStatusMessage(message.state));
            } else if (message.state.gamePhase === 1 && message.state.factionTurn === message.state.playerFaction) {
              setStatus("Your turn. Loading legal actions...");
            } else if (message.state.gamePhase === 1) {
              setStatus(`Waiting on ${factionLabels[message.state.factionTurn] ?? "another player"}.`);
            } else {
              setStatus(message.type === "game.start" ? "Multiplayer game started." : "Received multiplayer update.");
            }
          }
          return;
        }

        if (message.type === "battle.prompt") {
          setMultiplayerNotice(null);
          setMultiplayerBattlePrompt(message.prompt ?? null);
          if (!message.prompt) {
            return;
          }

          if (message.prompt.stage === "ready_to_resolve") {
            setStatus("Battle choices locked in. Resolve when ready.");
            return;
          }

          if (message.prompt.waitingOnFaction === parsedState.playerFaction) {
            setStatus("Battle response needed.");
          } else {
            setStatus(`Waiting on ${factionLabels[message.prompt.waitingOnFaction] ?? "another player"} for battle response.`);
          }
          return;
        }

        if (message.type === "session.error") {
          setMultiplayerNotice({
            level: "error",
            title: "Session Error",
            detail: message.error
          });
          setStatus(message.error);
        }
      }
    });

    client.connect();

    return () => {
      client.disconnect();
    };
  }, [multiplayerToken, parsedState.playerFaction]);

  useEffect(() => {
    const previous = previousConnectionStatus.current;
    previousConnectionStatus.current = multiplayerConnectionStatus;

    if (!multiplayerToken) {
      return;
    }
    if (multiplayerConnectionStatus === previous) {
      return;
    }
    if (multiplayerConnectionStatus === "reconnecting") {
      setMultiplayerNotice({
        level: "warning",
        title: "Reconnecting",
        detail: "Realtime connection lost. Waiting for the websocket session to recover."
      });
      setStatus("Realtime connection lost. Reconnecting...");
      return;
    }
    if (multiplayerConnectionStatus === "connected" && (previous === "connecting" || previous === "reconnecting")) {
      setMultiplayerNotice(null);
      setStatus("Realtime multiplayer connection active.");
    }
  }, [multiplayerConnectionStatus, multiplayerToken]);

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
    if (!multiplayerToken || !serverGameID || showSetup) {
      return;
    }
    const gameID = serverGameID;

    let cancelled = false;

    async function loadActiveBattleSession() {
      try {
        const session = await fetchBattleSession(gameID, multiplayerToken);
        if (cancelled) {
          return;
        }
        if (session.revision !== null) {
          setServerRevision(session.revision);
        }
        setMultiplayerBattlePrompt(session.prompt);
      } catch {
        if (!cancelled) {
          setMultiplayerBattlePrompt(null);
        }
      }
    }

    void loadActiveBattleSession();

    return () => {
      cancelled = true;
    };
  }, [multiplayerToken, serverGameID, showSetup]);

  useEffect(() => {
    if (!multiplayerToken || showSetup || parsedState.gamePhase !== 1) {
      autoLoadedActionKey.current = "";
      return;
    }
    if (multiplayerBattlePrompt || parsedState.factionTurn !== parsedState.playerFaction) {
      autoLoadedActionKey.current = "";
      return;
    }

    const loadKey = [
      serverGameID ?? "",
      serverRevision ?? "na",
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

    async function loadTurnActions() {
      try {
        setStatus("Your turn. Loading legal actions...");
        await loadActionsForState(parsedState);
      } catch (err) {
        if (cancelled) {
          return;
        }
        autoLoadedActionKey.current = "";
        setStatus(err instanceof Error ? err.message : "Failed to refresh legal actions");
      }
    }

    void loadTurnActions();

    return () => {
      cancelled = true;
    };
  }, [
    multiplayerBattlePrompt,
    multiplayerToken,
    parsedState,
    serverGameID,
    serverRevision,
    showSetup
  ]);

  function resetToSetup(options?: { clearSaved?: boolean; status?: string }) {
    if (options?.clearSaved) {
      clearSavedSession();
      setHasSavedSession(false);
      setSavedSessionInfo(null);
    }
    setServerGameID(null);
    setServerRevision(null);
    setShowSetup(true);
    setActiveModal(null);
    setShowAdvancedTurnPanel(false);
    setShowBoardEditor(false);
    setShowContextDrawer(false);
    setShowWorkspaceTools(false);
    setShowRecoveryTools(false);
    setShowGuideHelp(false);
    setStatus(options?.status ?? "Choose factions and create a new game.");
  }

  function clearMultiplayerState() {
    setMultiplayerSession(null);
    setMultiplayerLobby(null);
    setMultiplayerSelf(null);
    setMultiplayerBattlePrompt(null);
    setMultiplayerNotice(null);
    setMultiplayerConnectionStatus("disconnected");
    setSetupScreen("wizard");
  }

  function enterLoadedGame(nextState: GameState, gameID: string | null, revision: number | null, nextStatus: string) {
    syncState(nextState);
    setServerGameID(gameID);
    setServerRevision(revision);
    setMultiplayerBattlePrompt(null);
    setMultiplayerNotice(null);
    setShowSetup(false);
    setShowBoardEditor(false);
    setShowContextDrawer(nextState.gamePhase === 2);
    setShowWorkspaceTools(nextState.gamePhase !== 1);
    setShowRecoveryTools(false);
    setActiveModal(null);
    setShowGuideHelp(false);
    setStatus(nextStatus);
  }

  function syncState(nextState: GameState) {
    const normalizedState = normalizeState(nextState);
    currentStateRef.current = normalizedState;
    startTransition(() => {
      setParsedState(normalizedState);
      setStateText(stringifyState(normalizedState));
      setActions([]);
      setSelectedBattleIndex(null);
      setHoveredActionIndex(null);
      setBattleModifiers(emptyBattleModifiers);
      setBattleContext(null);
      setAssistDefenderAmbushChoice(null);
      setError("");
    });
  }

  async function loadActionsForState(baseState: GameState, options?: { successStatus?: string }) {
    const requestState = normalizeState(baseState);
    currentStateRef.current = requestState;
    const { actions: nextActions, revision } = await fetchValidActions(requestState, serverGameID, multiplayerToken);
    if (revision !== null) {
      setServerRevision(revision);
    }

    startTransition(() => {
      setParsedState(requestState);
      setStateText(stringifyState(requestState));
      setActions(nextActions);
      setSelectedBattleIndex(null);
      setHoveredActionIndex(null);
      setBattleModifiers(emptyBattleModifiers);
      setBattleContext(null);
      setAssistDefenderAmbushChoice(null);
      setError("");
    });

    setStatus(
      options?.successStatus ??
        (nextActions.length > 0 ? `Loaded ${nextActions.length} action(s).` : zeroActionHint(requestState))
    );

    return nextActions;
  }

  function updateState(mutator: (draft: GameState) => void) {
    const nextState = structuredClone(parsedState);
    mutator(nextState);
    syncState(nextState);
  }

  function updateClearing(clearingID: number, mutator: (clearing: Clearing) => void) {
    updateState((draft) => {
      const clearing = draft.map.clearings.find((item) => item.id === clearingID);
      if (!clearing) {
        return;
      }
      mutator(clearing);
    });
  }

  async function refreshActions() {
    if (error) {
      setStatus("Fix the JSON before requesting actions.");
      return;
    }

    if (boardIsEmpty && parsedState.gamePhase !== 0) {
      setStatus("Enter the current board state first.");
      setShowGuideHelp(true);
      return;
    }

    try {
      setStatus("Fetching valid actions...");
      await loadActionsForState(currentStateRef.current);
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to fetch actions";
      setStatus(message);
    }
  }

  function needsStandAndDeliverObservation(action: Action): boolean {
    const currentState = currentStateRef.current;
    return (
      currentState.gameMode === 1 &&
      !multiplayerToken &&
      action.usePersistentEffect?.effectID === "stand_and_deliver" &&
      action.usePersistentEffect.targetFaction !== action.usePersistentEffect.faction &&
      factionHandSize(currentState, action.usePersistentEffect.targetFaction) > 0 &&
      (action.usePersistentEffect.observedCardID ?? 0) <= 0
    );
  }

  function isImpossibleStandAndDeliver(action: Action): boolean {
    const currentState = currentStateRef.current;
    return (
      action.usePersistentEffect?.effectID === "stand_and_deliver" &&
      action.usePersistentEffect.targetFaction !== action.usePersistentEffect.faction &&
      factionHandSize(currentState, action.usePersistentEffect.targetFaction) === 0
    );
  }

  async function applyFinalizedAction(actionToApply: Action) {
    try {
      const currentState = currentStateRef.current;
      setStatus("Applying action...");
      const { state: nextState, effectResult, revision } = await applyAction(
        currentState,
        actionToApply,
        serverGameID,
        serverRevision,
        multiplayerToken
      );
      if (revision !== null) {
        setServerRevision(revision);
      }
      if (nextState.gamePhase === 0) {
        await loadActionsForState(nextState, { successStatus: "Setup step applied." });
        setActiveModal(null);
        return;
      }

      syncState(nextState);
      setStatus(nextState.gamePhase === 2 ? gameOverStatusMessage(nextState) : effectResult?.message ?? "Action applied.");
      setActiveModal(null);
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to apply action";
      setStatus(message);
    }
  }

  async function handleApply(action: Action) {
    if (isImpossibleStandAndDeliver(action)) {
      setStatus("Stand and Deliver cannot target a faction with no recorded cards.");
      return;
    }

    if (needsStandAndDeliverObservation(action)) {
      setPendingStandAndDeliverAction(action);
      setStandAndDeliverCardID("");
      setActiveModal("standAndDeliver");
      setStatus("Confirm the Stand and Deliver observation before applying.");
      return;
    }

    await applyFinalizedAction(action);
  }

  async function handleResolveAndApply() {
    const battleAction = multiplayerBattlePrompt?.action ?? (selectedBattleIndex !== null ? actions[selectedBattleIndex] : null);
    if (!battleAction) {
      setStatus("Select a battle action first.");
      return;
    }

    const action = battleAction;
    if (action.type !== ACTION_TYPE.BATTLE) {
      setStatus("Selected action is not a battle.");
      return;
    }
    if (multiplayerToken && multiplayerBattlePrompt && multiplayerBattlePrompt.stage !== "ready_to_resolve") {
      if (multiplayerBattlePrompt.waitingOnFaction === parsedState.playerFaction) {
        setStatus("Respond to the current battle prompt before resolving.");
      } else {
        setStatus(`Waiting on ${factionLabels[multiplayerBattlePrompt.waitingOnFaction] ?? "another player"} before resolving.`);
      }
      return;
    }
    if (assistDefenderAmbushPromptRequired && assistDefenderAmbushChoice === null) {
      setStatus("Answer whether the defender used an ambush before resolving the battle.");
      return;
    }

    try {
      setStatus("Resolving battle...");
      const resolved = await resolveBattle(
        parsedState,
        action,
        multiplayerToken ? 0 : Number(attackerRoll),
        multiplayerToken ? 0 : Number(defenderRoll),
        multiplayerToken
          ? undefined
          : {
              ...battleModifiers,
              defenderAmbush: assistDefenderAmbushPromptRequired
                ? assistDefenderAmbushChoice === true
                : battleModifiers.defenderAmbush
            },
        serverGameID,
        multiplayerToken
      );
      if (resolved.revision !== null) {
        setServerRevision(resolved.revision);
      }
      const requestRevision = resolved.revision ?? serverRevision;
      const { state: nextState, effectResult, revision } = await applyAction(
        parsedState,
        resolved.action,
        serverGameID,
        requestRevision,
        multiplayerToken
      );
      if (revision !== null) {
        setServerRevision(revision);
      }
      setMultiplayerBattlePrompt(null);
      syncState(nextState);
      if (!multiplayerToken && nextState.gamePhase === 1) {
        await loadActionsForState(nextState, {
          successStatus: effectResult?.message ?? "Battle resolved. Loaded the next legal actions."
        });
        setActiveModal(null);
        return;
      }
      setStatus(nextState.gamePhase === 2 ? gameOverStatusMessage(nextState) : effectResult?.message ?? "Battle resolved and applied.");
      setActiveModal(null);
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to resolve battle";
      setStatus(message);
    }
  }

  async function handleSubmitBattleResponse() {
    if (!multiplayerToken || !multiplayerBattlePrompt?.gameID) {
      setStatus("No multiplayer battle prompt is active.");
      return;
    }

    const request = {
      gameID: multiplayerBattlePrompt.gameID,
      useAmbush: multiplayerBattlePrompt.stage === "defender_response" ? battleModifiers.defenderAmbush : undefined,
      useDefenderArmorers:
        multiplayerBattlePrompt.stage === "defender_response" ? battleModifiers.defenderUsesArmorers : undefined,
      useSappers: multiplayerBattlePrompt.stage === "defender_response" ? battleModifiers.defenderUsesSappers : undefined,
      useCounterAmbush:
        multiplayerBattlePrompt.stage === "attacker_response" ? battleModifiers.attackerCounterAmbush : undefined,
      useAttackerArmorers:
        multiplayerBattlePrompt.stage === "attacker_response" ? battleModifiers.attackerUsesArmorers : undefined,
      useBrutalTactics:
        multiplayerBattlePrompt.stage === "attacker_response" ? battleModifiers.attackerUsesBrutalTactics : undefined
    };

    try {
      setMultiplayerSubmitting(true);
      setStatus("Submitting battle response...");
      const response = await submitBattleResponse(request, multiplayerToken);
      if (response.revision !== null) {
        setServerRevision(response.revision);
      }
      setMultiplayerBattlePrompt(response.prompt);
      if (response.prompt?.stage === "ready_to_resolve") {
        setStatus("Battle choices locked in. Resolve when ready.");
      } else {
        setStatus("Battle response submitted.");
      }
    } catch (err) {
      setStatus(err instanceof Error ? err.message : "Failed to submit battle response");
    } finally {
      setMultiplayerSubmitting(false);
    }
  }

  function openBattleForActionIndex(actionIndex: number) {
    setSelectedBattleIndex(actionIndex);
    setHoveredActionIndex(actionIndex);
    setBattleModifiers(emptyBattleModifiers);
    setAssistDefenderAmbushChoice(null);
    setActiveModal(null);
    const action = actions[actionIndex];
    if (multiplayerToken && serverGameID && action?.type === ACTION_TYPE.BATTLE) {
      void (async () => {
        try {
          setMultiplayerSubmitting(true);
          setStatus("Opening multiplayer battle flow...");
          const response = await openBattle(parsedState, action, serverGameID, multiplayerToken);
          if (response.revision !== null) {
            setServerRevision(response.revision);
          }
          setMultiplayerBattlePrompt(response.prompt);
          if (response.prompt?.stage === "ready_to_resolve") {
            setStatus("Battle is ready to resolve.");
            return;
          }
          setStatus("Battle selected. Follow the multiplayer response prompt in Battle Flow.");
        } catch (err) {
          setStatus(err instanceof Error ? err.message : "Failed to open multiplayer battle flow");
        } finally {
          setMultiplayerSubmitting(false);
        }
      })();
      return;
    }

    setStatus("Battle selected. Resolve it from Battle Flow.");
  }

  async function handleResumeSavedGame() {
    const savedSession = loadSavedSession();
    if (!savedSession) {
      throw new Error("No saved game found.");
    }

    const loaded = savedSession.gameID
      ? await loadGame(savedSession.gameID, multiplayerToken)
      : { state: savedSession.state, gameID: null, revision: savedSession.revision };
    setSavedSessionInfo({
      state: loaded.state,
      gameID: loaded.gameID,
      revision: loaded.revision,
      savedAt: savedSession.savedAt
    });
    syncState(loaded.state);
    setServerGameID(loaded.gameID);
    setServerRevision(loaded.revision);
    setShowSetup(false);
    setShowBoardEditor(false);
    setShowGuideHelp(loaded.state.gamePhase === 1);
    setStatus(
      loaded.state.gamePhase === 2
        ? "Reviewing saved final result."
        : loaded.state.gamePhase === 0
          ? "Resumed setup."
          : "Resumed saved game."
    );

    if (loaded.state.gamePhase === 0) {
      try {
        await loadActionsForState(loaded.state, { successStatus: "Resumed setup. Choose a highlighted setup target." });
      } catch (err) {
        const message = err instanceof Error ? err.message : "Failed to load resumed setup actions";
        setStatus(message);
      }
    }
  }

  async function handleCreateLobby(request: {
    displayName: string;
    factions: number[];
    eyrieLeader: number;
    vagabondCharacter: number;
  }) {
    try {
      setMultiplayerSubmitting(true);
      setStatus("Creating multiplayer lobby...");
      const result = await createLobby(request);
      setMultiplayerSession({
        playerToken: result.playerToken,
        displayName: request.displayName,
        joinCode: result.lobby.joinCode,
        gameID: result.lobby.gameID ?? null
      });
      setMultiplayerLobby(result.lobby);
      setMultiplayerSelf(result.self);
      setServerGameID(null);
      setServerRevision(null);
      setShowSetup(true);
      setSetupScreen("wizard");
      setStatus(`Lobby ${result.lobby.joinCode} created.`);
    } catch (err) {
      setStatus(err instanceof Error ? err.message : "Failed to create lobby");
    } finally {
      setMultiplayerSubmitting(false);
    }
  }

  async function handleJoinLobby(request: { displayName: string; joinCode: string }) {
    try {
      setMultiplayerSubmitting(true);
      setStatus(`Joining lobby ${request.joinCode}...`);
      const result = await joinLobby(request);
      setMultiplayerSession({
        playerToken: result.playerToken,
        displayName: request.displayName,
        joinCode: result.lobby.joinCode,
        gameID: result.lobby.gameID ?? null
      });
      setMultiplayerLobby(result.lobby);
      setMultiplayerSelf(result.self);
      setServerGameID(result.lobby.gameID ?? null);
      setShowSetup(true);
      setSetupScreen("wizard");
      setStatus(`Joined lobby ${result.lobby.joinCode}.`);
    } catch (err) {
      setStatus(err instanceof Error ? err.message : "Failed to join lobby");
    } finally {
      setMultiplayerSubmitting(false);
    }
  }

  async function handleClaimLobby(nextFaction: number | null) {
    if (!multiplayerToken) {
      setStatus("No multiplayer session is active.");
      return;
    }

    try {
      setMultiplayerSubmitting(true);
      const result = await claimLobbyFaction(multiplayerToken, nextFaction);
      setMultiplayerLobby(result.lobby);
      setMultiplayerSelf(result.self);
      setStatus(nextFaction === null ? "Released faction claim." : "Faction claim updated.");
    } catch (err) {
      setStatus(err instanceof Error ? err.message : "Failed to claim faction");
    } finally {
      setMultiplayerSubmitting(false);
    }
  }

  async function handleSetLobbyReady(isReady: boolean) {
    if (!multiplayerToken) {
      setStatus("No multiplayer session is active.");
      return;
    }

    try {
      setMultiplayerSubmitting(true);
      const result = await setLobbyReady(multiplayerToken, isReady);
      setMultiplayerLobby(result.lobby);
      setMultiplayerSelf(result.self);
      setStatus(isReady ? "Marked ready." : "Marked not ready.");
    } catch (err) {
      setStatus(err instanceof Error ? err.message : "Failed to update ready state");
    } finally {
      setMultiplayerSubmitting(false);
    }
  }

  async function handleStartLobby() {
    if (!multiplayerToken) {
      setStatus("No multiplayer session is active.");
      return;
    }

    try {
      setMultiplayerSubmitting(true);
      setStatus("Starting multiplayer game...");
      const result = await startLobby(multiplayerToken);
      setMultiplayerLobby(result.lobby);
      setMultiplayerSelf(result.self);
      setMultiplayerSession((current) =>
        current
          ? {
              ...current,
              joinCode: result.lobby.joinCode,
              gameID: result.gameID
            }
          : current
      );
      enterLoadedGame(result.state, result.gameID, result.revision, "Multiplayer game started.");
    } catch (err) {
      setStatus(err instanceof Error ? err.message : "Failed to start lobby");
    } finally {
      setMultiplayerSubmitting(false);
    }
  }

  async function handleLeaveLobby() {
    if (!multiplayerToken) {
      clearMultiplayerState();
      resetToSetup({ status: "Choose factions and create a new game." });
      return;
    }

    try {
      setMultiplayerSubmitting(true);
      await leaveLobby(multiplayerToken);
      clearMultiplayerState();
      setServerGameID(null);
      setServerRevision(null);
      setSetupScreen("wizard");
      setShowSetup(true);
      setStatus("Left multiplayer lobby.");
    } catch (err) {
      setStatus(err instanceof Error ? err.message : "Cannot leave this multiplayer session right now.");
    } finally {
      setMultiplayerSubmitting(false);
    }
  }

  const selectedClearing =
    parsedState.map.clearings.find((clearing) => clearing.id === selectedClearingID) ??
    parsedState.map.clearings[0];
  const selectedClearingRuler = selectedClearing ? rulerOfClearing(selectedClearing) : null;
  const boardLayout = boardLayoutForState(parsedState);
  const boardIsEmpty = isBoardEmpty(parsedState);
  const previewedAction =
    actions[hoveredActionIndex ?? selectedBattleIndex ?? -1] ?? null;
  const {
    battleCandidates: assistBattleCandidates,
    movementCandidates: assistMovementCandidates,
    buildRecruitCandidates: assistBuildRecruitCandidates,
    factionSpatialCandidates: assistFactionSpatialCandidates
  } = buildAssistBoardCandidates({
    battleCandidates: assistBattleCandidateRefs,
    movementCandidates: [...assistMovementCandidateRefs, ...playerMovementCandidateRefs],
    buildRecruitCandidates: [...assistBuildRecruitCandidateRefs, ...playerBuildRecruitCandidateRefs],
    factionSpatialCandidates: [...assistFactionSpatialCandidateRefs, ...playerFactionSpatialCandidateRefs]
  });
  const activeMovementSource = assistMovementSource ?? playerMovementSource;
  const highlightedClearings = previewedAction
    ? affectedClearings(previewedAction)
    : assistBoardHighlights({
        battleCandidates: assistBattleCandidates,
        movementCandidates: assistMovementCandidates,
        buildRecruitCandidates: assistBuildRecruitCandidates,
        factionSpatialCandidates: assistFactionSpatialCandidates,
        movementSource: activeMovementSource
      });
  const selectedBattleAction =
    selectedBattleIndex !== null && actions[selectedBattleIndex]?.type === ACTION_TYPE.BATTLE
      ? actions[selectedBattleIndex]
      : null;
  const activeBattleAction = multiplayerBattlePrompt?.action ?? selectedBattleAction;
  const activeBattleContext = multiplayerBattlePrompt?.battleContext ?? battleContext;
  const attackerFaction = activeBattleAction?.battle?.faction ?? -1;
  const defenderFaction = activeBattleAction?.battle?.targetFaction ?? -1;
  const attackerHasScoutingParty = activeBattleContext?.attackerHasScoutingParty ?? false;
  const canDefenderAmbush = activeBattleContext?.canDefenderAmbush ?? false;
  const assistDefenderAmbushPromptRequired = activeBattleContext?.assistDefenderAmbushPromptRequired ?? false;
  const canAttackerCounterAmbush = activeBattleContext?.canAttackerCounterAmbush ?? false;
  const canAttackerArmorers = activeBattleContext?.canAttackerArmorers ?? false;
  const canDefenderArmorers = activeBattleContext?.canDefenderArmorers ?? false;
  const canAttackerBrutalTactics = activeBattleContext?.canAttackerBrutalTactics ?? false;
  const canDefenderSappers = activeBattleContext?.canDefenderSappers ?? false;
  const marquiseSetupActions = actions.filter((action) => action.type === ACTION_TYPE.MARQUISE_SETUP);
  const eyrieSetupActions = actions.filter((action) => action.type === ACTION_TYPE.EYRIE_SETUP);
  const vagabondSetupActions = actions.filter((action) => action.type === ACTION_TYPE.VAGABOND_SETUP);
  const isMultiplayerGame = multiplayerToken !== null;
  const hasPrimaryBattleFlow = Boolean(activeBattleAction?.battle);
  const showPrimarySetupFlow = parsedState.gamePhase === 0;
  const showPrimaryReviewFlow = parsedState.gamePhase === 2;
  const showPrimaryAssistFlow =
    parsedState.gamePhase === 1 &&
    parsedState.gameMode === 1 &&
    parsedState.factionTurn !== parsedState.playerFaction &&
    !hasPrimaryBattleFlow;
  const showPrimaryPlayerFlow =
    parsedState.gamePhase === 1 &&
    parsedState.factionTurn === parsedState.playerFaction &&
    !hasPrimaryBattleFlow;
  const primaryFlowLabel = showPrimaryReviewFlow
    ? "Review"
    : hasPrimaryBattleFlow
      ? "Battle"
      : showPrimarySetupFlow
        ? "Setup"
        : showPrimaryAssistFlow
          ? "Observed Turn"
          : showPrimaryPlayerFlow
            ? "Player Turn"
            : isMultiplayerGame
              ? "Waiting"
              : "Flow";
  const primaryFlowSummary = (() => {
    if (showPrimaryReviewFlow) {
      return "The match is complete. Review the result here before touching restart or recovery tools.";
    }
    if (hasPrimaryBattleFlow) {
      if (multiplayerBattlePrompt) {
        if (
          (multiplayerBattlePrompt.stage === "defender_response" || multiplayerBattlePrompt.stage === "attacker_response") &&
          multiplayerBattlePrompt.waitingOnFaction === parsedState.playerFaction
        ) {
          return "Your battle response is blocking the turn. Submit it here to continue.";
        }
        if (
          (multiplayerBattlePrompt.stage === "defender_response" || multiplayerBattlePrompt.stage === "attacker_response") &&
          multiplayerBattlePrompt.waitingOnFaction !== parsedState.playerFaction
        ) {
          return `Battle progress is waiting on ${factionLabels[multiplayerBattlePrompt.waitingOnFaction] ?? "another player"}.`;
        }
        if (multiplayerBattlePrompt.stage === "ready_to_resolve") {
          return multiplayerBattlePrompt.action.battle?.faction === parsedState.playerFaction
            ? "All battle responses are in. Resolve here to return to the turn flow."
            : `All battle responses are in. Waiting on ${factionLabels[multiplayerBattlePrompt.action.battle?.faction ?? -1] ?? "the attacker"} to resolve.`;
        }
      }
      return "Battle flow is blocking the turn. Finish it here before returning to other actions.";
    }
    if (showPrimarySetupFlow) {
      return actions.length > 0
        ? "Highlighted setup targets are ready on the board."
        : "Load setup choices, then complete the staged board selections here.";
    }
    if (showPrimaryAssistFlow) {
      return actions.length > 0
        ? `${actions.length} public action(s) are loaded for the observed turn.`
        : "Load public actions or record observed events manually here.";
    }
    if (showPrimaryPlayerFlow) {
      if (actions.length > 0) {
        return `${actions.length} legal action(s) are loaded and ready.`;
      }
      return isMultiplayerGame
        ? "Your turn has priority. Legal actions will refresh automatically here."
        : "Load legal actions here to continue the turn.";
    }
    if (isMultiplayerGame) {
      if (multiplayerConnectionStatus === "reconnecting" || multiplayerConnectionStatus === "connecting") {
        return "Realtime connection is recovering. Stay here while the server resynchronizes state.";
      }
      return `Waiting on ${factionLabels[parsedState.factionTurn] ?? "another player"} until the server hands your faction priority.`;
    }
    return "Use this area for the active workflow before opening any secondary tools.";
  })();

  const handleAssistBattleCandidatesChange = useCallback((candidates: AssistActionCandidateRef[]) => {
    setAssistBattleCandidateRefs((current) => (sameCandidateRefs(current, candidates) ? current : candidates));
  }, []);

  const handleAssistMovementCandidatesChange = useCallback((candidates: AssistActionCandidateRef[]) => {
    setAssistMovementCandidateRefs((current) => (sameCandidateRefs(current, candidates) ? current : candidates));
    setAssistMovementSource(null);
  }, []);

  const handlePlayerMovementCandidatesChange = useCallback((candidates: AssistActionCandidateRef[]) => {
    setPlayerMovementCandidateRefs((current) => (sameCandidateRefs(current, candidates) ? current : candidates));
    setPlayerMovementSource(null);
  }, []);

  const handleAssistBuildRecruitCandidatesChange = useCallback((candidates: AssistActionCandidateRef[]) => {
    setAssistBuildRecruitCandidateRefs((current) => (sameCandidateRefs(current, candidates) ? current : candidates));
  }, []);

  const handlePlayerBuildRecruitCandidatesChange = useCallback((candidates: AssistActionCandidateRef[]) => {
    setPlayerBuildRecruitCandidateRefs((current) => (sameCandidateRefs(current, candidates) ? current : candidates));
  }, []);

  const handleAssistFactionSpatialCandidatesChange = useCallback((candidates: AssistActionCandidateRef[]) => {
    setAssistFactionSpatialCandidateRefs((current) => (sameCandidateRefs(current, candidates) ? current : candidates));
  }, []);

  const handlePlayerFactionSpatialCandidatesChange = useCallback((candidates: AssistActionCandidateRef[]) => {
    setPlayerFactionSpatialCandidateRefs((current) => (sameCandidateRefs(current, candidates) ? current : candidates));
  }, []);

  const legalSetupClearingIDs =
    parsedState.gamePhase !== 0
      ? []
      : parsedState.setupStage === 1
        ? (() => {
            if (marquiseSetupDraft.keepClearingID === null) {
              return Array.from(new Set(marquiseSetupActions.map((action) => action.marquiseSetup?.keepClearingID ?? 0))).filter(
                (value) => value > 0
              );
            }

            const filteredByKeep = marquiseSetupActions.filter((action) =>
              marquiseSetupMatches(action, marquiseSetupDraft)
            );
            if (marquiseSetupDraft.sawmillClearingID === null) {
              return Array.from(new Set(filteredByKeep.map((action) => action.marquiseSetup?.sawmillClearingID ?? 0))).filter(
                (value) => value > 0
              );
            }

            const filteredBySawmill = filteredByKeep.filter((action) =>
              marquiseSetupMatches(action, marquiseSetupDraft)
            );
            if (marquiseSetupDraft.workshopClearingID === null) {
              return Array.from(new Set(filteredBySawmill.map((action) => action.marquiseSetup?.workshopClearingID ?? 0))).filter(
                (value) => value > 0
              );
            }

            return Array.from(
              new Set(
                filteredBySawmill
                  .filter((action) => marquiseSetupMatches(action, marquiseSetupDraft))
                  .map((action) => action.marquiseSetup?.recruiterClearingID ?? 0)
              )
            ).filter((value) => value > 0);
          })()
        : parsedState.setupStage === 2
          ? eyrieSetupActions.map((action) => action.eyrieSetup?.clearingID ?? 0).filter((value) => value > 0)
          : [];

  const selectedSetupClearingIDs =
    parsedState.gamePhase === 0 && parsedState.setupStage === 1
      ? [
          marquiseSetupDraft.keepClearingID,
          marquiseSetupDraft.sawmillClearingID,
          marquiseSetupDraft.workshopClearingID,
          marquiseSetupDraft.recruiterClearingID
        ].filter(
          (value): value is number => value !== null
        )
      : parsedState.gamePhase === 0 && parsedState.setupStage === 2 && eyrieSetupDraftClearingID !== null
        ? [eyrieSetupDraftClearingID]
      : [];

  const setupPreviewPiecesByClearing = (() => {
    const previews: Record<number, ClearingPreviewPiece[]> = {};
    const addPreview = (clearingID: number | null, piece: ClearingPreviewPiece) => {
      if (clearingID === null) {
        return;
      }
      previews[clearingID] ??= [];
      previews[clearingID].push(piece);
    };

    if (parsedState.gamePhase === 0 && parsedState.setupStage === 1) {
      addPreview(marquiseSetupDraft.keepClearingID, { kind: "keep", label: "Pending Keep", preview: true });
      addPreview(marquiseSetupDraft.sawmillClearingID, { kind: "sawmill", label: "Pending sawmill", preview: true });
      addPreview(marquiseSetupDraft.workshopClearingID, { kind: "workshop", label: "Pending workshop", preview: true });
      addPreview(marquiseSetupDraft.recruiterClearingID, { kind: "recruiter", label: "Pending recruiter", preview: true });
    }

    if (parsedState.gamePhase === 0 && parsedState.setupStage === 2) {
      addPreview(eyrieSetupDraftClearingID, { kind: "roost", label: "Pending roost", preview: true });
      addPreview(eyrieSetupDraftClearingID, { kind: "eyrie", count: 6, label: "Pending Eyrie warriors 6", preview: true });
    }

    return previews;
  })();

  const forestTargets =
    parsedState.gamePhase === 0 && parsedState.setupStage === 3
      ? parsedState.map.forests.map((forest) => ({
          forestID: forest.id,
          label: `Forest ${forest.id}`,
          legal: vagabondSetupActions.some((action) => action.vagabondSetup?.forestID === forest.id),
          selected: vagabondSetupDraftForestID === forest.id
        }))
      : assistMovementCandidates.length > 0 && parsedState.gamePhase === 1
        ? parsedState.map.forests
            .map((forest) => {
              const legal =
                activeMovementSource === null
                  ? assistMovementCandidates.some((candidate) => candidate.endpoints.fromForestID === forest.id)
                  : assistMovementCandidates.some(
                      (candidate) => movementSourceMatches(candidate.endpoints, activeMovementSource) && candidate.endpoints.toForestID === forest.id
                    );
              return {
                forestID: forest.id,
                label: `Forest ${forest.id}`,
                legal,
                selected: activeMovementSource?.kind === "forest" && activeMovementSource.id === forest.id
              };
            })
            .filter((forest) => forest.legal || forest.selected)
        : [];
  const setupPrompt = showPrimarySetupFlow ? setupBoardPrompt(parsedState.setupStage, marquiseSetupDraft) : null;
  const setupLegalChoiceCount = parsedState.setupStage === 3 ? forestTargets.filter((target) => target.legal).length : legalSetupClearingIDs.length;
  const hasMarquiseDraftSelection =
    marquiseSetupDraft.keepClearingID !== null ||
    marquiseSetupDraft.sawmillClearingID !== null ||
    marquiseSetupDraft.workshopClearingID !== null ||
    marquiseSetupDraft.recruiterClearingID !== null;

  useEffect(() => {
    if (assistDefenderAmbushPromptRequired) {
      setAssistDefenderAmbushChoice(null);
      setBattleModifiers((current) => ({
        ...current,
        defenderAmbush: false,
        attackerCounterAmbush: false
      }));
      return;
    }

    setAssistDefenderAmbushChoice(false);
  }, [assistDefenderAmbushPromptRequired, activeBattleAction, multiplayerBattlePrompt]);

  useEffect(() => {
    if (!multiplayerBattlePrompt) {
      return;
    }

    setBattleModifiers({
      ...emptyBattleModifiers,
      defenderAmbush: multiplayerBattlePrompt.defenderAmbush ?? false,
      defenderUsesArmorers: multiplayerBattlePrompt.defenderUsedArmorers ?? false,
      defenderUsesSappers: multiplayerBattlePrompt.defenderUsedSappers ?? false,
      attackerCounterAmbush: multiplayerBattlePrompt.attackerCounterAmbush ?? false,
      attackerUsesArmorers: multiplayerBattlePrompt.attackerUsedArmorers ?? false,
      attackerUsesBrutalTactics: multiplayerBattlePrompt.attackerUsedBrutalTactics ?? false
    });
  }, [multiplayerBattlePrompt]);

  useEffect(() => {
    let cancelled = false;

    async function loadBattleContext() {
      if (multiplayerBattlePrompt?.battleContext) {
        setBattleContext(multiplayerBattlePrompt.battleContext);
        return;
      }
      if (!selectedBattleAction?.battle) {
        setBattleContext(null);
        return;
      }

      try {
        const nextContext = await fetchBattleContext(parsedState, selectedBattleAction, serverGameID, multiplayerToken);
        if (!cancelled) {
          if (nextContext.revision !== null) {
            setServerRevision(nextContext.revision);
          }
          setBattleContext(nextContext.battleContext);
        }
      } catch {
        if (!cancelled) {
          setBattleContext(null);
        }
      }
    }

    void loadBattleContext();

    return () => {
      cancelled = true;
    };
  }, [multiplayerBattlePrompt, multiplayerToken, parsedState, selectedBattleAction, serverGameID]);

  function setBoardMovementSource(source: AssistMovementSource | null) {
    if (parsedState.factionTurn === parsedState.playerFaction) {
      setPlayerMovementSource(source);
      setAssistMovementSource(null);
      return;
    }
    setAssistMovementSource(source);
    setPlayerMovementSource(null);
  }

  async function handleSetupClearingClick(clearingID: number) {
    if (assistBattleCandidates.length > 0 && parsedState.gamePhase === 1 && parsedState.factionTurn !== parsedState.playerFaction) {
      setSelectedClearingID(clearingID);
      const matchingBattles = assistBattleCandidates.filter((candidate) => candidate.action.battle?.clearingID === clearingID);
      if (matchingBattles.length === 1) {
        openBattleForActionIndex(matchingBattles[0].actionIndex);
        setStatus(`Battle selected in clearing ${clearingID}. Resolve it from Battle Flow.`);
        return;
      }
      if (matchingBattles.length > 1) {
        setStatus(`Clearing ${clearingID} has multiple battle targets. Choose the observed defender in the Battle prompt.`);
        return;
      }
      setStatus("Choose one of the highlighted battle clearings.");
      return;
    }

    if (assistMovementCandidates.length > 0 && parsedState.gamePhase === 1) {
      setSelectedClearingID(clearingID);
      if (activeMovementSource === null) {
        const sourceMatches = assistMovementCandidates.filter((candidate) => candidate.endpoints.fromClearingID === clearingID);
        if (sourceMatches.length === 0) {
          setStatus("Choose one of the highlighted move source clearings.");
          return;
        }
        const uniqueClearingTargets = Array.from(new Set(sourceMatches.map((candidate) => candidate.endpoints.toClearingID))).filter((value) => value > 0);
        const uniqueForestTargets = Array.from(new Set(sourceMatches.map((candidate) => candidate.endpoints.toForestID))).filter((value) => value > 0);
        if (sourceMatches.length === 1 && uniqueClearingTargets.length === 1 && uniqueForestTargets.length === 0) {
          setStatus(`Recording movement from clearing ${clearingID} to clearing ${uniqueClearingTargets[0]}...`);
          await handleApply(sourceMatches[0].action);
          return;
        }
        setBoardMovementSource({ kind: "clearing", id: clearingID });
        setStatus(`Move source selected: clearing ${clearingID}. Choose a highlighted destination.`);
        return;
      }

      const matchingMoves = assistMovementCandidates.filter(
        (candidate) => movementSourceMatches(candidate.endpoints, activeMovementSource) && candidate.endpoints.toClearingID === clearingID
      );
      if (matchingMoves.length === 1) {
        const sourceLabel = activeMovementSource.kind === "clearing" ? `clearing ${activeMovementSource.id}` : `forest ${activeMovementSource.id}`;
        setStatus(`Recording movement from ${sourceLabel} to clearing ${clearingID}...`);
        await handleApply(matchingMoves[0].action);
        setBoardMovementSource(null);
        return;
      }
      if (matchingMoves.length > 1) {
        setStatus(`Multiple moves match that route. Choose the exact candidate from the Move list.`);
        return;
      }
      if (assistMovementCandidates.some((candidate) => candidate.endpoints.fromClearingID === clearingID)) {
        setBoardMovementSource({ kind: "clearing", id: clearingID });
        setStatus(`Move source changed to clearing ${clearingID}. Choose a highlighted destination.`);
        return;
      }
      setStatus("Choose a highlighted move destination, or click another highlighted source to restart the route.");
      return;
    }

    if (assistBuildRecruitCandidates.length > 0 && parsedState.gamePhase === 1) {
      setSelectedClearingID(clearingID);
      const matchingCandidates = assistBuildRecruitCandidates.filter((candidate) => candidate.clearingIDs.includes(clearingID));
      if (matchingCandidates.length === 1) {
        setStatus(`Recording ${actionHeadline(matchingCandidates[0].action).toLowerCase()} at clearing ${clearingID}...`);
        await handleApply(matchingCandidates[0].action);
        return;
      }
      if (matchingCandidates.length > 1) {
        setStatus(`Clearing ${clearingID} matches multiple Build / Recruit candidates. Choose the exact candidate from the list.`);
        return;
      }
      setStatus("Choose one of the highlighted build, recruit, or overwork clearings.");
      return;
    }

    if (assistFactionSpatialCandidates.length > 0 && parsedState.gamePhase === 1) {
      setSelectedClearingID(clearingID);
      const matchingCandidates = assistFactionSpatialCandidates.filter((candidate) => candidate.clearingIDs.includes(clearingID));
      if (matchingCandidates.length === 1) {
        setStatus(`Recording ${actionHeadline(matchingCandidates[0].action).toLowerCase()} at clearing ${clearingID}...`);
        await handleApply(matchingCandidates[0].action);
        return;
      }
      if (matchingCandidates.length > 1) {
        setStatus(`Clearing ${clearingID} matches multiple faction-action candidates. Choose the exact candidate from the list.`);
        return;
      }
      setStatus("Choose one of the highlighted faction-action clearings.");
      return;
    }

    if (parsedState.gamePhase !== 0) {
      setSelectedClearingID(clearingID);
      if (multiplayerToken) {
        setStatus("Board editing is disabled in multiplayer.");
        return;
      }
      setShowBoardEditor(true);
      setStatus(`Selected clearing ${clearingID} for board editing.`);
      return;
    }

    if (!legalSetupClearingIDs.includes(clearingID)) {
      setStatus("Choose one of the highlighted setup targets.");
      return;
    }

    if (parsedState.setupStage === 1) {
      if (marquiseSetupDraft.keepClearingID === null) {
        setMarquiseSetupDraft({ ...emptyMarquiseSetupDraft, keepClearingID: clearingID });
        setStatus("Choose the starting sawmill location.");
        return;
      }
      if (marquiseSetupDraft.sawmillClearingID === null) {
        setMarquiseSetupDraft({ ...marquiseSetupDraft, sawmillClearingID: clearingID });
        setStatus("Choose the starting workshop location.");
        return;
      }
      if (marquiseSetupDraft.workshopClearingID === null) {
        setMarquiseSetupDraft({ ...marquiseSetupDraft, workshopClearingID: clearingID });
        setStatus("Choose the starting recruiter location.");
        return;
      }

      const finalDraft = { ...marquiseSetupDraft, recruiterClearingID: clearingID };
      const action = marquiseSetupActions.find((candidate) =>
        marquiseSetupMatches(candidate, finalDraft)
      );
      if (!action) {
        setStatus("That building placement is not legal.");
        return;
      }
      setMarquiseSetupDraft(finalDraft);
      setStatus("Applying Marquise setup...");
      await handleApply(action);
      return;
    }

    if (parsedState.setupStage === 2) {
      const action = eyrieSetupActions.find((candidate) => candidate.eyrieSetup?.clearingID === clearingID);
      if (!action) {
        setStatus("That starting clearing is not legal for the Eyrie.");
        return;
      }
      setEyrieSetupDraftClearingID(clearingID);
      setStatus("Applying Eyrie setup...");
      await handleApply(action);
    }
  }

  async function handleSetupForestClick(forestID: number) {
    if (assistMovementCandidates.length > 0 && parsedState.gamePhase === 1) {
      if (activeMovementSource === null) {
        const sourceMatches = assistMovementCandidates.filter((candidate) => candidate.endpoints.fromForestID === forestID);
        if (sourceMatches.length === 0) {
          setStatus("Choose one of the highlighted move source forests.");
          return;
        }
        const uniqueClearingTargets = Array.from(new Set(sourceMatches.map((candidate) => candidate.endpoints.toClearingID))).filter((value) => value > 0);
        const uniqueForestTargets = Array.from(new Set(sourceMatches.map((candidate) => candidate.endpoints.toForestID))).filter((value) => value > 0);
        if (sourceMatches.length === 1 && uniqueClearingTargets.length === 1 && uniqueForestTargets.length === 0) {
          setStatus(`Recording movement from forest ${forestID} to clearing ${uniqueClearingTargets[0]}...`);
          await handleApply(sourceMatches[0].action);
          return;
        }
        setBoardMovementSource({ kind: "forest", id: forestID });
        setStatus(`Move source selected: forest ${forestID}. Choose a highlighted destination.`);
        return;
      }

      const matchingMoves = assistMovementCandidates.filter(
        (candidate) => movementSourceMatches(candidate.endpoints, activeMovementSource) && candidate.endpoints.toForestID === forestID
      );
      if (matchingMoves.length === 1) {
        const sourceLabel = activeMovementSource.kind === "clearing" ? `clearing ${activeMovementSource.id}` : `forest ${activeMovementSource.id}`;
        setStatus(`Recording movement from ${sourceLabel} to forest ${forestID}...`);
        await handleApply(matchingMoves[0].action);
        setBoardMovementSource(null);
        return;
      }
      if (matchingMoves.length > 1) {
        setStatus("Multiple moves match that forest route. Choose the exact candidate from the Move list.");
        return;
      }
      if (assistMovementCandidates.some((candidate) => candidate.endpoints.fromForestID === forestID)) {
        setBoardMovementSource({ kind: "forest", id: forestID });
        setStatus(`Move source changed to forest ${forestID}. Choose a highlighted destination.`);
        return;
      }
      setStatus("Choose a highlighted move forest, or click another highlighted source to restart the route.");
      return;
    }

    if (parsedState.gamePhase !== 0 || parsedState.setupStage !== 3) {
      return;
    }

    const action = vagabondSetupActions.find((candidate) => candidate.vagabondSetup?.forestID === forestID);
    if (!action) {
      setStatus("That forest is not a legal Vagabond starting forest.");
      return;
    }
    setVagabondSetupDraftForestID(forestID);
    setStatus("Applying Vagabond setup...");
    await handleApply(action);
  }

  if (showSetup) {
    if (multiplayerLobby) {
      return (
        <LobbyScreen
          lobby={multiplayerLobby}
          self={multiplayerSelf}
          connectionStatus={multiplayerConnectionStatus}
          status={status}
          submitting={multiplayerSubmitting}
          onClaimFaction={handleClaimLobby}
          onReady={handleSetLobbyReady}
          onStart={handleStartLobby}
          onLeave={handleLeaveLobby}
        />
      );
    }

    if (setupScreen === "create-lobby" || setupScreen === "join-lobby") {
      return (
        <JoinScreen
          mode={setupScreen === "create-lobby" ? "create" : "join"}
          submitting={multiplayerSubmitting}
          status={status}
          onBack={() => {
            setSetupScreen("wizard");
            setStatus("Choose how you want to play.");
          }}
          onCreateLobby={handleCreateLobby}
          onJoinLobby={handleJoinLobby}
        />
      );
    }

    return (
      <SetupWizard
        canResume={hasSavedSession}
        savedSessionInfo={savedSessionInfo}
        onStart={async (state, gameID, revision) => {
          clearMultiplayerState();
          enterLoadedGame(state, gameID, revision, state.gamePhase === 0 ? "Setup started." : "Game created.");
          if (state.gamePhase === 0) {
            try {
              await loadActionsForState(state, { successStatus: "Choose a highlighted setup target." });
            } catch (err) {
              const message = err instanceof Error ? err.message : "Failed to load setup actions";
              setStatus(message);
            }
          }
        }}
        onUseSample={() => {
          clearMultiplayerState();
          syncState(initialState);
          setServerGameID(null);
          setServerRevision(null);
          setShowSetup(false);
          setShowBoardEditor(false);
          setStatus("Loaded sample state.");
          setShowGuideHelp(true);
        }}
        onOpenCreateLobby={() => {
          setSetupScreen("create-lobby");
          setStatus("Enter your display name to create a lobby.");
        }}
        onOpenJoinLobby={() => {
          setSetupScreen("join-lobby");
          setStatus("Enter your display name and join code.");
        }}
        onClearSavedSession={() => {
          clearSavedSession();
          setHasSavedSession(false);
          setSavedSessionInfo(null);
          setStatus("Cleared saved game.");
        }}
        onResume={handleResumeSavedGame}
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
        {setupPrompt ? (
          <>
            <div className="board-phase-banner">
              <span>{setupPrompt.title}</span>
            </div>
            <div className="board-setup-prompt">
              <p className="eyebrow">Setup</p>
              <strong>{setupPrompt.instruction}</strong>
              <span>{setupPrompt.detail}</span>
              <span>{setupLegalChoiceCount} legal choice{setupLegalChoiceCount === 1 ? "" : "s"}</span>
              {parsedState.setupStage === 1 && hasMarquiseDraftSelection ? (
                <button
                  type="button"
                  className="secondary"
                  onClick={() => {
                    setMarquiseSetupDraft(emptyMarquiseSetupDraft);
                    setStatus("Marquise setup draft reset.");
                  }}
                >
                  Reset Marquise setup
                </button>
              ) : null}
            </div>
          </>
        ) : null}
        {parsedState.gamePhase === 2 ? (
          <div className="board-hint endgame-banner">
            {gameOverHeadline(parsedState)}. Open the Game Over panel for the final standings.
          </div>
        ) : null}
        <BoardPanel
          state={parsedState}
          clearings={parsedState.map.clearings}
          forests={parsedState.map.forests}
          boardLayout={boardLayout}
          selectedClearingID={selectedClearingID}
          keepClearingID={parsedState.marquise.keepClearingID}
          vagabondClearingID={parsedState.vagabond.clearingID}
          vagabondInForest={parsedState.vagabond.inForest}
          highlightedClearings={highlightedClearings}
          previewedAction={previewedAction}
          setupLegalClearingIDs={legalSetupClearingIDs}
          setupSelectedClearingIDs={selectedSetupClearingIDs}
          setupPreviewPiecesByClearing={setupPreviewPiecesByClearing}
          forestTargets={forestTargets}
          onSelectClearing={handleSetupClearingClick}
          onSelectForest={handleSetupForestClick}
        />

        {boardIsEmpty && parsedState.gamePhase !== 0 ? <div className="board-hint">Click a clearing to select it for board editing.</div> : null}
      </div>

      <aside className={`app-command-deck ${showPrimarySetupFlow ? "setup-deck" : ""}`} aria-label="Play controls">
        <section className="panel sidebar-panel">
          <p className="eyebrow">RootBuddy</p>
          <div className="status-block">
            <div className="status-block-main">
              <strong>{factionLabels[parsedState.factionTurn] ?? "Unknown"}</strong>
              <span>
                {parsedState.gamePhase === 0
                  ? setupStageLabels[parsedState.setupStage] ?? "Setup"
                  : `${phaseLabels[parsedState.currentPhase] ?? "Unknown"} / ${stepLabels[parsedState.currentStep] ?? "Unknown"}`}
              </span>
            </div>
            {multiplayerToken ? (
              <span className={`connection-pill compact ${multiplayerConnectionStatus}`}>
                {multiplayerConnectionStatus === "connected"
                  ? "Live"
                  : multiplayerConnectionStatus === "reconnecting"
                    ? "Reconnecting"
                    : multiplayerConnectionStatus === "connecting"
                      ? "Connecting"
                      : "Offline"}
              </span>
            ) : null}
          </div>
          <span className={error ? "message error" : "message"}>{error || status}</span>
        </section>

        {!showPrimarySetupFlow ? (
          <div className="primary-flow-stack">
            <section className="panel sidebar-panel primary-flow-panel">
              <p className="eyebrow">Primary Workflow</p>
              <span className="summary-label">{primaryFlowLabel}</span>
              <span className="summary-line">{primaryFlowSummary}</span>
            </section>

            {multiplayerNotice ? (
            <section className={`panel sidebar-panel notice-panel ${multiplayerNotice.level}`}>
              <p className="eyebrow">{multiplayerNotice.title}</p>
              <span className="summary-line">{multiplayerNotice.detail}</span>
            </section>
            ) : null}

            {!showPrimaryAssistFlow ? (
            <FlowGuidePanel
              state={parsedState}
              loadedActionCount={actions.length}
              selectedBattleAction={selectedBattleAction}
              isMultiplayer={isMultiplayerGame}
              multiplayerConnectionStatus={multiplayerConnectionStatus}
              multiplayerBattlePrompt={multiplayerBattlePrompt}
              onGenerateActions={refreshActions}
              onOpenHelp={() => setShowGuideHelp(true)}
            />
            ) : null}

            {hasPrimaryBattleFlow ? (
            <BattleFlowPanel
              selectedBattleIndex={selectedBattleIndex}
              selectedBattleAction={selectedBattleAction}
              multiplayerBattlePrompt={multiplayerBattlePrompt}
              multiplayerPerspectiveFaction={parsedState.playerFaction}
              multiplayerSubmitting={multiplayerSubmitting}
              attackerFaction={attackerFaction}
              defenderFaction={defenderFaction}
              attackerRoll={attackerRoll}
              defenderRoll={defenderRoll}
              battleModifiers={battleModifiers}
              battleContext={battleContext}
              assistDefenderAmbushChoice={assistDefenderAmbushChoice}
              onSetAttackerRoll={setAttackerRoll}
              onSetDefenderRoll={setDefenderRoll}
              onSetBattleModifiers={(updater) => setBattleModifiers((current) => updater(current))}
              onSetAssistDefenderAmbushChoice={setAssistDefenderAmbushChoice}
              onSubmitMultiplayerResponse={handleSubmitBattleResponse}
              onResolveAndApply={handleResolveAndApply}
              onClearSelection={() => {
                setSelectedBattleIndex(null);
                setHoveredActionIndex(null);
                setBattleContext(null);
                setMultiplayerBattlePrompt(null);
                setBattleModifiers(emptyBattleModifiers);
                setAssistDefenderAmbushChoice(null);
                setStatus("Cleared selected battle.");
              }}
            />
            ) : null}

            {showPrimaryReviewFlow ? (
            <EndgamePanel
              state={parsedState}
              hasSavedSession={hasSavedSession}
              serverGameID={serverGameID}
              onNewGame={() => {
                if (multiplayerToken) {
                  setStatus("Starting a new game from the in-game multiplayer workspace is not supported.");
                  return;
                }
                resetToSetup({ clearSaved: true, status: "Start a new game." });
              }}
              onReturnToSetup={() => {
                if (multiplayerToken) {
                  setStatus("Return to setup is disabled while a multiplayer session is active.");
                  return;
                }
                resetToSetup({ status: "Returned to setup. Resume is still available until you clear it." });
              }}
              onClearSavedSession={() => {
                clearSavedSession();
                setHasSavedSession(false);
                setSavedSessionInfo(null);
                setStatus("Cleared the saved endgame result.");
              }}
              onOpenDebug={() => {
                if (multiplayerToken) {
                  setStatus("Debug JSON is disabled in multiplayer.");
                  return;
                }
                setActiveModal("json");
              }}
            />
            ) : null}

            {showPrimaryAssistFlow ? (
            <AssistWorkflowPanel
              state={parsedState}
              actions={actions}
              onApply={handleApply}
              onGenerateActions={refreshActions}
              onOpenTurnState={() => setShowAdvancedTurnPanel(true)}
              onOpenBattle={openBattleForActionIndex}
              onBattleCandidatesChange={handleAssistBattleCandidatesChange}
              onMovementCandidatesChange={handleAssistMovementCandidatesChange}
              onBuildRecruitCandidatesChange={handleAssistBuildRecruitCandidatesChange}
              onFactionSpatialCandidatesChange={handleAssistFactionSpatialCandidatesChange}
            />
            ) : null}

            {showPrimaryPlayerFlow ? (
            <PlayerActionsPanel
              state={parsedState}
              actions={actions}
              isMultiplayer={isMultiplayerGame}
              onApply={handleApply}
              onGenerateActions={refreshActions}
              onOpenBattle={openBattleForActionIndex}
              onPreviewAction={setHoveredActionIndex}
              onMovementCandidatesChange={handlePlayerMovementCandidatesChange}
              onBuildRecruitCandidatesChange={handlePlayerBuildRecruitCandidatesChange}
              onFactionSpatialCandidatesChange={handlePlayerFactionSpatialCandidatesChange}
            />
            ) : null}
          </div>
        ) : null}

        {!showPrimarySetupFlow && showGuideHelp ? <GuideHelpPanel gamePhase={parsedState.gamePhase} onClose={() => setShowGuideHelp(false)} /> : null}
        {parsedState.gamePhase === 1 && !multiplayerToken ? (
          <details
            className="panel sidebar-panel board-editor-drawer"
            open={showBoardEditor}
            onToggle={(event) => setShowBoardEditor(event.currentTarget.open)}
          >
            <summary className="panel-summary">
              <span className="summary-label">Board Editor</span>
              <span className="summary-line">
                {showBoardEditor
                  ? `Editing clearing ${selectedClearing?.id ?? "?"}.`
                  : `Click a clearing to open correction controls for clearing ${selectedClearing?.id ?? "?"}.`}
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
                  onUpdateClearing={updateClearing}
                  onSetKeepClearing={(clearingID) =>
                    updateState((draft) => {
                      draft.marquise.keepClearingID = clearingID;
                    })
                  }
                  onSetVagabondClearing={(clearingID, inForest) =>
                    updateState((draft) => {
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

        {!showPrimarySetupFlow ? (
        <details
          className="panel sidebar-panel context-drawer secondary-drawer"
          open={showContextDrawer}
          onToggle={(event) => setShowContextDrawer(event.currentTarget.open)}
        >
          <summary className="panel-summary">
            <span className="summary-label">Context & Reference</span>
            <span className="summary-line">Turn summary, card visibility, and session details live here as secondary reference.</span>
          </summary>
          <div className="context-drawer-body">
            <TurnSummaryPanel state={parsedState} />
            <CardVisibilityPanel state={parsedState} />
            <SessionStatusPanel
              state={parsedState}
              hasSavedSession={hasSavedSession}
              serverGameID={serverGameID}
              savedSessionInfo={savedSessionInfo}
              multiplayerSession={multiplayerSession}
              multiplayerConnectionStatus={multiplayerConnectionStatus}
              multiplayerBattlePrompt={multiplayerBattlePrompt}
            />
          </div>
        </details>
        ) : null}

        {!showPrimarySetupFlow ? (
        <details
          className="panel sidebar-panel sidebar-actions-panel secondary-drawer"
          open={showWorkspaceTools}
          onToggle={(event) => setShowWorkspaceTools(event.currentTarget.open)}
        >
          <summary className="panel-summary">
            <span className="summary-label">{parsedState.gamePhase === 2 ? "Review Workspace" : "Workspace Tools"}</span>
            <span className="summary-line">
              {parsedState.gamePhase === 2
                ? "Restart and review controls live here after the match ends."
                : "Board inspection and setup/recovery controls live here as secondary tools."}
            </span>
          </summary>
          {parsedState.gamePhase === 2 ? (
            <div className="summary-stack" style={{ margin: "0.9rem 0" }}>
              <span className="summary-line">The match is finished. Use these controls for review, restart, or recovery only.</span>
            </div>
          ) : (
            <div className="summary-stack" style={{ margin: "0.9rem 0" }}>
              <span className="summary-label">Board</span>
              <span className="summary-line">Selected clearing: {selectedClearing?.id ?? "None"}</span>
              {selectedClearing ? (
                <>
                  <span className="summary-line">Suit: {suitLabels[selectedClearing.suit] ?? "Unknown"}</span>
                  <span className="summary-line">
                    Ruler: {selectedClearingRuler === null ? "None" : factionLabels[selectedClearingRuler] ?? "Unknown"}
                  </span>
                  <span className="summary-line">
                    Paths / slots: {selectedClearing.adj.length} / {usedBuildSlots(selectedClearing)}/{selectedClearing.buildSlots}
                  </span>
                </>
              ) : null}
            </div>
          )}
          <div className="sidebar-actions">
            <button
              type="button"
              className="secondary"
              onClick={() => {
                if (multiplayerToken) {
                  setStatus("Return to setup is disabled while a multiplayer session is active.");
                  return;
                }
                resetToSetup({
                  clearSaved: parsedState.gamePhase !== 2,
                  status: parsedState.gamePhase === 2 ? "Returned to setup. Resume is still available until you clear it." : "Start a new game."
                });
              }}
            >
              {parsedState.gamePhase === 2 ? "Return to Setup" : "Setup"}
            </button>
          </div>
          {!multiplayerToken ? (
            <details
              className="advanced-tools secondary-drawer"
              style={{ marginTop: "0.9rem" }}
              open={showRecoveryTools}
              onToggle={(event) => setShowRecoveryTools(event.currentTarget.open)}
            >
              <summary className="panel-summary">
                <span className="summary-label">Recovery Tools</span>
                <span className="summary-line">Manual correction and recovery controls live here, away from the primary play flow.</span>
              </summary>
              {parsedState.gamePhase !== 2 ? (
                <div style={{ marginTop: "0.9rem" }}>
                  <TurnFlowPanel
                    state={parsedState}
                    onApply={handleApply}
                    onGenerateActions={refreshActions}
                    onOpenAdvanced={() => setShowAdvancedTurnPanel(true)}
                    onUpdateState={updateState}
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
                    onUpdateState={updateState}
                    title="Advanced Turn"
                    showCloseButton={false}
                    onClose={() => setShowAdvancedTurnPanel(false)}
                  />
                </div>
              ) : null}
            </details>
          ) : null}
          <div className="sidebar-actions footer" style={{ marginTop: "0.9rem" }}>
            {!multiplayerToken && parsedState.gamePhase !== 2 ? (
              <button
                type="button"
                className="secondary"
                onClick={() => {
                  syncState(initialState);
                  setServerGameID(null);
                  setServerRevision(null);
                  setShowBoardEditor(false);
                  setStatus("Board reset. Click a clearing to start setting the board.");
                  setShowGuideHelp(true);
                }}
              >
                Reset
              </button>
            ) : null}
            {parsedState.gamePhase === 2 ? (
              <button type="button" onClick={() => resetToSetup({ clearSaved: true, status: "Start a new game." })}>
                New Game
              </button>
            ) : null}
          </div>
        </details>
        ) : null}
      </aside>

      {activeModal ? (
        <div className="modal-backdrop" onClick={() => setActiveModal(null)}>
          <div className="modal-shell" onClick={(event) => event.stopPropagation()}>
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
                  value={stateText}
                  onChange={(event) => setStateText(event.target.value)}
                  spellCheck={false}
                />
              </section>
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
                      setStatus("Stand and Deliver cancelled.");
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
                      await applyFinalizedAction(actionToApply);
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
