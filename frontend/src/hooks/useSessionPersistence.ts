import { useEffect, useState } from "react";
import { clearSavedSession, loadSavedSession, saveSavedSession, type SavedSession } from "../localSession";
import type { Action, GameState } from "../types";
import type { SetupScreen } from "../multiplayer";

type UseSessionPersistenceOptions = {
  getMultiplayerToken: () => string | null;
  loadSavedGame: (savedSession: SavedSession) => Promise<{ state: GameState; gameID: string | null; revision: number | null }>;
  loadActionsForState: (state: GameState, options?: { successStatus?: string }) => Promise<Action[]>;
  parsedState: GameState;
  serverGameID: string | null;
  serverRevision: number | null;
  setActiveModal: (value: null) => void;
  setServerGameID: (value: string | null) => void;
  setServerRevision: (value: number | null) => void;
  setShowAdvancedTurnPanel: (value: boolean) => void;
  setShowBoardEditor: (value: boolean) => void;
  setShowContextDrawer: (value: boolean) => void;
  setShowGuideHelp: (value: boolean) => void;
  setShowRecoveryTools: (value: boolean) => void;
  setShowWorkspaceTools: (value: boolean) => void;
  setStatus: (status: string) => void;
  syncState: (state: GameState) => void;
};

export function useSessionPersistence({
  getMultiplayerToken,
  loadSavedGame,
  loadActionsForState,
  parsedState,
  serverGameID,
  serverRevision,
  setActiveModal,
  setServerGameID,
  setServerRevision,
  setShowAdvancedTurnPanel,
  setShowBoardEditor,
  setShowContextDrawer,
  setShowGuideHelp,
  setShowRecoveryTools,
  setShowWorkspaceTools,
  setStatus,
  syncState
}: UseSessionPersistenceOptions) {
  const initialSavedSession = loadSavedSession();
  const [showSetup, setShowSetup] = useState(true);
  const [setupScreen, setSetupScreen] = useState<SetupScreen>("wizard");
  const [hasSavedSession, setHasSavedSession] = useState(() => initialSavedSession !== null);
  const [savedSessionInfo, setSavedSessionInfo] = useState<SavedSession | null>(initialSavedSession);

  useEffect(() => {
    if (showSetup || getMultiplayerToken()) {
      return;
    }
    const session: SavedSession = {
      state: parsedState,
      gameID: serverGameID,
      revision: serverRevision,
      savedAt: new Date().toISOString()
    };
    if (saveSavedSession(session)) {
      setHasSavedSession(true);
      setSavedSessionInfo(session);
    }
  }, [parsedState, serverGameID, serverRevision, showSetup]);

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

  function enterLoadedGame(nextState: GameState, gameID: string | null, revision: number | null, nextStatus: string) {
    syncState(nextState);
    setServerGameID(gameID);
    setServerRevision(revision);
    setShowSetup(false);
    setShowBoardEditor(false);
    setShowContextDrawer(nextState.gamePhase === 2);
    setShowWorkspaceTools(nextState.gamePhase !== 1);
    setShowRecoveryTools(false);
    setActiveModal(null);
    setShowGuideHelp(false);
    setStatus(nextStatus);
  }

  async function handleResumeSavedGame() {
    const savedSession = loadSavedSession();
    if (!savedSession) {
      throw new Error("No saved game found.");
    }

    const loaded = await loadSavedGame(savedSession);
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
        setStatus(err instanceof Error ? err.message : "Failed to load resumed setup actions");
      }
    }
  }

  return {
    showSetup,
    setShowSetup,
    setupScreen,
    setSetupScreen,
    hasSavedSession,
    setHasSavedSession,
    savedSessionInfo,
    setSavedSessionInfo,
    resetToSetup,
    enterLoadedGame,
    handleResumeSavedGame
  };
}
