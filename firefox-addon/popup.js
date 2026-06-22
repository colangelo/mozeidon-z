// About popup — shows the extension version + links.
// Firefox MV2: the `browser` API is available natively in popup pages (no polyfill needed).
const manifest = browser.runtime.getManifest()

document.getElementById("version").textContent = "v" + manifest.version
document.getElementById("repo").href =
  manifest.homepage_url || "https://github.com/colangelo/mozeidon-z"
document.getElementById("amo").href =
  "https://addons.mozilla.org/firefox/addon/mozeidon-z/"
document.getElementById("foot").textContent = manifest.name + " · MIT"

// Connection status — ask the background for the current native-app port state.
function setStatus(connected) {
  document.getElementById("dot").classList.add(connected ? "ok" : "bad")
  document.getElementById("status-text").textContent = connected
    ? "Connected"
    : "Not connected"
}

browser.runtime
  .sendMessage({ type: "status" })
  .then((res) => setStatus(!!(res && res.connected)))
  .catch(() => setStatus(false))
