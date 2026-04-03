import { ClearingMarker } from "./ClearingMarker";
import { actionTouchesClearing, countBuildings, countTokens, countWarriors, clearingPosition, rulerOfClearing, rulingPresence, suitClass } from "../gameHelpers";
import { ACTION_TYPE, describeAction, factionLabels, suitLabels } from "../labels";
import { useState } from "react";
import type { BoardLayout } from "../boardLayouts";
import type { Action, Clearing, Forest, HighlightedClearing } from "../types";

type BoardPanelProps = {
  clearings: Clearing[];
  forests: Forest[];
  boardLayout: BoardLayout;
  selectedClearingID: number;
  keepClearingID: number;
  vagabondClearingID: number;
  vagabondInForest: boolean;
  highlightedClearings?: HighlightedClearing[];
  actions?: Action[];
  previewedAction?: Action | null;
  setupLegalClearingIDs?: number[];
  setupSelectedClearingIDs?: number[];
  forestTargets?: Array<{
    forestID: number;
    label: string;
    legal: boolean;
    selected: boolean;
  }>;
  onSelectClearing: (clearingID: number) => void;
  onSelectForest?: (forestID: number) => void;
};

export function BoardPanel({
  clearings,
  forests,
  boardLayout,
  selectedClearingID,
  keepClearingID,
  vagabondClearingID,
  vagabondInForest,
  highlightedClearings = [],
  actions = [],
  previewedAction = null,
  setupLegalClearingIDs = [],
  setupSelectedClearingIDs = [],
  forestTargets = [],
  onSelectClearing
  ,
  onSelectForest
}: BoardPanelProps) {
  const [hoveredClearingID, setHoveredClearingID] = useState<number | null>(null);
  const highlightByClearing = new Map(
    highlightedClearings.map((highlight) => [highlight.clearingID, highlight.role])
  );
  const legalSetupClearings = new Set(setupLegalClearingIDs);
  const selectedSetupClearings = new Set(setupSelectedClearingIDs);
  const selectedClearing = clearings.find((clearing) => clearing.id === selectedClearingID) ?? clearings[0] ?? null;
  const focusedClearing =
    clearings.find((clearing) => clearing.id === hoveredClearingID) ??
    selectedClearing;
  const focusedClearingID = focusedClearing?.id ?? null;

  const forestPosition = (forestID: number) => {
    const forest = forests.find((entry) => entry.id === forestID);
    if (forest && forest.adjacentClearings.length > 0) {
      const adjacentPositions = forest.adjacentClearings
        .map((clearingID) => boardLayout.clearingPositions[clearingID])
        .filter((position): position is NonNullable<typeof position> => position !== undefined);

      if (adjacentPositions.length > 0) {
        const totals = adjacentPositions.reduce(
          (sum, position) => ({
            left: sum.left + position.left,
            top: sum.top + position.top
          }),
          { left: 0, top: 0 }
        );

        return {
          left: totals.left / adjacentPositions.length,
          top: totals.top / adjacentPositions.length
        };
      }
    }

    return boardLayout.forestPositions[forestID] ?? null;
  };

  const adjacencySegments = clearings.flatMap((clearing) =>
    clearing.adj
      .filter((adjacentID) => clearing.id < adjacentID)
      .map((adjacentID) => {
        const from = boardLayout.clearingPositions[clearing.id];
        const to = boardLayout.clearingPositions[adjacentID];
        if (!from || !to) {
          return null;
        }

        return {
          key: `${clearing.id}-${adjacentID}`,
          x1: from.left,
          y1: from.top,
          x2: to.left,
          y2: to.top,
          connectedToFocused:
            focusedClearingID !== null &&
            (clearing.id === focusedClearingID || adjacentID === focusedClearingID)
        };
      })
      .filter((segment): segment is NonNullable<typeof segment> => segment !== null)
  );

  const focusedClearingPresence =
    focusedClearing === null
      ? []
      : [
          {
            key: "marquise",
            label: factionLabels[0] ?? "Marquise",
            total:
              countWarriors(focusedClearing.warriors, 0) +
              countBuildings(focusedClearing.buildings, 0) +
              (focusedClearing.id === keepClearingID ? 1 : 0),
            detail: `${countWarriors(focusedClearing.warriors, 0)} warriors, ${countBuildings(focusedClearing.buildings, 0)} buildings${focusedClearing.id === keepClearingID ? ", keep" : ""}`
          },
          {
            key: "alliance",
            label: factionLabels[1] ?? "Woodland Alliance",
            total: countWarriors(focusedClearing.warriors, 1) + countBuildings(focusedClearing.buildings, 1) + countTokens(focusedClearing.tokens, 1),
            detail: `${countWarriors(focusedClearing.warriors, 1)} warriors, ${countBuildings(focusedClearing.buildings, 1)} bases, ${countTokens(focusedClearing.tokens, 1)} tokens`
          },
          {
            key: "eyrie",
            label: factionLabels[2] ?? "Eyrie",
            total: countWarriors(focusedClearing.warriors, 2) + countBuildings(focusedClearing.buildings, 2),
            detail: `${countWarriors(focusedClearing.warriors, 2)} warriors, ${countBuildings(focusedClearing.buildings, 2)} roosts`
          },
          {
            key: "vagabond",
            label: factionLabels[3] ?? "Vagabond",
            total: !vagabondInForest && focusedClearing.id === vagabondClearingID ? 1 : 0,
            detail: !vagabondInForest && focusedClearing.id === vagabondClearingID ? "Present in clearing" : "Not present"
          }
        ].filter((entry) => entry.total > 0);
  const focusedRuler = focusedClearing ? rulerOfClearing(focusedClearing) : null;
  const adjacentClearings =
    focusedClearing === null
      ? []
      : focusedClearing.adj
          .map((adjacentID) => clearings.find((clearing) => clearing.id === adjacentID))
          .filter((clearing): clearing is Clearing => clearing !== undefined);
  const boardControlSummary = [
    { key: "marquise", label: factionLabels[0] ?? "Marquise", count: clearings.filter((clearing) => rulerOfClearing(clearing) === 0).length },
    { key: "alliance", label: factionLabels[1] ?? "Woodland Alliance", count: clearings.filter((clearing) => rulerOfClearing(clearing) === 1).length },
    { key: "eyrie", label: factionLabels[2] ?? "Eyrie", count: clearings.filter((clearing) => rulerOfClearing(clearing) === 2).length },
    { key: "contested", label: "Unruled", count: clearings.filter((clearing) => rulerOfClearing(clearing) === null).length }
  ];
  const focusedClearingActions =
    focusedClearing === null
      ? []
      : actions
          .map((action, index) => ({ action, index }))
          .filter(({ action }) => actionTouchesClearing(action, focusedClearing.id));
  const previewedClearingIDs = new Set(highlightedClearings.map((highlight) => highlight.clearingID));
  const previewFootprint = highlightedClearings.reduce(
    (summary, highlight) => {
      summary.total += 1;
      if (highlight.role === "source") {
        summary.source += 1;
      } else if (highlight.role === "target") {
        summary.target += 1;
      } else {
        summary.affected += 1;
      }
      return summary;
    },
    { source: 0, target: 0, affected: 0, total: 0 }
  );

  const clearingBoardPosition = (clearingID: number) => {
    const known = boardLayout.clearingPositions[clearingID];
    if (known) {
      return known;
    }

    const index = clearings.findIndex((clearing) => clearing.id === clearingID);
    if (index < 0) {
      return null;
    }

    const fallback = clearingPosition(clearingID, index, boardLayout.clearingPositions);
    return {
      left: Number.parseFloat(fallback.left),
      top: Number.parseFloat(fallback.top)
    };
  };

  const previewOverlayTitle = (() => {
    if (!previewedAction) {
      return null;
    }

    switch (previewedAction.type) {
      case ACTION_TYPE.MOVEMENT:
      case ACTION_TYPE.SLIP:
        return "Route Preview";
      case ACTION_TYPE.BATTLE:
      case ACTION_TYPE.BATTLE_RESOLUTION:
        return "Conflict Preview";
      case ACTION_TYPE.BUILD:
        return "Build Preview";
      case ACTION_TYPE.CRAFT:
        return "Craft Preview";
      case ACTION_TYPE.RECRUIT:
        return "Recruit Preview";
      default:
        return "Action Preview";
    }
  })();

  const previewOverlayDetail = (() => {
    if (!previewedAction) {
      return null;
    }

    switch (previewedAction.type) {
      case ACTION_TYPE.MOVEMENT:
        return `Moving from clearing ${previewedAction.movement?.from ?? "?"} to clearing ${previewedAction.movement?.to ?? "?"}.`;
      case ACTION_TYPE.SLIP:
        return `Slipping from ${previewedAction.slip?.fromForestID ? `forest ${previewedAction.slip.fromForestID}` : `clearing ${previewedAction.slip?.from ?? "?"}`} to ${previewedAction.slip?.toForestID ? `forest ${previewedAction.slip.toForestID}` : `clearing ${previewedAction.slip?.to ?? "?"}`}.`;
      case ACTION_TYPE.BATTLE:
        return `${factionLabels[previewedAction.battle?.faction ?? 0] ?? "Unknown"} initiates battle against ${factionLabels[previewedAction.battle?.targetFaction ?? 0] ?? "Unknown"} in clearing ${previewedAction.battle?.clearingID ?? "?"}.`;
      case ACTION_TYPE.BUILD:
        return `Building in clearing ${previewedAction.build?.clearingID ?? "?"} using ${previewedAction.build?.woodSources?.length ?? 0} wood source(s).`;
      case ACTION_TYPE.CRAFT:
        return `Crafting uses workshop access from ${(previewedAction.craft?.usedWorkshopClearings ?? []).join(", ") || "no"} clearings.`;
      case ACTION_TYPE.RECRUIT:
        return `Recruit affects ${(previewedAction.recruit?.clearingIDs ?? []).length} clearing(s).`;
      default:
        return describeAction(previewedAction);
    }
  })();

  const previewRoutes =
    !previewedAction
      ? []
      : (() => {
          switch (previewedAction.type) {
            case ACTION_TYPE.MOVEMENT: {
              const from = clearingBoardPosition(previewedAction.movement?.from ?? -1);
              const to = clearingBoardPosition(previewedAction.movement?.to ?? -1);
              return from && to
                ? [{ key: "movement-route", x1: from.left, y1: from.top, x2: to.left, y2: to.top, tone: "movement" as const }]
                : [];
            }
            case ACTION_TYPE.SLIP: {
              const from = clearingBoardPosition(previewedAction.slip?.from ?? -1);
              const to = clearingBoardPosition(previewedAction.slip?.to ?? -1);
              return from && to
                ? [{ key: "slip-route", x1: from.left, y1: from.top, x2: to.left, y2: to.top, tone: "movement" as const }]
                : [];
            }
            case ACTION_TYPE.BUILD: {
              const target = clearingBoardPosition(previewedAction.build?.clearingID ?? -1);
              if (!target) {
                return [];
              }
              return (previewedAction.build?.woodSources ?? [])
                .map((source) => {
                  const position = clearingBoardPosition(source.clearingID);
                  if (!position) {
                    return null;
                  }
                  return {
                    key: `build-${source.clearingID}`,
                    x1: position.left,
                    y1: position.top,
                    x2: target.left,
                    y2: target.top,
                    tone: "supply" as const
                  };
                })
                .filter((route): route is NonNullable<typeof route> => route !== null);
            }
            default:
              return [];
          }
        })();

  const previewAnnotations =
    !previewedAction
      ? []
      : highlightedClearings
          .map((highlight) => {
            const position = clearingBoardPosition(highlight.clearingID);
            if (!position) {
              return null;
            }

            let label = highlight.role === "source" ? "Source" : highlight.role === "target" ? "Target" : "Affected";
            let detail = "";

            switch (previewedAction.type) {
              case ACTION_TYPE.MOVEMENT:
              case ACTION_TYPE.SLIP:
                label = highlight.role === "source" ? "From" : highlight.role === "target" ? "To" : label;
                detail = `Clearing ${highlight.clearingID}`;
                break;
              case ACTION_TYPE.BUILD: {
                const woodSource = previewedAction.build?.woodSources?.find((source) => source.clearingID === highlight.clearingID);
                label = woodSource ? "Wood Source" : "Build Site";
                detail = woodSource ? `${woodSource.amount} wood` : `Clearing ${highlight.clearingID}`;
                break;
              }
              case ACTION_TYPE.BATTLE:
              case ACTION_TYPE.BATTLE_RESOLUTION:
                label = "Battle Site";
                detail = `${factionLabels[previewedAction.battle?.targetFaction ?? previewedAction.battleResolution?.targetFaction ?? 0] ?? "Unknown"} targeted`;
                break;
              case ACTION_TYPE.CRAFT:
                label = "Workshop";
                detail = `Craft support`;
                break;
              case ACTION_TYPE.RECRUIT:
                label = "Recruit";
                detail = `Clearing ${highlight.clearingID}`;
                break;
              default:
                detail = `Clearing ${highlight.clearingID}`;
                break;
            }

            return {
              key: `${highlight.clearingID}-${highlight.role}`,
              left: position.left,
              top: position.top,
              role: highlight.role,
              label,
              detail
            };
          })
          .filter((annotation): annotation is NonNullable<typeof annotation> => annotation !== null);

  return (
    <section className="board-panel">
      <div className={`board-canvas ${previewedAction ? "preview-active" : ""}`}>
        <img
          className={`board-map-art ${previewedAction ? "preview-active" : ""}`}
          src={boardLayout.imagePath}
          alt={`${boardLayout.label} board`}
        />
        {focusedClearing ? (
          <div className="board-chrome">
            <div className="board-focus-card">
              <p className="board-kicker">{boardLayout.label} Woodland</p>
              <div className="board-focus-header">
                <div>
                  <h2>
                    Clearing {focusedClearing.id}
                    {hoveredClearingID !== null && hoveredClearingID !== selectedClearingID ? (
                      <span className="board-focus-mode"> Preview</span>
                    ) : null}
                  </h2>
                  <div className="board-focus-meta">
                    <span className={`board-suit-pill ${suitClass(focusedClearing.suit)}`}>
                      {suitLabels[focusedClearing.suit] ?? "Unknown"} Suit
                    </span>
                    <span>{focusedClearing.adj.length} paths</span>
                    <span>
                      {focusedRuler === null
                        ? "No ruler"
                        : `${factionLabels[focusedRuler] ?? "Unknown"} rule`}
                    </span>
                    <span>
                      {focusedClearing.buildings.length}/{focusedClearing.buildSlots} build slots filled
                    </span>
                    {focusedClearing.wood > 0 ? <span>{focusedClearing.wood} wood</span> : null}
                    {focusedClearing.ruins ? <span>Ruins</span> : null}
                  </div>
                </div>
                <div className="board-focus-totals">
                  <span>
                    {Object.values(focusedClearing.warriors).reduce((sum, count) => sum + count, 0)} warriors
                  </span>
                  <span>{focusedClearing.tokens.length} tokens</span>
                  {focusedRuler !== null ? (
                    <span>
                      {rulingPresence(focusedClearing, focusedRuler)} rule strength
                    </span>
                  ) : null}
                </div>
              </div>
              <div className="board-control-strip">
                {boardControlSummary.map((entry) => (
                  <div key={entry.key} className={`board-control-card ${entry.key}`}>
                    <span>{entry.label}</span>
                    <strong>{entry.count}</strong>
                  </div>
                ))}
              </div>
              <div className="board-presence-grid">
                {focusedClearingPresence.length > 0 ? (
                  focusedClearingPresence.map((presence) => (
                    <div key={presence.key} className={`board-presence-card ${presence.key}`}>
                      <span className="board-presence-label">{presence.label}</span>
                      <strong>{presence.total}</strong>
                      <span>{presence.detail}</span>
                    </div>
                  ))
                ) : (
                  <div className="board-presence-empty">
                    No faction presence here yet. Use this clearing for setup inspection or future action previews.
                  </div>
                )}
              </div>
              {adjacentClearings.length > 0 ? (
                <div className="board-adjacency-panel">
                  <span className="board-presence-label">Adjacent Clearings</span>
                  <div className="board-adjacency-list">
                    {adjacentClearings.map((clearing) => {
                      const ruler = rulerOfClearing(clearing);
                      return (
                        <button
                          key={clearing.id}
                          type="button"
                          className={`board-adjacency-chip ${clearing.id === selectedClearingID ? "selected" : ""}`}
                          onClick={() => onSelectClearing(clearing.id)}
                        >
                          <span className={`board-adjacency-suit ${suitClass(clearing.suit)}`}>
                            {suitLabels[clearing.suit] ?? "Unknown"}
                          </span>
                          <strong>{clearing.id}</strong>
                          <span>{ruler === null ? "No ruler" : factionLabels[ruler] ?? "Unknown"}</span>
                        </button>
                      );
                    })}
                  </div>
                </div>
              ) : null}
              <div className="board-action-panel">
                <span className="board-presence-label">Loaded Actions Here</span>
                {focusedClearingActions.length > 0 ? (
                  <div className="board-action-list">
                    {focusedClearingActions.slice(0, 4).map(({ action, index }) => (
                      <div
                        key={`${action.type}-${index}`}
                        className={`board-action-chip ${previewedAction === action ? "previewed" : ""}`}
                      >
                        <strong>#{index + 1}</strong>
                        <span>{describeAction(action)}</span>
                      </div>
                    ))}
                    {focusedClearingActions.length > 4 ? (
                      <div className="board-action-more">
                        +{focusedClearingActions.length - 4} more loaded action(s) affect this clearing
                      </div>
                    ) : null}
                  </div>
                ) : (
                  <div className="board-action-empty">
                    None of the currently loaded actions target this clearing.
                  </div>
                )}
              </div>
            </div>
          </div>
        ) : null}
        {previewedAction && previewOverlayTitle ? (
          <div className="board-preview-overlay">
            <div className="board-preview-card">
              <p className="board-kicker">{previewOverlayTitle}</p>
              <strong>{describeAction(previewedAction)}</strong>
              {previewOverlayDetail ? <span>{previewOverlayDetail}</span> : null}
              {previewFootprint.total > 0 ? (
                <div className="board-preview-summary">
                  <span>{previewFootprint.total} touched</span>
                  {previewFootprint.source > 0 ? <span>{previewFootprint.source} source</span> : null}
                  {previewFootprint.target > 0 ? <span>{previewFootprint.target} target</span> : null}
                  {previewFootprint.affected > 0 ? <span>{previewFootprint.affected} affected</span> : null}
                </div>
              ) : null}
            </div>
          </div>
        ) : null}
        <div className="board-overlay">
          {previewRoutes.length > 0 ? (
            <svg className="board-preview-routes" viewBox="0 0 100 100" preserveAspectRatio="none">
              {previewRoutes.map((route) => (
                <line
                  key={route.key}
                  x1={route.x1}
                  y1={route.y1}
                  x2={route.x2}
                  y2={route.y2}
                  className={route.tone}
                />
              ))}
            </svg>
          ) : null}
          {!boardLayout.imagePath ? (
            <svg className="board-paths" viewBox="0 0 100 100" preserveAspectRatio="none">
              {adjacencySegments.map((segment) => (
                <line
                  key={segment.key}
                  x1={segment.x1}
                  y1={segment.y1}
                  x2={segment.x2}
                  y2={segment.y2}
                  className={segment.connectedToFocused ? "connected-to-focus" : undefined}
                />
              ))}
            </svg>
          ) : (
            <svg className="board-paths board-paths-focus" viewBox="0 0 100 100" preserveAspectRatio="none">
              {adjacencySegments
                .filter((segment) => segment.connectedToFocused)
                .map((segment) => (
                  <line
                    key={segment.key}
                    x1={segment.x1}
                    y1={segment.y1}
                    x2={segment.x2}
                    y2={segment.y2}
                    className="connected-to-focus"
                  />
              ))}
            </svg>
          )}
          {previewAnnotations.map((annotation) => (
            <div
              key={annotation.key}
              className={`board-preview-badge ${annotation.role}`}
              style={{ left: `${annotation.left}%`, top: `${annotation.top}%` }}
            >
              <strong>{annotation.label}</strong>
              <span>{annotation.detail}</span>
            </div>
          ))}
          {clearings.map((clearing, index) => (
            <ClearingMarker
              key={clearing.id}
              clearing={clearing}
              position={clearingPosition(clearing.id, index, boardLayout.clearingPositions)}
              isSelected={clearing.id === selectedClearingID}
              isFocused={clearing.id === focusedClearingID}
              isAdjacentToFocus={focusedClearing ? focusedClearing.adj.includes(clearing.id) : false}
              isDimmed={
                previewedAction !== null &&
                !previewedClearingIDs.has(clearing.id) &&
                clearing.id !== selectedClearingID &&
                clearing.id !== focusedClearingID
              }
              hasKeep={clearing.id === keepClearingID}
              hasVagabond={!vagabondInForest && clearing.id === vagabondClearingID}
              highlightRole={highlightByClearing.get(clearing.id)}
              isSetupLegal={legalSetupClearings.has(clearing.id)}
              isSetupChosen={selectedSetupClearings.has(clearing.id)}
              onClick={() => onSelectClearing(clearing.id)}
              onHover={(hovered) => setHoveredClearingID(hovered ? clearing.id : null)}
            />
          ))}
          {forestTargets.map((forest) => {
            const position = forestPosition(forest.forestID);
            if (!position) {
              return null;
            }

            return (
              <button
                key={forest.forestID}
                type="button"
                className={`forest-marker ${forest.legal ? "legal" : ""} ${forest.selected ? "selected" : ""}`}
                style={{ left: `${position.left}%`, top: `${position.top}%` }}
                onClick={() => onSelectForest?.(forest.forestID)}
              >
                <span className="forest-marker-label">{forest.label}</span>
              </button>
            );
          })}
        </div>
      </div>
    </section>
  );
}
