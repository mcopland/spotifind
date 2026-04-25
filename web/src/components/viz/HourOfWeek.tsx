// 7x24 hour-of-week listening grid using deterministic sample data

const CELL_W = 18;
const CELL_H = 10;
const GAP = 2;
const DAYS = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];
const HOURS = Array.from({ length: 24 }, (_, i) => i);

// Deterministic activity level 0-1 per (day, hour)
function activity(day: number, hour: number): number {
  // Peaks in evenings, lower on weekends mornings, near-zero late nights
  const evening = Math.max(0, 1 - Math.abs(hour - 20) / 6);
  const morning = Math.max(0, 1 - Math.abs(hour - 8) / 4) * 0.4;
  const weekend = day >= 5 ? 0.7 : 1;
  const night = hour < 6 ? 0.05 : 1;
  return Math.min(1, (evening + morning) * weekend * night);
}

export default function HourOfWeek() {
  const labelH = 14;
  const labelW = 30;
  const width = labelW + HOURS.length * (CELL_W + GAP) - GAP;
  const height = labelH + DAYS.length * (CELL_H + GAP) - GAP;

  return (
    <div style={{ overflowX: "auto" }}>
      <svg
        width={width}
        height={height}
        viewBox={`0 0 ${String(width)} ${String(height)}`}
        style={{ display: "block" }}
      >
        {/* Hour labels */}
        {[0, 6, 12, 18, 23].map((h) => (
          <text
            key={h}
            x={labelW + h * (CELL_W + GAP)}
            y={10}
            fontSize={8}
            fill="var(--fg-3)"
            fontFamily="var(--font-mono)"
          >
            {h === 0 ? "12am" : h === 12 ? "12pm" : h < 12 ? `${String(h)}am` : `${String(h - 12)}pm`}
          </text>
        ))}

        {/* Day labels */}
        {DAYS.map((day, di) => (
          <text
            key={day}
            x={labelW - 4}
            y={labelH + di * (CELL_H + GAP) + CELL_H / 2 + 3}
            fontSize={8}
            fill="var(--fg-3)"
            textAnchor="end"
            fontFamily="var(--font-mono)"
          >
            {day}
          </text>
        ))}

        {/* Grid */}
        {DAYS.map((_, di) =>
          HOURS.map((h) => {
            const val = activity(di, h);
            return (
              <rect
                key={`${String(di)}-${String(h)}`}
                x={labelW + h * (CELL_W + GAP)}
                y={labelH + di * (CELL_H + GAP)}
                width={CELL_W}
                height={CELL_H}
                rx={2}
                fill={`color-mix(in oklch, var(--acc) ${String(Math.round(val * 70))}%, var(--bg-2))`}
              />
            );
          })
        )}
      </svg>
    </div>
  );
}
