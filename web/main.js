"use strict";

import {
  buildAPICompatibilityMessage,
  buildRuntimeExitMessage,
  missingAPIMethods,
} from "./runtime.mjs";

const go = new Go();
const DEFAULT_WIDTH = 960;
const DEFAULT_HEIGHT = 540;

let api;
let runtimeExited = false;
let runtimeExitMessage = buildRuntimeExitMessage();
let currentDemoID = "";

async function init() {
  const canvas = document.getElementById("plotCanvas");
  if (!canvas) {
    updateStatus("Failed to load WASM: canvas element was not found");
    return;
  }

  try {
    let result;
    try {
      result = await WebAssembly.instantiateStreaming(
        fetch("main.wasm"),
        go.importObject,
      );
    } catch (streamError) {
      const response = await fetch("main.wasm");
      const bytes = await response.arrayBuffer();
      result = await WebAssembly.instantiate(bytes, go.importObject);
      console.warn(
        "instantiateStreaming failed, fell back to ArrayBuffer",
        streamError,
      );
    }

    monitorRuntime(go.run(result.instance));
    api = await waitForAPI();
    ensureCompatibleAPI(api);

    const demos = api.listDemos();
    populateSelector(demos, api.defaultDemoID());

    document
      .getElementById("renderBtn")
      .addEventListener("click", () => mountSelectedDemo());
    document
      .getElementById("downloadBtn")
      .addEventListener("click", () => downloadCurrentPlot());
    document
      .getElementById("demoSelector")
      .addEventListener("change", () => mountSelectedDemo());

    mountSelectedDemo();
  } catch (error) {
    updateStatus(`Failed to load WASM: ${error.message}`);
    console.error(error);
  }
}

function waitForAPI() {
  return new Promise((resolve, reject) => {
    const startedAt = Date.now();

    function poll() {
      if (runtimeExited) {
        reject(new Error(runtimeExitMessage));
        return;
      }
      if (window.matplotlibGoWASM) {
        resolve(window.matplotlibGoWASM);
        return;
      }
      if (Date.now() - startedAt > 5000) {
        reject(new Error("matplotlibGoWASM API was not registered"));
        return;
      }
      window.setTimeout(poll, 20);
    }

    poll();
  });
}

function ensureCompatibleAPI(candidate) {
  const missingMethods = missingAPIMethods(candidate);
  if (missingMethods.length > 0) {
    throw new Error(buildAPICompatibilityMessage(missingMethods));
  }
}

function monitorRuntime(runPromise) {
  Promise.resolve(runPromise)
    .then(() => {
      runtimeExited = true;
      runtimeExitMessage = buildRuntimeExitMessage();
      updateStatus(runtimeExitMessage);
    })
    .catch((error) => {
      runtimeExited = true;
      runtimeExitMessage = buildRuntimeExitMessage(error);
      updateStatus(runtimeExitMessage);
      console.error(error);
    });
}

function populateSelector(demos, selectedID) {
  const selector = document.getElementById("demoSelector");
  selector.innerHTML = "";

  for (const demo of demos) {
    const option = document.createElement("option");
    option.value = demo.id;
    option.textContent = demo.title;
    if (demo.id === selectedID) {
      option.selected = true;
    }
    selector.appendChild(option);
  }
}

function mountSelectedDemo() {
  if (runtimeExited) {
    updateStatus(runtimeExitMessage);
    return;
  }

  const selector = document.getElementById("demoSelector");
  const demoID = selector.value || api.defaultDemoID();
  const size = plotSize();

  updateStatus(`Rendering ${demoID}…`);
  setDownloadEnabled(false);

  const startedAt = performance.now();
  const result = api.mountDemo("plotCanvas", demoID, size.width, size.height);
  if (result.error) {
    updateStatus(result.error);
    return;
  }
  const elapsedMs = performance.now() - startedAt;

  currentDemoID = result.id;
  document.getElementById("demoTitle").textContent = result.title;
  document.getElementById("demoDescription").textContent = result.description;
  setDownloadEnabled(true);
  updateStatus(
    `Rendered ${result.id} in ${elapsedMs.toFixed(1)} ms at ${result.width}×${result.height}`,
  );
}

function plotSize() {
  const canvas = document.getElementById("plotCanvas");
  const width = Math.max(1, Math.round(canvas.clientWidth || DEFAULT_WIDTH));
  const height = Math.max(1, Math.round(canvas.clientHeight || DEFAULT_HEIGHT));
  return { width, height };
}

function downloadCurrentPlot() {
  const canvas = document.getElementById("plotCanvas");
  if (!canvas || !currentDemoID) {
    return;
  }

  canvas.toBlob((blob) => {
    if (!blob) {
      updateStatus("Failed to prepare PNG download");
      return;
    }

    const link = document.createElement("a");
    const url = URL.createObjectURL(blob);
    link.href = url;
    link.download = `matplotlib-go-${currentDemoID}.png`;
    document.body.appendChild(link);
    link.click();
    link.remove();
    URL.revokeObjectURL(url);
  }, "image/png");
}

function updateStatus(message) {
  document.getElementById("statusMsg").textContent = message;
}

function setDownloadEnabled(enabled) {
  document.getElementById("downloadBtn").disabled = !enabled;
}

init();
