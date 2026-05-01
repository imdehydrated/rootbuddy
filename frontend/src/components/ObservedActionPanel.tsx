import { useEffect, useState } from "react";
import { describeKnownCardID } from "../cardCatalog";
import { ACTION_TYPE, factionLabels, suitLabels } from "../labels";
import type { Action, GameState } from "../types";
import { ReferenceCard } from "./CardUi";
import { TokenListEditor } from "./TokenListEditor";

type ObservedActionPanelProps = {
  state: GameState;
  onApply: (action: Action) => Promise<void>;
  onClose?: () => void;
  embedded?: boolean;
  preferredActorFaction?: number | null;
  preferredTemplate?: ObservedTemplateKey | null;
};

export type ObservedTemplateKey =
  | "battle_resolution"
  | "pass_phase"
  | "add_to_decree"
  | "craft"
  | "overwork"
  | "spread_sympathy"
  | "revolt"
  | "mobilize"
  | "train"
  | "aid"
  | "evening_discard"
  | "other_player_draw"
  | "other_player_play"
  | "activate_dominance"
  | "take_dominance";

type ObservedFormState = {
  actorFaction: number;
  template: ObservedTemplateKey;
  cardID: string;
  itemIndex: string;
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
  usedWorkshopClearings: number[];
  supporterCardIDs: number[];
  discardCardIDs: number[];
  decreeCardIDs: number[];
  decreeColumns: number[];
  defenderAmbushed: boolean;
  attackerCounterAmbush: boolean;
  attackerUsedArmorers: boolean;
  defenderUsedArmorers: boolean;
  attackerUsedBrutalTactics: boolean;
  defenderUsedSappers: boolean;
};

type ObservedIssue = {
  message: string;
  blocking?: boolean;
};

const templateLabels: Record<ObservedTemplateKey, string> = {
  battle_resolution: "Battle",
  pass_phase: "Advance Phase",
  add_to_decree: "Add To Decree",
  craft: "Craft",
  overwork: "Overwork",
  spread_sympathy: "Spread Sympathy",
  revolt: "Revolt",
  mobilize: "Mobilize",
  train: "Train",
  aid: "Aid",
  evening_discard: "Evening Discard",
  other_player_draw: "Other Player Draw",
  other_player_play: "Other Player Play",
  activate_dominance: "Activate Dominance",
  take_dominance: "Take Dominance"
};

function parseNumber(value: string, fallback = 0): number {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
}

