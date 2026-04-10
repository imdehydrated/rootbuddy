import type { CSSProperties } from "react";
import { describeKnownCardID } from "../cardCatalog";
import { factionLabels } from "../labels";

type VPTrackerProps = {
  victoryPoints: Record<string, number>;
  turnOrder: number[];
  dominance: Record<string, number>;
  coalitionActive: boolean;
  coalitionPartner: number;
};

const factionAccentColors = ["#b14d36", "#4c7a45", "#496aa0", "#8a6842"];

function factionAccentColor(faction: number) {
  return factionAccentColors[faction] ?? "#7a6045";
}

function orderedFactions(turnOrder: number[]) {
  const valid = turnOrder.filter((faction, index, source) => factionLabels[faction] && source.indexOf(faction) === index);
  const missing = factionLabels
    .map((_, index) => index)
    .filter((faction) => !valid.includes(faction));
  return [...valid, ...missing];
}

export function VPTracker({
  victoryPoints,
  turnOrder,
  dominance,
  coalitionActive,
  coalitionPartner
}: VPTrackerProps) {
  return (
    <section className="vp-tracker" aria-label="Victory point tracker">
      <div className="vp-tracker-header">
        <span className="phase-bar-kicker">Score Track</span>
        <strong>Victory Points</strong>
        <span>First to 30 wins unless dominance changes the condition.</span>
      </div>
      <div className="vp-tracker-row">
        {orderedFactions(turnOrder).map((faction) => {
          const score = Math.max(0, victoryPoints[String(faction)] ?? 0);
          const dominanceCardID = dominance[String(faction)];
          const isVagabondCoalition = coalitionActive && faction === 3;
          const isCoalitionPartner = coalitionActive && faction === coalitionPartner;
          const style = {
            "--faction-color": factionAccentColor(faction),
            "--vp-progress": `${Math.min(score, 30) / 30 * 100}%`
          } as CSSProperties;

          return (
            <article
              key={faction}
              className={`vp-tracker-segment ${isCoalitionPartner ? "coalition-partner" : ""} ${isVagabondCoalition ? "coalition-vagabond" : ""}`.trim()}
              style={style}
            >
              <div className="vp-tracker-segment-top">
                <span className="vp-tracker-label">{factionLabels[faction] ?? `Faction ${faction}`}</span>
                <strong>{score}</strong>
              </div>
              <div className="vp-tracker-progress" aria-hidden="true">
                <span className="vp-tracker-fill" />
              </div>
              <span className="vp-tracker-total">{score} / 30 VP</span>
              {typeof dominanceCardID === "number" ? (
                <span className="vp-tracker-note">Dominance: {describeKnownCardID(dominanceCardID)}</span>
              ) : null}
              {isCoalitionPartner ? <span className="vp-tracker-note">Coalition partner</span> : null}
              {isVagabondCoalition ? <span className="vp-tracker-note">Coalition active</span> : null}
            </article>
          );
        })}
      </div>
    </section>
  );
}
