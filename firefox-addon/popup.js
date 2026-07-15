// About popup — shows the extension version + links.
// Firefox MV2: the `browser` API is available natively in popup pages (no polyfill needed).
const manifest = browser.runtime.getManifest()

document.getElementById("version").textContent = "v" + manifest.version
document.getElementById("repo").href =
  manifest.homepage_url || "https://github.com/colangelo/mozeidon-z"
document.getElementById("amo").href =
  "https://addons.mozilla.org/firefox/addon/mozeidon-z/"
document.getElementById("foot").textContent = manifest.name + " · MIT"

// Current window id — matches the `windowId` half of Mozeidon's `windowId:tabId`
// tab-ID format, so you can tell which Firefox window a tab like `3289:596` is in.
browser.windows
  .getCurrent()
  .then((w) => {
    document.getElementById("window").textContent = "Window " + w.id
  })
  .catch(() => {})

// Connection status — ask the background for the current native-app port state.
function setStatus(connected) {
  const dot = document.getElementById("dot")
  // Reset both states first so a bad→ok (or ok→bad) transition flips cleanly
  // rather than stacking classes — matters now that we poll live below.
  dot.classList.remove("ok", "bad")
  dot.classList.add(connected ? "ok" : "bad")
  document.getElementById("status-text").textContent = connected
    ? "Connected"
    : "Not connected"
}

function refreshStatus() {
  browser.runtime
    .sendMessage({ type: "status" })
    .then((res) => setStatus(!!(res && res.connected)))
    .catch(() => setStatus(false))
}

// The background reconnects to the native app every ~1s, so a popup opened while
// disconnected would otherwise stay red until reopened. Poll so it goes green
// (or red) on its own while open.
refreshStatus()
const statusTimer = setInterval(refreshStatus, 1000)
window.addEventListener("unload", () => clearInterval(statusTimer))
