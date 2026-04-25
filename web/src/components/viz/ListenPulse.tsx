// Monthly listening bar chart using deterministic sample data

const MONTHS = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];

// Deterministic play counts that look like realistic listening data
const SAMPLE_VALUES = [
  820, 710, 940, 1050, 1180, 1320, 980, 860, 1100, 1240, 1050, 1380,
];

interface ListenPulseProps {
  width?: number;
  height?: number;
}

export default function ListenPulse({ width = 340, height = 100 }: ListenPulseProps) {
  const paddingBottom = 18;
  const paddingTop = 8;
  const chartH = height - paddingBottom - paddingTop;
  const barW = Math.floor((width - 4) / MONTHS.length) - 3;
  const max = Math.max(...SAMPLE_VALUES);

  return (
    <svg
      width={width}
      height={height}
      viewBox={`0 0 ${String(width)} ${String(height)}`}
      style={{ display: "block" }}
    >
      {MONTHS.map((month, i) => {
        const barH = Math.round((SAMPLE_VALUES[i] / max) * chartH);
        const x = i * (barW + 3) + 2;
        const y = paddingTop + chartH - barH;
        return (
          <g key={month}>
            <rect
              x={x}
              y={y}
              width={barW}
              height={barH}
              rx={2}
              fill="color-mix(in oklch, var(--acc) 70%, var(--bg-2))"
            />
            <text
              x={x + barW / 2}
              y={height - 4}
              fontSize={7}
              fill="var(--fg-3)"
              textAnchor="middle"
              fontFamily="var(--font-mono)"
            >
              {month.slice(0, 1)}
            </text>
          </g>
        );
      })}
    </svg>
  );
}
