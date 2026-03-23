import { countBuildings, suitClass } from "../gameHelpers";
import { suitLabels } from "../labels";
import type { Clearing } from "../types";

type InspectorPanelProps = {
  clearing?: Clearing;
  keepClearingID: number;
  onUpdateClearing: (clearingID: number, mutator: (clearing: Clearing) => void) => void;
  onSetKeepClearing: (clearingID: number) => void;
  onClose: () => void;
};

function setMarquiseBuildingCount(clearing: Clearing, type: number, count: number) {
  clearing.buildings = [
    ...clearing.buildings.filter(
      (building) => !(building.faction === 0 && building.type === type)
    ),
    ...Array.from({ length: count }, () => ({ faction: 0, type }))
  ];
}

function setEnemyBuildingCount(clearing: Clearing, count: number) {
  clearing.buildings = [
    ...clearing.buildings.filter((building) => building.faction !== 2),
    ...Array.from({ length: count }, () => ({ faction: 2, type: 0 }))
  ];
}

type CounterEditorProps = {
  label: string;
  value: number;
  onDecrease: () => void;
  onIncrease: () => void;
};

function CounterEditor({ label, value, onDecrease, onIncrease }: CounterEditorProps) {
  return (
    <div className="counter-editor">
      <span className="counter-label">{label}</span>
      <div className="counter-controls">
        <button type="button" className="counter-button secondary" onClick={onDecrease}>
          -
        </button>
        <strong className="counter-value">{value}</strong>
        <button type="button" className="counter-button" onClick={onIncrease}>
          +
        </button>
      </div>
    </div>
  );
}

