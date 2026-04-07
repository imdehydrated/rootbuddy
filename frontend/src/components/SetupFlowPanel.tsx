import { factionLabels } from "../labels";

type MarquiseSetupDraft = {
  keepClearingID: number | null;
  sawmillClearingID: number | null;
  workshopClearingID: number | null;
  recruiterClearingID: number | null;
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
      if (draft.recruiterClearingID === null) {
        return "Click a legal clearing for the starting recruiter.";
      }
      return "Applying the Marquise setup.";
    case 2:
      return "Click a highlighted corner clearing.";
    case 3:
      return "Click a forest marker on the board.";
    default:
      return "Follow the highlighted setup choices.";
  }
}

function setupStages() {
  return [
    { stage: 1, label: "Marquise" },
    { stage: 2, label: "Eyrie" },
    { stage: 3, label: "Vagabond" }
  ];
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

      <div className="setup-stage-strip">
        {setupStages().map((entry) => {
          const stateLabel = entry.stage < stage ? "done" : entry.stage === stage ? "active" : "upcoming";
          return (
            <div key={entry.stage} className={`setup-stage-pill ${stateLabel}`}>
              <span>{entry.label}</span>
              <strong>{entry.stage < stage ? "Done" : entry.stage === stage ? "Current" : "Pending"}</strong>
            </div>
          );
        })}
      </div>

      {stage === 1 ? (
        <div className="summary-stack" style={{ marginTop: "0.9rem" }}>
          <span className="summary-label">Marquise Draft</span>
          <div className="flow-step-list">
            <div className={`flow-step-card ${marquiseDraft.keepClearingID === null ? "active" : "done"}`}>
              <strong>Keep</strong>
              <span className="summary-line">{marquiseDraft.keepClearingID ?? "Pending"}</span>
            </div>
            <div className={`flow-step-card ${marquiseDraft.sawmillClearingID === null ? "active" : "done"}`}>
              <strong>Sawmill</strong>
              <span className="summary-line">{marquiseDraft.sawmillClearingID ?? "Pending"}</span>
            </div>
            <div className={`flow-step-card ${marquiseDraft.workshopClearingID === null ? "active" : "done"}`}>
              <strong>Workshop</strong>
              <span className="summary-line">{marquiseDraft.workshopClearingID ?? "Pending"}</span>
            </div>
            <div className={`flow-step-card ${marquiseDraft.recruiterClearingID === null ? "active" : "done"}`}>
              <strong>Recruiter</strong>
              <span className="summary-line">{marquiseDraft.recruiterClearingID ?? "Pending"}</span>
            </div>
          </div>
          <button type="button" className="secondary" onClick={onResetMarquiseDraft}>
            Reset Marquise Choice
          </button>
        </div>
      ) : null}
    </section>
  );
}
