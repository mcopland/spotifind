import { Plus, Trash2 } from "lucide-react";
import { useReducer } from "react";

type EntityType = "tracks" | "artists" | "playlists";
type Operator = "=" | "!=" | ">" | "<" | ">=" | "<=" | "contains" | "not contains";
type Logic = "AND" | "OR";

interface FilterCondition {
  id: number;
  field: string;
  operator: Operator;
  value: string;
  logic: Logic;
}

interface QueryState {
  entity: EntityType;
  conditions: FilterCondition[];
  sortField: string;
  sortDir: "asc" | "desc";
  limit: number;
  nextId: number;
}

type Action =
  | { type: "SET_ENTITY"; entity: EntityType }
  | { type: "ADD_CONDITION" }
  | { type: "UPDATE_CONDITION"; id: number; patch: Partial<FilterCondition> }
  | { type: "REMOVE_CONDITION"; id: number }
  | { type: "SET_SORT"; field: string; dir: "asc" | "desc" }
  | { type: "SET_LIMIT"; limit: number };

const ENTITY_FIELDS: Record<EntityType, string[]> = {
  tracks: ["name", "artist", "album", "genre", "popularity", "duration_ms", "saved_at", "explicit"],
  artists: ["name", "genre", "popularity", "followers"],
  playlists: ["name", "track_count", "is_public", "collaborative"],
};

const OPERATORS: Operator[] = ["=", "!=", ">", "<", ">=", "<=", "contains", "not contains"];

function reducer(state: QueryState, action: Action): QueryState {
  switch (action.type) {
    case "SET_ENTITY":
      return { ...state, entity: action.entity, conditions: [] };
    case "ADD_CONDITION":
      return {
        ...state,
        conditions: [
          ...state.conditions,
          {
            id: state.nextId,
            field: ENTITY_FIELDS[state.entity][0],
            operator: "=",
            value: "",
            logic: "AND",
          },
        ],
        nextId: state.nextId + 1,
      };
    case "UPDATE_CONDITION":
      return {
        ...state,
        conditions: state.conditions.map((c) =>
          c.id === action.id ? { ...c, ...action.patch } : c
        ),
      };
    case "REMOVE_CONDITION":
      return { ...state, conditions: state.conditions.filter((c) => c.id !== action.id) };
    case "SET_SORT":
      return { ...state, sortField: action.field, sortDir: action.dir };
    case "SET_LIMIT":
      return { ...state, limit: action.limit };
    default:
      return state;
  }
}

const initial: QueryState = {
  entity: "tracks",
  conditions: [],
  sortField: "name",
  sortDir: "asc",
  limit: 50,
  nextId: 1,
};

const selectStyle: React.CSSProperties = {
  height: 28,
  padding: "0 8px",
  border: "1px solid var(--hair)",
  borderRadius: "var(--radius-sm)",
  background: "var(--bg)",
  color: "var(--fg)",
  fontSize: 12,
  outline: "none",
};

const inputStyle: React.CSSProperties = {
  height: 28,
  padding: "0 8px",
  border: "1px solid var(--hair)",
  borderRadius: "var(--radius-sm)",
  background: "var(--bg)",
  color: "var(--fg)",
  fontSize: 12,
  outline: "none",
  minWidth: 120,
};

