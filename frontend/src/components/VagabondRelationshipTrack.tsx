import { factionLabels, relationshipLabels } from "../labels";
import type { GameState } from "../types";

type VagabondRelationshipTrackProps = {
  state: GameState;
  editable?: boolean;
  onSetRelationship?: (faction: number, relationship: number) => void;
};

const nonVagabondFactions = [0, 1, 2];
const alliedTrack = [1, 2, 3, 4];
const aidCosts: Record<number, number> = {
  1: 1,
  2: 2,
  3: 3
};
const relationshipRewards: Record<number, string> = {
  2: "+1 VP",
  3: "+2 VP",
  4: "+2 VP"
};

function aidProgressLabel(relationship: number, progress: number): string {
  if (relationship === 0) {
    return "Hostile Aid allowed; relationship stays Hostile.";
  }
  if (relationship === 4) {
    return "Allied Aid scores 2 VP.";
  }

  const cost = aidCosts[relationship] ?? 0;
  const nextRelationship = relationship + 1;
  if (!cost || !relationshipLabels[nextRelationship]) {
    return "No Aid progress.";
  }

  return `${progress}/${cost} Aid this turn toward ${relationshipLabels[nextRelationship]}.`;
}

function alliedWarriorsHere(state: GameState, faction: number): number {
  if (state.vagabond.inForest) {
    return 0;
  }
  const clearing = state.map.clearings.find((candidate) => candidate.id === state.vagabond.clearingID);
  return clearing?.warriors?.[String(faction)] ?? 0;
}

export function VagabondRelationshipTrack({ state, editable = false, onSetRelationship }: VagabondRelationshipTrackProps) {
  return (
    <div className="relationship-track-list">
      {nonVagabondFactions.map((faction) => {
        const relationship = state.vagabond.relationships[String(faction)] ?? 1;
        const progress = state.turnProgress.vagabondAidCounts?.[String(faction)] ?? 0;
        const alliedWarriors = relationship === 4 ? alliedWarriorsHere(state, faction) : 0;

        return (
          <div className="relationship-track-row" key={faction}>
            <div className={`relationship-hostile ${relationship === 0 ? "active" : ""}`}>
              {editable ? (
                <button type="button" onClick={() => onSetRelationship?.(faction, 0)}>
                  Hostile
                </button>
              ) : (
                <span>Hostile</span>
              )}
            </div>
            <div className="relationship-track-main">
              <div className="relationship-track-heading">
                <strong>{factionLabels[faction] ?? `Faction ${faction}`}</strong>
                <span>{aidProgressLabel(relationship, progress)}</span>
              </div>
              {relationship === 4 ? (
                <div className={`relationship-ally-command ${alliedWarriors > 0 ? "ready" : ""}`}>
                  <strong>{alliedWarriors}</strong>
                  <span>{alliedWarriors === 1 ? "warrior" : "warriors"} with the Vagabond</span>
                  <small>{alliedWarriors > 0 ? "Move or battle with this ally" : "No allied warriors in this clearing"}</small>
                </div>
              ) : null}
              <div className="relationship-nodes" aria-label={`${factionLabels[faction]} relationship`}>
                {alliedTrack.map((level) => {
                  const active = relationship === level;
                  const node = (
                    <>
                      <span>{relationshipLabels[level] ?? `Level ${level}`}</span>
                      <small>{relationshipRewards[level] ?? "0 VP"}</small>
                    </>
                  );

                  return editable ? (
                    <button
                      className={`relationship-node ${active ? "active" : ""}`}
                      key={level}
                      type="button"
                      onClick={() => onSetRelationship?.(faction, level)}
                    >
                      {node}
                    </button>
                  ) : (
                    <span className={`relationship-node ${active ? "active" : ""}`} key={level}>
                      {node}
                    </span>
                  );
                })}
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
}
