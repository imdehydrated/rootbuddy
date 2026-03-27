import { useEffect, useState } from "react";
import { factionLabels, suitLabels } from "../labels";
import type { Action, GameState } from "../types";

type ObservedActionPanelProps = {
  state: GameState;
  onApply: (action: Action) => Promise<void>;
  onClose: () => void;
};

type ObservedTemplateKey =
  | "battle_resolution"
  | "add_to_decree"
  | "craft"
  | "overwork"
  | "spread_sympathy"
  | "revolt"
  | "mobilize"
  | "train"
  | "aid"
  | "other_player_draw"
  | "other_player_play"
  | "activate_dominance"
  | "take_dominance";

type ObservedFormState = {
  actorFaction: number;
  template: ObservedTemplateKey;
  cardID: string;
  count: string;
  clearingID: string;
  targetFaction: number;
  attackerRoll: string;
  defenderRoll: string;
  attackerLosses: string;
  defenderLosses: string;
  decreeCardID: string;
  sourceEffectID: string;
  baseSuit: number;
  spentCardID: string;
  dominanceCardID: string;
  usedWorkshopClearings: string;
  supporterCardIDs: string;
  decreeCardIDs: string;
  decreeColumns: string;
  defenderAmbushed: boolean;
  attackerCounterAmbush: boolean;
  attackerUsedArmorers: boolean;
  defenderUsedArmorers: boolean;
  attackerUsedBrutalTactics: boolean;
  defenderUsedSappers: boolean;
};

const templateLabels: Record<ObservedTemplateKey, string> = {
  battle_resolution: "Battle",
  add_to_decree: "Add To Decree",
  craft: "Craft",
  overwork: "Overwork",
  spread_sympathy: "Spread Sympathy",
  revolt: "Revolt",
  mobilize: "Mobilize",
  train: "Train",
  aid: "Aid",
  other_player_draw: "Other Player Draw",
  other_player_play: "Other Player Play",
  activate_dominance: "Activate Dominance",
  take_dominance: "Take Dominance"
};

function parseNumber(value: string, fallback = 0): number {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
}

function parseNumberList(value: string): number[] {
  return value
    .split(",")
    .map((part) => Number(part.trim()))
    .filter((entry) => Number.isFinite(entry) && entry > 0);
}

function formatNumberList(values: number[]): string {
  return values.join(", ");
}

function templatesForFaction(faction: number): ObservedTemplateKey[] {
  switch (faction) {
    case 0:
      return ["battle_resolution", "craft", "overwork", "other_player_draw", "other_player_play", "activate_dominance", "take_dominance"];
    case 1:
      return ["battle_resolution", "spread_sympathy", "revolt", "mobilize", "train", "other_player_draw", "other_player_play", "activate_dominance", "take_dominance"];
    case 2:
      return ["battle_resolution", "add_to_decree", "craft", "other_player_draw", "other_player_play", "activate_dominance", "take_dominance"];
    case 3:
      return ["battle_resolution", "aid", "craft", "other_player_draw", "other_player_play", "activate_dominance", "take_dominance"];
    default:
      return ["battle_resolution", "other_player_draw", "other_player_play"];
  }
}

function availableTargetFactions(actorFaction: number): number[] {
  return factionLabels.map((_, index) => index).filter((index) => index !== actorFaction);
}

function initialFormState(state: GameState): ObservedFormState {
  const actorFaction = state.factionTurn;
  const targetFaction = availableTargetFactions(actorFaction)[0] ?? state.playerFaction;
  return {
    actorFaction,
    template: templatesForFaction(actorFaction)[0],
    cardID: "24",
    count: "1",
    clearingID: String(state.map.clearings[0]?.id ?? 1),
    targetFaction,
    attackerRoll: "0",
    defenderRoll: "0",
    attackerLosses: "0",
    defenderLosses: "0",
    decreeCardID: "0",
    sourceEffectID: "",
    baseSuit: 0,
    spentCardID: "24",
    dominanceCardID: "14",
    usedWorkshopClearings: "",
    supporterCardIDs: "24",
    decreeCardIDs: "24",
    decreeColumns: "1",
    defenderAmbushed: false,
    attackerCounterAmbush: false,
    attackerUsedArmorers: false,
    defenderUsedArmorers: false,
    attackerUsedBrutalTactics: false,
    defenderUsedSappers: false
  };
}