export function InspectorPanel({
  clearing,
  keepClearingID,
  onUpdateClearing,
  onSetKeepClearing,
  onClose
}: InspectorPanelProps) {
  if (!clearing) {
    return null;
  }

  const hasKeep = keepClearingID === clearing.id;

  return (
    <section className="panel inspector-panel">
      <div className="panel-header">
        <h2>Selected Clearing</h2>
        <div className="inspector-header-actions">
          <span className={`pill suit-pill ${suitClass(clearing.suit)}`}>
            {suitLabels[clearing.suit] ?? "Unknown"} {clearing.id}
          </span>
          <button type="button" className="secondary inspector-close" onClick={onClose}>
            Close
          </button>
        </div>
      </div>

      <div className="inspector-meta">
        <span>Adjacency: {clearing.adj.join(", ") || "None"}</span>
        <span>Build Slots: {clearing.buildSlots}</span>
      </div>

      <div className="inspector-summary">
        {(clearing.warriors["0"] ?? 0) > 0 ? (
          <span className="indicator-pill marquise">
            Marquise Warriors {clearing.warriors["0"] ?? 0}
          </span>
        ) : null}
        {(clearing.warriors["2"] ?? 0) > 0 ? (
          <span className="indicator-pill eyrie">
            Eyrie Warriors {clearing.warriors["2"] ?? 0}
          </span>
        ) : null}
        {clearing.wood > 0 ? (
          <span className="indicator-pill wood">Wood {clearing.wood}</span>
        ) : null}
        {countBuildings(clearing.buildings, 0, 0) > 0 ? (
          <span className="indicator-square sawmill">
            Sawmill {countBuildings(clearing.buildings, 0, 0)}
          </span>
        ) : null}
        {countBuildings(clearing.buildings, 0, 1) > 0 ? (
          <span className="indicator-square workshop">
            Workshop {countBuildings(clearing.buildings, 0, 1)}
          </span>
        ) : null}
        {countBuildings(clearing.buildings, 0, 2) > 0 ? (
          <span className="indicator-square recruiter">
            Recruiter {countBuildings(clearing.buildings, 0, 2)}
          </span>
        ) : null}
        {countBuildings(clearing.buildings, 2) > 0 ? (
          <span className="indicator-square enemy">
            Enemy Buildings {countBuildings(clearing.buildings, 2)}
          </span>
        ) : null}
        {hasKeep ? <span className="indicator-square keep">Keep</span> : null}
        {clearing.ruins ? <span className="indicator-square ruins">Ruins</span> : null}
      </div>

      <div className="setup-note">
        Use quick controls to edit the selected clearing directly from the board state.
      </div>

      <div className="editor-grid">
        <CounterEditor
          label="Wood"
          value={clearing.wood}
          onDecrease={() =>
            onUpdateClearing(clearing.id, (draft) => {
              draft.wood = Math.max(0, draft.wood - 1);
            })
          }
          onIncrease={() =>
            onUpdateClearing(clearing.id, (draft) => {
              draft.wood += 1;
            })
          }
        />
        <CounterEditor
          label="Marquise Warriors"
          value={clearing.warriors["0"] ?? 0}
          onDecrease={() =>
            onUpdateClearing(clearing.id, (draft) => {
              draft.warriors["0"] = Math.max(0, (draft.warriors["0"] ?? 0) - 1);
            })
          }
          onIncrease={() =>
            onUpdateClearing(clearing.id, (draft) => {
              draft.warriors["0"] = (draft.warriors["0"] ?? 0) + 1;
            })
          }
        />
        <CounterEditor
          label="Eyrie Warriors"
          value={clearing.warriors["2"] ?? 0}
          onDecrease={() =>
            onUpdateClearing(clearing.id, (draft) => {
              draft.warriors["2"] = Math.max(0, (draft.warriors["2"] ?? 0) - 1);
            })
          }
          onIncrease={() =>
            onUpdateClearing(clearing.id, (draft) => {
              draft.warriors["2"] = (draft.warriors["2"] ?? 0) + 1;
            })
          }
        />
        <CounterEditor
          label="Sawmills"
          value={countBuildings(clearing.buildings, 0, 0)}
          onDecrease={() =>
            onUpdateClearing(clearing.id, (draft) => {
              setMarquiseBuildingCount(draft, 0, Math.max(0, countBuildings(draft.buildings, 0, 0) - 1));
            })
          }
          onIncrease={() =>
            onUpdateClearing(clearing.id, (draft) => {
              setMarquiseBuildingCount(draft, 0, countBuildings(draft.buildings, 0, 0) + 1);
            })
          }
        />
        <CounterEditor
          label="Workshops"
          value={countBuildings(clearing.buildings, 0, 1)}
          onDecrease={() =>
            onUpdateClearing(clearing.id, (draft) => {
              setMarquiseBuildingCount(draft, 1, Math.max(0, countBuildings(draft.buildings, 0, 1) - 1));
            })
          }
          onIncrease={() =>
            onUpdateClearing(clearing.id, (draft) => {
              setMarquiseBuildingCount(draft, 1, countBuildings(draft.buildings, 0, 1) + 1);
            })
          }
        />
        <CounterEditor
          label="Recruiters"
          value={countBuildings(clearing.buildings, 0, 2)}
          onDecrease={() =>
            onUpdateClearing(clearing.id, (draft) => {
              setMarquiseBuildingCount(draft, 2, Math.max(0, countBuildings(draft.buildings, 0, 2) - 1));
            })
          }
          onIncrease={() =>
            onUpdateClearing(clearing.id, (draft) => {
              setMarquiseBuildingCount(draft, 2, countBuildings(draft.buildings, 0, 2) + 1);
            })
          }
        />
        <CounterEditor
          label="Enemy Buildings"
          value={countBuildings(clearing.buildings, 2)}
          onDecrease={() =>
            onUpdateClearing(clearing.id, (draft) => {
              setEnemyBuildingCount(draft, Math.max(0, countBuildings(draft.buildings, 2) - 1));
            })
          }
          onIncrease={() =>
            onUpdateClearing(clearing.id, (draft) => {
              setEnemyBuildingCount(draft, countBuildings(draft.buildings, 2) + 1);
            })
          }
        />
        <div className="toggle-editor">
          <span className="counter-label">Keep</span>
          <button
            type="button"
            className={`toggle-button ${hasKeep ? "active" : "secondary"}`}
            onClick={() => onSetKeepClearing(hasKeep ? 0 : clearing.id)}
          >
            {hasKeep ? "Present" : "Set Keep Here"}
          </button>
        </div>
        <div className="toggle-editor">
          <span className="counter-label">Ruins</span>
          <button
            type="button"
            className={`toggle-button ${clearing.ruins ? "active" : "secondary"}`}
            onClick={() =>
              onUpdateClearing(clearing.id, (draft) => {
                draft.ruins = !draft.ruins;
              })
            }
          >
            {clearing.ruins ? "Present" : "Absent"}
          </button>
        </div>
      </div>
    </section>
  );
}
