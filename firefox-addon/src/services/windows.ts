import { Port } from "../models/port"
import { Command } from "../models/command"
import { log } from "../logger"
import { Response } from "../models/response"
import { delay, handleError } from "../utils"
import browser from "webextension-polyfill"

export async function getWindows(port: Port, {}: Command) {
  try {
    log(`Starting getWindows`)
    // TODO : filtering certain window types ?
    // populate:true gives us each window's tabs, so we can report a tab count and
    // the active tab's title/url — that turns an opaque window id into something a
    // human can recognize (pairs with the about-popup showing "Window <id>").
    const [windows, lastFocused] = await Promise.all([
      browser.windows.getAll({ populate: true }),
      browser.windows.getLastFocused(),
    ])
    log("Sending back ", windows.length, " windows")

    const returnedWindows = windows.map((window) => {
      const tabs = window.tabs ?? []
      const activeTab = tabs.find((t) => t.active)
      return {
        id: window.id,
        // isLastFocused kept for backward-compat: `tabs get -w` reads this field.
        isLastFocused: window.id === lastFocused.id,
        focused: window.focused ?? false,
        type: window.type ?? "",
        state: window.state ?? "",
        incognito: window.incognito ?? false,
        tabCount: tabs.length,
        activeTabTitle: activeTab?.title ?? "",
        activeTabUrl: activeTab?.url ?? "",
        top: window.top ?? -1,
        left: window.left ?? -1,
        width: window.width ?? -1,
        height: window.height ?? -1,
      }
    })

    port.postMessage(Response.data(returnedWindows))
    // pause 5ms, or this end message may be received before the message above
    await delay(5)
    return port.postMessage(Response.end())
  } catch (e) {
    return handleError(e, port)
  }
}

// args: "<windowId>"
export async function focusWindow(port: Port, { args }: Command) {
  try {
    if (!args) {
      log("missing args in focus-window")
      return port.postMessage(Response.end())
    }
    const windowId = Number.parseInt(args)
    await browser.windows.update(windowId, { focused: true })
    // Return the active tab title so the CLI can raise the specific Firefox window
    // to the foreground on macOS (via AppleScript), same as `tabs activate`.
    const win = await browser.windows.get(windowId, { populate: true })
    const activeTab = (win.tabs ?? []).find((t) => t.active)
    log("focused window ", windowId)
    port.postMessage(
      Response.data({
        success: true,
        windowId,
        title: activeTab?.title ?? "",
      })
    )
    // pause 50ms, or this end message may be received before the message above
    await delay(50)
    return port.postMessage(Response.end())
  } catch (e) {
    return handleError(e, port)
  }
}

// args: "<incognito>:<state>|<url1>\n<url2>..."  (state or geometry-less; urls optional)
export async function newWindow(port: Port, { args }: Command) {
  try {
    let incognito = false
    let state: string | undefined = undefined
    let urls: string[] = []

    if (args) {
      // urls may contain ":" and "," so split flags from urls on the first "|",
      // and separate individual urls by newline (which urls cannot contain).
      const sep = args.indexOf("|")
      const flagsPart = sep >= 0 ? args.slice(0, sep) : args
      const urlsPart = sep >= 0 ? args.slice(sep + 1) : ""
      const [inc, st] = flagsPart.split(":")
      incognito = inc === "true"
      state = st && st !== "none" ? st : undefined
      urls = urlsPart ? urlsPart.split("\n").filter((u) => u.length > 0) : []
    }

    const createData: any = {}
    if (urls.length > 0) createData.url = urls
    if (incognito) createData.incognito = true
    if (state) createData.state = state

    const win = await browser.windows.create(createData)
    log("created window ", JSON.stringify(win?.id))
    port.postMessage(Response.data({ id: win?.id }))
    // pause 10ms, or this end message may be received before the message above
    await delay(10)
    return port.postMessage(Response.end())
  } catch (e) {
    return handleError(e, port)
  }
}

// args: "<windowId>"
export async function closeWindow(port: Port, { args }: Command) {
  try {
    if (!args) {
      log("missing args in close-window")
      return port.postMessage(Response.end())
    }
    const windowId = Number.parseInt(args)
    await browser.windows.remove(windowId)
    log("closed window ", windowId)
    return port.postMessage(Response.end())
  } catch (e) {
    return handleError(e, port)
  }
}

// args: "<id>:<state>:<top>:<left>:<width>:<height>"  ("none" = leave unchanged)
export async function updateWindow(port: Port, { args }: Command) {
  try {
    if (!args) {
      log("missing args in update-window")
      return port.postMessage(Response.end())
    }
    const userArgs = args.split(":")
    const windowId = Number.parseInt(userArgs[0])
    const [, state, top, left, width, height] = userArgs

    const updateInfo: any = {}
    if (state && state !== "none") updateInfo.state = state
    if (top && top !== "none") updateInfo.top = Number.parseInt(top)
    if (left && left !== "none") updateInfo.left = Number.parseInt(left)
    if (width && width !== "none") updateInfo.width = Number.parseInt(width)
    if (height && height !== "none") updateInfo.height = Number.parseInt(height)

    const win = await browser.windows.update(windowId, updateInfo)
    log("updated window ", JSON.stringify(win?.id))
    return port.postMessage(Response.end())
  } catch (e) {
    return handleError(e, port)
  }
}
