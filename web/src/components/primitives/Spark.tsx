interface SparkProps {
  data: number[];
  width?: number;
  height?: number;
  stroke?: string;
  fill?: string;
}

export default function Spark({
  data,
  width = 120,
  height = 22,
  stroke = "var(--acc)",
  fill = "color-mix(in oklch, var(--acc) 15%, transparent)",
}: SparkProps) {
  if (data.length < 2) return null;
  const min = Math.min(...data);
  const max = Math.max(...data);
  const sx = (i: number) => (i * width) / (data.length - 1);
  const sy = (v: number) =>
    height - 2 - ((v - min) / (max - min || 1)) * (height - 4);

  const d = data
    .map((v, i) => `${i === 0 ? "M" : "L"} ${sx(i).toFixed(1)} ${sy(v).toFixed(1)}`)
    .join(" ");
  const area = `${d} L ${String(width)} ${String(height)} L 0 ${String(height)} Z`;

  return (
    <svg
      width={width}
      height={height}
      viewBox={`0 0 ${String(width)} ${String(height)}`}
      style={{ overflow: "visible" }}
    >
      <path d={area} fill={fill} />
      <path
        d={d}
        fill="none"
        stroke={stroke}
        strokeWidth="1.25"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}
