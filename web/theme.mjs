export const THEME_KEY = "matplotlib-go-theme";
export const THEME_MODES = Object.freeze(["auto", "light", "dark"]);

const THEME_LABELS = Object.freeze({
  auto: "Auto",
  light: "Light",
  dark: "Dark",
});

const THEME_ICONS = Object.freeze({
  auto: `
    <rect x="3" y="4" width="18" height="12" rx="2"></rect>
    <path d="M8 20h8"></path>
    <path d="M12 16v4"></path>
  `,
  light: `
    <circle cx="12" cy="12" r="5"></circle>
    <line x1="12" y1="1" x2="12" y2="3"></line>
    <line x1="12" y1="21" x2="12" y2="23"></line>
    <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line>
    <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line>
    <line x1="1" y1="12" x2="3" y2="12"></line>
    <line x1="21" y1="12" x2="23" y2="12"></line>
    <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line>
    <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line>
  `,
  dark: `
    <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"></path>
  `,
});

export function normalizeThemeMode(mode) {
  return THEME_MODES.includes(mode) ? mode : "auto";
}

export function resolveThemeMode(mode, systemTheme) {
  const normalizedMode = normalizeThemeMode(mode);
  if (normalizedMode !== "auto") {
    return normalizedMode;
  }
  return systemTheme === "dark" ? "dark" : "light";
}

export function nextThemeMode(mode) {
  const normalizedMode = normalizeThemeMode(mode);
  const currentIndex = THEME_MODES.indexOf(normalizedMode);
  return THEME_MODES[(currentIndex + 1) % THEME_MODES.length];
}

export function themeLabel(mode) {
  return THEME_LABELS[normalizeThemeMode(mode)];
}

export function themeIcon(mode) {
  return THEME_ICONS[normalizeThemeMode(mode)];
}