export default function QueryBuilderPage() {
  const [state, dispatch] = useReducer(reducer, initial);
  const fields = ENTITY_FIELDS[state.entity];

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
          Query builder
        </h1>
        <div style={{ marginTop: 2, fontSize: 12, color: "var(--fg-2)" }}>
          Build custom filters to explore your library. Execution requires backend support.
        </div>
      </div>

      {/* FROM block */}
      <div
        style={{
          background: "var(--bg)",
          border: "1px solid var(--hair)",
          borderRadius: "var(--radius)",
          padding: "14px 16px",
          marginBottom: 12,
          display: "flex",
          alignItems: "center",
          gap: 12,
        }}
      >
        <span
          style={{
            fontSize: 10,
            fontWeight: 700,
            fontFamily: "var(--font-mono)",
            color: "var(--fg-3)",
            width: 40,
            flexShrink: 0,
          }}
        >
          FROM
        </span>
        <select
          value={state.entity}
          onChange={(e) => {
            dispatch({ type: "SET_ENTITY", entity: e.target.value as EntityType });
          }}
          style={{ ...selectStyle, fontWeight: 500 }}
        >
          <option value="tracks">tracks</option>
          <option value="artists">artists</option>
          <option value="playlists">playlists</option>
        </select>
      </div>

      {/* WHERE block */}
      <div
        style={{
          background: "var(--bg)",
          border: "1px solid var(--hair)",
          borderRadius: "var(--radius)",
          overflow: "hidden",
          marginBottom: 12,
        }}
      >
        <div
          style={{
            padding: "10px 16px",
            borderBottom: state.conditions.length > 0 ? "1px solid var(--hair)" : undefined,
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
          }}
        >
          <span
            style={{
              fontSize: 10,
              fontWeight: 700,
              fontFamily: "var(--font-mono)",
              color: "var(--fg-3)",
            }}
          >
            WHERE
          </span>
          <button
            onClick={() => { dispatch({ type: "ADD_CONDITION" }); }}
            style={{
              display: "inline-flex",
              alignItems: "center",
              gap: 5,
              fontSize: 11,
              color: "var(--acc-ink)",
              cursor: "pointer",
              padding: "2px 8px",
              border: "1px solid color-mix(in oklch, var(--acc) 30%, var(--hair))",
              borderRadius: "var(--radius-sm)",
              background: "var(--acc-soft)",
            }}
          >
            <Plus size={11} /> Add filter
          </button>
        </div>
        {state.conditions.length === 0 && (
          <div style={{ padding: "16px", fontSize: 12, color: "var(--fg-3)" }}>
            No filters — all {state.entity} will be returned.
          </div>
        )}
        {state.conditions.map((cond, idx) => (
          <div
            key={cond.id}
            style={{
              display: "flex",
              alignItems: "center",
              gap: 8,
              padding: "10px 16px",
              borderBottom: idx < state.conditions.length - 1 ? "1px solid var(--hair)" : undefined,
            }}
          >
            {/* AND/OR toggle */}
            <select
              value={cond.logic}
              onChange={(e) => {
                dispatch({ type: "UPDATE_CONDITION", id: cond.id, patch: { logic: e.target.value as Logic } });
              }}
              style={{
                ...selectStyle,
                width: 52,
                fontFamily: "var(--font-mono)",
                fontSize: 10,
                fontWeight: 600,
                color: cond.logic === "AND" ? "var(--info)" : "var(--warn)",
                display: idx === 0 ? "none" : undefined,
              }}
            >
              <option value="AND">AND</option>
              <option value="OR">OR</option>
            </select>
            {idx === 0 && <div style={{ width: 52, flexShrink: 0 }} />}

            {/* Field */}
            <select
              value={cond.field}
              onChange={(e) => {
                dispatch({ type: "UPDATE_CONDITION", id: cond.id, patch: { field: e.target.value } });
              }}
              style={selectStyle}
            >
              {fields.map((f) => (
                <option key={f} value={f}>{f}</option>
              ))}
            </select>

            {/* Operator */}
            <select
              value={cond.operator}
              onChange={(e) => {
                dispatch({ type: "UPDATE_CONDITION", id: cond.id, patch: { operator: e.target.value as Operator } });
              }}
              style={{ ...selectStyle, fontFamily: "var(--font-mono)" }}
            >
              {OPERATORS.map((op) => (
                <option key={op} value={op}>{op}</option>
              ))}
            </select>

            {/* Value */}
            <input
              value={cond.value}
              onChange={(e) => {
                dispatch({ type: "UPDATE_CONDITION", id: cond.id, patch: { value: e.target.value } });
              }}
              placeholder="value"
              style={inputStyle}
            />

            {/* Remove */}
            <button
              onClick={() => { dispatch({ type: "REMOVE_CONDITION", id: cond.id }); }}
              style={{
                display: "flex",
                alignItems: "center",
                color: "var(--fg-3)",
                cursor: "pointer",
                padding: 4,
              }}
              onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.color = "var(--err)"; }}
              onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.color = "var(--fg-3)"; }}
            >
              <Trash2 size={13} />
            </button>
          </div>
        ))}
      </div>

      {/* SORT & LIMIT */}
      <div
        style={{
          background: "var(--bg)",
          border: "1px solid var(--hair)",
          borderRadius: "var(--radius)",
          padding: "14px 16px",
          marginBottom: 20,
          display: "flex",
          alignItems: "center",
          gap: 16,
        }}
      >
        <span
          style={{
            fontSize: 10,
            fontWeight: 700,
            fontFamily: "var(--font-mono)",
            color: "var(--fg-3)",
            width: 40,
            flexShrink: 0,
          }}
        >
          ORDER
        </span>
        <select
          value={state.sortField}
          onChange={(e) => { dispatch({ type: "SET_SORT", field: e.target.value, dir: state.sortDir }); }}
          style={selectStyle}
        >
          {fields.map((f) => (
            <option key={f} value={f}>{f}</option>
          ))}
        </select>
        <select
          value={state.sortDir}
          onChange={(e) => { dispatch({ type: "SET_SORT", field: state.sortField, dir: e.target.value as "asc" | "desc" }); }}
          style={{ ...selectStyle, fontFamily: "var(--font-mono)" }}
        >
          <option value="asc">ASC</option>
          <option value="desc">DESC</option>
        </select>
        <span style={{ marginLeft: 16, fontSize: 10, fontWeight: 700, fontFamily: "var(--font-mono)", color: "var(--fg-3)" }}>
          LIMIT
        </span>
        <select
          value={state.limit}
          onChange={(e) => { dispatch({ type: "SET_LIMIT", limit: Number(e.target.value) }); }}
          style={{ ...selectStyle, fontFamily: "var(--font-mono)" }}
        >
          {[25, 50, 100, 250, 500].map((n) => (
            <option key={n} value={n}>{n}</option>
          ))}
        </select>
      </div>

      {/* Run button */}
      <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
        <button
          style={{
            padding: "7px 18px",
            background: "var(--acc)",
            color: "black",
            border: "none",
            borderRadius: "var(--radius-sm)",
            fontSize: 13,
            fontWeight: 600,
            cursor: "pointer",
          }}
          onClick={() => { /* no-op: backend not yet supported */ }}
        >
          Run query
        </button>
        <span
          style={{
            fontSize: 11,
            color: "var(--fg-3)",
            fontFamily: "var(--font-mono)",
          }}
        >
          backend execution coming soon
        </span>
      </div>
    </div>
  );
}
