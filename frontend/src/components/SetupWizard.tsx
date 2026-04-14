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
      : "Create or join a multiplayer lobby for a normal in-browser game.";

  return (
    <main className="app-shell entry-shell">
      <section className="panel modal-panel entry-panel" style={{ maxWidth: 720 }}>
        {selectedMode === null ? (
          <>
            <div className="landing-hero pregame-hero">
              <div className="landing-hero-copy">
                <p className="eyebrow">Woodland Command Desk</p>
                <span className="landing-kicker">RootBuddy</span>
                <h1 className="landing-title">Take your seat at the woodland table.</h1>
                <p className="landing-copy">
                  Run a live browser match, reopen a suspended board, or use the adjudication companion for shared-table play.
                  RootBuddy keeps the map, turns, and setup flow in one place instead of scattering them across notes and chat.
                </p>
              </div>
              <div className="landing-feature-row" aria-label="Available entry points">
                <article className="landing-feature">
                  <span className="summary-label">Online Tables</span>
                  <strong>Live lobby flow</strong>
                  <span>Host a match, pass around a join code, and keep every faction on the same stateful board.</span>
                </article>
                <article className="landing-feature">
                  <span className="summary-label">Assist Play</span>
                  <strong>Shared reference</strong>
                  <span>Guide setup, action resolution, and board corrections when the physical game needs a rules-aware companion.</span>
                </article>
                <article className="landing-feature">
                  <span className="summary-label">Saved Boards</span>
                  <strong>Resume without guesswork</strong>
                  <span>Return to unfinished setups and finished results with the same perspective and turn context intact.</span>
                </article>
              </div>
            </div>

            <div className="landing-grid pregame-mode-grid">
              <button type="button" className="mode-choice-card pregame-mode-card online-mode-card" onClick={() => setSelectedMode(0)} disabled={submitting}>
                <span className="mode-choice-ribbon">Digital Match</span>
                <strong>Online Play</strong>
                <span className="mode-choice-flavor">Open a browser-native Root table for remote players.</span>
                <div className="mode-choice-details">
                  <span>Lobby hosting and join codes</span>
                  <span>Seat claiming, ready checks, and shared board state</span>
                </div>
              </button>
              <button type="button" className="mode-choice-card secondary pregame-mode-card assist-mode-card" onClick={() => setSelectedMode(1)} disabled={submitting}>
                <span className="mode-choice-ribbon">Table Companion</span>
                <strong>Assist Mode</strong>
                <span className="mode-choice-flavor">Keep the board honest during in-person or shared-reference play.</span>
                <div className="mode-choice-details">
                  <span>Faction setup and action prompting</span>
                  <span>Guided resolution when the table needs a neutral rules helper</span>
                </div>
              </button>
            </div>

            {canResume && savedSessionInfo ? (
              <div className="saved-session-card pregame-saved-card">
                <div className="saved-session-head">
                  <div>
                    <span className="summary-label">Saved Chronicle</span>
                    <strong>{savedSessionTitle(savedSessionInfo)}</strong>
                  </div>
                  <span className="saved-session-badge">{savedFinishedResult ? "Review" : "Resume"}</span>
                </div>
                <div className="saved-session-details">
                  {savedSessionDetail(savedSessionInfo).map((line) => (
                    <span key={line} className="summary-line">
                      {line}
                    </span>
                  ))}
                </div>
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

            <div className="sidebar-actions footer landing-utility-actions">
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
                <p className="eyebrow">Pregame</p>
                <h2>{savedFinishedResult ? "Review or Start New Game" : modeLabel}</h2>
              </div>
              <button type="button" className="secondary" onClick={() => setSelectedMode(null)} disabled={submitting}>
                Back
              </button>
            </div>

            <div className="flow-guide-hero pregame-selected-mode">
              <span className="summary-label">Selected Mode</span>
              <span className="summary-line">{modeDescription}</span>
            </div>

            {selectedMode === 0 ? (
              <div className="pregame-lobby-shell">
                <div className="pregame-seat-grid">
                  <article className="pregame-seat-card host-seat">
                    <span className="summary-label">Host A Match</span>
                    <strong>Create Lobby</strong>
                    <span className="summary-line">Open a new online table, share the join code, and start when every seat is ready.</span>
                    <button type="button" className="secondary" onClick={onOpenCreateLobby} disabled={submitting}>
                      Create Lobby
                    </button>
                  </article>
                  <article className="pregame-seat-card join-seat">
                    <span className="summary-label">Join A Match</span>
                    <strong>Join Lobby</strong>
                    <span className="summary-line">Enter a join code to sit at an existing table and claim a faction.</span>
                    <button type="button" className="secondary" onClick={onOpenJoinLobby} disabled={submitting}>
                      Join Lobby
                    </button>
                  </article>
                </div>
                <span className="summary-line">Online play is lobby-based. Create a table or join one that already exists.</span>
              </div>
            ) : null}

            {selectedMode === 1 ? (
              <>
                <div className="summary-stack pregame-assist-config">
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

                <div className="summary-stack pregame-assist-config">
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
                  <div className="summary-stack pregame-assist-config">
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
                  <div className="summary-stack pregame-assist-config">
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
                    Start Assist Game
                  </button>
                </div>
                <span className="message">{status || "Choose factions and continue."}</span>
              </>
            ) : (
              <span className="message">{status || "Choose Create Lobby or Join Lobby to continue."}</span>
            )}
          </>
        )}
      </section>
    </main>
  );
}
