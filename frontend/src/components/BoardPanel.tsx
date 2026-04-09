import { ClearingMarker, type ClearingPreviewPiece } from "./ClearingMarker";
import { clearingPosition } from "../gameHelpers";
import { ACTION_TYPE } from "../labels";
import { useState } from "react";
import type { BoardLayout } from "../boardLayouts";
import type { Action, Clearing, Forest, GameState, HighlightedClearing } from "../types";

type BoardPanelProps = {
  state: GameState;
  clearings: Clearing[];
  forests: Forest[];
  boardLayout: BoardLayout;
  selectedClearingID: number;
  keepClearingID: number;
  vagabondClearingID: number;
  vagabondInForest: boolean;
  highlightedClearings?: HighlightedClearing[];
  previewedAction?: Action | null;
  setupLegalClearingIDs?: number[];
  setupSelectedClearingIDs?: number[];
  setupPreviewPiecesByClearing?: Record<number, ClearingPreviewPiece[]>;
  forestTargets?: Array<{
    forestID: number;
    label: string;
    legal: boolean;
    selected: boolean;
  }>;
  onSelectClearing: (clearingID: number) => void;
  onSelectForest?: (forestID: number) => void;
};

type BoardViewport = {
  x: number;
  y: number;
  scale: number;
};

type DragState = {
  startX: number;
  startY: number;
  originX: number;
  originY: number;
} | null;

const MIN_BOARD_SCALE = 1;
const MAX_BOARD_SCALE = 2.25;
const BOARD_ZOOM_STEP = 0.12;

function clampBoardScale(nextScale: number) {
  return Math.min(MAX_BOARD_SCALE, Math.max(MIN_BOARD_SCALE, nextScale));
}

export function BoardPanel({
  state,
  clearings,
  forests,
  boardLayout,
  selectedClearingID,
  keepClearingID,
  vagabondClearingID,
  vagabondInForest,
  highlightedClearings = [],
  previewedAction = null,
  setupLegalClearingIDs = [],
  setupSelectedClearingIDs = [],
  setupPreviewPiecesByClearing = {},
  forestTargets = [],
  onSelectClearing
  ,
  onSelectForest
}: BoardPanelProps) {
  const [hoveredClearingID, setHoveredClearingID] = useState<number | null>(null);
  const [viewport, setViewport] = useState<BoardViewport>({ x: 0, y: 0, scale: 1 });
  const [dragState, setDragState] = useState<DragState>(null);
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

  const previewedClearingIDs = new Set(highlightedClearings.map((highlight) => highlight.clearingID));
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

  function handleBoardMouseDown(event: React.MouseEvent<HTMLDivElement>) {
    if (event.button !== 0) {
      return;
    }
    const target = event.target as HTMLElement | null;
    if (target?.closest("button")) {
      return;
    }
    setDragState({
      startX: event.clientX,
      startY: event.clientY,
      originX: viewport.x,
      originY: viewport.y
    });
  }

  function handleBoardMouseMove(event: React.MouseEvent<HTMLDivElement>) {
    if (!dragState) {
      return;
    }
    setViewport((current) => ({
      ...current,
      x: dragState.originX + (event.clientX - dragState.startX),
      y: dragState.originY + (event.clientY - dragState.startY)
    }));
  }

  function handleBoardMouseUp() {
    setDragState(null);
  }

  function handleBoardWheel(event: React.WheelEvent<HTMLDivElement>) {
    event.preventDefault();
    setViewport((current) => ({
      ...current,
      scale: clampBoardScale(current.scale + (event.deltaY < 0 ? BOARD_ZOOM_STEP : -BOARD_ZOOM_STEP))
    }));
  }

  return (
    <section className="board-panel">
      <div
        className={`board-canvas ${previewedAction ? "preview-active" : ""} ${dragState ? "dragging" : "pannable"}`}
        data-testid="board-canvas"
        onMouseDown={handleBoardMouseDown}
        onMouseMove={handleBoardMouseMove}
        onMouseUp={handleBoardMouseUp}
        onMouseLeave={handleBoardMouseUp}
        onWheel={handleBoardWheel}
      >
        <div
          className="board-view"
          data-testid="board-view"
          style={{ transform: `translate(${viewport.x}px, ${viewport.y}px) scale(${viewport.scale})` }}
        >
          <img
            className={`board-map-art ${previewedAction ? "preview-active" : ""}`}
            src={boardLayout.imagePath}
            alt={`${boardLayout.label} board`}
          />
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
                previewPieces={setupPreviewPiecesByClearing[clearing.id] ?? []}
                onClick={(event) => {
                  event.stopPropagation();
                  onSelectClearing(clearing.id);
                }}
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
                  onClick={(event) => {
                    event.stopPropagation();
                    onSelectForest?.(forest.forestID);
                  }}
                >
                  <span className="forest-marker-label">{forest.label}</span>
                </button>
              );
            })}
          </div>
        </div>
      </div>
    </section>
  );
}
