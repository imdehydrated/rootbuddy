import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { ACTION_TYPE } from "../labels";
import { sampleState } from "../sampleState";
import type { Action, GameState } from "../types";
import type { AssistActionCandidateRef } from "../assistDirector";
import { PlayerActionsPanel } from "./PlayerActionsPanel";

function activeState(overrides: Partial<GameState> = {}): GameState {
  return {
    ...structuredClone(sampleState),
    gamePhase: 1,
    gameMode: 1,
    playerFaction: 2,
    factionTurn: 2,
    currentPhase: 1,
    currentStep: 3,
    ...overrides
  };
}

function renderPanel(options: {
  state?: GameState;
  actions?: Action[];
  isMultiplayer?: boolean;
  onApply?: (action: Action) => Promise<void>;
  onGenerateActions?: () => Promise<void>;
  onOpenBattle?: (actionIndex: number) => void;
  onPreviewAction?: (actionIndex: number | null) => void;
  onMovementCandidatesChange?: (candidates: AssistActionCandidateRef[]) => void;
  onBuildRecruitCandidatesChange?: (candidates: AssistActionCandidateRef[]) => void;
  onFactionSpatialCandidatesChange?: (candidates: AssistActionCandidateRef[]) => void;
} = {}) {
  return render(
    <PlayerActionsPanel
      state={options.state ?? activeState()}
      actions={options.actions ?? []}
      isMultiplayer={options.isMultiplayer ?? false}
      onApply={options.onApply ?? vi.fn(async () => undefined)}
      onGenerateActions={options.onGenerateActions ?? vi.fn(async () => undefined)}
      onOpenBattle={options.onOpenBattle ?? vi.fn()}
      onPreviewAction={options.onPreviewAction}
      onMovementCandidatesChange={options.onMovementCandidatesChange}
      onBuildRecruitCandidatesChange={options.onBuildRecruitCandidatesChange}
      onFactionSpatialCandidatesChange={options.onFactionSpatialCandidatesChange}
    />
  );
}

