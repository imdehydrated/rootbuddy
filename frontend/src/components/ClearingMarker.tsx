import { countBuildings, countTokens, hasAnyIndicators, suitClass } from "../gameHelpers";
import { suitLabels } from "../labels";
import type { Clearing, HighlightedClearing } from "../types";

type ClearingMarkerProps = {
  clearing: Clearing;
  position: { left: string; top: string };
  isSelected: boolean;
  isFocused?: boolean;
  isAdjacentToFocus?: boolean;
  isDimmed?: boolean;
  hasKeep: boolean;
  hasVagabond: boolean;
  highlightRole?: HighlightedClearing["role"];
  isSetupLegal?: boolean;
  isSetupChosen?: boolean;
  onClick: () => void;
  onHover?: (hovered: boolean) => void;
};

type TokenChipProps = {
  kind:
    | "marquise"
    | "alliance"
    | "eyrie"
    | "vagabond"
    | "wood"
    | "sawmill"
    | "workshop"
    | "recruiter"
    | "roost"
    | "base"
    | "sympathy"
    | "keep"
    | "ruins";
  count?: number;
  label: string;
};

type TokenChipDatum = TokenChipProps;

function TokenChip({ kind, count, label }: TokenChipProps) {
  return (
    <span className={`token-chip ${kind}`} aria-label={label} title={label}>
      <span className={`token-glyph ${kind}`} aria-hidden="true" />
      {count && count > 1 ? <span className="token-count">{count}</span> : null}
    </span>
  );
}

function renderTokenRow(chips: TokenChipDatum[], maxVisible: number, compact = false) {
  if (chips.length === 0) {
    return null;
  }

  const visible = chips.slice(0, maxVisible);
  const hidden = chips.slice(maxVisible);
  const hiddenCount = hidden.length;

  return (
    <span className={`indicator-row ${compact ? "compact" : ""}`}>
      {visible.map((chip) => (
        <TokenChip key={`${chip.kind}-${chip.label}`} kind={chip.kind} count={chip.count} label={chip.label} />
      ))}
      {hiddenCount > 0 ? (
        <span
          className="marker-overflow-chip"
          aria-label={`${hiddenCount} more clearing indicators`}
          title={hidden.map((chip) => chip.label).join(", ")}
        >
          +{hiddenCount}
        </span>
      ) : null}
    </span>
  );
}

export function ClearingMarker({
  clearing,
  position,
  isSelected,
  isFocused = false,
  isAdjacentToFocus = false,
  isDimmed = false,
  hasKeep,
  hasVagabond,
  highlightRole,
  isSetupLegal = false,
  isSetupChosen = false,
  onClick,
  onHover
}: ClearingMarkerProps) {
  const marquiseWarriors = clearing.warriors["0"] ?? 0;
  const allianceWarriors = clearing.warriors["1"] ?? 0;
  const eyrieWarriors = clearing.warriors["2"] ?? 0;
  const sawmills = countBuildings(clearing.buildings, 0, 0);
  const workshops = countBuildings(clearing.buildings, 0, 1);
  const recruiters = countBuildings(clearing.buildings, 0, 2);
  const roosts = countBuildings(clearing.buildings, 2, 3);
  const allianceBases = countBuildings(clearing.buildings, 1, 4);
  const sympathy = countTokens(clearing.tokens, 1, 1);
  const structureChips: TokenChipDatum[] = [
    sawmills > 0 ? { kind: "sawmill", count: sawmills, label: `Sawmills ${sawmills}` } : null,
    workshops > 0 ? { kind: "workshop", count: workshops, label: `Workshops ${workshops}` } : null,
    recruiters > 0 ? { kind: "recruiter", count: recruiters, label: `Recruiters ${recruiters}` } : null,
    roosts > 0 ? { kind: "roost", count: roosts, label: `Roosts ${roosts}` } : null,
    allianceBases > 0 ? { kind: "base", count: allianceBases, label: `Alliance bases ${allianceBases}` } : null,
    sympathy > 0 ? { kind: "sympathy", count: sympathy, label: `Sympathy ${sympathy}` } : null,
    hasKeep ? { kind: "keep", label: "Keep" } : null,
    clearing.ruins ? { kind: "ruins", label: "Ruins" } : null
  ].filter((chip): chip is TokenChipDatum => chip !== null);
  const pieceChips: TokenChipDatum[] = [
    marquiseWarriors > 0 ? { kind: "marquise", count: marquiseWarriors, label: `Marquise warriors ${marquiseWarriors}` } : null,
    allianceWarriors > 0 ? { kind: "alliance", count: allianceWarriors, label: `Alliance warriors ${allianceWarriors}` } : null,
    eyrieWarriors > 0 ? { kind: "eyrie", count: eyrieWarriors, label: `Eyrie warriors ${eyrieWarriors}` } : null,
    hasVagabond ? { kind: "vagabond", label: "Vagabond" } : null,
    clearing.wood > 0 ? { kind: "wood", count: clearing.wood, label: `Wood ${clearing.wood}` } : null
  ].filter((chip): chip is TokenChipDatum => chip !== null);
  const denseStructures = structureChips.length > 4;
  const densePieces = pieceChips.length > 4;
  const denseMarker = denseStructures || densePieces;

  const classNames = ["clearing-marker", suitClass(clearing.suit)];
  if (isSelected) {
    classNames.push("selected");
  }
  if (isFocused) {
    classNames.push("focused");
  }
  if (isAdjacentToFocus) {
    classNames.push("adjacent-focus");
  }
  if (isDimmed) {
    classNames.push("dimmed");
  }
  if (denseMarker) {
    classNames.push("dense-marker");
  }
  if (highlightRole) {
    classNames.push(`highlight-${highlightRole}`);
  }
  if (isSetupLegal) {
    classNames.push("setup-legal");
  }
  if (isSetupChosen) {
    classNames.push("setup-chosen");
  }

  return (
    <button
      type="button"
      className={classNames.join(" ")}
      style={position}
      onClick={onClick}
      onMouseEnter={() => onHover?.(true)}
      onMouseLeave={() => onHover?.(false)}
      onFocus={() => onHover?.(true)}
      onBlur={() => onHover?.(false)}
      aria-label={`Clearing ${clearing.id}`}
      title={`Clearing ${clearing.id}`}
    >
      <span className="marker-clearing-id">{clearing.id}</span>
      <span className={`marker-suit-badge ${suitClass(clearing.suit)}`}>
        {suitLabels[clearing.suit] ?? "Unknown"}
      </span>
      <span className="marker-token-cluster marker-structures">
        {hasAnyIndicators([sawmills, workshops, recruiters, roosts, allianceBases, sympathy], clearing.ruins || hasKeep)
          ? renderTokenRow(structureChips, denseStructures ? 3 : 5, true)
          : null}
      </span>
      <span className="marker-token-cluster marker-pieces">
        {hasAnyIndicators([marquiseWarriors, allianceWarriors, eyrieWarriors, clearing.wood], hasVagabond)
          ? renderTokenRow(pieceChips, densePieces ? 3 : 5)
          : null}
      </span>
      <span className="marker-footer">
        <span>{clearing.adj.length} paths</span>
        <span>{clearing.buildings.length}/{clearing.buildSlots} slots</span>
      </span>
    </button>
  );
}
