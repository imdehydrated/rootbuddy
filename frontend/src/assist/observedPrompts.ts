import type { ObservedTemplateKey } from "../components/ObservedActionPanel";
import type { GameState } from "../types";

export function observedPromptTemplates(state: GameState): Array<{ label: string; template: ObservedTemplateKey }> {
  const prompts: Array<{ label: string; template: ObservedTemplateKey }> = [
    { label: "Hidden Draw", template: "other_player_draw" },
    { label: "Known Card Play", template: "other_player_play" }
  ];

  if (state.factionTurn === 2) {
    prompts.push({ label: "Decree Choice", template: "add_to_decree" });
  }
  if (state.factionTurn === 1) {
    prompts.push({ label: "Supporter Spend", template: "spread_sympathy" });
    prompts.push({ label: "Revolt", template: "revolt" });
  }
  if (state.factionTurn === 3) {
    prompts.push({ label: "Aid", template: "aid" });
  }

  prompts.push({ label: "Battle Result", template: "battle_resolution" });
  prompts.push({ label: "Dominance", template: "activate_dominance" });

  return prompts;
}
