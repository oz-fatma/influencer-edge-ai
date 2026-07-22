import assert from "node:assert/strict";
import type { MLCEngine } from "@mlc-ai/web-llm";
import {
  __resetWebLLMEngineForTests,
  __setCreateMLCEngineForTests,
  initWebLLMEngine,
  isWebLLMLoading,
  isWebLLMReady,
} from "./webllm";

const mockEngine = { id: "mock-engine" } as unknown as MLCEngine;

async function runTests() {
  __resetWebLLMEngineForTests();

  const originalWindow = globalThis.window;
  Object.defineProperty(globalThis, "window", {
    configurable: true,
    value: globalThis,
  });

  const originalNavigator = globalThis.navigator;
  Object.defineProperty(globalThis, "navigator", {
    configurable: true,
    value: {
      ...originalNavigator,
      gpu: {
        requestAdapter: async () => ({}),
      },
    },
  });

  let createCalls = 0;
  let releaseLoad!: () => void;
  const loadGate = new Promise<void>((resolve) => {
    releaseLoad = resolve;
  });

  __setCreateMLCEngineForTests(async (_modelId, opts) => {
    createCalls += 1;
    opts?.initProgressCallback?.({ progress: 0.5, text: "Loading weights" });
    await loadGate;
    opts?.initProgressCallback?.({ progress: 1, text: "Ready to go" });
    return mockEngine;
  });

  const progressReports: string[] = [];
  const first = initWebLLMEngine((report) => {
    progressReports.push(report.text);
  });
  const second = initWebLLMEngine((report) => {
    progressReports.push(`wait:${report.text}`);
  });

  while (!isWebLLMLoading()) {
    await new Promise((resolve) => setTimeout(resolve, 5));
  }

  releaseLoad();
  const [engineA, engineB] = await Promise.all([first, second]);

  assert.equal(engineA, engineB);
  assert.equal(createCalls, 1, "CreateMLCEngine should run only once");
  assert.equal(isWebLLMReady(), true);
  assert.equal(isWebLLMLoading(), false);

  const third = await initWebLLMEngine((report) => {
    progressReports.push(`ready:${report.text}`);
  });

  assert.equal(third, mockEngine);
  assert.equal(createCalls, 1, "initialized engine should be reused");
  assert.ok(progressReports.some((text) => text.startsWith("wait:")));
  assert.ok(progressReports.includes("ready:Model ready"));

  __resetWebLLMEngineForTests();
  Object.defineProperty(globalThis, "navigator", {
    configurable: true,
    value: originalNavigator,
  });
  Object.defineProperty(globalThis, "window", {
    configurable: true,
    value: originalWindow,
  });

  console.log("webllm engine singleton tests passed");
}

runTests().catch((error) => {
  console.error(error);
  process.exit(1);
});
