import { factionLabels, itemStatusLabels, itemTypeLabels, relationshipLabels, vagabondCharacterLabels } from "../../labels";
import type { GameState } from "../../types";

type VagabondEditorProps = {
  state: GameState;
  onUpdateState: (mutator: (draft: GameState) => void) => void;
};

export function VagabondEditor({ state, onUpdateState }: VagabondEditorProps) {
  return (
    <>
      <div className="summary-section">
        <h3>Vagabond</h3>
        <div className="control-grid">
          <label>
            <span>Vagabond Character</span>
            <select
              value={state.vagabond.character}
              onChange={(event) =>
                onUpdateState((draft) => {
                  draft.vagabond.character = Number(event.target.value);
                })
              }
            >
              {vagabondCharacterLabels.map((label, index) => (
                <option key={label} value={index}>
                  {label}
                </option>
              ))}
            </select>
          </label>
          <label>
            <span>Vagabond Clearing</span>
            <input
              type="number"
              value={state.vagabond.clearingID}
              onChange={(event) =>
                onUpdateState((draft) => {
                  draft.vagabond.clearingID = Number(event.target.value);
                })
              }
            />
          </label>
          <label>
            <span>Vagabond Forest</span>
            <input
              type="number"
              value={state.vagabond.forestID}
              onChange={(event) =>
                onUpdateState((draft) => {
                  draft.vagabond.forestID = Number(event.target.value);
                })
              }
            />
          </label>
          <label className="checkbox">
            <input
              type="checkbox"
              checked={state.vagabond.inForest}
              onChange={(event) =>
                onUpdateState((draft) => {
                  draft.vagabond.inForest = event.target.checked;
                })
              }
            />
            Vagabond In Forest
          </label>
        </div>
      </div>

      <div className="summary-section">
        <h3>Readout</h3>
        <div className="summary-grid dense">
          <div className="summary-card">
            <span className="summary-label">Vagabond Items</span>
            {state.vagabond.items.length === 0 ? (
              <span>None</span>
            ) : (
              state.vagabond.items.map((item, index) => (
                <span key={`${item.type}-${index}`}>
                  {index + 1}. {itemStatusLabels[item.status] ?? "Unknown"} {itemTypeLabels[item.type] ?? `Item ${item.type}`}
                </span>
              ))
            )}
          </div>
          <div className="summary-card">
            <span className="summary-label">Relationships</span>
            {Object.entries(state.vagabond.relationships).length === 0 ? (
              <span>None</span>
            ) : (
              Object.entries(state.vagabond.relationships).map(([faction, level]) => (
                <span key={faction}>
                  {factionLabels[Number(faction)]}: {relationshipLabels[level] ?? "Unknown"}
                </span>
              ))
            )}
          </div>
        </div>
      </div>
    </>
  );
}
