import type { GameState } from "../../types";
import { ReferenceCard } from "../CardUi";
import { describeVisibleCard } from "./utils";

type AllianceEditorProps = {
  state: GameState;
  onUpdateState: (mutator: (draft: GameState) => void) => void;
};

export function AllianceEditor({ state, onUpdateState }: AllianceEditorProps) {
  const allianceSupporterLabels = state.alliance.supporters.map(describeVisibleCard);

  return (
    <div className="summary-section">
      <h3>Alliance</h3>
      <div className="control-grid">
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
      </div>

      <div className="summary-stack" style={{ marginTop: "1rem" }}>
        <span className="summary-label">Supporter Readout</span>
        <div className="observed-reference-grid">
          <ReferenceCard
            label="Alliance Supporters"
            items={allianceSupporterLabels.map((label, index) => ({ key: `${label}-${index}`, label }))}
            emptyLabel="None visible"
          />
        </div>
      </div>
    </div>
  );
}
