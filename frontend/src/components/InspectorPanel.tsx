import { countBuildings, countTokens, suitClass } from "../gameHelpers";
import { suitLabels } from "../labels";
import type { Clearing } from "../types";

type InspectorPanelProps = {
  clearing?: Clearing;
  keepClearingID: number;
  vagabondClearingID: number;
  vagabondInForest: boolean;
  title?: string;
  showCloseButton?: boolean;
  onUpdateClearing: (clearingID: number, mutator: (clearing: Clearing) => void) => void;
  onSetKeepClearing: (clearingID: number) => void;
  onSetVagabondClearing: (clearingID: number, inForest: boolean) => void;
  onClose: () => void;
};

function setBuildingCount(clearing: Clearing, faction: number, type: number, count: number) {
  clearing.buildings = [
    ...clearing.buildings.filter((building) => !(building.faction === faction && building.type === type)),
    ...Array.from({ length: count }, () => ({ faction, type }))
  ];
}

function setTokenCount(clearing: Clearing, faction: number, type: number, count: number) {
  clearing.tokens = [
    ...clearing.tokens.filter((token) => !(token.faction === faction && token.type === type)),
    ...Array.from({ length: count }, () => ({ faction, type }))
  ];
}

function setWarriorCount(clearing: Clearing, faction: number, count: number) {
  clearing.warriors[String(faction)] = Math.max(0, count);
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
  vagabondClearingID,
  vagabondInForest,
  title = "Selected Clearing",
  showCloseButton = true,
  onUpdateClearing,
  onSetKeepClearing,
  onSetVagabondClearing,
  onClose
}: InspectorPanelProps) {
  if (!clearing) {
    return null;
  }

  const hasKeep = keepClearingID === clearing.id;
  const hasVagabond = !vagabondInForest && vagabondClearingID === clearing.id;

  return (
    <section className="panel inspector-panel">
      <div className="panel-header">
        <h2>{title}</h2>
        <div className="inspector-header-actions">
          <span className={`pill suit-pill ${suitClass(clearing.suit)}`}>
            {suitLabels[clearing.suit] ?? "Unknown"} {clearing.id}
          </span>
          {showCloseButton ? (
            <button type="button" className="secondary inspector-close" onClick={onClose}>
              Close
            </button>
          ) : null}
        </div>
      </div>

      <div className="inspector-meta">
        <span>Adjacency: {clearing.adj.join(", ") || "None"}</span>
        <span>Build Slots: {clearing.buildSlots}</span>
      </div>

      <div className="inspector-summary">
        {(clearing.warriors["0"] ?? 0) > 0 ? (
          <span className="indicator-pill marquise">Marquise Warriors {clearing.warriors["0"] ?? 0}</span>
        ) : null}
        {(clearing.warriors["1"] ?? 0) > 0 ? (
          <span className="indicator-pill alliance">Alliance Warriors {clearing.warriors["1"] ?? 0}</span>
        ) : null}
        {(clearing.warriors["2"] ?? 0) > 0 ? (
          <span className="indicator-pill eyrie">Eyrie Warriors {clearing.warriors["2"] ?? 0}</span>
        ) : null}
        {hasVagabond ? <span className="indicator-pill vagabond">Vagabond</span> : null}
        {clearing.wood > 0 ? <span className="indicator-pill wood">Wood {clearing.wood}</span> : null}
        {countBuildings(clearing.buildings, 0, 0) > 0 ? (
          <span className="indicator-square sawmill">Sawmill {countBuildings(clearing.buildings, 0, 0)}</span>
        ) : null}
        {countBuildings(clearing.buildings, 0, 1) > 0 ? (
          <span className="indicator-square workshop">Workshop {countBuildings(clearing.buildings, 0, 1)}</span>
        ) : null}
        {countBuildings(clearing.buildings, 0, 2) > 0 ? (
          <span className="indicator-square recruiter">Recruiter {countBuildings(clearing.buildings, 0, 2)}</span>
        ) : null}
        {countBuildings(clearing.buildings, 2, 3) > 0 ? (
          <span className="indicator-square roost">Roost {countBuildings(clearing.buildings, 2, 3)}</span>
        ) : null}
        {countBuildings(clearing.buildings, 1, 4) > 0 ? (
          <span className="indicator-square base">Base {countBuildings(clearing.buildings, 1, 4)}</span>
        ) : null}
        {countTokens(clearing.tokens, 1, 1) > 0 ? (
          <span className="indicator-square sympathy">Sympathy {countTokens(clearing.tokens, 1, 1)}</span>
        ) : null}
        {hasKeep ? <span className="indicator-square keep">Keep</span> : null}
        {clearing.ruins ? <span className="indicator-square ruins">Ruins</span> : null}
      </div>

      <div className="setup-note">Use quick controls to edit the selected clearing directly from the board state.</div>

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
              setWarriorCount(draft, 0, (draft.warriors["0"] ?? 0) - 1);
            })
          }
          onIncrease={() =>
            onUpdateClearing(clearing.id, (draft) => {
              setWarriorCount(draft, 0, (draft.warriors["0"] ?? 0) + 1);
            })
          }
        />
        <CounterEditor
          label="Alliance Warriors"
          value={clearing.warriors["1"] ?? 0}
          onDecrease={() =>
            onUpdateClearing(clearing.id, (draft) => {
              setWarriorCount(draft, 1, (draft.warriors["1"] ?? 0) - 1);
            })
          }
          onIncrease={() =>
            onUpdateClearing(clearing.id, (draft) => {
              setWarriorCount(draft, 1, (draft.warriors["1"] ?? 0) + 1);
            })
          }
        />
        <CounterEditor
          label="Eyrie Warriors"
          value={clearing.warriors["2"] ?? 0}
          onDecrease={() =>
            onUpdateClearing(clearing.id, (draft) => {
              setWarriorCount(draft, 2, (draft.warriors["2"] ?? 0) - 1);
            })
          }
          onIncrease={() =>
            onUpdateClearing(clearing.id, (draft) => {
              setWarriorCount(draft, 2, (draft.warriors["2"] ?? 0) + 1);
            })
          }
        />
        <CounterEditor
          label="Sawmills"
          value={countBuildings(clearing.buildings, 0, 0)}
          onDecrease={() =>
            onUpdateClearing(clearing.id, (draft) => {
              setBuildingCount(draft, 0, 0, Math.max(0, countBuildings(draft.buildings, 0, 0) - 1));
            })
          }
          onIncrease={() =>
            onUpdateClearing(clearing.id, (draft) => {
              setBuildingCount(draft, 0, 0, countBuildings(draft.buildings, 0, 0) + 1);
            })
          }
        />
        <CounterEditor
          label="Workshops"
          value={countBuildings(clearing.buildings, 0, 1)}
          onDecrease={() =>
            onUpdateClearing(clearing.id, (draft) => {
              setBuildingCount(draft, 0, 1, Math.max(0, countBuildings(draft.buildings, 0, 1) - 1));
            })
          }
          onIncrease={() =>
            onUpdateClearing(clearing.id, (draft) => {
              setBuildingCount(draft, 0, 1, countBuildings(draft.buildings, 0, 1) + 1);
            })
          }
        />
        <CounterEditor
          label="Recruiters"
          value={countBuildings(clearing.buildings, 0, 2)}
          onDecrease={() =>
            onUpdateClearing(clearing.id, (draft) => {
              setBuildingCount(draft, 0, 2, Math.max(0, countBuildings(draft.buildings, 0, 2) - 1));
            })
          }
          onIncrease={() =>
            onUpdateClearing(clearing.id, (draft) => {
              setBuildingCount(draft, 0, 2, countBuildings(draft.buildings, 0, 2) + 1);
            })
          }
        />
        <CounterEditor
          label="Roosts"
          value={countBuildings(clearing.buildings, 2, 3)}
          onDecrease={() =>
            onUpdateClearing(clearing.id, (draft) => {
              setBuildingCount(draft, 2, 3, Math.max(0, countBuildings(draft.buildings, 2, 3) - 1));
            })
          }
          onIncrease={() =>
            onUpdateClearing(clearing.id, (draft) => {
              setBuildingCount(draft, 2, 3, countBuildings(draft.buildings, 2, 3) + 1);
            })
          }
        />
        <CounterEditor
          label="Alliance Bases"
          value={countBuildings(clearing.buildings, 1, 4)}
          onDecrease={() =>
            onUpdateClearing(clearing.id, (draft) => {
              setBuildingCount(draft, 1, 4, Math.max(0, countBuildings(draft.buildings, 1, 4) - 1));
            })
          }
          onIncrease={() =>
            onUpdateClearing(clearing.id, (draft) => {
              setBuildingCount(draft, 1, 4, countBuildings(draft.buildings, 1, 4) + 1);
            })
          }
        />
        <CounterEditor
          label="Sympathy"
          value={countTokens(clearing.tokens, 1, 1)}
          onDecrease={() =>
            onUpdateClearing(clearing.id, (draft) => {
              setTokenCount(draft, 1, 1, Math.max(0, countTokens(draft.tokens, 1, 1) - 1));
            })
          }
          onIncrease={() =>
            onUpdateClearing(clearing.id, (draft) => {
              setTokenCount(draft, 1, 1, countTokens(draft.tokens, 1, 1) + 1);
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
          <span className="counter-label">Vagabond</span>
          <button
            type="button"
            className={`toggle-button ${hasVagabond ? "active" : "secondary"}`}
            onClick={() => onSetVagabondClearing(hasVagabond ? 0 : clearing.id, false)}
          >
            {hasVagabond ? "Present" : "Place Here"}
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
