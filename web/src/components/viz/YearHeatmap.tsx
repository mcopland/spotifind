// 365-day listening heatmap using deterministic sample data

const CELL = 10;
const GAP = 2;
const WEEKS = 53;
const DAYS = 7;

// Deterministic sample values 0-4 for each day in the grid
function sampleValue(week: number, day: number): number {
  const n = week * 7 + day;
  // Blend a weekly pattern (weekdays busier) with a slow wave
  const weekday = day < 5 ? 1 : 0.3;
  const wave = 0.5 + 0.5 * Math.sin(n / 12);
  const raw = weekday * wave + 0.2 * Math.sin(n / 3);
  if (raw < 0.2) return 0;
  if (raw < 0.45) return 1;
  if (raw < 0.65) return 2;
  if (raw < 0.82) return 3;
  return 4;
}

const LEVELS = [
  "var(--bg-2)",
  "color-mix(in oklch, var(--acc) 20%, var(--bg-2))",
  "color-mix(in oklch, var(--acc) 40%, var(--bg-2))",
  "color-mix(in oklch, var(--acc) 65%, var(--bg-2))",
  "var(--acc)",
];

const MONTH_LABELS = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
const DAY_LABELS = ["", "Mon", "", "Wed", "", "Fri", ""];

export default function YearHeatmap() {
  const width = WEEKS * (CELL + GAP) - GAP;
  const height = DAYS * (CELL + GAP) - GAP;

  return (
    <div style={{ overflowX: "auto" }}>
      <svg
        width={width + 28}
        height={height + 20}
        viewBox={`0 0 ${String(width + 28)} ${String(height + 20)}`}
        style={{ display: "block" }}
      >
        {/* Day labels */}
        {DAY_LABELS.map((label, i) => (
          <text
            key={i}
            x={24}
            y={20 + i * (CELL + GAP) + CELL / 2 + 3}
            fontSize={8}
            fill="var(--fg-3)"
            textAnchor="end"
            fontFamily="var(--font-mono)"
          >
            {label}
          </text>
        ))}

        {/* Month labels - approximately every 4-5 weeks */}
        {MONTH_LABELS.map((label, mi) => (
          <text
            key={mi}
            x={28 + Math.round(mi * (WEEKS / 12)) * (CELL + GAP)}
            y={10}
            fontSize={8}
            fill="var(--fg-3)"
            fontFamily="var(--font-mono)"
          >
            {label}
          </text>
        ))}

        {/* Grid */}
        {Array.from({ length: WEEKS }, (_, w) =>
          Array.from({ length: DAYS }, (_, d) => {
            const val = sampleValue(w, d);
            return (
              <rect
                key={`${String(w)}-${String(d)}`}
                x={28 + w * (CELL + GAP)}
                y={20 + d * (CELL + GAP)}
                width={CELL}
                height={CELL}
                rx={2}
                fill={LEVELS[val]}
              />
            );
          })
        )}
      </svg>
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: 6,
          marginTop: 4,
          paddingLeft: 28,
          fontSize: 10,
          color: "var(--fg-3)",
          fontFamily: "var(--font-mono)",
        }}
      >
        <span>Less</span>
        {LEVELS.map((color, i) => (
          <div
            key={i}
            style={{ width: CELL, height: CELL, borderRadius: 2, background: color }}
          />
        ))}
        <span>More</span>
      </div>
    </div>
  );
}
