import { factionLabels, phaseLabels, setupStageLabels } from "../labels";
import type { Action, GameState } from "../types";

type FlowGuidePanelProps = {
  state: GameState;
  loadedActionCount: number;
  selectedBattleAction: Action | null;
  onGenerateActions: () => Promise<void>;
  onOpenHelp: () => void;
};

type GuideContent = {
  eyebrow: string;
  title: string;
  detail: string;
  checklist: string[];
  primaryLabel?: string;
  primaryAction?: () => void;
};

function setupGuide(state: GameState, loadedActionCount: number, onGenerateActions: () => Promise<void>): GuideContent {
  const stageLabel = setupStageLabels[state.setupStage] ?? "Setup";
  return {
    eyebrow: "Flow Guide",
    title: stageLabel,
    detail:
      loadedActionCount > 0
        ? "Choose one of the highlighted legal setup targets on the board."
        : "Load the legal setup choices, then follow the highlighted targets on the board.",
    checklist:
      state.setupStage === 1
        ? ["Choose the keep corner, then place the sawmill, workshop, and recruiter.", "Apply the staged setup choices directly from the board."]
        : state.setupStage === 2
          ? ["Choose the Eyrie starting clearing from the highlighted corners.", "The game will advance to Vagabond setup automatically."]
          : state.setupStage === 3
            ? ["Choose the Vagabond starting forest from the highlighted forest markers.", "Opening hands are dealt after setup is complete."]
            : ["Follow the highlighted setup targets in order."],
    primaryLabel: loadedActionCount > 0 ? undefined : "Load Setup Choices",
    primaryAction: loadedActionCount > 0 ? undefined : () => void onGenerateActions()
  };
}

function playerTurnGuide(
  state: GameState,
  loadedActionCount: number,
  onGenerateActions: () => Promise<void>
): GuideContent {
  const phaseLabel = phaseLabels[state.currentPhase] ?? "Unknown";
  return {
    eyebrow: "Flow Guide",
    title: `Your Turn: ${phaseLabel}`,
    detail:
      loadedActionCount > 0
        ? "Apply one of the loaded legal actions below. Battles still resolve through the battle flow."
        : "Generate legal actions for the current board and turn state, then apply them from the sidebar.",
    checklist: [
      "Keep the board state current before refreshing actions.",
      "Use Player Turn for routine play.",
      "Click a clearing only when you need a board correction; editing lives outside the normal action flow."
    ],
    primaryLabel: loadedActionCount > 0 ? "Refresh Actions" : "Load Actions",
    primaryAction: () => void onGenerateActions()
  };
}

function observedTurnGuide(
  state: GameState,
  loadedActionCount: number,
  onGenerateActions: () => Promise<void>
): GuideContent {
  const actingFaction = factionLabels[state.factionTurn] ?? "Unknown";
  const phaseLabel = phaseLabels[state.currentPhase] ?? "Unknown";
  return {
    eyebrow: "Flow Guide",
    title: `Observed Turn: ${actingFaction}`,
    detail:
      loadedActionCount > 0
        ? `Record ${actingFaction}'s ${phaseLabel.toLowerCase()} using the generated public actions or the observed-action form below.`
        : `Load generated public actions when the board is current, or record hidden/public events directly in the Observed Turn panel.`,
    checklist: [
      "Use shortcuts for common public actions before filling the observed form manually.",
      "Advance the phase only after the physical board and the app match.",
      "Use Advanced Tools only when you need to correct turn state manually."
    ],
    primaryLabel: loadedActionCount > 0 ? "Refresh Public Actions" : "Load Public Actions",
    primaryAction: () => void onGenerateActions()
  };
}

function battleGuide(selectedBattleAction: Action): GuideContent {
  const battle = selectedBattleAction.battle;
  return {
    eyebrow: "Flow Guide",
    title: "Resolve Selected Battle",
    detail: `${factionLabels[battle?.faction ?? -1] ?? "Attacker"} vs ${factionLabels[battle?.targetFaction ?? -1] ?? "Defender"} in clearing ${battle?.clearingID ?? "?"}. This is the blocking step before the turn can continue.`,
    checklist: [
      "Enter the rolls and any visible effect choices in Battle Flow.",
      "In Assist mode, answer the Ambush prompt before resolving.",
      "Resolve and apply the battle before loading or applying more actions."
    ]
  };
}

function reviewGuide(): GuideContent {
  return {
    eyebrow: "Flow Guide",
    title: "Final Result Review",
    detail: "The match is over. Review the win, final standings, and saved-result options from the Game Over panel.",
    checklist: [
      "Return to Setup keeps this result available for review.",
      "Clear Saved Result removes the resumable copy.",
      "New Game replaces the finished match with a fresh setup."
    ]
  };
}

export function FlowGuidePanel({
  state,
  loadedActionCount,
  selectedBattleAction,
  onGenerateActions,
  onOpenHelp
}: FlowGuidePanelProps) {
  const guide =
    state.gamePhase === 0
      ? setupGuide(state, loadedActionCount, onGenerateActions)
      : state.gamePhase === 2
        ? reviewGuide()
        : selectedBattleAction?.battle
          ? battleGuide(selectedBattleAction)
        : state.factionTurn === state.playerFaction
          ? playerTurnGuide(state, loadedActionCount, onGenerateActions)
          : observedTurnGuide(state, loadedActionCount, onGenerateActions);

  return (
    <section className="panel sidebar-panel flow-guide-panel">
      <p className="eyebrow">{guide.eyebrow}</p>
      <div className="flow-guide-hero">
        <span className="summary-label">{guide.title}</span>
        <span className="summary-line">{guide.detail}</span>
      </div>

      <div className="summary-stack" style={{ marginTop: "0.9rem" }}>
        <span className="summary-label">Next</span>
        {guide.checklist.map((item) => (
          <span key={item} className="summary-line">
            {item}
          </span>
        ))}
      </div>

      <div className="sidebar-actions footer" style={{ marginTop: "0.9rem" }}>
        {guide.primaryLabel && guide.primaryAction ? (
          <button type="button" onClick={guide.primaryAction}>
            {guide.primaryLabel}
          </button>
        ) : null}
        <button type="button" className="secondary" onClick={onOpenHelp}>
          {state.gamePhase === 2 ? "Review Help" : "Help"}
        </button>
      </div>
    </section>
  );
}
