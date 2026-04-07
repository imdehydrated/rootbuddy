import type { MultiplayerConnectionStatus } from "../multiplayer";
import { factionLabels, phaseLabels, setupStageLabels } from "../labels";
import type { Action, BattlePrompt, GameState } from "../types";

type FlowGuidePanelProps = {
  state: GameState;
  loadedActionCount: number;
  selectedBattleAction: Action | null;
  isMultiplayer: boolean;
  multiplayerConnectionStatus: MultiplayerConnectionStatus;
  multiplayerBattlePrompt: BattlePrompt | null;
  onGenerateActions: () => Promise<void>;
  onOpenHelp: () => void;
};

type GuideContent = {
  eyebrow: string;
  title: string;
  detail: string;
  checklist: Array<{
    title: string;
    detail: string;
    tone?: "active" | "waiting" | "note";
  }>;
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
        ? [
            {
              title: "Stage the Marquise corner",
              detail: "Choose the keep corner first, then complete the sawmill, workshop, and recruiter placements.",
              tone: "active"
            },
            {
              title: "Use the board highlights",
              detail: "Apply the staged setup choices directly from the board instead of editing state manually.",
              tone: "note"
            }
          ]
        : state.setupStage === 2
          ? [
              {
                title: "Choose the Eyrie roost corner",
                detail: "Pick the starting clearing from the highlighted corner options.",
                tone: "active"
              },
              {
                title: "Advance is automatic",
                detail: "The game moves into Vagabond setup immediately after the Eyrie choice resolves.",
                tone: "note"
              }
            ]
        : state.setupStage === 3
            ? [
                {
                  title: "Choose the starting forest",
                  detail: "Select one of the highlighted forest markers on the board.",
                  tone: "active"
                },
                {
                  title: "Hands are dealt after setup",
                  detail: "Once the forest is chosen, setup finishes and opening hands are assigned.",
                  tone: "note"
                }
              ]
            : [{ title: "Follow the highlighted targets", detail: "Apply each setup choice in order.", tone: "active" }],
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
        : "Generate legal actions for the current board and turn state, then apply them from the guided controls.",
    checklist: [
      {
        title: "Use loaded legal actions",
        detail: "Routine play should come from Player Turn instead of manual state edits.",
        tone: "active"
      },
      {
        title: "Refresh only from current board state",
        detail: "If the board is stale, refresh after correcting it so the legal action set stays trustworthy.",
        tone: "note"
      },
      {
        title: "Keep edits exceptional",
        detail: "Click a clearing only when you need a correction. Board editing stays outside the normal flow.",
        tone: "waiting"
      }
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
      {
        title: "Use shortcuts first",
        detail: "Reach for common public actions before filling the observed-action form manually.",
        tone: "active"
      },
      {
        title: "Advance only when synced",
        detail: "Move the phase forward only after the physical board and the app state match.",
        tone: "note"
      },
      {
        title: "Keep Advanced Tools as fallback",
        detail: "Use the deeper editor only when the normal observed-turn path is not enough.",
        tone: "waiting"
      }
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
      {
        title: "Finish the battle first",
        detail: "Battle resolution is the blocking step before the turn can continue.",
        tone: "active"
      },
      {
        title: "Enter rolls and visible effects",
        detail: "Use Battle Flow for the current modifiers and outcome inputs.",
        tone: "note"
      },
      {
        title: "Handle assist prompts before resolve",
        detail: "If Ambush or other prompts are pending, answer them before applying the battle.",
        tone: "waiting"
      }
    ]
  };
}

