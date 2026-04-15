import { useEffect, useState, type Dispatch, type SetStateAction } from "react";
import { clearSavedSession, type SavedSession } from "../localSession";
import type { ActiveModal } from "../modalState";
import type { Action, GameState } from "../types";
import type { SetupScreen } from "../multiplayer";

type UseSessionPersistenceOptions = {
  getMultiplayerToken: () => string | null;
  loadSavedGame: (savedSession: SavedSession) => Promise<{ state: GameState; gameID: string | null; revision: number | null }>;
  loadActionsForState: (state: GameState, options?: { successStatus?: string }) => Promise<Action[]>;
  parsedState: GameState;
  serverGameID: string | null;
  serverRevision: number | null;
  setActiveModal: Dispatch<SetStateAction<ActiveModal>>;
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
  const [showSetup, setShowSetup] = useState(true);
  const [setupScreen, setSetupScreen] = useState<SetupScreen>("wizard");
  const [hasSavedSession, setHasSavedSession] = useState(false);
  const [savedSessionInfo, setSavedSessionInfo] = useState<SavedSession | null>(null);

  useEffect(() => {
    // Local assist/offline autosave is intentionally disabled.
    clearSavedSession();
    setHasSavedSession(false);
    setSavedSessionInfo(null);
  }, []);

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
    throw new Error("Local saved-game resume is no longer supported. Multiplayer reconnect still resumes automatically in this browser.");
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
