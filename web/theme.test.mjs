import test from "node:test";
import assert from "node:assert/strict";

import {
  nextThemeMode,
  normalizeThemeMode,
  resolveThemeMode,
  themeIcon,
  themeLabel,
} from "./theme.mjs";

test("normalizeThemeMode accepts only supported theme modes", () => {
  assert.equal(normalizeThemeMode("auto"), "auto");
  assert.equal(normalizeThemeMode("light"), "light");
  assert.equal(normalizeThemeMode("dark"), "dark");
  assert.equal(normalizeThemeMode("sepia"), "auto");
  assert.equal(normalizeThemeMode(null), "auto");
});

test("resolveThemeMode follows the system only in auto mode", () => {
  assert.equal(resolveThemeMode("auto", "dark"), "dark");
  assert.equal(resolveThemeMode("auto", "light"), "light");
  assert.equal(resolveThemeMode("auto", "unknown"), "light");
  assert.equal(resolveThemeMode("light", "dark"), "light");
  assert.equal(resolveThemeMode("dark", "light"), "dark");
});

test("nextThemeMode cycles through auto, light, and dark", () => {
  assert.equal(nextThemeMode("auto"), "light");
  assert.equal(nextThemeMode("light"), "dark");
  assert.equal(nextThemeMode("dark"), "auto");
  assert.equal(nextThemeMode("unknown"), "light");
});

test("theme toggle helpers return display content for supported modes", () => {
  assert.equal(themeLabel("auto"), "Auto");
  assert.equal(themeLabel("light"), "Light");
  assert.equal(themeLabel("dark"), "Dark");
  assert.match(themeIcon("auto"), /rect/);
  assert.match(themeIcon("light"), /circle/);
  assert.match(themeIcon("dark"), /path/);
});
