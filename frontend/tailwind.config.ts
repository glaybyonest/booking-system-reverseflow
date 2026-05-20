import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      fontFamily: {
        sans: ["Inter Variable", "Inter", "ui-sans-serif", "system-ui", "sans-serif"]
      },
      colors: {
        // PDF UI Kit exact colors
        ink: "#0b0d10",
        "ink-2": "#2a2d32",
        mute: "#6b7280",
        "mute-2": "#9ca3af",
        border: "#e4e6ea",
        bg: "#f8f9fa",
        accent: {
          DEFAULT: "#ff5a1f",
          soft: "#fff0eb",
          fg: "#c7380d"
        },
        ok: {
          DEFAULT: "#0e7a4e",
          soft: "#d1fae5",
          fg: "#065f3a"
        },
        warn: {
          DEFAULT: "#b45309",
          soft: "#fef3c7",
          fg: "#92400e"
        },
        err: {
          DEFAULT: "#c7382d",
          soft: "#fee2e2",
          fg: "#991b1b"
        },
        info: {
          DEFAULT: "#2a6fdb",
          soft: "#dbeafe",
          fg: "#1d4ed8"
        }
      },
      boxShadow: {
        soft: "0 24px 70px rgba(11, 13, 16, 0.08)",
        card: "0 1px 3px rgba(0,0,0,0.06), 0 0 0 1px rgba(0,0,0,0.04)",
        "card-hover": "0 8px 24px rgba(11, 13, 16, 0.12)"
      },
      borderRadius: {
        "4xl": "2rem",
        "5xl": "2.5rem"
      }
    }
  },
  plugins: []
};

export default config;
