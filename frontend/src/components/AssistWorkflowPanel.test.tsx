import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { ACTION_TYPE } from "../labels";
import { sampleState } from "../sampleState";
import type { Action, GameState } from "../types";
import type { AssistActionCandidateRef } from "../assistDirector";
import { AssistWorkflowPanel } from "./AssistWorkflowPanel";

function observedState(overrides: Partial<GameState> = {}): GameState {
  return {
    ...structuredClone(sampleState),
    gameMode: 1,
    gamePhase: 1,
    playerFaction: 3,
    factionTurn: 2,
    currentPhase: 1,
    currentStep: 3,
    ...overrides
  };
}

function renderPanel(options: {
  state?: GameState;
  actions?: Action[];
  surface?: "sidebar" | "tray";
  onApply?: (action: Action) => Promise<void>;
  onGenerateActions?: () => Promise<void>;
  onOpenTurnState?: () => void;
  onOpenBattle?: (actionIndex: number) => void;
  onBattleCandidatesChange?: (candidates: AssistActionCandidateRef[]) => void;
  onMovementCandidatesChange?: (candidates: AssistActionCandidateRef[]) => void;
  onBuildRecruitCandidatesChange?: (candidates: AssistActionCandidateRef[]) => void;
  onFactionSpatialCandidatesChange?: (candidates: AssistActionCandidateRef[]) => void;
} = {}) {
  return render(
    <AssistWorkflowPanel
      state={options.state ?? observedState()}
      actions={options.actions ?? []}
      surface={options.surface}
      onApply={options.onApply ?? vi.fn(async () => undefined)}
      onGenerateActions={options.onGenerateActions ?? vi.fn(async () => undefined)}
      onOpenTurnState={options.onOpenTurnState ?? vi.fn()}
      onOpenBattle={options.onOpenBattle ?? vi.fn()}
      onBattleCandidatesChange={options.onBattleCandidatesChange}
      onMovementCandidatesChange={options.onMovementCandidatesChange}
      onBuildRecruitCandidatesChange={options.onBuildRecruitCandidatesChange}
      onFactionSpatialCandidatesChange={options.onFactionSpatialCandidatesChange}
    />
  );
}

