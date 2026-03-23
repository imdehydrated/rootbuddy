import { countBuildings, hasAnyIndicators, suitClass } from "../gameHelpers";
import { suitLabels } from "../labels";
import type { Clearing, HighlightedClearing } from "../types";

type ClearingMarkerProps = {
  clearing: Clearing;
  position: { left: string; top: string };
  isSelected: boolean;
  hasKeep: boolean;
  highlightRole?: HighlightedClearing["role"];
  onClick: () => void;
};

type TokenChipProps = {
  kind:
    | "marquise"
    | "eyrie"
    | "wood"
    | "sawmill"
    | "workshop"
    | "recruiter"
    | "enemy"
    | "keep"
    | "ruins";
  count?: number;
  label: string;
};

function TokenChip({ kind, count, label }: TokenChipProps) {
  return (
    <span className={`token-chip ${kind}`} aria-label={label} title={label}>
      <span className={`token-glyph ${kind}`} aria-hidden="true" />
      {count && count > 1 ? <span className="token-count">{count}</span> : null}
    </span>
  );
}

export function ClearingMarker({
  clearing,
  position,
  isSelected,
  hasKeep,
  highlightRole,
  onClick
}: ClearingMarkerProps) {
  const marquiseWarriors = clearing.warriors["0"] ?? 0;
  const eyrieWarriors = clearing.warriors["2"] ?? 0;
  const sawmills = countBuildings(clearing.buildings, 0, 0);
  const workshops = countBuildings(clearing.buildings, 0, 1);
  const recruiters = countBuildings(clearing.buildings, 0, 2);
  const enemyBuildings = countBuildings(clearing.buildings, 2);

  const classNames = ["clearing-marker", suitClass(clearing.suit)];
  if (isSelected) {
    classNames.push("selected");
  }
  if (highlightRole) {
    classNames.push(`highlight-${highlightRole}`);
  }

  return (
    <button
      type="button"
      className={classNames.join(" ")}
      style={position}
      onClick={onClick}
      aria-label={`Clearing ${clearing.id}`}
      title={`Clearing ${clearing.id}`}
    >
      <span className={`marker-suit-badge ${suitClass(clearing.suit)}`}>
        {suitLabels[clearing.suit] ?? "Unknown"}
      </span>
      <span className="marker-token-cluster marker-structures">
        {hasAnyIndicators([sawmills, workshops, recruiters, enemyBuildings], clearing.ruins || hasKeep) ? (
          <span className="indicator-row compact">
            {sawmills > 0 ? (
              <TokenChip kind="sawmill" count={sawmills} label={`Sawmills ${sawmills}`} />
            ) : null}
            {workshops > 0 ? (
              <TokenChip kind="workshop" count={workshops} label={`Workshops ${workshops}`} />
            ) : null}
            {recruiters > 0 ? (
              <TokenChip kind="recruiter" count={recruiters} label={`Recruiters ${recruiters}`} />
            ) : null}
            {enemyBuildings > 0 ? (
              <TokenChip
                kind="enemy"
                count={enemyBuildings}
                label={`Enemy buildings ${enemyBuildings}`}
              />
            ) : null}
            {hasKeep ? <TokenChip kind="keep" label="Keep" /> : null}
            {clearing.ruins ? <TokenChip kind="ruins" label="Ruins" /> : null}
          </span>
        ) : null}
      </span>
      <span className="marker-token-cluster marker-pieces">
        {hasAnyIndicators([marquiseWarriors, eyrieWarriors, clearing.wood], false) ? (
          <span className="indicator-row">
            {marquiseWarriors > 0 ? (
              <TokenChip
                kind="marquise"
                count={marquiseWarriors}
                label={`Marquise warriors ${marquiseWarriors}`}
              />
            ) : null}
            {eyrieWarriors > 0 ? (
              <TokenChip kind="eyrie" count={eyrieWarriors} label={`Eyrie warriors ${eyrieWarriors}`} />
            ) : null}
            {clearing.wood > 0 ? (
              <TokenChip kind="wood" count={clearing.wood} label={`Wood ${clearing.wood}`} />
            ) : null}
          </span>
        ) : null}
      </span>
    </button>
  );
}
