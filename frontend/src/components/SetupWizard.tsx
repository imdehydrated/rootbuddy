import { useState, type CSSProperties } from "react";
import { setupGame } from "../api";
import { eyrieLeaderLabels, factionLabels, vagabondCharacterLabels } from "../labels";
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
const factionAccentColors = ["#b14d36", "#496aa0", "#4c7a45", "#8a6842"];

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
  const selectedFactionLabels = factions.map((faction) => factionLabels[faction]);

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
      <section
        className={`${selectedMode === null ? "landing-screen-panel" : "panel modal-panel pregame-modal-panel wide-entry-panel"} entry-panel`}
      >
        {selectedMode === null ? (
          <div className="landing-stack">
            <div className="landing-backdrop" aria-label="Landing screen">
              <div className="landing-backdrop-overlay">
                <div className="landing-command-shell">
                  <div className="landing-hero-copy">
                    <span className="landing-kicker">RootBuddy</span>
                    <h1 className="landing-title">Choose how you want to begin.</h1>
                    <p className="landing-copy">
                      Start an online lobby for remote play or launch assist mode for shared-table adjudication.
                    </p>
                  </div>

                  <div className="landing-grid pregame-mode-grid">
                    <button type="button" className="mode-choice-card pregame-mode-card online-mode-card" onClick={() => setSelectedMode(0)} disabled={submitting}>
                      Online Play
                    </button>
                    <button type="button" className="mode-choice-card secondary pregame-mode-card assist-mode-card" onClick={() => setSelectedMode(1)} disabled={submitting}>
                      Assist Mode
                    </button>
                  </div>

                  <span className="message landing-message">{status || "Choose Online Play or Assist Mode to continue."}</span>
                </div>
              </div>
            </div>
          </div>
        ) : (
          <>
            <div className="panel-header pregame-panel-header">
              <div>
                <p className="eyebrow">Pregame</p>
                <h2>{modeLabel}</h2>
              </div>
              <button type="button" className="secondary" onClick={() => setSelectedMode(null)} disabled={submitting}>
                Back
              </button>
            </div>

            <div className={`pregame-command-surface ${selectedMode === 1 ? "assist" : "online"}`}>
              <div className="pregame-command-hero">
                <div className="pregame-command-copy">
                  <span className="summary-label">Selected Mode</span>
                  <h3 className="pregame-command-title">{modeLabel}</h3>
                  <p className="pregame-command-copyline">{modeDescription}</p>
                </div>
                <div className="pregame-command-meta">
                  {selectedMode === 0 ? (
                    <>
                      <article className="pregame-meta-card">
                        <span className="summary-label">Flow</span>
                        <strong>Lobby-first</strong>
                        <span className="summary-line">Create a table, share the code, and claim faction seats before starting.</span>
                      </article>
                      <article className="pregame-meta-card">
                        <span className="summary-label">Reconnect</span>
                        <strong>Session-safe</strong>
                        <span className="summary-line">Players can rejoin multiplayer tables after disconnecting without local save-state setup.</span>
                      </article>
                    </>
                  ) : (
                    <>
                      <article className="pregame-meta-card">
                        <span className="summary-label">Factions Ready</span>
                        <strong>{factions.length} in the woodland</strong>
                        <span className="summary-line">{selectedFactionLabels.join(" / ")}</span>
                      </article>
                      <article className="pregame-meta-card">
                        <span className="summary-label">Perspective</span>
                        <strong>{factionLabels[playerFaction]}</strong>
                        <span className="summary-line">Configure the player seat and any faction-specific setup before launch.</span>
                      </article>
                    </>
                  )}
                </div>
              </div>

              {selectedMode === 0 ? (
                <div className="pregame-lobby-shell">
                  <div className="pregame-table-ribbon">
                    <span className="summary-label">Online Table</span>
                    <span className="summary-line">Choose whether you are opening the room or arriving with a code from the host.</span>
                  </div>
                  <div className="pregame-seat-grid">
                    <article className="pregame-seat-card host-seat">
                      <span className="summary-label">Host A Match</span>
                      <strong>Create Lobby</strong>
                      <span className="summary-line">Open a new online table, share the join code, and start when every seat is ready.</span>
                      <span className="pregame-seat-note">Best when you are organizing the table and want RootBuddy to generate the room code.</span>
                      <button type="button" className="secondary" onClick={onOpenCreateLobby} disabled={submitting}>
                        Create Lobby
                      </button>
                    </article>
                    <article className="pregame-seat-card join-seat">
                      <span className="summary-label">Join A Match</span>
                      <strong>Join Lobby</strong>
                      <span className="summary-line">Enter a join code to sit at an existing table and claim a faction.</span>
                      <span className="pregame-seat-note">Best when a host has already opened the room and you are taking a seat at their table.</span>
                      <button type="button" className="secondary" onClick={onOpenJoinLobby} disabled={submitting}>
                        Join Lobby
                      </button>
                    </article>
                  </div>
                </div>
              ) : null}

              {selectedMode === 1 ? (
                <>
                  <div className="pregame-selection-strip">
                    <span className="pregame-selection-pill">{factions.length} factions selected</span>
                    <span className="pregame-selection-pill">Perspective: {factionLabels[playerFaction]}</span>
                    <span className="pregame-selection-pill">Map: Autumn</span>
                  </div>

                  <div className="pregame-assist-grid">
                    <section className="summary-stack pregame-assist-config pregame-config-card">
                      <span className="summary-label">Factions In Game</span>
                      <div className="pregame-faction-grid">
                        {allFactions.map((faction) => {
                          const included = factions.includes(faction);
                          const style = {
                            "--faction-color": factionAccentColors[faction] ?? "#7a6045"
                          } as CSSProperties;

                          return (
                            <label
                              key={faction}
                              className={`pregame-faction-card ${included ? "selected" : ""}`}
                              style={style}
                            >
                              <input
                                type="checkbox"
                                checked={included}
                                onChange={() => toggleFaction(faction)}
                              />
                              <span className="pregame-faction-card-name">{factionLabels[faction]}</span>
                              <span className="pregame-faction-card-state">{included ? "Included" : "Not in game"}</span>
                            </label>
                          );
                        })}
                      </div>
                    </section>

                    <section className="summary-stack pregame-assist-config pregame-config-card">
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
                      <span className="summary-line">Set the point-of-view faction so RootBuddy can present the right prompts first.</span>
                    </section>

                    {factions.includes(2) ? (
                      <section className="summary-stack pregame-assist-config pregame-config-card">
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
                        <span className="summary-line">Choose the initial Eyrie leader so decree guidance starts from the right posture.</span>
                      </section>
                    ) : null}

                    {factions.includes(3) ? (
                      <section className="summary-stack pregame-assist-config pregame-config-card">
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
                        <span className="summary-line">Pick the starting Vagabond so item prompts and quests line up with the chosen character.</span>
                      </section>
                    ) : null}
                  </div>

                  <div className="pregame-action-strip">
                    <span className="summary-line">Choose the factions at the table, set any faction-specific setup choices, and then launch directly into assist mode.</span>
                    <div className="sidebar-actions footer">
                      <button type="button" className="secondary" onClick={onUseSample} disabled={submitting}>
                        Use Sample State
                      </button>
                      <button type="button" onClick={handleStart} disabled={submitting}>
                        Start Assist Game
                      </button>
                    </div>
                  </div>
                </>
              ) : (
                <span className="message pregame-status-message">{status || "Choose Create Lobby or Join Lobby to continue."}</span>
              )}
            </div>

            {selectedMode === 1 ? <span className="message pregame-status-message">{status || "Choose factions and continue."}</span> : null}
          </>
        )}
      </section>
    </main>
  );
}