describe("AssistWorkflowPanel", () => {
  it("auto-loads public candidates once for an observed turn", async () => {
    const onGenerateActions = vi.fn(async () => undefined);

    renderPanel({ actions: [], onGenerateActions });

    expect(screen.getAllByText(/Reading the public board state/i)).toHaveLength(2);
    await waitFor(() => expect(onGenerateActions).toHaveBeenCalledTimes(1));
  });

  it("uses a compact tray summary with secondary table-notes controls in tray mode", () => {
    const moveAction: Action = {
      type: ACTION_TYPE.MOVEMENT,
      movement: {
        faction: 2,
        count: 1,
        maxCount: 1,
        from: 3,
        to: 7,
        fromForestID: 0,
        toForestID: 0,
        decreeCardID: 0,
        sourceEffectID: ""
      }
    };

    renderPanel({ surface: "tray", actions: [moveAction] });

    expect(screen.getByText(/Record what happened on the table/i)).toBeInTheDocument();
    expect(screen.getByText(/Table Notes & Hidden Info/i)).toBeInTheDocument();
    expect(screen.queryByText(/^Assist Workflow$/i)).not.toBeInTheDocument();
  });

  it("uses a craft choice prompt instead of applying from a raw candidate list", async () => {
    const craftAction: Action = {
      type: ACTION_TYPE.CRAFT,
      craft: {
        faction: 2,
        cardID: 7,
        usedWorkshopClearings: [3]
      }
    };
    const onApply = vi.fn(async () => undefined);

    renderPanel({ actions: [craftAction], onApply });

    fireEvent.click(screen.getByRole("button", { name: /Craft/i }));

    expect(screen.getByText(/Choose the crafted card first/i)).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: /Royal Claim/i }));

    await waitFor(() => expect(onApply).toHaveBeenCalledWith(craftAction));
  });

  it("resolves ambiguous craft routes with a workshop-route choice", async () => {
    const workshopThree: Action = {
      type: ACTION_TYPE.CRAFT,
      craft: {
        faction: 2,
        cardID: 7,
        usedWorkshopClearings: [3]
      }
    };
    const workshopEleven: Action = {
      type: ACTION_TYPE.CRAFT,
      craft: {
        faction: 2,
        cardID: 7,
        usedWorkshopClearings: [11]
      }
    };
    const onApply = vi.fn(async () => undefined);

    renderPanel({ actions: [workshopThree, workshopEleven], onApply });

    fireEvent.click(screen.getByRole("button", { name: /Craft/i }));
    fireEvent.click(screen.getByRole("button", { name: /Royal Claim.*2 workshop paths/i }));

    expect(onApply).not.toHaveBeenCalled();
    expect(screen.getByRole("button", { name: /Workshops 3/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Workshops 11/i })).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /Workshops 11/i }));

    await waitFor(() => expect(onApply).toHaveBeenCalledWith(workshopEleven));
  });

  it("reports movement candidates only after the movement intent is selected", async () => {
    const movementAction: Action = {
      type: ACTION_TYPE.MOVEMENT,
      movement: {
        faction: 2,
        count: 1,
        maxCount: 1,
        from: 3,
        to: 7,
        fromForestID: 0,
        toForestID: 0,
        decreeCardID: 0,
        sourceEffectID: ""
      }
    };
    const onMovementCandidatesChange = vi.fn();

    renderPanel({ actions: [movementAction], onMovementCandidatesChange });

    await waitFor(() => expect(onMovementCandidatesChange).toHaveBeenLastCalledWith([]));

    fireEvent.click(screen.getByRole("button", { name: /Move/i }));

    await waitFor(() => expect(onMovementCandidatesChange).toHaveBeenLastCalledWith([{ actionIndex: 0, action: movementAction }]));
  });

  it("keeps exact generated candidates closed by default for normal observed intents", () => {
    const movementAction: Action = {
      type: ACTION_TYPE.MOVEMENT,
      movement: {
        faction: 2,
        count: 1,
        maxCount: 1,
        from: 3,
        to: 7,
        fromForestID: 0,
        toForestID: 0,
        decreeCardID: 0,
        sourceEffectID: ""
      }
    };

    const { container } = renderPanel({ actions: [movementAction] });

    fireEvent.click(screen.getByRole("button", { name: /Move/i }));

    expect(container.querySelector(".assist-exact-candidate-drawer")?.hasAttribute("open")).toBe(false);
  });

  it("reports build and recruit candidates only after the Build / Recruit intent is selected", async () => {
    const buildAction: Action = {
      type: ACTION_TYPE.BUILD,
      build: {
        faction: 0,
        clearingID: 1,
        buildingType: 0,
        woodSources: [{ clearingID: 1, amount: 1 }],
        decreeCardID: 0
      }
    };
    const onBuildRecruitCandidatesChange = vi.fn();

    renderPanel({ actions: [buildAction], onBuildRecruitCandidatesChange });

    await waitFor(() => expect(onBuildRecruitCandidatesChange).toHaveBeenLastCalledWith([]));

    fireEvent.click(screen.getByRole("button", { name: /Build \/ Recruit/i }));

    await waitFor(() => expect(onBuildRecruitCandidatesChange).toHaveBeenLastCalledWith([{ actionIndex: 0, action: buildAction }]));
  });

  it("reports clearing-based faction action candidates only after Faction Action is selected", async () => {
    const organizeAction: Action = {
      type: ACTION_TYPE.ORGANIZE,
      organize: {
        faction: 1,
        clearingID: 4
      }
    };
    const onFactionSpatialCandidatesChange = vi.fn();

    renderPanel({ state: observedState({ factionTurn: 1 }), actions: [organizeAction], onFactionSpatialCandidatesChange });

    await waitFor(() => expect(onFactionSpatialCandidatesChange).toHaveBeenLastCalledWith([]));

    fireEvent.click(screen.getByRole("button", { name: /Faction Action/i }));

    await waitFor(() => expect(onFactionSpatialCandidatesChange).toHaveBeenLastCalledWith([{ actionIndex: 0, action: organizeAction }]));
  });

  it("applies ambiguous Spread Sympathy choices from card-specific faction prompts", async () => {
    const royalSupporter: Action = {
      type: ACTION_TYPE.SPREAD_SYMPATHY,
      spreadSympathy: {
        faction: 1,
        clearingID: 4,
        supporterCardIDs: [7]
      }
    };
    const bankSupporter: Action = {
      type: ACTION_TYPE.SPREAD_SYMPATHY,
      spreadSympathy: {
        faction: 1,
        clearingID: 4,
        supporterCardIDs: [15]
      }
    };
    const onApply = vi.fn(async () => undefined);

    renderPanel({ state: observedState({ factionTurn: 1 }), actions: [royalSupporter, bankSupporter], onApply });

    fireEvent.click(screen.getByRole("button", { name: /Faction Action/i }));
    fireEvent.click(screen.getByRole("button", { name: /Spread Sympathy to clearing 4.*Better Burrow Bank/i }));

    await waitFor(() => expect(onApply).toHaveBeenCalledWith(bankSupporter));
  });

  it("applies ambiguous Aid choices from card-specific faction prompts", async () => {
    const royalAid: Action = {
      type: ACTION_TYPE.AID,
      aid: {
        faction: 3,
        targetFaction: 0,
        clearingID: 6,
        cardID: 7
      }
    };
    const bankAid: Action = {
      type: ACTION_TYPE.AID,
      aid: {
        faction: 3,
        targetFaction: 0,
        clearingID: 6,
        cardID: 15
      }
    };
    const onApply = vi.fn(async () => undefined);

    renderPanel({ state: observedState({ factionTurn: 3, playerFaction: 1 }), actions: [royalAid, bankAid], onApply });

    fireEvent.click(screen.getByRole("button", { name: /Faction Action/i }));
    fireEvent.click(screen.getByRole("button", { name: /Aid Marquise in clearing 6.*Royal Claim/i }));

    await waitFor(() => expect(onApply).toHaveBeenCalledWith(royalAid));
  });

  it("opens battle flow from a defender target choice", async () => {
    const battleAction: Action = {
      type: ACTION_TYPE.BATTLE,
      battle: {
        faction: 2,
        clearingID: 3,
        targetFaction: 0,
        decreeCardID: 0,
        sourceEffectID: ""
      }
    };
    const onOpenBattle = vi.fn();

    renderPanel({ actions: [battleAction], onOpenBattle });

    fireEvent.click(screen.getByRole("button", { name: /A battle started/i }));
    expect(screen.getByText(/Choose the observed battle target/i)).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /Clearing 3: Marquise/i }));

    expect(onOpenBattle).toHaveBeenCalledWith(0);
  });

  it("resolves ambiguous decree additions with a card choice before column assignment", async () => {
    const addToRecruit: Action = {
      type: ACTION_TYPE.ADD_TO_DECREE,
      addToDecree: {
        faction: 2,
        cardIDs: [7],
        columns: [0]
      }
    };
    const addToMove: Action = {
      type: ACTION_TYPE.ADD_TO_DECREE,
      addToDecree: {
        faction: 2,
        cardIDs: [7],
        columns: [1]
      }
    };
    const onApply = vi.fn(async () => undefined);

    renderPanel({ actions: [addToRecruit, addToMove], onApply });

    fireEvent.click(screen.getByRole("button", { name: /Faction Action/i }));
    fireEvent.click(screen.getByRole("button", { name: /Royal Claim.*column choices/i }));

    expect(onApply).not.toHaveBeenCalled();
    expect(screen.getByRole("button", { name: /Royal Claim.*Recruit/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Royal Claim.*Move/i })).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /Royal Claim.*Move/i }));

    await waitFor(() => expect(onApply).toHaveBeenCalledWith(addToMove));
  });

  it("applies Draw / Advance choices from direct choice cards", async () => {
    const passAction: Action = {
      type: ACTION_TYPE.PASS_PHASE,
      passPhase: {
        faction: 2
      }
    };
    const onApply = vi.fn(async () => undefined);

    renderPanel({ actions: [passAction], onApply });

    fireEvent.click(screen.getByRole("button", { name: /Draw \/ Advance/i }));
    fireEvent.click(screen.getByRole("button", { name: /Advance phase/i }));

    await waitFor(() => expect(onApply).toHaveBeenCalledWith(passAction));
  });

  it("applies Card Effect choices from direct choice cards", async () => {
    const playAction: Action = {
      type: ACTION_TYPE.OTHER_PLAYER_PLAY,
      otherPlayerPlay: {
        faction: 2,
        cardID: 7
      }
    };
    const onApply = vi.fn(async () => undefined);

    renderPanel({ actions: [playAction], onApply });

    fireEvent.click(screen.getByRole("button", { name: /Card Effect/i }));
    fireEvent.click(screen.getByRole("button", { name: /Play Royal Claim/i }));

    await waitFor(() => expect(onApply).toHaveBeenCalledWith(playAction));
  });

  it("applies Vagabond quest reward choices from direct faction choice cards", async () => {
    const questAction: Action = {
      type: ACTION_TYPE.QUEST,
      quest: {
        faction: 3,
        questID: 1,
        itemIndexes: [0, 1],
        reward: 1
      }
    };
    const onApply = vi.fn(async () => undefined);

    renderPanel({ state: observedState({ factionTurn: 3, playerFaction: 0 }), actions: [questAction], onApply });

    fireEvent.click(screen.getByRole("button", { name: /Faction Action/i }));
    fireEvent.click(screen.getByRole("button", { name: /Expel Bandits: Draw 2 cards/i }));

    await waitFor(() => expect(onApply).toHaveBeenCalledWith(questAction));
  });

  it("applies Vagabond repair choices from direct faction choice cards", async () => {
    const repairAction: Action = {
      type: ACTION_TYPE.REPAIR,
      repair: {
        faction: 3,
        itemIndex: 4
      }
    };
    const onApply = vi.fn(async () => undefined);

    renderPanel({ state: observedState({ factionTurn: 3, playerFaction: 0 }), actions: [repairAction], onApply });

    fireEvent.click(screen.getByRole("button", { name: /Faction Action/i }));
    fireEvent.click(screen.getByRole("button", { name: /Repair Tea #5/i }));

    await waitFor(() => expect(onApply).toHaveBeenCalledWith(repairAction));
  });

  it("keeps hidden observed events behind manual capture", () => {
    renderPanel({ actions: [] });

    fireEvent.click(screen.getByRole("button", { name: /Hidden Draw/i }));

    expect(screen.getByRole("heading", { name: /Observed Turn Tools/i })).toBeInTheDocument();
    expect(screen.getByText(/Use for hidden draws when only the count is known/i)).toBeInTheDocument();
  });
});
