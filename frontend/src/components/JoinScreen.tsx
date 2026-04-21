import { useState } from "react";
import { factionLabels } from "../labels";

type JoinScreenProps = {
  mode: "create" | "join";
  submitting: boolean;
  status: string;
  onBack: () => void;
  onCreateLobby: (request: { displayName: string; factions: number[] }) => Promise<void>;
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
        factions
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
