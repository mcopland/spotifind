import { Component, type ReactNode } from "react";

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
}

export default class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false };

  static getDerivedStateFromError(): State {
    return { hasError: true };
  }

  render() {
    if (this.state.hasError) {
      return (
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            alignItems: "center",
            justifyContent: "center",
            height: 240,
            gap: 12,
          }}
        >
          <p style={{ color: "var(--err)", fontSize: 13 }}>Something went wrong.</p>
          <button
            onClick={() => { window.location.reload(); }}
            style={{
              padding: "4px 14px",
              background: "var(--acc)",
              color: "black",
              fontSize: 12,
              fontWeight: 500,
              borderRadius: "var(--radius-sm)",
              cursor: "pointer",
            }}
          >
            Reload
          </button>
        </div>
      );
    }
    return this.props.children;
  }
}
