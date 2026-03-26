import { ClearingMarker } from "./ClearingMarker";
import { clearingPosition } from "../gameHelpers";
import type { BoardLayout } from "../boardLayouts";
import type { Clearing, HighlightedClearing } from "../types";

type BoardPanelProps = {
  clearings: Clearing[];
  boardLayout: BoardLayout;
  selectedClearingID: number;
  keepClearingID: number;
  vagabondClearingID: number;
  vagabondInForest: boolean;
  highlightedClearings?: HighlightedClearing[];
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
  boardLayout,
  selectedClearingID,
  keepClearingID,
  vagabondClearingID,
  vagabondInForest,
  highlightedClearings = [],
  setupLegalClearingIDs = [],
  setupSelectedClearingIDs = [],
  forestTargets = [],
  onSelectClearing
  ,
  onSelectForest
}: BoardPanelProps) {
  const highlightByClearing = new Map(
    highlightedClearings.map((highlight) => [highlight.clearingID, highlight.role])
  );
  const legalSetupClearings = new Set(setupLegalClearingIDs);
  const selectedSetupClearings = new Set(setupSelectedClearingIDs);

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
          y2: to.top
        };
      })
      .filter((segment): segment is NonNullable<typeof segment> => segment !== null)
  );

  return (
    <section className="board-panel">
      <div className="board-canvas">
        <div className="board-overlay">
          <svg className="board-paths" viewBox="0 0 100 100" preserveAspectRatio="none">
            {adjacencySegments.map((segment) => (
              <line
                key={segment.key}
                x1={segment.x1}
                y1={segment.y1}
                x2={segment.x2}
                y2={segment.y2}
              />
            ))}
          </svg>
          {clearings.map((clearing, index) => (
            <ClearingMarker
              key={clearing.id}
              clearing={clearing}
              position={clearingPosition(clearing.id, index, boardLayout.clearingPositions)}
              isSelected={clearing.id === selectedClearingID}
              hasKeep={clearing.id === keepClearingID}
              hasVagabond={!vagabondInForest && clearing.id === vagabondClearingID}
              highlightRole={highlightByClearing.get(clearing.id)}
              isSetupLegal={legalSetupClearings.has(clearing.id)}
              isSetupChosen={selectedSetupClearings.has(clearing.id)}
              onClick={() => onSelectClearing(clearing.id)}
            />
          ))}
          {forestTargets.map((forest) => {
            const position = boardLayout.forestPositions[forest.forestID];
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
