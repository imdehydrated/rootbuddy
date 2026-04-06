import { useMemo, useState } from "react";

type TokenListEditorProps = {
  label: string;
  values: number[];
  onChange: (values: number[]) => void;
  formatValue?: (value: number) => string;
  minValue?: number;
  placeholder?: string;
  allowDuplicates?: boolean;
  validateValue?: (value: number) => boolean;
};

function analyzeDraft(draft: string, minValue: number, validateValue?: (value: number) => boolean) {
  const rawTokens = draft
    .split(",")
    .map((token) => token.trim())
    .filter((token) => token.length > 0);

  const validValues: number[] = [];
  const invalidTokens: string[] = [];

  rawTokens.forEach((token) => {
    const parsed = Number(token);
    if (!Number.isInteger(parsed) || parsed < minValue || (validateValue && !validateValue(parsed))) {
      invalidTokens.push(token);
      return;
    }
    validValues.push(parsed);
  });

  return { rawTokens, validValues, invalidTokens };
}

export function TokenListEditor({
  label,
  values,
  onChange,
  formatValue = (value) => String(value),
  minValue = 1,
  placeholder = "Add comma-separated values",
  allowDuplicates = false,
  validateValue
}: TokenListEditorProps) {
  const [draft, setDraft] = useState("");
  const analysis = useMemo(() => analyzeDraft(draft, minValue, validateValue), [draft, minValue, validateValue]);
  const draftDuplicates = allowDuplicates
    ? []
    : analysis.validValues.filter((value, index, list) => list.indexOf(value) !== index);
  const alreadyPresent = allowDuplicates ? [] : analysis.validValues.filter((value) => values.includes(value));

  const addDraftValues = () => {
    if (analysis.validValues.length === 0) {
      return;
    }

    const additions = allowDuplicates ? analysis.validValues : analysis.validValues.filter((value) => !values.includes(value));
    if (additions.length === 0) {
      return;
    }

    onChange([...values, ...additions]);
    setDraft("");
  };

  return (
    <div className="token-list-editor">
      <span className="summary-label">{label}</span>
      <div className="token-list-entry">
        <input
          type="text"
          value={draft}
          placeholder={placeholder}
          onChange={(event) => setDraft(event.target.value)}
          onKeyDown={(event) => {
            if (event.key === "Enter") {
              event.preventDefault();
              addDraftValues();
            }
          }}
        />
        <button
          type="button"
          className="secondary"
          onClick={addDraftValues}
          disabled={analysis.validValues.length === 0 || (!allowDuplicates && analysis.validValues.every((value) => values.includes(value)))}
        >
          Add
        </button>
      </div>
      <div className="known-card-pill-list">
        {values.length > 0 ? (
          values.map((value, index) => (
            <button
              key={`${label}-${value}-${index}`}
              type="button"
              className="known-card-pill token-list-pill"
              onClick={() => onChange(values.filter((_, itemIndex) => itemIndex !== index))}
            >
              {formatValue(value)} <span className="token-list-remove">x</span>
            </button>
          ))
        ) : (
          <span className="summary-line">No entries added.</span>
        )}
      </div>
      {draft.length > 0 ? (
        <div className="summary-stack">
          {analysis.validValues.length > 0 ? (
            <span className="summary-line">Ready: {analysis.validValues.map((value) => formatValue(value)).join(", ")}</span>
          ) : null}
          {analysis.invalidTokens.length > 0 ? (
            <span className="message">Ignored invalid values: {analysis.invalidTokens.join(", ")}</span>
          ) : null}
          {!allowDuplicates && draftDuplicates.length > 0 ? (
            <span className="message">Duplicate entries in draft: {draftDuplicates.join(", ")}</span>
          ) : null}
          {!allowDuplicates && alreadyPresent.length > 0 ? (
            <span className="message">Already present: {alreadyPresent.join(", ")}</span>
          ) : null}
        </div>
      ) : null}
    </div>
  );
}
