package core

import (
	"github.com/egovelox/mozeidon/browser/core"
	"github.com/egovelox/mozeidon/profiles"
)

type App struct {
	browser *core.BrowserService
	Profile *profiles.Profile
}

func newApp(profile *profiles.Profile) (*App, error) {
	browser := core.NewBrowserService(profile.IpcName)

	return &App{browser: browser, Profile: profile}, nil
}

func NewAppWithProfile(profileId string) (*App, error) {
	profile, err := profiles.GetProfileById(profileId)
	if err != nil {
		// Fallback to legacy IPC name when no profiles are registered
		// (e.g. when using a pre-v4 native app)
		profile = &profiles.Profile{
			IpcName: "mozeidon_native_app",
		}
	}

	return newApp(profile)
}