describe("PlayerActionsPanel", () => {
  it("uses an intent-first active-turn flow before applying movement", async () => {
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
    const onApply = vi.fn(async () => undefined);

    renderPanel({ actions: [movementAction], onApply });

    expect(screen.getByRole("button", { name: /Move.*Pieces changed clearings/i })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Move up to 1 from clearing 3 to clearing 7/i })).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /Move.*Pieces changed clearings/i }));
    fireEvent.click(screen.getByRole("button", { name: /Move up to 1 from clearing 3 to clearing 7/i }));

    await waitFor(() => expect(onApply).toHaveBeenCalledWith(movementAction));
  });

  it("reports active movement candidates only after the Move intent is selected", async () => {
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

    fireEvent.click(screen.getByRole("button", { name: /Move.*Pieces changed clearings/i }));

    await waitFor(() => expect(onMovementCandidatesChange).toHaveBeenLastCalledWith([{ actionIndex: 0, action: movementAction }]));
  });

  it("keeps exact legal actions closed by default for normal active intents", () => {
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

    fireEvent.click(screen.getByRole("button", { name: /Move.*Pieces changed clearings/i }));

    expect(container.querySelector(".assist-exact-candidate-drawer")?.hasAttribute("open")).toBe(false);
  });

  it("reports active Build / Recruit candidates only after the intent is selected", async () => {
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

    renderPanel({ state: activeState({ playerFaction: 0, factionTurn: 0 }), actions: [buildAction], onBuildRecruitCandidatesChange });

    await waitFor(() => expect(onBuildRecruitCandidatesChange).toHaveBeenLastCalledWith([]));

    fireEvent.click(screen.getByRole("button", { name: /Build \/ Recruit.*Pieces, wood, or buildings/i }));

    await waitFor(() => expect(onBuildRecruitCandidatesChange).toHaveBeenLastCalledWith([{ actionIndex: 0, action: buildAction }]));
  });

  it("reports active clearing-based Faction Action candidates only after the intent is selected", async () => {
    const organizeAction: Action = {
      type: ACTION_TYPE.ORGANIZE,
      organize: {
        faction: 1,
        clearingID: 4
      }
    };
    const onFactionSpatialCandidatesChange = vi.fn();

    renderPanel({ state: activeState({ playerFaction: 1, factionTurn: 1 }), actions: [organizeAction], onFactionSpatialCandidatesChange });

    await waitFor(() => expect(onFactionSpatialCandidatesChange).toHaveBeenLastCalledWith([]));

    fireEvent.click(screen.getByRole("button", { name: /Faction Action.*faction-specific public step/i }));

    await waitFor(() => expect(onFactionSpatialCandidatesChange).toHaveBeenLastCalledWith([{ actionIndex: 0, action: organizeAction }]));
  });

  it("opens battle flow from an active-turn battle target prompt", () => {
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

    fireEvent.click(screen.getByRole("button", { name: /Battle.*A battle started/i }));
    expect(screen.getByText(/Choose the battle target first/i)).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: /Clearing 3: Marquise/i }));

    expect(onOpenBattle).toHaveBeenCalledWith(0);
  });

  it("keeps ambiguous active-turn battle targets in the exact legal fallback", () => {
    const firstBattle: Action = {
      type: ACTION_TYPE.BATTLE,
      battle: {
        faction: 2,
        clearingID: 3,
        targetFaction: 0,
        decreeCardID: 0,
        sourceEffectID: ""
      }
    };
    const decreeBattle: Action = {
      type: ACTION_TYPE.BATTLE,
      battle: {
        faction: 2,
        clearingID: 3,
        targetFaction: 0,
        decreeCardID: 7,
        sourceEffectID: ""
      }
    };
    const onOpenBattle = vi.fn();

    renderPanel({ actions: [firstBattle, decreeBattle], onOpenBattle });

    fireEvent.click(screen.getByRole("button", { name: /Battle.*A battle started/i }));
    fireEvent.click(screen.getByRole("button", { name: /Clearing 3: Marquise.*2 battle options/i }));

    expect(onOpenBattle).not.toHaveBeenCalled();
    expect(screen.getByText(/still maps to 2 legal battle options/i)).toBeInTheDocument();
    expect(screen.getByText(/Exact Legal Actions/i)).toBeInTheDocument();
  });

  it("applies active-turn craft from a card-first prompt", async () => {
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

    fireEvent.click(screen.getByRole("button", { name: /Craft.*A known card was crafted/i }));
    fireEvent.click(screen.getByRole("button", { name: /Royal Claim/i }));

    await waitFor(() => expect(onApply).toHaveBeenCalledWith(craftAction));
  });

  it("resolves ambiguous active-turn craft routes with a workshop-route prompt", async () => {
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

    fireEvent.click(screen.getByRole("button", { name: /Craft.*A known card was crafted/i }));
    fireEvent.click(screen.getByRole("button", { name: /Royal Claim.*2 craft routes/i }));

    expect(onApply).not.toHaveBeenCalled();
    expect(screen.getByRole("button", { name: /Workshops 3/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Workshops 11/i })).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /Workshops 11/i }));

    await waitFor(() => expect(onApply).toHaveBeenCalledWith(workshopEleven));
  });

  it("applies active-turn Draw / Advance choices from direct choice cards", async () => {
    const passAction: Action = {
      type: ACTION_TYPE.PASS_PHASE,
      passPhase: {
        faction: 2
      }
    };
    const onApply = vi.fn(async () => undefined);

    renderPanel({ actions: [passAction], onApply });

    fireEvent.click(screen.getByRole("button", { name: /Draw \/ Advance.*phase advanced/i }));
    fireEvent.click(screen.getByRole("button", { name: /Advance phase/i }));

    await waitFor(() => expect(onApply).toHaveBeenCalledWith(passAction));
  });

  it("applies active-turn Card Effect choices from direct choice cards", async () => {
    const dominanceAction: Action = {
      type: ACTION_TYPE.ACTIVATE_DOMINANCE,
      activateDominance: {
        faction: 2,
        cardID: 14,
        targetFaction: 2
      }
    };
    const onApply = vi.fn(async () => undefined);

    renderPanel({ actions: [dominanceAction], onApply });

    fireEvent.click(screen.getByRole("button", { name: /Card Effect.*persistent effect/i }));
    fireEvent.click(screen.getByRole("button", { name: /Activate Dominance/i }));

    await waitFor(() => expect(onApply).toHaveBeenCalledWith(dominanceAction));
  });

  it("applies active-turn non-spatial faction choices from direct choice cards", async () => {
    const mobilizeAction: Action = {
      type: ACTION_TYPE.MOBILIZE,
      mobilize: {
        faction: 2,
        cardID: 7
      }
    };
    const onApply = vi.fn(async () => undefined);

    renderPanel({ actions: [mobilizeAction], onApply });

    fireEvent.click(screen.getByRole("button", { name: /Faction Action.*faction-specific public step/i }));
    fireEvent.click(screen.getByRole("button", { name: /Mobilize Royal Claim/i }));

    await waitFor(() => expect(onApply).toHaveBeenCalledWith(mobilizeAction));
  });
});
