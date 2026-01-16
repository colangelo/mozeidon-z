# 🔱 Mozeidon Z

> **This is a fork of [egovelox/mozeidon](https://github.com/egovelox/mozeidon)** with additional features for macOS users.

## What's New in This Fork

- **`tabs activate`** - Activate a tab AND bring its Firefox window to the foreground, even across macOS Spaces
- **`tabs pick`** - Interactive TUI tab picker using [bubbletea](https://github.com/charmbracelet/bubbletea)
- **AppleScript integration** - Proper window activation on macOS that works with multiple windows/Spaces
- **`just` recipes** - Simplified setup and development workflow (`just setup-all`, `just package-firefox`, etc.)

## Firefox Extension

Install the **Mozeidon Z** extension from AMO:

**[https://addons.mozilla.org/en-US/firefox/addon/mozeidon-z/](https://addons.mozilla.org/en-US/firefox/addon/mozeidon-z/)**

---

## TLDR

- Handle your tabs, groups, bookmarks and history from outside of your web-browser.
- [🤓 Installation guide](#installation)
- [📖 CLI Reference Documentation](CLI_REFERENCE.md)
- [✨ Desktop applications](#desktop-applications-based-on-mozeidon-cli)

## Intro

Mozeidon is essentially a CLI written in [Go](https://go.dev/) to handle [Mozilla Firefox](https://www.mozilla.org/firefox/) tabs, history, and bookmarks.

Here you'll find:
- A guide to complete the [installation](#installation) of the mozeidon components (see [architecture](#architecture))
- [Advanced examples](#examples) of the CLI usage (including integration with `fzf` and `fzf-tmux`)
- [A Raycast extension](#raycast-extension) built around Mozeidon CLI (for macOS only)

All the code is available here as open-source. You can be sure that:
- Your browsing data (tabs, bookmarks, etc) will remain private and safe: mozeidon will never share anything outside of your system.
- At any time, stopping or removing the mozeidon firefox addon extension will stop or remove all related processes on your machine.

Using the `mozeidon` CLI (see [CLI reference](CLI_REFERENCE.md)), you can:
- List all currently opened tabs
- List recently-closed tabs
- List, delete current history
- List current bookmarks
- **Activate a tab and bring its window to foreground** (new in this fork)
- **Pick a tab interactively with TUI** (new in this fork)
- Switch to a currently opened tab
- Open a new tab (empty tab or with target url)
- Close a currently opened tab
- Pin/unpin a currently opened tab
- Group/ungroup a currently opened tab
- Create, delete, update a bookmark

| <img width="1512" height="910" alt="mozeidon-cli" src="https://github.com/user-attachments/assets/32b49616-5129-479c-aea6-9490395464c9" /> |
|:--:|
| *Example output showing tabs with `jq` formatting* |

<br/>

| <img width="1512" height="910" alt="mozeidon-cli" src="https://github.com/user-attachments/assets/a3757c8b-652a-4a59-b0b7-5f7dc06a79a6" /> |
|:--:|
| *Example output showing tabs with `--go-template` formatting* |

<br/>

| <img width="1512" alt="mozeidon-cli-2" src="https://github.com/egovelox/mozeidon/assets/56078155/9ba5c99b-0436-433c-9b73-427f2b3c897f"> |
|:--:|
| *Example output showing tabs within custom `fzf-tmux` script* |

<br/>

## Architecture


<img width="788" alt="mozeidon-architecture" src="https://github.com/egovelox/mozeidon/assets/56078155/15192276-e85f-4de0-956d-6eba0517303b">
<br/><br/>

Mozeidon is built on IPC and native-messaging protocols, using the following components:

- **[Mozeidon Z Firefox Extension](#mozeidon-z-firefox-extension)** - A TypeScript WebExtension running in Firefox that receives commands and sends back data (tabs, bookmarks, etc.) by leveraging browser APIs.

- **[Mozeidon native-app](#mozeidon-native-app)** - A Go program that acts as an IPC broker. It communicates with the browser extension via [native-messaging](https://developer.mozilla.org/docs/Mozilla/Add-ons/WebExtensions/Native_messaging) protocol.

- **[Mozeidon CLI](#mozeidon-cli)** - A Go CLI that communicates with the native-app via IPC protocol. This is what you use from the terminal.


## Installation

You need to install 3 components:

1. **[Mozeidon Z Firefox Extension](#mozeidon-z-firefox-extension)** - Install from AMO
2. **[Mozeidon native-app](#mozeidon-native-app)** - Install via Homebrew
3. **[Mozeidon CLI](#mozeidon-cli)** - Install via Homebrew or build from source

### Quick Start (macOS with Homebrew)

```bash
# Install native app and CLI
brew tap egovelox/homebrew-mozeidon
brew install egovelox/mozeidon/mozeidon-native-app
brew install egovelox/mozeidon/mozeidon

# Configure native messaging (or use: just setup-native-messaging)
mkdir -p ~/Library/Application\ Support/Mozilla/NativeMessagingHosts
cat > ~/Library/Application\ Support/Mozilla/NativeMessagingHosts/mozeidon.json << 'EOF'
{
  "name": "mozeidon",
  "description": "Mozeidon native messaging host",
  "path": "/opt/homebrew/bin/mozeidon-native-app",
  "type": "stdio",
  "allowed_extensions": ["mozeidon-addon@egovelox.com", "mozeidon-z@a-layer.io"]
}
EOF

# Restart Firefox, then test
mozeidon tabs get
```

## Mozeidon Z Firefox Extension

The **Mozeidon Z** addon for Mozilla Firefox (this fork) can be found here:

**[https://addons.mozilla.org/en-US/firefox/addon/mozeidon-z/](https://addons.mozilla.org/en-US/firefox/addon/mozeidon-z/)**

Latest version: `3.2.2`

> Note: The original [mozeidon extension](https://addons.mozilla.org/en-US/firefox/addon/mozeidon) will also work, but you'll miss the improvements in this fork.

## Mozeidon native-app

The [mozeidon native-app](https://github.com/egovelox/mozeidon-native-app) is an IPC broker that connects the CLI to the browser extension.

**Install via Homebrew (macOS/Linux):**
```bash
brew tap egovelox/homebrew-mozeidon
brew install egovelox/mozeidon/mozeidon-native-app
```

Or download from the [release page](https://github.com/egovelox/mozeidon-native-app/releases), or build from source.

### Configure Native Messaging (Required)

The native-app must be registered with Firefox.

**Option 1: Use just (recommended)**
```bash
just setup-native-messaging
```

**Option 2: Manual setup**

Create `~/Library/Application Support/Mozilla/NativeMessagingHosts/mozeidon.json`:

```json
{
  "name": "mozeidon",
  "description": "Mozeidon native messaging host",
  "path": "/opt/homebrew/bin/mozeidon-native-app",
  "type": "stdio",
  "allowed_extensions": [
    "mozeidon-addon@egovelox.com",
    "mozeidon-z@a-layer.io"
  ]
}
```

**Important:** Restart Firefox after creating this file.

For other OS, see the [Mozilla documentation](https://developer.mozilla.org/en-US/docs/Mozilla/Add-ons/WebExtensions/Native_manifests#manifest_location) for the correct `NativeMessagingHosts` location.

## Mozeidon CLI

The Mozeidon CLI is a lightweight CLI written in Go.

📖 **[Complete CLI Reference Documentation](CLI_REFERENCE.md)**

### New Commands in This Fork

```bash
# Activate a tab AND bring its window to foreground (works across macOS Spaces)
mozeidon tabs activate 3289:596

# Interactive TUI tab picker
mozeidon tabs pick
```

### Install via Homebrew

```bash
brew tap egovelox/homebrew-mozeidon
brew install egovelox/mozeidon/mozeidon
```

Or download from the [release page](https://github.com/egovelox/mozeidon/releases).

### Build from Source (This Fork)

```bash
git clone https://github.com/anthropics/mozeidon-z.git  # TODO: update with actual fork URL
cd mozeidon-z/cli && go build
```

## Examples 

📖 **[Complete CLI Reference Documentation](CLI_REFERENCE.md)**

### How to use the Mozeidon CLI with ``go-template`` syntax for customized output :

```bash
# get maximum 10 of latest bookmarks, title and url

mozeidon bookmarks -m 10 --go-template '{{range .Items}}{{.Title}} {{.Url}}{{"\n"}}{{end}}'
```

```bash
# get opened tabs, with 📌 icon if pinned

mozeidon tabs get --go-template '{{range .Items}}{{.WindowId}}:{{.Id}} {{.Url}} {{if .Pinned}}📌{{else}}🦊{{end}} {{"\\u001b[38;5;109m"}} {{.Domain}}{{"\\033[0m"}} {{.Title}}{{"\n"}}{{end}}'
```

### Customized tabs output with a pipe into ``fzf``

If you've installed [fzf](https://github.com/junegunn/fzf) you can use it as a kind of UI for mozeidon CLI.

The below `bash` command shows how `fzf` can be used to select a tab, and to open it in your browser.

```bash
mozeidon tabs get --go-template '{{range .Items}}{{.WindowId}}:{{.Id}} {{.Url}} {{if .Pinned}}📌{{else}}🦊{{end}} {{"\u001b[38;5;109m"}} {{.Domain}}{{"\033[0m"}} {{.Title}}{{"\n"}}{{end}}' \
| fzf --ansi --with-nth 3.. --bind=enter:accept-non-empty \
| cut -d ' ' -f1 \
| xargs -n1 -I % sh -c 'mozeidon tabs switch % && open -a firefox'
```

note : ``xargs -n1`` prevents to run any command if no tab was chosen with fzf ( say, for example, that you exited fzf with ctrl-c )

note : ``mozeidon tabs switch`` is used to switch to the tab you chose in fzf

### Same as previous, but tailored for tmux

As an example, let's bind our mozeidon script with the tmux shortcut ``Prefix-t``

```bash
# in $HOME/.tmux.conf
bind t run-shell -b "bash $HOME/.tmux/mozeidon_tabs.sh"
```

Now create the script ``$HOME/.tmux/mozeidon_tabs.sh`` :

```bash
#!/bin/bash
mozeidon tabs get --go-template \
'{{range .Items}}{{.WindowId}}:{{.Id}} {{.Url}} {{if .Pinned}}📌{{else}}🦊{{end}} {{"\u001b[38;5;109m"}} {{.Domain}}{{"\033[0m"}}  {{.Title}}{{"\n"}}{{end}}' \
| fzf-tmux -p 60% -- \
--no-bold --layout=reverse --margin 0% --no-separator --no-info --black --color bg+:black,hl:reverse,hl+:reverse,gutter:black --ansi --with-nth 3.. --bind=enter:accept-non-empty \
| cut -d ' ' -f1 \
| xargs -n1 -I % sh -c '$HOME/bin/mozeidon tabs switch % && open -a firefox'
```

### Another advanced fzf-tmux script

This more advanced script will allow to :
- open a new tab (empty or with search query)
- switch to a currently open tab
- close one or many tabs

```bash
#!/bin/bash
$HOME/bin/mozeidon tabs get --go-template \
  '{{range .Items}}{{.WindowId}}:{{.Id}} {{.Url}} {{if .Pinned}}📌{{else}}🦊{{end}} {{"\u001b[38;5;109m"}}  {{.Domain}}{{"\033[0m"}}  {{.Title}}{{"\n"}}{{end}}'\
  | fzf-tmux -p 60% -- \
  --border-label=TABS \
  --no-bold \
  --layout=reverse \
  --margin 0% \
  --no-separator \
  --no-info \
  --black \
  --color bg+:black,hl:reverse,hl+:reverse,gutter:black \
  --with-nth 3.. \
  --bind="enter:accept+execute($HOME/bin/mozeidon tabs switch {1} && open -a firefox)" \
  --multi \
  --marker=❌ \
  --bind="ctrl-p:accept-non-empty+execute($HOME/bin/mozeidon tabs close {+1})" \
  --bind="ctrl-o:print-query" \
  --header-first \
  --color=header:#5e6b6b \
  '--header=close tab(s) [C-p] 
open new tab [C-o]'\
  | grep -v "[🦊📌]" \
  | xargs -r -I {} sh -c '$HOME/bin/mozeidon tabs new "{}" && open -a firefox'
```

## Desktop applications based on mozeidon CLI

### Swell

[Swell](https://github.com/egovelox/swell) 🏄 🌊 is very fast and has plenty of features  
( custom shortcuts, chromium or mozilla compatibility, tabs, groups, bookmarks, history ).

<img width="1141" height="686" alt="swell_tabs_panel" src="https://github.com/user-attachments/assets/ccfc7aac-dc02-4dba-97dd-25ee29da3957" />


### Raycast extension

For MacOS and Firefox users only : see [the Mozeidon Raycast extension](https://www.raycast.com/egovelox/mozeidon).

This Raycast extension will not work with Chrome browser. Better see [Swell](https://github.com/egovelox/swell).

Note that you'll first need to complete the installation of Mozeidon components ([Mozeidon firefox add-on](https://github.com/egovelox/mozeidon/tree/main?tab=readme-ov-file#mozeidon-firefox-addon), [Mozeidon native-app](https://github.com/egovelox/mozeidon/tree/main?tab=readme-ov-file#mozeidon-native-app) and [Mozeidon CLI](https://github.com/egovelox/mozeidon/tree/main?tab=readme-ov-file#mozeidon-cli)).

Note that you cannot list **history items** with this Raycast extension : only **tabs**, **recently-closed tabs**, and **bookmarks**.

![mozeidon-4](https://github.com/egovelox/mozeidon/assets/56078155/a3b8d378-7fe2-4062-9722-15b4cf7f9d6f)


### MacOS swift app-agent

**Not maintained anymore**.  

Please now switch to [Swell](https://github.com/egovelox/swell) 🏄 🌊  

**Not maintained anymore**.  

If you ask for something faster than [Raycast](https://github.com/egovelox/mozeidon/tree/main?tab=readme-ov-file#raycast-extension) ( which I find quite slow to trigger the search list ),  
you might take a look at this macOS app [mozeidon-macos-ui](https://github.com/egovelox/mozeidon-macos-ui)

[🆕 Quick install of the mozeidon-macos-ui brew cask](https://github.com/egovelox/mozeidon-macos-ui?tab=readme-ov-file#homebrew)

<img width="640" alt="mozeidon-macos-ui" src="https://github.com/user-attachments/assets/8590a296-3a4d-4287-b362-83804893710e" />


## Releases

Various releases of the Mozeidon CLI can be found on the [releases page](https://github.com/egovelox/mozeidon/releases).

Releases are managed with github-actions and [goreleaser](https://github.com/goreleaser/goreleaser).

A release will be auto-published when a new git tag is pushed,
e.g :

```bash
git clone https://github.com/egovelox/mozeidon.git && cd mozeidon;

git tag -a v2.0.0 -m "A new mozeidon (CLI) release"

git push origin v2.0.0
```

## Local development setup

We'll assume that you installed and followed the steps described in the `Mozeidon native-app` paragraph above.
In fact, you rarely need to modify this component, it's just a message broker (see the `architecture` paragraph above ).

First clone this repository.

### Quick setup with just

If you have [just](https://github.com/casey/just) installed, you can set up everything in one command:

```bash
just setup-all
```

This will build the CLI, extensions, Raycast extension, and configure native messaging.

Other useful just commands:
- `just --list` - Show all available commands
- `just setup-native-messaging` - Configure Firefox native messaging
- `just check-native-messaging` - Verify native messaging is configured
- `just package-firefox` - Package extension for AMO upload
- `just tabs-get` - Test getting open tabs

### Manual setup

Build the cli and the extensions locally :

```bash
make all
```
Now, before loading the local extension in your browser, don't forget to disable any running instance of the `mozeidon` extension. 

Then, in Firefox ( or any web-browser ), via `Extensions > Debug Addons > Load Temporary Add-on`, select the manifest file in `firefox-addon/manifest.json`.  
This will load the local extension.

From there, you may want to go further and build the CLI also :

You should now be able to execute the CLI using the local binary :

```bash
./cli/mozeidon tabs get
```

## Notes and well-known limitations

#### `mozeidon` cannot work simultaneously in different web-browsers

For users who installed both the firefox-addon AND the chrome-addon, or for those who use multiple browsers, each loading the mozeidon extension :

`mozeidon` CLI will not work properly when multiple instances of the mozeidon browser-extension are activated at the same time.  
To overcome this limitation, keep one extension activated (e.g firefox-addon)  
and deactivate the other extension (e.g chrome-addon).

If you notice any error during this operation, try to deactivate/reactivate the browser extension 🙏.

This is currently not planned for resolution. Technically, to overcome this limitation, we would need a particular native-messaging-host (`mozeidon-native-app`) 
for each web-browser.  
The installation would surely be too complex for most users. 
(see [architecture](https://github.com/egovelox/mozeidon?tab=readme-ov-file#architecture)),  

#### For a given web-browser, `mozeidon` cannot work correctly if you use multiple browser windows.

Mozeidon cannot guess in which browser window a given tab-id is located.  
Consequently, if you have 2 browser windows,  
`mozeidon tabs switch [tab-id]` will switch to the last active browser window,  
where, possibly, the tab-id you targeted is not located.

This is by design, `mozeidon` was precisely meant to facilitate browsing within a unique browser window.

See [https://github.com/egovelox/mozeidon/issues/6](https://github.com/egovelox/mozeidon/issues/6)

#### `mozeidon bookmark update --folder-path` and browsers' *default bookmark folder*.

##### TLDR

By design, the `mozeidon` cli can only `move` bookmarks ( e.g changing their folder location ) to places located under the *Bookmarks Bar* tree. 

For `mozeidon` cli, there is no valid `--folder-path` for places located outside of the *Bookmarks Bar* tree.

The cli still allows `get`, `update`, `delete` operations for those bookmarks located outside of the *Bookmarks Bar* tree, but as for `moving` them, it can only `move` them inside the *Bookmarks Bar* tree. 

This is by design : the cli uses a `folder-path` model, with the root, represented as `/`, being actually the *Bookmarks Bar* itself.

##### Details

Each browser has a *default bookmark folder*, e.g `Other Bookmarks` in Firefox, `Other Favorites` in Edge, etc.  
But internally, in the browser, this *default bookmark folder* IS NOT LOCATED INSIDE THE *Bookmarks Bar* tree ( though this folder may be displayed by the browser in the *Bookmarks Bar* ).

This *default bookmark folder* contravenes `mozeidon` design for bookmarks folder locations (`--folder-path`):  
In `mozeidon`, a `--folder-path` value:
- is always a path starting from the root of the *Bookmarks Bar* tree ( aka *Favorites Bar* ) represented as `/`
- should always start with `/`
- should always end with `/`


E.g in `mozeidon` :
- `--folderPath "//surf/"` represents a `surf` folder, inside a `""` (no title) folder, inside the *Bookmarks Bar*
- `--folderPath "//Other Bookmarks/"` represents a `Other Bookmarks` folder, inside a `""` (no title) folder, inside the *Bookmarks Bar*
  ( and it should not be confused with Firefox's *default bookmark folder* )

Because internally, inside the browser, this *default bookmark folder* is not located in the *Bookmarks Bar* tree - though this folder may be displayed by the browser in the *Bookmarks Bar*, the `--folder-path` flag on the `mozeidon bookmark` commands cannot reference such *default bookmark folder*,
meaning that :
- You can create a bookmark in the *default bookmark folder* with `mozeidon bookmark new` by omitting the `--folder-path` flag.
- Such created bookmark will appear in the results of `mozeidon bookmarks` with a `parent` field of `"parent":"//Other Bookmarks/"`
- You cannot move such bookmark in a *child* folder ( i.e inside the *default bookmark folder* ) with e.g (let's say our bookmark-id is `42`) `mozeidon bookmark update 42 --folder-path '//Other Bookmarks/surf/'`
- Instead, that command will move the bookmark in a `surf` folder, inside a `Other Bookmarks` folder, inside a `""` (no title) folder, inside the *Bookmarks Bar*
- But still, you can move such bookmark in a folder located inside the *Bookmarks Bar*, with e.g `mozeidon bookmark update 42 --folder-path '/surf/'`