function templatesForFaction(faction: number): ObservedTemplateKey[] {
  switch (faction) {
    case 0:
      return ["battle_resolution", "craft", "overwork", "evening_discard", "other_player_draw", "other_player_play", "pass_phase", "activate_dominance", "take_dominance"];
    case 1:
      return ["battle_resolution", "spread_sympathy", "revolt", "mobilize", "train", "evening_discard", "other_player_draw", "other_player_play", "pass_phase", "activate_dominance", "take_dominance"];
    case 2:
      return ["battle_resolution", "add_to_decree", "craft", "evening_discard", "other_player_draw", "other_player_play", "pass_phase", "activate_dominance", "take_dominance"];
    case 3:
      return ["battle_resolution", "aid", "craft", "evening_discard", "other_player_draw", "other_player_play", "pass_phase", "activate_dominance", "take_dominance"];
    default:
      return ["battle_resolution", "other_player_draw", "other_player_play", "pass_phase"];
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
    itemIndex: "0",
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
    usedWorkshopClearings: [],
    supporterCardIDs: [24],
    discardCardIDs: [24],
    decreeCardIDs: [24],
    decreeColumns: [1],
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
    case "pass_phase":
      return {
        type: 24,
        passPhase: {
          faction
        }
      };
    case "add_to_decree":
      return {
        type: 7,
        addToDecree: {
          faction,
          cardIDs: form.decreeCardIDs,
          columns: form.decreeColumns
        }
      };
    case "craft":
      return {
        type: 6,
        craft: {
          faction,
          cardID: parseNumber(form.cardID),
          usedWorkshopClearings: form.usedWorkshopClearings
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
          supporterCardIDs: form.supporterCardIDs
        }
      };
    case "revolt":
      return {
        type: 9,
        revolt: {
          faction,
          clearingID: parseNumber(form.clearingID),
          baseSuit: form.baseSuit,
          supporterCardIDs: form.supporterCardIDs
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
          cardID: parseNumber(form.cardID),
          itemIndex: parseNumber(form.itemIndex, 0)
        }
      };
    case "evening_discard":
      if (faction === 3) {
        return {
          type: ACTION_TYPE.VAGABOND_DISCARD,
          vagabondDiscard: {
            faction,
            cardIDs: form.discardCardIDs
          }
        };
      }
      return {
        type: ACTION_TYPE.EVENING_DISCARD,
        eveningDiscard: {
          faction,
          cardIDs: form.discardCardIDs,
          count: form.discardCardIDs.length
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
    case "pass_phase":
      return "Use when the acting faction is done with the current step or phase and you just need to advance turn flow.";
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
    case "evening_discard":
      return "Use when a faction discarded public card identities down to the Evening hand limit.";
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

function integerFieldIssue(label: string, value: string, minValue: number): ObservedIssue | null {
  const parsed = Number(value);
  if (!Number.isInteger(parsed) || parsed < minValue) {
    return {
      message: `${label} must be an integer ${minValue === 0 ? "0 or greater" : `at least ${minValue}`}.`,
      blocking: true
    };
  }
  return null;
}

function observedFormIssues(form: ObservedFormState): ObservedIssue[] {
  const issues: ObservedIssue[] = [];

  if (form.template === "add_to_decree" && form.decreeCardIDs.length !== form.decreeColumns.length) {
    issues.push({
      message: "Each decree card needs a matching decree column.",
      blocking: true
    });
  }

  if ((form.template === "spread_sympathy" || form.template === "revolt") && form.supporterCardIDs.length === 0) {
    issues.push({
      message: "Add the supporter cards spent for this Alliance action.",
      blocking: true
    });
  }

  if (form.template === "craft" && form.usedWorkshopClearings.length === 0) {
    issues.push({
      message: "Add workshop clearings if the crafted card had a cost; leave this empty only for no-cost effects."
    });
  }

  if (form.template === "other_player_draw") {
    const drawIssue = integerFieldIssue("Draw count", form.count, 1);
    if (drawIssue) {
      issues.push(drawIssue);
    }
  }

  if (form.template === "evening_discard" && form.discardCardIDs.length === 0) {
    issues.push({
      message: "Add the public discarded card IDs.",
      blocking: true
    });
  }

  if (form.template === "battle_resolution") {
    const battleIssues = [
      integerFieldIssue("Attacker roll", form.attackerRoll, 0),
      integerFieldIssue("Defender roll", form.defenderRoll, 0),
      integerFieldIssue("Attacker losses", form.attackerLosses, 0),
      integerFieldIssue("Defender losses", form.defenderLosses, 0)
    ].filter((issue): issue is ObservedIssue => issue !== null);
    issues.push(...battleIssues);
  }

  if (form.template === "aid") {
    const itemIssue = integerFieldIssue("Item slot", form.itemIndex, 0);
    if (itemIssue) {
      issues.push(itemIssue);
    }
  }

  return issues;
}

function previewCardLabel(cardID: number): string {
  return describeKnownCardID(cardID);
}

export function ObservedActionPanel({
  state,
  onApply,
  onClose,
  embedded = false,
  preferredActorFaction = null,
  preferredTemplate = null
}: ObservedActionPanelProps) {
  const [form, setForm] = useState<ObservedFormState>(() => initialFormState(state));
  const availableTemplates = templatesForFaction(form.actorFaction);
  const targetFactions = availableTargetFactions(form.actorFaction);
  const validClearingIDs = new Set(state.map.clearings.map((clearing) => clearing.id));

  useEffect(() => {
    setForm(initialFormState(state));
  }, [state.factionTurn, state.playerFaction, state.map.clearings]);

  useEffect(() => {
    if (preferredActorFaction === null) {
      return;
    }

    setForm((current) =>
      current.actorFaction === preferredActorFaction
        ? current
        : {
            ...current,
            actorFaction: preferredActorFaction
          }
    );
  }, [preferredActorFaction]);

  useEffect(() => {
    if (preferredTemplate === null) {
      return;
    }

    setForm((current) =>
      current.template === preferredTemplate
        ? current
        : {
            ...current,
            template: preferredTemplate
          }
    );
  }, [preferredTemplate]);

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
  const observedIssues = observedFormIssues(form);
  const hasBlockingIssue = observedIssues.some((issue) => issue.blocking);
  const enteredCardID = parseNumber(form.cardID);
  const enteredBattleDecreeCardID = parseNumber(form.decreeCardID);
  const enteredDominanceCardID = parseNumber(form.dominanceCardID);
  const enteredSpentCardID = parseNumber(form.spentCardID);
  const enteredSupporterCardIDs = form.supporterCardIDs;
  const enteredDiscardCardIDs = form.discardCardIDs;
  const enteredDecreeCardIDs = form.decreeCardIDs;
  const enteredWorkshopClearings = form.usedWorkshopClearings;
  const enteredDecreeColumns = form.decreeColumns;
  const referenceGroups = [
    enteredCardID > 0 ? { label: "Card", items: [previewCardLabel(enteredCardID)] } : null,
    form.template === "battle_resolution" && enteredBattleDecreeCardID > 0
      ? { label: "Battle Decree Card", items: [previewCardLabel(enteredBattleDecreeCardID)] }
      : null,
    enteredSupporterCardIDs.length > 0
      ? { label: "Supporters", items: enteredSupporterCardIDs.map(previewCardLabel) }
      : null,
    enteredDiscardCardIDs.length > 0
      ? { label: "Discarded Cards", items: enteredDiscardCardIDs.map(previewCardLabel) }
      : null,
    enteredDecreeCardIDs.length > 0
      ? { label: "Decree Cards", items: enteredDecreeCardIDs.map(previewCardLabel) }
      : null,
    (form.template === "activate_dominance" || form.template === "take_dominance") && enteredDominanceCardID > 0
      ? { label: "Dominance Card", items: [previewCardLabel(enteredDominanceCardID)] }
      : null,
    form.template === "take_dominance" && enteredSpentCardID > 0
      ? { label: "Spent Card", items: [previewCardLabel(enteredSpentCardID)] }
      : null,
    form.template === "craft" && enteredWorkshopClearings.length > 0
      ? { label: "Workshops", items: enteredWorkshopClearings.map((clearingID) => `Clearing ${clearingID}`) }
      : null,
    form.template === "aid"
      ? { label: "Item", items: [`Slot ${parseNumber(form.itemIndex, 0)}`] }
      : null,
    form.template === "add_to_decree" && enteredDecreeColumns.length > 0
      ? { label: "Columns", items: enteredDecreeColumns.map((column) => `Column ${column}`) }
      : null
  ].filter((group): group is { label: string; items: string[] } => group !== null);

  const updateForm = <K extends keyof ObservedFormState>(key: K, value: ObservedFormState[K]) => {
    setForm((current) => ({
      ...current,
      [key]: value
    }));
  };

  return (
    <section className={embedded ? "observed-embedded" : "panel modal-panel"}>
      <div className="panel-header">
        <h2>{embedded ? "Observed Turn Tools" : "Observed Action"}</h2>
        {!embedded && onClose ? (
          <button type="button" className="secondary" onClick={onClose}>
            Close
          </button>
        ) : null}
      </div>

      <p className="message">
        {embedded
          ? "Record what the current faction did on the physical board without leaving the main board flow."
          : "Record public table observations when assist mode cannot infer the event cleanly. Use the battle template for public battle outcomes and visible effect usage."}
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

      <div className="flow-step-card note observed-template-guide">
        <strong>{templateLabels[form.template]}</strong>
        <span className="summary-line">{templateHint(form.template)}</span>
      </div>

      {observedIssues.length > 0 ? (
        <div className="observed-issue-list" aria-live="polite">
          {observedIssues.map((issue) => (
            <span key={issue.message} className={`message observed-issue${issue.blocking ? " error" : ""}`}>
              {issue.message}
            </span>
          ))}
        </div>
      ) : null}

      <div className="observed-form-grid">
        {(form.template === "craft" ||
          form.template === "overwork" ||
          form.template === "mobilize" ||
          form.template === "train" ||
          form.template === "aid" ||
          form.template === "other_player_play") ? (
          <label>
            <span>Card</span>
            <input value={form.cardID} onChange={(event) => updateForm("cardID", event.target.value)} />
          </label>
        ) : null}

        {form.template === "other_player_draw" ? (
          <label>
            <span>Draw Count</span>
            <input value={form.count} onChange={(event) => updateForm("count", event.target.value)} />
          </label>
        ) : null}

        {form.template === "aid" ? (
          <label>
            <span>Item Slot</span>
            <input value={form.itemIndex} onChange={(event) => updateForm("itemIndex", event.target.value)} />
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
              <span>Eyrie Decree Card</span>
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
          <TokenListEditor
            label="Supporter Cards"
            values={form.supporterCardIDs}
            onChange={(values) => updateForm("supporterCardIDs", values)}
            formatValue={previewCardLabel}
            placeholder="Add supporter card IDs"
          />
        ) : null}

        {form.template === "evening_discard" ? (
          <TokenListEditor
            label="Discarded Cards"
            values={form.discardCardIDs}
            onChange={(values) => updateForm("discardCardIDs", values)}
            formatValue={previewCardLabel}
            placeholder="Add discarded card IDs"
          />
        ) : null}

        {form.template === "add_to_decree" ? (
          <>
            <TokenListEditor
              label="Decree Cards"
              values={form.decreeCardIDs}
              onChange={(values) => updateForm("decreeCardIDs", values)}
              formatValue={previewCardLabel}
              placeholder="Add decree card IDs"
            />
            <TokenListEditor
              label="Decree Columns"
              values={form.decreeColumns}
              onChange={(values) => updateForm("decreeColumns", values)}
              formatValue={(column) => `Column ${column}`}
              placeholder="Add decree columns"
              allowDuplicates
              validateValue={(column) => column >= 1 && column <= 4}
            />
          </>
        ) : null}

        {form.template === "craft" ? (
          <TokenListEditor
            label="Used Workshop Clearings"
            values={form.usedWorkshopClearings}
            onChange={(values) => updateForm("usedWorkshopClearings", values)}
            formatValue={(clearingID) => `Clearing ${clearingID}`}
            placeholder="Add workshop clearings"
            allowDuplicates
            validateValue={(clearingID) => validClearingIDs.has(clearingID)}
          />
        ) : null}

        {(form.template === "activate_dominance" || form.template === "take_dominance") ? (
          <label>
            <span>Dominance Card</span>
            <input
              value={form.dominanceCardID}
              onChange={(event) => updateForm("dominanceCardID", event.target.value)}
            />
          </label>
        ) : null}

        {form.template === "take_dominance" ? (
          <label>
            <span>Spent Card</span>
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

      {referenceGroups.length > 0 ? (
        <div className="summary-stack" style={{ marginTop: "1rem" }}>
          <span className="summary-label">Card References</span>
          <div className="observed-reference-grid">
            {referenceGroups.map((group) => (
              <ReferenceCard
                key={group.label}
                label={group.label}
                items={group.items.map((item, index) => ({ key: `${group.label}-${item}-${index}`, label: item }))}
              />
            ))}
          </div>
        </div>
      ) : null}

      <details className="secondary-drawer observed-preview-drawer">
        <summary className="panel-summary">
          <span className="summary-label">Raw Action Preview</span>
          <span className="summary-line">Audit the generated payload only when troubleshooting assist recordkeeping.</span>
        </summary>
        <textarea className="state-editor observed-preview" value={actionPreview} readOnly spellCheck={false} />
      </details>

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
          disabled={hasBlockingIssue}
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