function multiplayerGuide(
  state: GameState,
  loadedActionCount: number,
  connectionStatus: MultiplayerConnectionStatus,
  battlePrompt: BattlePrompt | null,
  onGenerateActions: () => Promise<void>
): GuideContent {
  if (connectionStatus === "reconnecting" || connectionStatus === "disconnected") {
    return {
      eyebrow: "Flow Guide",
      title: "Realtime Sync Interrupted",
      detail: "The multiplayer session is reconnecting. Wait for the live connection to recover before expecting prompts or fresh actions.",
      checklist: [
        {
          title: "Treat the current UI as stale",
          detail: "Do not rely on the existing action buttons while the connection is recovering.",
          tone: "waiting"
        },
        {
          title: "Wait for the server push",
          detail: "The server remains authoritative and will send the latest state after reconnect.",
          tone: "active"
        },
        {
          title: "Fallback if recovery fails",
          detail: "Return to the lobby or reload using the saved browser session if reconnect does not recover.",
          tone: "note"
        }
      ]
    };
  }

  if (battlePrompt) {
    if (battlePrompt.stage === "defender_response" || battlePrompt.stage === "attacker_response") {
      return {
        eyebrow: "Flow Guide",
        title: battlePrompt.waitingOnFaction === state.playerFaction ? "Your Battle Response" : "Waiting On Battle Response",
        detail:
          battlePrompt.waitingOnFaction === state.playerFaction
            ? "Use Battle Flow to submit the response owned by your faction before the turn can continue."
            : `Waiting on ${factionLabels[battlePrompt.waitingOnFaction] ?? "another player"} to answer the current battle prompt.`,
        checklist: [
          {
            title: "Use the prompt the server exposed",
            detail: "Battle Flow only shows the choices valid for your current perspective.",
            tone: battlePrompt.waitingOnFaction === state.playerFaction ? "active" : "waiting"
          },
          {
            title: "Resolution comes after responses",
            detail: "Once all responses are in, the attacker receives the resolve step.",
            tone: "note"
          },
          {
            title: "Dice stay server-owned",
            detail: "Multiplayer battle randomness remains authoritative on the server.",
            tone: "note"
          }
        ]
      };
    }

    if (battlePrompt.stage === "ready_to_resolve") {
      const attackerFaction = battlePrompt.action.battle?.faction ?? -1;
      return {
        eyebrow: "Flow Guide",
        title: attackerFaction === state.playerFaction ? "Resolve Battle" : "Waiting For Battle Resolution",
        detail:
          attackerFaction === state.playerFaction
            ? "Battle responses are complete. Resolve the battle from Battle Flow to continue the turn."
            : `Battle responses are complete. Waiting on ${factionLabels[attackerFaction] ?? "the attacker"} to resolve.`,
        checklist: [
          {
            title: "Resolve before doing anything else",
            detail: "Do not refresh or apply unrelated actions until the battle fully finishes.",
            tone: attackerFaction === state.playerFaction ? "active" : "waiting"
          },
          {
            title: "Next actions load after server advance",
            detail: "The next legal action set appears automatically after the server updates turn state.",
            tone: "note"
          },
          {
            title: "Server authority still applies",
            detail: "Battle resolution stays authoritative even after all player responses are collected.",
            tone: "note"
          }
        ]
      };
    }
  }

  if (state.factionTurn === state.playerFaction) {
    const phaseLabel = phaseLabels[state.currentPhase] ?? "Unknown";
    return {
      eyebrow: "Flow Guide",
      title: `Your Turn: ${phaseLabel}`,
      detail:
        loadedActionCount > 0
          ? "Apply one of the server-authoritative legal actions below. Battles still use Battle Flow."
          : "Legal actions are loading automatically for your turn. You can still refresh manually if needed.",
      checklist: [
        {
          title: "Trust the server-authoritative action list",
          detail: "The server rejects stale or out-of-turn actions, so keep the flow inside the loaded list.",
          tone: "active"
        },
        {
          title: "Wait for pushed state after each action",
          detail: "Continue only after the websocket update lands and confirms the new authoritative state.",
          tone: "note"
        },
        {
          title: "Expect battle interruptions",
          detail: "Battle prompts can interrupt normal action flow when another faction must respond.",
          tone: "waiting"
        }
      ],
      primaryLabel: loadedActionCount > 0 ? "Refresh Actions" : "Refresh Actions",
      primaryAction: () => void onGenerateActions()
    };
  }

  const actingFaction = factionLabels[state.factionTurn] ?? "Unknown";
  const phaseLabel = phaseLabels[state.currentPhase] ?? "Unknown";
  return {
    eyebrow: "Flow Guide",
    title: `Waiting On ${actingFaction}`,
    detail: `${actingFaction} is taking their ${phaseLabel.toLowerCase()}. Multiplayer clients are passive until the server gives your faction priority.`,
    checklist: [
      {
        title: "Watch for battle prompts",
        detail: "Your faction may still need to answer a reaction window while another player is active.",
        tone: "active"
      },
      {
        title: "Actions will load on handoff",
        detail: "Your legal actions should populate automatically when the turn passes to you.",
        tone: "waiting"
      },
      {
        title: "Use session status while idle",
        detail: "Confirm connection health in the session panel if the game appears stuck.",
        tone: "note"
      }
    ]
  };
}

function reviewGuide(): GuideContent {
  return {
    eyebrow: "Flow Guide",
    title: "Final Result Review",
    detail: "The match is over. Review the win, final standings, and saved-result options from the Game Over panel.",
    checklist: [
      {
        title: "Return to Setup keeps review available",
        detail: "Use it when you want to leave the board but retain the final result for later inspection.",
        tone: "note"
      },
      {
        title: "Clear Saved Result removes the review copy",
        detail: "This deletes the resumable local review state for the finished match.",
        tone: "waiting"
      },
      {
        title: "New Game replaces the result",
        detail: "Start fresh only when you are done reviewing the finished state.",
        tone: "active"
      }
    ]
  };
}

export function FlowGuidePanel({
  state,
  loadedActionCount,
  selectedBattleAction,
  isMultiplayer,
  multiplayerConnectionStatus,
  multiplayerBattlePrompt,
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
        : isMultiplayer
          ? multiplayerGuide(state, loadedActionCount, multiplayerConnectionStatus, multiplayerBattlePrompt, onGenerateActions)
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
        <div className="flow-step-list">
          {guide.checklist.map((item) => (
            <div key={`${item.title}-${item.detail}`} className={`flow-step-card ${item.tone ?? "note"}`}>
              <strong>{item.title}</strong>
              <span className="summary-line">{item.detail}</span>
            </div>
          ))}
        </div>
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
