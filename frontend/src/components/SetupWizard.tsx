import { useState } from "react";
import { setupGame } from "../api";
import { eyrieLeaderLabels, factionLabels, phaseLabels, setupStageLabels, vagabondCharacterLabels } from "../labels";
import type { SavedSession } from "../localSession";
import type { GameState } from "../types";

type SetupWizardProps = {
  onStart: (state: GameState, gameID: string | null, revision: number | null) => void;
  onUseSample: () => void;
  onResume: () => Promise<void>;
  onClearSavedSession: () => void;
  onOpenCreateLobby: () => void;
  onOpenJoinLobby: () => void;
  canResume: boolean;
  savedSessionInfo: SavedSession | null;
};

const allFactions = [0, 2, 1, 3];

function formatSavedAt(savedAt: string | undefined): string {
  if (!savedAt) {
    return "";
  }

  const parsed = new Date(savedAt);
  if (Number.isNaN(parsed.getTime())) {
    return "";
  }

  return parsed.toLocaleString();
}

function savedSessionTitle(savedSession: SavedSession | null): string {
  if (!savedSession) {
    return "";
  }

  if (savedSession.state.gamePhase === 2) {
    return "Finished result ready to review";
  }
  if (savedSession.state.gamePhase === 0) {
    return "Setup in progress";
  }
  return "Active saved game";
}

function resumeButtonLabel(savedSession: SavedSession | null): string {
  if (!savedSession) {
    return "Resume Saved Game";
  }

  if (savedSession.state.gamePhase === 2) {
    return "Review Finished Game";
  }
  if (savedSession.state.gamePhase === 0) {
    return "Resume Setup";
  }
  return "Resume Saved Game";
}

function savedSessionDetail(savedSession: SavedSession | null): string[] {
  if (!savedSession) {
    return [];
  }

  const modeLabel = savedSession.state.gameMode === 0 ? "Online" : "Assist";
  const lines = [
    `Mode: ${modeLabel}`,
    `Perspective: ${factionLabels[savedSession.state.playerFaction] ?? "Unknown"}`
  ];

  if (savedSession.state.gamePhase === 2) {
    lines.push(`Winner: ${factionLabels[savedSession.state.winner] ?? "Unknown"}`);
  } else if (savedSession.state.gamePhase === 0) {
    lines.push(`Stage: ${setupStageLabels[savedSession.state.setupStage] ?? "Setup"}`);
  } else {
    lines.push(`Turn: ${factionLabels[savedSession.state.factionTurn] ?? "Unknown"} - ${phaseLabels[savedSession.state.currentPhase] ?? "Unknown"}`);
  }

  const savedAt = formatSavedAt(savedSession.savedAt);
  if (savedAt) {
    lines.push(`Saved: ${savedAt}`);
  }

  return lines;
}

