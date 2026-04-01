import { useState } from "react";
import { eyrieLeaderLabels, factionLabels, vagabondCharacterLabels } from "../labels";

type JoinScreenProps = {
  mode: "create" | "join";
  submitting: boolean;
  status: string;
  onBack: () => void;
  onCreateLobby: (request: {
    displayName: string;
    factions: number[];
    eyrieLeader: number;
    vagabondCharacter: number;
  }) => Promise<void>;
  onJoinLobby: (request: { displayName: string; joinCode: string }) => Promise<void>;
};

const allFactions = [0, 2, 1, 3];

export function JoinScreen({
  mode,
  submitting,
  status,
  onBack,
  onCreateLobby,
  onJoinLobby
}: JoinScreenProps) {
  const [displayName, setDisplayName] = useState("");
  const [joinCode, setJoinCode] = useState("");
  const [factions, setFactions] = useState<number[]>([0, 2, 1, 3]);
  const [vagabondCharacter, setVagabondCharacter] = useState(0);
  const [eyrieLeader, setEyrieLeader] = useState(0);

  function toggleFaction(faction: number) {
    setFactions((current) => {
      if (current.includes(faction)) {
        if (current.length <= 2) {
          return current;
        }
        return current.filter((value) => value !== faction);
      }
      if (current.length >= 4) {
        return current;
      }
      return [...current, faction];
    });
  }

  async function handleSubmit() {
    if (mode === "create") {
      await onCreateLobby({
        displayName,
        factions,
        eyrieLeader,
        vagabondCharacter
      });
      return;
    }

    await onJoinLobby({
      displayName,
      joinCode: joinCode.toUpperCase()
    });
  }

  return (
    <main className="app-shell entry-shell">
      <section className="panel modal-panel multiplayer-screen entry-panel">
        <div className="panel-header">
          <h2>{mode === "create" ? "Create Multiplayer Lobby" : "Join Multiplayer Lobby"}</h2>
        </div>

        <div className="summary-stack">
          <span className="summary-label">Display Name</span>
          <input value={displayName} onChange={(event) => setDisplayName(event.target.value)} placeholder="Enter your name" />
        </div>

        {mode === "join" ? (
          <div className="summary-stack">
            <span className="summary-label">Join Code</span>
            <input
              value={joinCode}
              onChange={(event) => setJoinCode(event.target.value.toUpperCase())}
              placeholder="ABC123"
              maxLength={6}
            />
          </div>
        ) : (
          <>
            <div className="summary-stack">
              <span className="summary-label">Factions In Lobby</span>
              <div className="multiplayer-choice-grid">
                {allFactions.map((faction) => (
                  <label key={faction} className="checkbox">
                    <input type="checkbox" checked={factions.includes(faction)} onChange={() => toggleFaction(faction)} />
                    <span>{factionLabels[faction]}</span>
                  </label>
                ))}
              </div>
            </div>

            {factions.includes(2) ? (
              <div className="summary-stack">
                <span className="summary-label">Eyrie Leader</span>
                <select value={eyrieLeader} onChange={(event) => setEyrieLeader(Number(event.target.value))}>
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
                <select value={vagabondCharacter} onChange={(event) => setVagabondCharacter(Number(event.target.value))}>
                  {vagabondCharacterLabels.map((label, index) => (
                    <option key={label} value={index}>
                      {label}
                    </option>
                  ))}
                </select>
              </div>
            ) : null}
          </>
        )}

        <div className="sidebar-actions footer">
          <button type="button" className="secondary" onClick={onBack} disabled={submitting}>
            Back
          </button>
          <button type="button" onClick={() => void handleSubmit()} disabled={submitting}>
            {mode === "create" ? "Create Lobby" : "Join Lobby"}
          </button>
        </div>
        <span className="message">{status || (mode === "create" ? "Create a live multiplayer lobby." : "Join with a code and reconnect later from this browser.")}</span>
      </section>
    </main>
  );
}
