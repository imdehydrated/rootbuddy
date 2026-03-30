import {
  eyrieLeaderLabels,
  factionLabels,
  itemStatusLabels,
  phaseLabels,
  relationshipLabels,
  stepLabels,
  vagabondCharacterLabels
} from "../labels";
import type { GameState } from "../types";

type TurnStatePanelProps = {
  state: GameState;
  onUpdateState: (mutator: (draft: GameState) => void) => void;
  onClose: () => void;
};

function parseNumberList(value: string): number[] {
  return value
    .split(",")
    .map((part) => Number(part.trim()))
    .filter((value) => Number.isFinite(value));
}

function formatNumberList(values: number[]): string {
  return values.join(", ");
}

export function TurnStatePanel({ state, onUpdateState, onClose }: TurnStatePanelProps) {
  return (
    <section className="panel modal-panel">
      <div className="panel-header">
        <h2>Advanced Turn State</h2>
        <button type="button" className="secondary" onClick={onClose}>
          Close
        </button>
      </div>

      <p className="message">
        Use this only when the quick sidebar flow is not enough and you need to correct the turn structure directly.
      </p>

      <div className="turn-state-actions">
        <button
          type="button"
          className="secondary"
          onClick={() =>
            onUpdateState((draft) => {
              draft.currentPhase = 0;
              draft.currentStep = 1;
            })
          }
        >
          Set Birdsong
        </button>
        <button
          type="button"
          className="secondary"
          onClick={() =>
            onUpdateState((draft) => {
              draft.currentPhase = 1;
              draft.currentStep = 3;
            })
          }
        >
          Set Daylight Actions
        </button>
        <button
          type="button"
          className="secondary"
          onClick={() =>
            onUpdateState((draft) => {
              draft.currentPhase = 2;
              draft.currentStep = 4;
            })
          }
        >
          Set Evening
        </button>
      </div>

      <div className="control-grid">
        <label>
          <span>Faction Turn</span>
          <select
            value={state.factionTurn}
            onChange={(event) =>
              onUpdateState((draft) => {
                draft.factionTurn = Number(event.target.value);
              })
            }
          >
            {factionLabels.map((label, index) => (
              <option key={label} value={index}>
                {label}
              </option>
            ))}
          </select>
        </label>

        <label>
          <span>Phase</span>
          <select
            value={state.currentPhase}
            onChange={(event) =>
              onUpdateState((draft) => {
                draft.currentPhase = Number(event.target.value);
              })
            }
          >
            {phaseLabels.map((label, index) => (
              <option key={label} value={index}>
                {label}
              </option>
            ))}
          </select>
        </label>

        <label>
          <span>Step</span>
          <select
            value={state.currentStep}
            onChange={(event) =>
              onUpdateState((draft) => {
                draft.currentStep = Number(event.target.value);
              })
            }
          >
            {stepLabels.map((label, index) => (
              <option key={label} value={index}>
                {label}
              </option>
            ))}
          </select>
        </label>

        <label>
          <span>Turn Order</span>
          <input
            type="text"
            value={formatNumberList(state.turnOrder)}
            onChange={(event) =>
              onUpdateState((draft) => {
                draft.turnOrder = parseNumberList(event.target.value);
              })
            }
          />
        </label>
      </div>

      <div className="summary-section">
        <h3>Victory Points</h3>
        <div className="control-grid">
          {factionLabels.map((label, index) => (
            <label key={label}>
              <span>{label}</span>
              <input
                type="number"
                value={state.victoryPoints[String(index)] ?? 0}
                onChange={(event) =>
                  onUpdateState((draft) => {
                    draft.victoryPoints[String(index)] = Number(event.target.value);
                  })
                }
              />
            </label>
          ))}
        </div>
      </div>

      <div className="summary-section">
        <h3>Turn Progress</h3>
        <div className="control-grid">
          <label>
            <span>Actions Used</span>
            <input
              type="number"
              value={state.turnProgress.actionsUsed}
              onChange={(event) =>
                onUpdateState((draft) => {
                  draft.turnProgress.actionsUsed = Number(event.target.value);
                })
              }
            />
          </label>
          <label>
            <span>Bonus Actions</span>
            <input
              type="number"
              value={state.turnProgress.bonusActions}
              onChange={(event) =>
                onUpdateState((draft) => {
                  draft.turnProgress.bonusActions = Number(event.target.value);
                })
              }
            />
          </label>
          <label>
            <span>Marches Used</span>
            <input
              type="number"
              value={state.turnProgress.marchesUsed}
              onChange={(event) =>
                onUpdateState((draft) => {
                  draft.turnProgress.marchesUsed = Number(event.target.value);
                })
              }
            />
          </label>
          <label>
            <span>Officer Actions</span>
            <input
              type="number"
              value={state.turnProgress.officerActionsUsed}
              onChange={(event) =>
                onUpdateState((draft) => {
                  draft.turnProgress.officerActionsUsed = Number(event.target.value);
                })
              }
            />
          </label>
          <label className="checkbox">
            <input
              type="checkbox"
              checked={state.turnProgress.recruitUsed}
              onChange={(event) =>
                onUpdateState((draft) => {
                  draft.turnProgress.recruitUsed = event.target.checked;
                })
              }
            />
            Recruit Used
          </label>
          <label className="checkbox">
            <input
              type="checkbox"
              checked={state.turnProgress.hasSlipped}
              onChange={(event) =>
                onUpdateState((draft) => {
                  draft.turnProgress.hasSlipped = event.target.checked;
                })
              }
            />
            Has Slipped
          </label>
        </div>
      </div>

      <div className="summary-section">
        <h3>Faction State</h3>
        <div className="control-grid">
          <label>
            <span>Marquise Supply</span>
            <input type="number" value={state.marquise.warriorSupply} readOnly />
          </label>
          <label>
            <span>Marquise Buildings</span>
            <input
              type="text"
              value={`${state.marquise.sawmillsPlaced}/${state.marquise.workshopsPlaced}/${state.marquise.recruitersPlaced}`}
              readOnly
            />
          </label>
          <label>
            <span>Eyrie Leader</span>
            <select
              value={state.eyrie.leader}
              onChange={(event) =>
                onUpdateState((draft) => {
                  draft.eyrie.leader = Number(event.target.value);
                })
              }
            >
              {eyrieLeaderLabels.map((label, index) => (
                <option key={label} value={index}>
                  {label}
                </option>
              ))}
            </select>
          </label>
          <label>
            <span>Eyrie Decree Recruit</span>
            <input
              type="text"
              value={formatNumberList(state.eyrie.decree.recruit)}
              onChange={(event) =>
                onUpdateState((draft) => {
                  draft.eyrie.decree.recruit = parseNumberList(event.target.value);
                })
              }
            />
          </label>
          <label>
            <span>Eyrie Decree Move</span>
            <input
              type="text"
              value={formatNumberList(state.eyrie.decree.move)}
              onChange={(event) =>
                onUpdateState((draft) => {
                  draft.eyrie.decree.move = parseNumberList(event.target.value);
                })
              }
            />
          </label>
          <label>
            <span>Eyrie Decree Battle</span>
            <input
              type="text"
              value={formatNumberList(state.eyrie.decree.battle)}
              onChange={(event) =>
                onUpdateState((draft) => {
                  draft.eyrie.decree.battle = parseNumberList(event.target.value);
                })
              }
            />
          </label>
          <label>
            <span>Eyrie Decree Build</span>
            <input
              type="text"
              value={formatNumberList(state.eyrie.decree.build)}
              onChange={(event) =>
                onUpdateState((draft) => {
                  draft.eyrie.decree.build = parseNumberList(event.target.value);
                })
              }
            />
          </label>
          <label>
            <span>Alliance Officers</span>
            <input
              type="number"
              value={state.alliance.officers}
              onChange={(event) =>
                onUpdateState((draft) => {
                  draft.alliance.officers = Number(event.target.value);
                })
              }
            />
          </label>
          <label>
            <span>Alliance Supporters</span>
            <input type="text" value={state.alliance.supporters.length} readOnly />
          </label>
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
                  {index + 1}. {itemStatusLabels[item.status] ?? "Unknown"} item {item.type}
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
    </section>
  );
}
