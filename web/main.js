"use strict";

const go = new Go();
const DEFAULT_WIDTH = 960;
const DEFAULT_HEIGHT = 540;

let api;
let canvas;

async function init() {
  canvas = document.getElementById("plotCanvas");

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
      console.warn("instantiateStreaming failed, fell back to ArrayBuffer", streamError);
    }

    go.run(result.instance);
    api = await waitForAPI();

    const demos = api.listDemos();
    populateSelector(demos, api.defaultDemoID());

    document
      .getElementById("renderBtn")
      .addEventListener("click", () => mountSelectedDemo());
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
  const selector = document.getElementById("demoSelector");
  const demoID = selector.value || api.defaultDemoID();

  updateStatus(`Rendering ${demoID}…`);

  const result = api.mountDemo("plotCanvas", demoID, DEFAULT_WIDTH, DEFAULT_HEIGHT);
  if (result.error) {
    updateStatus(result.error);
    return;
  }

  document.getElementById("demoTitle").textContent = result.title;
  document.getElementById("demoDescription").textContent = result.description;
  updateStatus(`Rendered ${result.id}`);
}

function updateStatus(message) {
  document.getElementById("statusMsg").textContent = message;
}

init();
