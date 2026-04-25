export default function SavedQueriesPage() {
  return (
    <div style={{ padding: "20px 24px", maxWidth: 860, margin: "0 auto" }}>
      <div style={{ marginBottom: 20 }}>
        <h1
          style={{
            margin: 0,
            fontSize: 18,
            fontWeight: 600,
            letterSpacing: "-0.015em",
            fontFamily: "var(--font-ui)",
          }}
        >
          Saved queries
        </h1>
        <div style={{ marginTop: 2, fontSize: 12, color: "var(--fg-2)" }}>
          Your saved Query Builder queries.
        </div>
      </div>

      {/* Empty state */}
      <div
        style={{
          background: "var(--bg)",
          border: "1px solid var(--hair)",
          borderRadius: "var(--radius)",
          padding: "48px 24px",
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          gap: 8,
        }}
      >
        <div
          style={{
            fontSize: 13,
            color: "var(--fg-2)",
            fontWeight: 500,
          }}
        >
          No saved queries yet.
        </div>
        <div style={{ fontSize: 12, color: "var(--fg-3)" }}>
          Build a query in the Query builder and save it here.
        </div>
      </div>
    </div>
  );
}
