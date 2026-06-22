import { ADDON_NAME } from "./config"
import { handler } from "./handler"
import { register } from "./services/registration"
import { log } from "./logger"
import { Payload } from "./models/payload"
import { delay } from "./utils"
import browser from "webextension-polyfill"

// `currentPort` is the live native-messaging port (undefined when disconnected) and
// doubles as the re-entrancy guard. `connected` is true only once registration has
// succeeded on a still-live port — it's what the about-popup status indicator reads.
let connected = false
let currentPort: browser.Runtime.Port | undefined

browser.runtime.onStartup.addListener(() => {
  log(`[${ADDON_NAME}] onStartup event fired`)
  connectAndListen()
})

browser.runtime.onInstalled.addListener(() => {
  log(`[${ADDON_NAME}] onInstalled event fired`)
  connectAndListen()
})

connectAndListen()

// Answer the about-popup's connection-status query with the current native-port state.
browser.runtime.onMessage.addListener((message: { type?: string }) => {
  if (message?.type === "status") {
    return Promise.resolve({ connected })
  }
  return undefined
})

function connectAndListen() {
  if (currentPort) return

  log(`Starting ${ADDON_NAME} add-on`)
  const port = browser.runtime.connectNative(ADDON_NAME)
  currentPort = port

  // Attach handlers IMMEDIATELY (before the async register) so a fast disconnect —
  // e.g. the native app missing or killed — is never missed. That's what keeps
  // `connected` accurate; postMessage on a dead port doesn't throw, so register()
  // resolving is NOT proof the port is alive.
  port.onMessage.addListener(async (payload: Payload) => {
    log(
      `[${ADDON_NAME}] Got message from native application: ${JSON.stringify(payload)}`
    )
    const { payload: command } = payload
    try {
      await handler(port, command)
    } catch (error) {
      if (error instanceof Error) {
        log(
          `[${ADDON_NAME}] Error while handling message`,
          error.message,
          error.stack
        )
      } else {
        log(`[${ADDON_NAME}] Error while handling message`, error)
      }
      throw error
    }
  })

  port.onDisconnect.addListener(async (disconnectedPort) => {
    if (currentPort !== port) return // stale event from a superseded port
    const errorMessage =
      disconnectedPort.error?.message || browser.runtime.lastError?.message
    log(
      `[${ADDON_NAME}] Disconnected with native application`,
      errorMessage ?? ""
    )
    connected = false
    currentPort = undefined

    const delayMs = 1000
    await delay(delayMs)
    log(`[${ADDON_NAME}] Reconnecting to native application...`)
    connectAndListen()
  })

  register(port)
    .then((registration) => {
      // Only mark connected if this port is still the live one — it may have
      // already disconnected while register() was running.
      if (currentPort === port) {
        connected = true
        log(
          `[${ADDON_NAME}] sent registration : ${JSON.stringify(registration)}`
        )
      }
    })
    .catch((error) => {
      log(`[${ADDON_NAME}] registration failed`, error)
    })
}
