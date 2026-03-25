import { useState } from "react";
import { setupGame } from "../api";
import { eyrieLeaderLabels, factionLabels, vagabondCharacterLabels } from "../labels";
import type { GameState } from "../types";

type SetupWizardProps = {
  onStart: (state: GameState) => void;
  onUseSample: () => void;
};

const allFactions = [0, 2, 1, 3];

export function SetupWizard({ onStart, onUseSample }: SetupWizardProps) {
  const [gameMode, setGameMode] = useState(0);
  const [playerFaction, setPlayerFaction] = useState(0);
  const [factions, setFactions] = useState<number[]>([0, 2, 1, 3]);
  const [vagabondCharacter, setVagabondCharacter] = useState(0);
  const [eyrieLeader, setEyrieLeader] = useState(0);
  const [status, setStatus] = useState("");
  const [submitting, setSubmitting] = useState(false);

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
      const state = await setupGame({
        gameMode,
        playerFaction,
        factions,
        mapID: "autumn",
        vagabondCharacter,
        eyrieLeader
      });
      onStart(state);
    } catch (error) {
      const message = error instanceof Error ? error.message : "Failed to create game";
      setStatus(message);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <main className="app-shell workspace-shell">
      <section className="panel modal-panel" style={{ maxWidth: 720, margin: "3rem auto" }}>
        <div className="panel-header">
          <h2>Setup Game</h2>
        </div>

        <div className="summary-stack">
          <span className="summary-label">Mode</span>
          <label>
            <input
              type="radio"
              checked={gameMode === 0}
              onChange={() => setGameMode(0)}
            />
            <span> Online</span>
          </label>
          <label>
            <input
              type="radio"
              checked={gameMode === 1}
              onChange={() => setGameMode(1)}
            />
            <span> Assist</span>
          </label>
        </div>

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
          <button type="button" className="secondary" onClick={onUseSample} disabled={submitting}>
            Use Sample State
          </button>
          <button type="button" onClick={handleStart} disabled={submitting}>
            Start Game
          </button>
        </div>
        <span className="message">{status || "Choose factions and create a new game."}</span>
      </section>
    </main>
  );
}
