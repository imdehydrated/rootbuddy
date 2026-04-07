import {
  eyrieLeaderLabels,
  factionLabels,
  itemTypeLabels,
  itemStatusLabels,
  phaseLabels,
  relationshipLabels,
  stepLabels,
  suitLabels,
  vagabondCharacterLabels
} from "../labels";
import { describeKnownCardID } from "../cardCatalog";
import type { Card, GameState } from "../types";
import { ReferenceCard } from "./CardUi";
import { TokenListEditor } from "./TokenListEditor";

type TurnStatePanelProps = {
  state: GameState;
  onUpdateState: (mutator: (draft: GameState) => void) => void;
  title?: string;
  showCloseButton?: boolean;
  onClose: () => void;
};

function parseNumberList(value: string): number[] {
  return value
    .split(",")
    .map((part) => part.trim())
    .filter((part) => part.length > 0)
    .map((part) => Number(part))
    .filter((value) => Number.isInteger(value));
}

function formatNumberList(values: number[]): string {
  return values.join(", ");
}

function isValidDecreeCardID(value: number): boolean {
  return value === -2 || value === -1 || value >= 1;
}

function describeVisibleCard(card: Card): string {
  return `${card.name} (${suitLabels[card.suit] ?? "Unknown"})`;
}

function duplicateValues(values: number[]): number[] {
  const seen = new Set<number>();
  const duplicates = new Set<number>();
  values.forEach((value) => {
    if (seen.has(value)) {
      duplicates.add(value);
      return;
    }
    seen.add(value);
  });
  return Array.from(duplicates);
}

function formatFactionList(factions: number[]): string {
  return factions.map((faction) => factionLabels[faction] ?? `Faction ${faction}`).join(", ");
}

export function TurnStatePanel({
  state,
  onUpdateState,
  title = "Advanced Turn State",
  showCloseButton = true,
  onClose
}: TurnStatePanelProps) {
  const decreePreviewGroups = [
    { label: "Recruit", cardIDs: state.eyrie.decree.recruit },
    { label: "Move", cardIDs: state.eyrie.decree.move },
    { label: "Battle", cardIDs: state.eyrie.decree.battle },
    { label: "Build", cardIDs: state.eyrie.decree.build }
  ];
  const allianceSupporterLabels = state.alliance.supporters.map(describeVisibleCard);
  const resolvedDecreeLabels = state.turnProgress.resolvedDecreeCardIDs.map(describeKnownCardID);
  const missingTurnOrderFactions = factionLabels.map((_, index) => index).filter((faction) => !state.turnOrder.includes(faction));
  const duplicateTurnOrderFactions = duplicateValues(state.turnOrder);
  const currentPhaseLabel = phaseLabels[state.currentPhase] ?? `Phase ${state.currentPhase}`;
  const currentStepLabel = stepLabels[state.currentStep] ?? `Step ${state.currentStep}`;

  return (
    <section className="panel modal-panel">
      <div className="panel-header">
        <h2>{title}</h2>
        {showCloseButton ? (
          <button type="button" className="secondary" onClick={onClose}>
            Close
          </button>
        ) : null}
      </div>

      <div className="flow-guide-hero turn-state-hero">
        <span className="summary-label">Assist Recovery Surface</span>
        <strong>{factionLabels[state.factionTurn] ?? "Unknown"} - {currentPhaseLabel}</strong>
        <span className="summary-line">
          Use this when the observed table state and the app's turn flow have drifted. Prefer checkpoint shortcuts before editing faction internals.
        </span>
      </div>

      <div className="flow-step-list turn-state-guide">
        <div className="flow-step-card active">
          <strong>1. Confirm the turn owner</strong>
          <span className="summary-line">
            Current target is {factionLabels[state.factionTurn] ?? "Unknown"} at {currentPhaseLabel} / {currentStepLabel}.
          </span>
        </div>
        <div className="flow-step-card note">
          <strong>2. Use checkpoint shortcuts first</strong>
          <span className="summary-line">
            Birdsong, Daylight, and Evening only reset broad phase/step position. They are safer than editing counters directly.
          </span>
        </div>
        <div className="flow-step-card waiting">
          <strong>3. Edit internals only for real drift</strong>
          <span className="summary-line">
            Decree cards, supporter visibility, Vagabond position, and turn-progress counters are direct recovery controls.
          </span>
        </div>
      </div>

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

      {missingTurnOrderFactions.length > 0 || duplicateTurnOrderFactions.length > 0 ? (
        <div className="observed-issue-list" aria-live="polite">
          {missingTurnOrderFactions.length > 0 ? (
            <span className="message observed-issue error">
              Turn order is missing: {formatFactionList(missingTurnOrderFactions)}.
            </span>
          ) : null}
          {duplicateTurnOrderFactions.length > 0 ? (
            <span className="message observed-issue error">
              Turn order has duplicates: {formatFactionList(duplicateTurnOrderFactions)}.
            </span>
          ) : null}
        </div>
      ) : null}

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
          <TokenListEditor
            label="Eyrie Decree Recruit"
            values={state.eyrie.decree.recruit}
            onChange={(values) =>
              onUpdateState((draft) => {
                draft.eyrie.decree.recruit = values;
              })
            }
            formatValue={describeKnownCardID}
            minValue={-2}
            validateValue={isValidDecreeCardID}
            placeholder="Add recruit decree cards"
          />
          <TokenListEditor
            label="Eyrie Decree Move"
            values={state.eyrie.decree.move}
            onChange={(values) =>
              onUpdateState((draft) => {
                draft.eyrie.decree.move = values;
              })
            }
            formatValue={describeKnownCardID}
            minValue={-2}
            validateValue={isValidDecreeCardID}
            placeholder="Add move decree cards"
          />
          <TokenListEditor
            label="Eyrie Decree Battle"
            values={state.eyrie.decree.battle}
            onChange={(values) =>
              onUpdateState((draft) => {
                draft.eyrie.decree.battle = values;
              })
            }
            formatValue={describeKnownCardID}
            minValue={-2}
            validateValue={isValidDecreeCardID}
            placeholder="Add battle decree cards"
          />
          <TokenListEditor
            label="Eyrie Decree Build"
            values={state.eyrie.decree.build}
            onChange={(values) =>
              onUpdateState((draft) => {
                draft.eyrie.decree.build = values;
              })
            }
            formatValue={describeKnownCardID}
            minValue={-2}
            validateValue={isValidDecreeCardID}
            placeholder="Add build decree cards"
          />
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

        <div className="summary-stack" style={{ marginTop: "1rem" }}>
          <span className="summary-label">Card Readout</span>
          <div className="observed-reference-grid">
            {decreePreviewGroups.map((group) => (
              <ReferenceCard
                key={group.label}
                label={`Decree ${group.label}`}
                items={group.cardIDs.map((cardID, index) => ({
                  key: `${group.label}-${cardID}-${index}`,
                  label: describeKnownCardID(cardID)
                }))}
              />
            ))}
            <ReferenceCard
              label="Alliance Supporters"
              items={allianceSupporterLabels.map((label, index) => ({ key: `${label}-${index}`, label }))}
              emptyLabel="None visible"
            />
            <ReferenceCard
              label="Resolved Decree Cards"
              items={resolvedDecreeLabels.map((label, index) => ({ key: `${label}-${index}`, label }))}
            />
            <ReferenceCard
              label="Used Workshop Clearings"
              items={state.turnProgress.usedWorkshopClearings.map((clearingID, index) => ({
                key: `${clearingID}-${index}`,
                label: `Clearing ${clearingID}`
              }))}
            />
          </div>
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
    </section>
  );
}
