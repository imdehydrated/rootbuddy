import { ACTION_TYPE } from "../labels";
import { actionCardReferences, actionContextTags, actionHeadline, createVisibleCardLookup, describeAction, formatCardReferenceLabel } from "../actionPresentation";
import type { AssistIntentGroup, AssistIntentKey } from "../assistDirector";
import type { Action, GameState } from "../types";

type IntentGridProps = {
  groups: AssistIntentGroup[];
  selectedIntent: AssistIntentKey | null;
  countLabel: string;
  onSelect: (intent: AssistIntentKey) => void;
};

type ActionOptionCardProps = {
  state: GameState;
  action: Action;
  actionIndex: number;
  variant: "choice" | "exact";
  onApply: (action: Action) => Promise<void>;
  onOpenBattle: (actionIndex: number) => void;
  onPreviewAction?: (actionIndex: number | null) => void;
  choiceClassName?: string;
};

type ExactActionDrawerProps = {
  title: string;
  summary: string;
  open: boolean;
  actions: Action[];
  allActions: Action[];
  state: GameState;
  onToggle: (open: boolean) => void;
  onApply: (action: Action) => Promise<void>;
  onOpenBattle: (actionIndex: number) => void;
  onPreviewAction?: (actionIndex: number | null) => void;
};

export function IntentGrid({ groups, selectedIntent, countLabel, onSelect }: IntentGridProps) {
  return (
    <div className="assist-intent-grid">
      {groups.map((group) => (
        <button
          key={group.key}
          type="button"
          className={`assist-intent-card ${selectedIntent === group.key ? "selected" : ""}`}
          onClick={() => onSelect(group.key)}
        >
          <strong>{group.label}</strong>
          <span>{group.detail}</span>
          <span className="summary-label">
            {group.actions.length} {countLabel}
          </span>
        </button>
      ))}
    </div>
  );
}

export function ActionOptionCard({
  state,
  action,
  actionIndex,
  variant,
  onApply,
  onOpenBattle,
  onPreviewAction,
  choiceClassName = "assist-choice-card"
}: ActionOptionCardProps) {
  const visibleCardLookup = createVisibleCardLookup(state);
  const Element = variant === "choice" ? "button" : "article";
  const cardClass = variant === "choice" ? choiceClassName : "embedded-action-card";

  return (
    <Element
      type={variant === "choice" ? "button" : undefined}
      className={cardClass}
      onMouseEnter={() => onPreviewAction?.(actionIndex)}
      onMouseLeave={() => onPreviewAction?.(null)}
      onFocus={() => onPreviewAction?.(actionIndex)}
      onBlur={() => onPreviewAction?.(null)}
      onClick={
        variant === "choice"
          ? () => {
              if (action.type === ACTION_TYPE.BATTLE) {
                onOpenBattle(actionIndex);
                return;
              }
              void onApply(action);
            }
          : undefined
      }
    >
      <div className="player-action-card-header">
        <strong>{actionHeadline(action)}</strong>
        <span className="player-action-index">#{actionIndex + 1}</span>
      </div>
      <span className="summary-line">{describeAction(action, state)}</span>
      {actionContextTags(action).length > 0 ? (
        <div className="player-action-chip-row">
          {actionContextTags(action).map((tag, tagIndex) => (
            <span key={`${actionIndex}-${tag}-${tagIndex}`} className="player-action-chip">
              {tag}
            </span>
          ))}
        </div>
      ) : null}
      {actionCardReferences(action).length > 0 ? (
        <div className="player-action-card-links">
          <span className="summary-label">Cards</span>
          <div className="known-card-pill-list">
            {actionCardReferences(action).map((reference, referenceIndex) => (
              <span key={`${actionIndex}-card-${reference.cardID}-${reference.zoneLabel}-${referenceIndex}`} className="known-card-pill">
                {formatCardReferenceLabel(reference, visibleCardLookup)}
              </span>
            ))}
          </div>
        </div>
      ) : null}
      {variant === "exact" ? (
        action.type === ACTION_TYPE.BATTLE ? (
          <button type="button" className="secondary" onClick={() => onOpenBattle(actionIndex)}>
            Resolve
          </button>
        ) : (
          <button type="button" className="secondary" onClick={() => void onApply(action)}>
            Apply
          </button>
        )
      ) : (
        <span className="summary-label">{action.type === ACTION_TYPE.BATTLE ? "Open Battle Flow" : "Apply this action"}</span>
      )}
    </Element>
  );
}

export function ExactActionDrawer({
  title,
  summary,
  open,
  actions,
  allActions,
  state,
  onToggle,
  onApply,
  onOpenBattle,
  onPreviewAction
}: ExactActionDrawerProps) {
  return (
    <details className="secondary-drawer assist-exact-candidate-drawer" open={open} onToggle={(event) => onToggle(event.currentTarget.open)}>
      <summary className="panel-summary">
        <span className="summary-label">{title}</span>
        <span className="summary-line">{summary}</span>
      </summary>
      <div className="embedded-action-list">
        {actions.map((action, index) => {
          const actionIndex = allActions.indexOf(action);
          return (
            <ActionOptionCard
              key={`${action.type}-${actionIndex}-${index}`}
              state={state}
              action={action}
              actionIndex={actionIndex}
              variant="exact"
              onApply={onApply}
              onOpenBattle={onOpenBattle}
              onPreviewAction={onPreviewAction}
            />
          );
        })}
      </div>
    </details>
  );
}
