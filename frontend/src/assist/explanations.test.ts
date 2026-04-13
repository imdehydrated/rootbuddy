import { describe, expect, it } from "vitest";

import { ACTION_TYPE } from "../labels";
import { sampleState } from "../sampleState";
import type { Action } from "../types";
import { actionExplanation } from "./explanations";

describe("actionExplanation", () => {
  it("describes Vagabond explore actions in board terms", () => {
    const explanation = actionExplanation(
      {
        type: ACTION_TYPE.EXPLORE,
        explore: {
          faction: 3,
          clearingID: 6,
          itemType: 3
        }
      } satisfies Action,
      sampleState
    );

    expect(explanation).toContain("clearing 6");
    expect(explanation).toContain("Hammer");
  });

  it("explains build actions as engine development", () => {
    const explanation = actionExplanation(
      {
        type: ACTION_TYPE.BUILD,
        build: {
          faction: 0,
          clearingID: 9,
          buildingType: 1,
          woodSources: [{ clearingID: 1, amount: 1 }],
          decreeCardID: 0
        }
      } satisfies Action,
      sampleState
    );

    expect(explanation).toContain("Workshop");
    expect(explanation).toContain("clearing 9");
  });

  it("explains persistent effects with their specific target", () => {
    const explanation = actionExplanation(
      {
        type: ACTION_TYPE.USE_PERSISTENT_EFFECT,
        usePersistentEffect: {
          faction: 0,
          effectID: "stand_and_deliver",
          targetFaction: 2,
          clearingID: 0,
          observedCardID: 0
        }
      } satisfies Action,
      sampleState
    );

    expect(explanation).toContain("Stand and Deliver!");
    expect(explanation).toContain("Eyrie");
  });
});
