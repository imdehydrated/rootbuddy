type CardPillItem = {
  key: string;
  label: string;
};

type KnownCardPillListProps = {
  items: CardPillItem[];
  emptyLabel?: string;
};

type ReferenceCardProps = {
  label: string;
  items: CardPillItem[];
  emptyLabel?: string;
};

export function KnownCardPillList({ items, emptyLabel = "None" }: KnownCardPillListProps) {
  if (items.length === 0) {
    return <span>{emptyLabel}</span>;
  }

  return (
    <div className="known-card-pill-list">
      {items.map((item) => (
        <span key={item.key} className="known-card-pill">
          {item.label}
        </span>
      ))}
    </div>
  );
}

export function ReferenceCard({ label, items, emptyLabel = "None" }: ReferenceCardProps) {
  return (
    <div className="observed-reference-card">
      <span className="summary-label">{label}</span>
      <KnownCardPillList items={items} emptyLabel={emptyLabel} />
    </div>
  );
}
