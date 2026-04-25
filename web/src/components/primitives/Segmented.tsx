interface Option<T extends string> {
  value: T;
  label: string;
}

interface SegmentedProps<T extends string> {
  options: Option<T>[];
  value: T;
  onChange: (v: T) => void;
}

export default function Segmented<T extends string>({
  options,
  value,
  onChange,
}: SegmentedProps<T>) {
  return (
    <div
      style={{
        display: "inline-flex",
        border: "1px solid var(--hair)",
        borderRadius: "var(--radius-sm)",
        overflow: "hidden",
        background: "var(--bg-1)",
      }}
    >
      {options.map((o) => (
        <button
          key={o.value}
          onClick={() => { onChange(o.value); }}
          style={{
            padding: "2px 8px",
            height: 24,
            fontSize: 11,
            fontFamily: "var(--font-ui)",
            borderRight: "1px solid var(--hair)",
            borderRadius: 0,
            background: value === o.value ? "var(--acc-soft)" : "transparent",
            color: value === o.value ? "var(--acc-ink)" : "var(--fg-1)",
            cursor: "pointer",
          }}
        >
          {o.label}
        </button>
      ))}
    </div>
  );
}