function buildObservedAction(form: ObservedFormState): Action {
  const faction = form.actorFaction;

  switch (form.template) {
    case "battle_resolution":
      return {
        type: 2,
        battleResolution: {
          faction,
          clearingID: parseNumber(form.clearingID),
          targetFaction: form.targetFaction,
          decreeCardID: parseNumber(form.decreeCardID),
          attackerRoll: parseNumber(form.attackerRoll),
          defenderRoll: parseNumber(form.defenderRoll),
          attackerHitModifier: 0,
          defenderHitModifier: 0,
          ignoreHitsToAttacker: false,
          ignoreHitsToDefender: false,
          defenderAmbushed: form.defenderAmbushed,
          attackerCounterAmbush: form.attackerCounterAmbush,
          attackerUsedArmorers: form.attackerUsedArmorers,
          defenderUsedArmorers: form.defenderUsedArmorers,
          attackerUsedBrutalTactics: form.attackerUsedBrutalTactics,
          defenderUsedSappers: form.defenderUsedSappers,
          ambushHitsToAttacker: form.defenderAmbushed && !form.attackerCounterAmbush ? 2 : 0,
          attackerLosses: parseNumber(form.attackerLosses),
          defenderLosses: parseNumber(form.defenderLosses),
          sourceEffectID: form.sourceEffectID.trim()
        }
      };
    case "add_to_decree":
      return {
        type: 7,
        addToDecree: {
          faction,
          cardIDs: parseNumberList(form.decreeCardIDs),
          columns: parseNumberList(form.decreeColumns)
        }
      };
    case "craft":
      return {
        type: 6,
        craft: {
          faction,
          cardID: parseNumber(form.cardID),
          usedWorkshopClearings: parseNumberList(form.usedWorkshopClearings)
        }
      };
    case "overwork":
      return {
        type: 5,
        overwork: {
          faction,
          clearingID: parseNumber(form.clearingID),
          cardID: parseNumber(form.cardID)
        }
      };
    case "spread_sympathy":
      return {
        type: 8,
        spreadSympathy: {
          faction,
          clearingID: parseNumber(form.clearingID),
          supporterCardIDs: parseNumberList(form.supporterCardIDs)
        }
      };
    case "revolt":
      return {
        type: 9,
        revolt: {
          faction,
          clearingID: parseNumber(form.clearingID),
          baseSuit: form.baseSuit,
          supporterCardIDs: parseNumberList(form.supporterCardIDs)
        }
      };
    case "mobilize":
      return {
        type: 10,
        mobilize: {
          faction,
          cardID: parseNumber(form.cardID)
        }
      };
    case "train":
      return {
        type: 11,
        train: {
          faction,
          cardID: parseNumber(form.cardID)
        }
      };
    case "aid":
      return {
        type: 15,
        aid: {
          faction,
          targetFaction: form.targetFaction,
          clearingID: parseNumber(form.clearingID),
          cardID: parseNumber(form.cardID)
        }
      };
    case "other_player_draw":
      return {
        type: 27,
        otherPlayerDraw: {
          faction,
          count: parseNumber(form.count, 1)
        }
      };
    case "other_player_play":
      return {
        type: 28,
        otherPlayerPlay: {
          faction,
          cardID: parseNumber(form.cardID)
        }
      };
    case "activate_dominance":
      return {
        type: 30,
        activateDominance: {
          faction,
          cardID: parseNumber(form.dominanceCardID),
          targetFaction: form.targetFaction
        }
      };
    case "take_dominance":
      return {
        type: 31,
        takeDominance: {
          faction,
          dominanceCardID: parseNumber(form.dominanceCardID),
          spentCardID: parseNumber(form.spentCardID)
        }
      };
    default:
      return {
        type: 24,
        passPhase: {
          faction
        }
      };
  }
}

function templateHint(template: ObservedTemplateKey): string {
  switch (template) {
    case "battle_resolution":
      return "Use when you observed a public battle and know the losses or effect usage. This records the resolved outcome directly.";
    case "add_to_decree":
      return "Use when you observed which card(s) the Eyrie added to decree.";
    case "craft":
      return "Use when another faction crafted a known public card.";
    case "overwork":
      return "Use when Marquise spent a known card to add wood.";
    case "spread_sympathy":
      return "Use when Alliance spent supporters to place sympathy.";
    case "revolt":
      return "Use when Alliance revolted with known supporter IDs and base suit.";
    case "mobilize":
      return "Use when Alliance moved a known hand card into supporters.";
    case "train":
      return "Use when Alliance spent a known hand card to gain an officer.";
    case "aid":
      return "Use when Vagabond gave a known card to another faction.";
    case "other_player_draw":
      return "Use for hidden draws when only the count is known.";
    case "other_player_play":
      return "Use for hidden plays/discards when you know the card ID that left hand.";
    case "activate_dominance":
      return "Use when a faction revealed and activated a dominance card.";
    case "take_dominance":
      return "Use when a faction spent a matching card to take an available dominance card.";
    default:
      return "";
  }
}

