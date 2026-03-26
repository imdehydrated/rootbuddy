import { factionLabels } from "../labels";

type MarquiseSetupDraft = {
  keepClearingID: number | null;
  sawmillClearingID: number | null;
  workshopClearingID: number | null;
};

type SetupFlowPanelProps = {
  stage: number;
  activeFaction: number;
  legalChoiceCount: number;
  marquiseDraft: MarquiseSetupDraft;
  onResetMarquiseDraft: () => void;
};

function stageTitle(stage: number): string {
  switch (stage) {
    case 1:
      return "Choose the Marquise setup";
    case 2:
      return "Choose the Eyrie starting clearing";
    case 3:
      return "Choose the Vagabond starting forest";
    default:
      return "Setup";
  }
}

function stageInstruction(stage: number, draft: MarquiseSetupDraft): string {
  switch (stage) {
    case 1:
      if (draft.keepClearingID === null) {
        return "Click a corner clearing for the Keep.";
      }
      if (draft.sawmillClearingID === null) {
        return "Click a legal clearing for the starting sawmill.";
      }
      if (draft.workshopClearingID === null) {
        return "Click a legal clearing for the starting workshop.";
      }
      return "Click a legal clearing for the starting recruiter.";
    case 2:
      return "Click a highlighted corner clearing.";
    case 3:
      return "Click a forest marker on the board.";
    default:
      return "Follow the highlighted setup choices.";
  }
}

export function SetupFlowPanel({
  stage,
  activeFaction,
  legalChoiceCount,
  marquiseDraft,
  onResetMarquiseDraft
}: SetupFlowPanelProps) {
  return (
    <section className="panel sidebar-panel">
      <p className="eyebrow">Setup</p>
      <div className="summary-stack">
        <span className="summary-label">{stageTitle(stage)}</span>
        <strong>{factionLabels[activeFaction] ?? "Unknown"}</strong>
        <span className="summary-line">{stageInstruction(stage, marquiseDraft)}</span>
        <span className="summary-line">Legal choices: {legalChoiceCount}</span>
      </div>

      {stage === 1 ? (
        <div className="summary-stack" style={{ marginTop: "0.9rem" }}>
          <span className="summary-label">Marquise Draft</span>
          <span className="summary-line">Keep: {marquiseDraft.keepClearingID ?? "Pending"}</span>
          <span className="summary-line">Sawmill: {marquiseDraft.sawmillClearingID ?? "Pending"}</span>
          <span className="summary-line">Workshop: {marquiseDraft.workshopClearingID ?? "Pending"}</span>
          <button type="button" className="secondary" onClick={onResetMarquiseDraft}>
            Reset Marquise Choice
          </button>
        </div>
      ) : null}
    </section>
  );
}
