import type { GameState } from "../../types";

type MarquiseEditorProps = {
  state: GameState;
};

export function MarquiseEditor({ state }: MarquiseEditorProps) {
  return (
    <div className="summary-section">
      <h3>Marquise</h3>
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
      </div>
    </div>
  );
}
