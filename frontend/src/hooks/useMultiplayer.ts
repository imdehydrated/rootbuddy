import { useEffect, useRef, useState } from "react";
import {
  claimLobbyFaction,
  createLobby,
  fetchBattleSession,
  fetchLobbyState,
  joinLobby,
  leaveLobby,
  loadGame,
  setLobbyReady,
  startLobby
} from "../api";
import { factionLabels } from "../labels";
import {
  clearSavedMultiplayerSession,
  loadSavedMultiplayerSession,
  saveSavedMultiplayerSession
} from "../localSession";
import type { MultiplayerConnectionStatus, MultiplayerSession, SetupScreen } from "../multiplayer";
import { RootBuddyWebSocketClient } from "../wsClient";
import { gameOverStatusMessage } from "./useGameState";
import type { Action, BattlePrompt, GameState, Lobby, LobbyPlayer } from "../types";

type MultiplayerNotice = {
  level: "warning" | "error";
  title: string;
  detail: string;
} | null;

type UseMultiplayerOptions = {
  enterLoadedGame: (state: GameState, gameID: string | null, revision: number | null, status: string) => void;
  loadActionsForState: (state: GameState, options?: { successStatus?: string }) => Promise<Action[]>;
  resetToSetup: (options?: { clearSaved?: boolean; status?: string }) => void;
  setServerGameID: (value: string | null) => void;
  setServerRevision: (value: number | null) => void;
  setShowBoardEditor: (value: boolean) => void;
  setShowGuideHelp: (value: boolean) => void;
  setShowSetup: (value: boolean) => void;
  setSetupScreen: (value: SetupScreen) => void;
  setStatus: (status: string) => void;
  syncState: (state: GameState) => void;
  getPlayerFaction: () => number;
  serverGameID: string | null;
  showSetup: boolean;
};

export function useMultiplayer({
  enterLoadedGame,
  loadActionsForState,
  resetToSetup,
  setServerGameID,
  setServerRevision,
  setShowBoardEditor,
  setShowGuideHelp,
  setShowSetup,
  setSetupScreen,
  setStatus,
  syncState,
  getPlayerFaction,
  serverGameID,
  showSetup
}: UseMultiplayerOptions) {
  const initialSavedMultiplayerSessionRef = useRef(loadSavedMultiplayerSession());
  const [multiplayerSession, setMultiplayerSession] = useState<MultiplayerSession | null>(
    initialSavedMultiplayerSessionRef.current
      ? {
          playerToken: initialSavedMultiplayerSessionRef.current.playerToken,
          displayName: initialSavedMultiplayerSessionRef.current.displayName,
          joinCode: initialSavedMultiplayerSessionRef.current.joinCode,
          gameID: initialSavedMultiplayerSessionRef.current.gameID
        }
      : null
  );
  const [multiplayerLobby, setMultiplayerLobby] = useState<Lobby | null>(null);
  const [multiplayerSelf, setMultiplayerSelf] = useState<LobbyPlayer | null>(null);
  const [multiplayerConnectionStatus, setMultiplayerConnectionStatus] = useState<MultiplayerConnectionStatus>("disconnected");
  const [multiplayerSubmitting, setMultiplayerSubmitting] = useState(false);
  const [multiplayerBattlePrompt, setMultiplayerBattlePrompt] = useState<BattlePrompt | null>(null);
  const [multiplayerNotice, setMultiplayerNotice] = useState<MultiplayerNotice>(null);
  const multiplayerSelfRef = useRef<LobbyPlayer | null>(null);
  const loadActionsForStateRef = useRef(loadActionsForState);
  const syncStateRef = useRef(syncState);
  const getPlayerFactionRef = useRef(getPlayerFaction);
  const previousConnectionStatus = useRef<MultiplayerConnectionStatus>("disconnected");
  const multiplayerToken = multiplayerSession?.playerToken ?? null;
  const perspectiveFaction = multiplayerSelf?.hasFaction ? multiplayerSelf.faction : getPlayerFaction();

  useEffect(() => {
    multiplayerSelfRef.current = multiplayerSelf;
  }, [multiplayerSelf]);

  useEffect(() => {
    loadActionsForStateRef.current = loadActionsForState;
    syncStateRef.current = syncState;
    getPlayerFactionRef.current = getPlayerFaction;
  }, [getPlayerFaction, loadActionsForState, syncState]);

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
    const savedSession = initialSavedMultiplayerSessionRef.current;
    if (!savedSession) {
      return;
    }
    let cancelled = false;

    async function resumeMultiplayerSession() {
      setMultiplayerSubmitting(true);
      setStatus("Rejoining multiplayer session...");
      try {
        const { lobby, self } = await fetchLobbyState(savedSession!.playerToken);
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
          const loaded = await loadGame(lobby.gameID, savedSession!.playerToken);
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
          syncStateRef.current(loaded.state);
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
  }, [setServerGameID, setServerRevision, setSetupScreen, setShowBoardEditor, setShowGuideHelp, setShowSetup, setStatus]);

  useEffect(() => {
    if (!multiplayerToken) {
      setMultiplayerConnectionStatus("disconnected");
      return;
    }

    const client = new RootBuddyWebSocketClient({
      token: multiplayerToken,
      onConnectionChange: setMultiplayerConnectionStatus,
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
          syncStateRef.current(message.state);
          setShowSetup(false);
          setShowBoardEditor(false);
          setShowGuideHelp(false);
          if (message.state.gamePhase === 0) {
            setStatus("Loading setup choices...");
            void loadActionsForStateRef.current(message.state, { successStatus: "Choose a highlighted setup target." });
            return;
          }
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

          const currentSelf = multiplayerSelfRef.current;
          const promptPerspectiveFaction = currentSelf?.hasFaction ? currentSelf.faction : getPlayerFactionRef.current();
          if (message.prompt.waitingOnFaction === promptPerspectiveFaction) {
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
  }, [multiplayerToken, setServerGameID, setServerRevision, setShowBoardEditor, setShowGuideHelp, setShowSetup, setStatus]);

  useEffect(() => {
    const previous = previousConnectionStatus.current;
    previousConnectionStatus.current = multiplayerConnectionStatus;

    if (!multiplayerToken || multiplayerConnectionStatus === previous) {
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
  }, [multiplayerConnectionStatus, multiplayerToken, setStatus]);

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
  }, [multiplayerToken, serverGameID, setServerRevision, showSetup]);

  function clearMultiplayerState() {
    setMultiplayerSession(null);
    setMultiplayerLobby(null);
    setMultiplayerSelf(null);
    setMultiplayerBattlePrompt(null);
    setMultiplayerNotice(null);
    setMultiplayerConnectionStatus("disconnected");
    setSetupScreen("wizard");
  }

  async function handleCreateLobby(request: { displayName: string; factions: number[]; eyrieLeader: number; vagabondCharacter: number }) {
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

  return {
    multiplayerSession,
    multiplayerLobby,
    multiplayerSelf,
    multiplayerConnectionStatus,
    multiplayerSubmitting,
    multiplayerBattlePrompt,
    multiplayerNotice,
    multiplayerToken,
    perspectiveFaction,
    setMultiplayerBattlePrompt,
    setMultiplayerSubmitting,
    clearMultiplayerState,
    handleCreateLobby,
    handleJoinLobby,
    handleClaimLobby,
    handleSetLobbyReady,
    handleStartLobby,
    handleLeaveLobby
  };
}
