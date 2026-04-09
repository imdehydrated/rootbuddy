import { describeKnownCardID } from "../../cardCatalog";
import { eyrieLeaderLabels } from "../../labels";
import type { GameState } from "../../types";
import { ReferenceCard } from "../CardUi";
import { TokenListEditor } from "../TokenListEditor";
import { isValidDecreeCardID, referenceItemsFromCardIDs } from "./utils";

type EyrieEditorProps = {
  state: GameState;
  onUpdateState: (mutator: (draft: GameState) => void) => void;
};

export function EyrieEditor({ state, onUpdateState }: EyrieEditorProps) {
  const decreePreviewGroups = [
    { label: "Recruit", cardIDs: state.eyrie.decree.recruit },
    { label: "Move", cardIDs: state.eyrie.decree.move },
    { label: "Battle", cardIDs: state.eyrie.decree.battle },
    { label: "Build", cardIDs: state.eyrie.decree.build }
  ];
  const resolvedDecreeLabels = state.turnProgress.resolvedDecreeCardIDs.map(describeKnownCardID);

  return (
    <div className="summary-section">
      <h3>Eyrie</h3>
      <div className="control-grid">
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
      </div>

      <div className="summary-stack" style={{ marginTop: "1rem" }}>
        <span className="summary-label">Decree Readout</span>
        <div className="observed-reference-grid">
          {decreePreviewGroups.map((group) => (
            <ReferenceCard key={group.label} label={`Decree ${group.label}`} items={referenceItemsFromCardIDs(group.label, group.cardIDs)} />
          ))}
          <ReferenceCard
            label="Resolved Decree Cards"
            items={resolvedDecreeLabels.map((label, index) => ({ key: `${label}-${index}`, label }))}
          />
        </div>
      </div>
    </div>
  );
}