export function SetupWizard({
  onStart,
  onUseSample,
  onResume,
  onClearSavedSession,
  onOpenCreateLobby,
  onOpenJoinLobby,
  canResume,
  savedSessionInfo
}: SetupWizardProps) {
  const [selectedMode, setSelectedMode] = useState<0 | 1 | null>(null);
  const [playerFaction, setPlayerFaction] = useState(0);
  const [factions, setFactions] = useState<number[]>([0, 2, 1, 3]);
  const [vagabondCharacter, setVagabondCharacter] = useState(0);
  const [eyrieLeader, setEyrieLeader] = useState(0);
  const [status, setStatus] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const savedFinishedResult = savedSessionInfo?.state.gamePhase === 2;

  function toggleFaction(faction: number) {
    setFactions((current) => {
      if (current.includes(faction)) {
        if (current.length <= 2) {
          return current;
        }
        const next = current.filter((value) => value !== faction);
        if (playerFaction === faction) {
          setPlayerFaction(next[0]);
        }
        return next;
      }

      if (current.length >= 4) {
        return current;
      }
      return [...current, faction];
    });
  }

  async function handleStart() {
    if (selectedMode === null) {
      setStatus("Choose Online Play or Assist Mode first.");
      return;
    }
    if (!factions.includes(playerFaction)) {
      setStatus("Your faction must be included in the game.");
      return;
    }
    if (factions.length < 2 || factions.length > 4) {
      setStatus("Choose between 2 and 4 factions.");
      return;
    }

    try {
      setSubmitting(true);
      setStatus("Creating game...");
      const result = await setupGame({
        gameMode: selectedMode,
        playerFaction,
        factions,
        mapID: "autumn",
        vagabondCharacter,
        eyrieLeader
      });
      onStart(result.state, result.gameID, result.revision);
    } catch (error) {
      const message = error instanceof Error ? error.message : "Failed to create game";
      setStatus(message);
    } finally {
      setSubmitting(false);
    }
  }

  async function handleResume() {
    try {
      setSubmitting(true);
      setStatus("Loading saved game...");
      await onResume();
    } catch (error) {
      const message = error instanceof Error ? error.message : "Failed to load saved game";
      setStatus(message);
    } finally {
      setSubmitting(false);
    }
  }

  const modeLabel = selectedMode === 1 ? "Assist Mode" : "Online Play";
  const modeDescription =
    selectedMode === 1
      ? "Use the companion flow for shared-table play, adjudication, and guided resolution."
      : "Start a standard online state or branch into a multiplayer lobby.";

  return (
    <main className="app-shell entry-shell">
      <section className="panel modal-panel entry-panel" style={{ maxWidth: 720 }}>
        {selectedMode === null ? (
          <>
            <div className="landing-hero">
              <p className="eyebrow">RootBuddy</p>
              <h1 className="landing-title">Root board state, multiplayer lobbies, and assist flow in one workspace.</h1>
              <p className="landing-copy">Choose how you want to begin. Online Play leads to standard online setup or multiplayer lobbies. Assist Mode keeps the adjudication companion flow.</p>
            </div>

            <div className="landing-grid">
              <button type="button" className="mode-choice-card" onClick={() => setSelectedMode(0)} disabled={submitting}>
                <span className="summary-label">Mode One</span>
                <strong>Online Play</strong>
                <span>Start a local online game, or branch into create/join lobby from the next step.</span>
              </button>
              <button type="button" className="mode-choice-card secondary" onClick={() => setSelectedMode(1)} disabled={submitting}>
                <span className="summary-label">Mode Two</span>
                <strong>Assist Mode</strong>
                <span>Use guided Root adjudication for in-person or shared-reference play.</span>
              </button>
            </div>

            {canResume && savedSessionInfo ? (
              <div className="saved-session-card">
                <span className="summary-label">{savedSessionTitle(savedSessionInfo)}</span>
                {savedSessionDetail(savedSessionInfo).map((line) => (
                  <span key={line} className="summary-line">
                    {line}
                  </span>
                ))}
                {savedFinishedResult ? (
                  <span className="summary-line">Use review to inspect the finished board, or start fresh to replace it.</span>
                ) : null}
                <div className="sidebar-actions footer">
                  <button type="button" className="secondary" onClick={handleResume} disabled={submitting}>
                    {resumeButtonLabel(savedSessionInfo)}
                  </button>
                  <button type="button" className="secondary" onClick={onClearSavedSession} disabled={submitting}>
                    Clear Saved Game
                  </button>
                </div>
              </div>
            ) : null}

            <div className="sidebar-actions footer">
              <button type="button" className="secondary" onClick={onUseSample} disabled={submitting}>
                Use Sample State
              </button>
            </div>
            <span className="message">{status || "Choose Online Play or Assist Mode to continue."}</span>
          </>
        ) : (
          <>
            <div className="panel-header">
              <div>
                <p className="eyebrow">Setup</p>
                <h2>{savedFinishedResult ? "Review or Start New Game" : modeLabel}</h2>
              </div>
              <button type="button" className="secondary" onClick={() => setSelectedMode(null)} disabled={submitting}>
                Back
              </button>
            </div>

            <div className="flow-guide-hero">
              <span className="summary-label">Selected Mode</span>
              <span className="summary-line">{modeDescription}</span>
            </div>

            {selectedMode === 0 ? (
              <div className="summary-stack">
                <span className="summary-label">Multiplayer Lobby</span>
                <div className="sidebar-actions">
                  <button type="button" className="secondary" onClick={onOpenCreateLobby} disabled={submitting}>
                    Create Lobby
                  </button>
                  <button type="button" className="secondary" onClick={onOpenJoinLobby} disabled={submitting}>
                    Join Lobby
                  </button>
                </div>
                <span className="summary-line">Or continue below to start a standard online game in this browser.</span>
              </div>
            ) : null}

            <div className="summary-stack">
              <span className="summary-label">Factions In Game</span>
              {allFactions.map((faction) => (
                <label key={faction}>
                  <input
                    type="checkbox"
                    checked={factions.includes(faction)}
                    onChange={() => toggleFaction(faction)}
                  />
                  <span> {factionLabels[faction]}</span>
                </label>
              ))}
            </div>

            <div className="summary-stack">
              <span className="summary-label">My Faction</span>
              <select
                value={playerFaction}
                onChange={(event) => setPlayerFaction(Number(event.target.value))}
              >
                {factions.map((faction) => (
                  <option key={faction} value={faction}>
                    {factionLabels[faction]}
                  </option>
                ))}
              </select>
            </div>

            {factions.includes(2) ? (
              <div className="summary-stack">
                <span className="summary-label">Eyrie Leader</span>
                <select
                  value={eyrieLeader}
                  onChange={(event) => setEyrieLeader(Number(event.target.value))}
                >
                  {eyrieLeaderLabels.map((label, index) => (
                    <option key={label} value={index}>
                      {label}
                    </option>
                  ))}
                </select>
              </div>
            ) : null}

            {factions.includes(3) ? (
              <div className="summary-stack">
                <span className="summary-label">Vagabond Character</span>
                <select
                  value={vagabondCharacter}
                  onChange={(event) => setVagabondCharacter(Number(event.target.value))}
                >
                  {vagabondCharacterLabels.map((label, index) => (
                    <option key={label} value={index}>
                      {label}
                    </option>
                  ))}
                </select>
              </div>
            ) : null}

            <div className="sidebar-actions footer">
              {canResume ? (
                <button type="button" className="secondary" onClick={handleResume} disabled={submitting}>
                  {resumeButtonLabel(savedSessionInfo)}
                </button>
              ) : null}
              {canResume ? (
                <button type="button" className="secondary" onClick={onClearSavedSession} disabled={submitting}>
                  Clear Saved Game
                </button>
              ) : null}
              <button type="button" className="secondary" onClick={onUseSample} disabled={submitting}>
                Use Sample State
              </button>
              <button type="button" onClick={handleStart} disabled={submitting}>
                {selectedMode === 1 ? "Start Assist Game" : "Start Online Game"}
              </button>
            </div>
            <span className="message">{status || "Choose factions and continue."}</span>
          </>
        )}
      </section>
    </main>
  );
}
