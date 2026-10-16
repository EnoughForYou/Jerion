/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  darkMode: ["class", '[data-theme="dark"]'],
  theme: {
    extend: {
      colors: {
        base: "var(--bg-base)",
        surface: "var(--bg-surface)",
        raised: "var(--bg-raised)",
        "text-primary": "var(--text-primary)",
        separator: {
          DEFAULT: "var(--separator)",
          subtle: "var(--separator-subtle)",
          muted: "var(--separator-muted)",
        },
        sev: {
          info: "var(--sev-info)",
          low: "var(--sev-low)",
          medium: "var(--sev-medium)",
          high: "var(--sev-high)",
          critical: "var(--sev-critical)",
        },
        ok: "var(--ok)",
      },
      fontFamily: {
        sans: ['"Inter Variable"', "Inter", "-apple-system", "BlinkMacSystemFont", "sans-serif"],
        mono: ['"JetBrains Mono"', "ui-monospace", "SFMono-Regular", "monospace"],
      },
      fontSize: {
        xs: ["11px", "14px"],
        sm: ["12px", "16px"],
        base: ["13px", "18px"], // PRD §5.3 desktop body font
        md: ["15px", "20px"],
        lg: ["17px", "22px"],
        xl: ["22px", "28px"],
        "2xl": ["28px", "34px"],
      },
      borderRadius: {
        sm: "var(--radius-sm, 6px)",
        md: "var(--radius-md, 10px)",
        lg: "var(--radius-lg, 14px)",
      },
    },
  },
  plugins: [],
};
