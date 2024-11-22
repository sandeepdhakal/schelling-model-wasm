importScripts("wasm_exec.js");

const go = new Go();
const waInit = WebAssembly.instantiateStreaming(
  fetch("schelling.wasm"),
  go.importObject,
).then((result) => {
  go.run(result.instance);
});

onmessage = async (e) => {
  await waInit;
  res = simulate(...e.data);
  postMessage(res);
};