export function ObservedActionPanel({ state, onApply, onClose }: ObservedActionPanelProps) {
  const [form, setForm] = useState<ObservedFormState>(() => initialFormState(state));
  const availableTemplates = templatesForFaction(form.actorFaction);
  const targetFactions = availableTargetFactions(form.actorFaction);

  useEffect(() => {
    setForm(initialFormState(state));
  }, [state.factionTurn, state.playerFaction, state.map.clearings]);

  useEffect(() => {
    if (availableTemplates.includes(form.template)) {
      return;
    }

    setForm((current) => ({
      ...current,
      template: availableTemplates[0]
    }));
  }, [availableTemplates, form.template]);

  useEffect(() => {
    if (targetFactions.includes(form.targetFaction)) {
      return;
    }

    setForm((current) => ({
      ...current,
      targetFaction: targetFactions[0] ?? current.targetFaction
    }));
  }, [form.targetFaction, targetFactions]);

  const action = buildObservedAction(form);
  const actionPreview = JSON.stringify(action, null, 2);

  const updateForm = <K extends keyof ObservedFormState>(key: K, value: ObservedFormState[K]) => {
    setForm((current) => ({
      ...current,
      [key]: value
    }));
  };

  return (
    <section className="panel modal-panel">
      <div className="panel-header">
        <h2>Observed Action</h2>
        <button type="button" className="secondary" onClick={onClose}>
          Close
        </button>
      </div>

      <p className="message">
        Record public table observations for factions whose full hands are hidden in assist mode. Use the battle template for
        observed public battle outcomes and effect usage.
      </p>

      <div className="observed-panel-grid">
        <div className="summary-stack">
          <span className="summary-label">Actor</span>
          <select value={form.actorFaction} onChange={(event) => updateForm("actorFaction", Number(event.target.value))}>
            {factionLabels.map((label, index) => (
              <option key={label} value={index}>
                {label}
              </option>
            ))}
          </select>
        </div>

        <div className="summary-stack">
          <span className="summary-label">Observed Action Type</span>
          <select
            value={form.template}
            onChange={(event) => updateForm("template", event.target.value as ObservedTemplateKey)}
          >
            {availableTemplates.map((key) => (
              <option key={key} value={key}>
                {templateLabels[key]}
              </option>
            ))}
          </select>
        </div>
      </div>

      <p className="message">{templateHint(form.template)}</p>

      <div className="observed-form-grid">
        {(form.template === "craft" ||
          form.template === "overwork" ||
          form.template === "mobilize" ||
          form.template === "train" ||
          form.template === "aid" ||
          form.template === "other_player_play") ? (
          <label>
            <span>Card ID</span>
            <input value={form.cardID} onChange={(event) => updateForm("cardID", event.target.value)} />
          </label>
        ) : null}

        {form.template === "other_player_draw" ? (
          <label>
            <span>Draw Count</span>
            <input value={form.count} onChange={(event) => updateForm("count", event.target.value)} />
          </label>
        ) : null}

        {(form.template === "battle_resolution" ||
          form.template === "overwork" ||
          form.template === "spread_sympathy" ||
          form.template === "revolt" ||
          form.template === "aid") ? (
          <label>
            <span>Clearing</span>
            <select value={form.clearingID} onChange={(event) => updateForm("clearingID", event.target.value)}>
              {state.map.clearings.map((clearing) => (
                <option key={clearing.id} value={String(clearing.id)}>
                  Clearing {clearing.id}
                </option>
              ))}
            </select>
          </label>
        ) : null}

        {(form.template === "battle_resolution" || form.template === "aid" || form.template === "activate_dominance") ? (
          <label>
            <span>Target Faction</span>
            <select
              value={form.targetFaction}
              onChange={(event) => updateForm("targetFaction", Number(event.target.value))}
            >
              {targetFactions.map((index) => (
                <option key={factionLabels[index]} value={index}>
                  {factionLabels[index]}
                </option>
              ))}
            </select>
          </label>
        ) : null}

        {form.template === "battle_resolution" ? (
          <>
            <label>
              <span>Attacker Losses</span>
              <input value={form.attackerLosses} onChange={(event) => updateForm("attackerLosses", event.target.value)} />
            </label>
            <label>
              <span>Defender Losses</span>
              <input value={form.defenderLosses} onChange={(event) => updateForm("defenderLosses", event.target.value)} />
            </label>
            <label>
              <span>Attacker Roll</span>
              <input value={form.attackerRoll} onChange={(event) => updateForm("attackerRoll", event.target.value)} />
            </label>
            <label>
              <span>Defender Roll</span>
              <input value={form.defenderRoll} onChange={(event) => updateForm("defenderRoll", event.target.value)} />
            </label>
            <label>
              <span>Eyrie Decree Card ID</span>
              <input value={form.decreeCardID} onChange={(event) => updateForm("decreeCardID", event.target.value)} />
            </label>
            <label>
              <span>Battle Source Effect</span>
              <select value={form.sourceEffectID} onChange={(event) => updateForm("sourceEffectID", event.target.value)}>
                <option value="">Normal Battle</option>
                <option value="command_warren">Command Warren</option>
              </select>
            </label>
          </>
        ) : null}

        {form.template === "revolt" ? (
          <label>
            <span>Base Suit</span>
            <select value={form.baseSuit} onChange={(event) => updateForm("baseSuit", Number(event.target.value))}>
              {suitLabels.slice(0, 3).map((label, index) => (
                <option key={label} value={index}>
                  {label}
                </option>
              ))}
            </select>
          </label>
        ) : null}

        {(form.template === "spread_sympathy" || form.template === "revolt") ? (
          <label>
            <span>Supporter Card IDs</span>
            <input
              value={form.supporterCardIDs}
              onChange={(event) => updateForm("supporterCardIDs", event.target.value)}
            />
          </label>
        ) : null}

        {form.template === "add_to_decree" ? (
          <>
            <label>
              <span>Decree Card IDs</span>
              <input value={form.decreeCardIDs} onChange={(event) => updateForm("decreeCardIDs", event.target.value)} />
            </label>
            <label>
              <span>Decree Columns</span>
              <input value={form.decreeColumns} onChange={(event) => updateForm("decreeColumns", event.target.value)} />
            </label>
          </>
        ) : null}

        {form.template === "craft" ? (
          <label>
            <span>Used Workshop Clearings</span>
            <input
              value={form.usedWorkshopClearings}
              onChange={(event) => updateForm("usedWorkshopClearings", event.target.value)}
            />
          </label>
        ) : null}

        {(form.template === "activate_dominance" || form.template === "take_dominance") ? (
          <label>
            <span>Dominance Card ID</span>
            <input
              value={form.dominanceCardID}
              onChange={(event) => updateForm("dominanceCardID", event.target.value)}
            />
          </label>
        ) : null}

        {form.template === "take_dominance" ? (
          <label>
            <span>Spent Card ID</span>
            <input value={form.spentCardID} onChange={(event) => updateForm("spentCardID", event.target.value)} />
          </label>
        ) : null}
      </div>

      {form.template === "battle_resolution" ? (
        <div className="observed-form-grid" style={{ marginTop: "0.75rem" }}>
          <label className="checkbox-row">
            <input
              type="checkbox"
              checked={form.defenderAmbushed}
              onChange={(event) => updateForm("defenderAmbushed", event.target.checked)}
            />
            <span>Defender Ambushed</span>
          </label>
          <label className="checkbox-row">
            <input
              type="checkbox"
              checked={form.attackerCounterAmbush}
              onChange={(event) => updateForm("attackerCounterAmbush", event.target.checked)}
            />
            <span>Attacker Counter-Ambushed</span>
          </label>
          <label className="checkbox-row">
            <input
              type="checkbox"
              checked={form.attackerUsedArmorers}
              onChange={(event) => updateForm("attackerUsedArmorers", event.target.checked)}
            />
            <span>Attacker Used Armorers</span>
          </label>
          <label className="checkbox-row">
            <input
              type="checkbox"
              checked={form.defenderUsedArmorers}
              onChange={(event) => updateForm("defenderUsedArmorers", event.target.checked)}
            />
            <span>Defender Used Armorers</span>
          </label>
          <label className="checkbox-row">
            <input
              type="checkbox"
              checked={form.attackerUsedBrutalTactics}
              onChange={(event) => updateForm("attackerUsedBrutalTactics", event.target.checked)}
            />
            <span>Attacker Used Brutal Tactics</span>
          </label>
          <label className="checkbox-row">
            <input
              type="checkbox"
              checked={form.defenderUsedSappers}
              onChange={(event) => updateForm("defenderUsedSappers", event.target.checked)}
            />
            <span>Defender Used Sappers</span>
          </label>
        </div>
      ) : null}

      <div className="summary-stack" style={{ marginTop: "1rem" }}>
        <span className="summary-label">Action Preview</span>
        <textarea className="state-editor observed-preview" value={actionPreview} readOnly spellCheck={false} />
      </div>

      <div className="sidebar-actions footer">
        <button
          type="button"
          className="secondary"
          onClick={() => setForm(initialFormState(state))}
        >
          Reset
        </button>
        <button
          type="button"
          onClick={async () => {
            await onApply(action);
          }}
        >
          Apply Observed Action
        </button>
      </div>
    </section>
  );
}
