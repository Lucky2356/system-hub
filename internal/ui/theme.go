package ui

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Palette shared by the custom theme and by widgets that draw their own colours
// (status text, gauges). Keeping them in one place stops the UI from drifting
// into a mix of ad-hoc RGB values.
var (
	AccentColor  = color.NRGBA{R: 56, G: 139, B: 253, A: 255}  // primary / selection
	SuccessColor = color.NRGBA{R: 63, G: 185, B: 80, A: 255}   // active, running
	WarningColor = color.NRGBA{R: 210, G: 153, B: 34, A: 255}  // inactive, dead
	DangerColor  = color.NRGBA{R: 248, G: 81, B: 73, A: 255}   // failed, exited
	MutedColor   = color.NRGBA{R: 139, G: 148, B: 158, A: 255} // unknown / secondary
)

// systemHubTheme is a dark-first theme with a calmer, higher-contrast palette
// than the Fyne defaults and slightly roomier spacing.
type systemHubTheme struct {
	variant fyne.ThemeVariant
}

func newAppTheme(variant fyne.ThemeVariant) fyne.Theme {
	return &systemHubTheme{variant: variant}
}

func (t *systemHubTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	v := t.variant

	switch name {
	case theme.ColorNamePrimary:
		return AccentColor
	case theme.ColorNameSuccess:
		return SuccessColor
	case theme.ColorNameWarning:
		return WarningColor
	case theme.ColorNameError:
		return DangerColor
	}

	if v == theme.VariantLight {
		switch name {
		case theme.ColorNameBackground:
			return color.NRGBA{R: 246, G: 248, B: 250, A: 255}
		case theme.ColorNameForeground:
			return color.NRGBA{R: 31, G: 35, B: 40, A: 255}
		case theme.ColorNameSeparator:
			return color.NRGBA{R: 216, G: 222, B: 228, A: 255}
		case theme.ColorNameInputBackground, theme.ColorNameOverlayBackground:
			return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		case theme.ColorNameDisabled:
			return color.NRGBA{R: 140, G: 149, B: 159, A: 255}
		}
		return theme.DefaultTheme().Color(name, theme.VariantLight)
	}

	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 13, G: 17, B: 23, A: 255}
	case theme.ColorNameForeground:
		return color.NRGBA{R: 230, G: 237, B: 243, A: 255}
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 48, G: 54, B: 61, A: 255}
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 22, G: 27, B: 34, A: 255}
	case theme.ColorNameOverlayBackground, theme.ColorNameMenuBackground:
		return color.NRGBA{R: 22, G: 27, B: 34, A: 255}
	case theme.ColorNameButton:
		return color.NRGBA{R: 33, G: 38, B: 45, A: 255}
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 110, G: 118, B: 129, A: 255}
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 26, G: 31, B: 38, A: 255}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 48, G: 54, B: 61, A: 255}
	case theme.ColorNameShadow:
		return color.NRGBA{R: 1, G: 4, B: 9, A: 140}
	}

	return theme.DefaultTheme().Color(name, theme.VariantDark)
}

func (t *systemHubTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *systemHubTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *systemHubTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 5
	case theme.SizeNameInnerPadding:
		return 9
	case theme.SizeNameInlineIcon:
		return 18
	case theme.SizeNameScrollBarSmall:
		return 5
	// widget.Card renders its title at heading size; the default dwarfs the
	// card's own content on a dense dashboard.
	case theme.SizeNameHeadingText:
		return 15
	case theme.SizeNameSubHeadingText:
		return 13
	}
	return theme.DefaultTheme().Size(name)
}

// StatusColor maps a systemd/docker state word to a palette colour. It is the
// single source of truth for state colouring across the dashboard, services and
// docker tabs, which previously each used slightly different RGB values.
func StatusColor(state string) color.Color {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "active", "running":
		return SuccessColor
	case "failed", "exited", "error":
		return DangerColor
	case "activating", "deactivating", "restarting", "paused", "created":
		return WarningColor
	// The dashboard passes a translated word for a missing service/container, so
	// both the English key and its Russian translation are matched here.
	case "inactive", "dead", "unavailable", "недоступен", "stopped":
		return MutedColor
	default:
		return MutedColor
	}
}

// StatusImportance maps a service/container state onto a widget Importance,
// which the theme renders with the matching palette colour.
//
// Prefer this over StatusColor for list rows: a canvas.Text placed in a border
// layout's side slot did not render inside widget.List rows, so states were
// invisible. A Label carries its colour through the theme and always draws.
func StatusImportance(state string) widget.Importance {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "active", "running":
		return widget.SuccessImportance
	case "failed", "exited", "error":
		return widget.DangerImportance
	case "activating", "deactivating", "restarting", "paused", "created":
		return widget.WarningImportance
	default:
		return widget.LowImportance
	}
}

// stateColumnWidth is the character width the state badge is padded to, so the
// service and container names line up in a column.
const stateColumnWidth = 12

// statePlaceholder sizes the list row template's state badge.
const statePlaceholder = "            " // stateColumnWidth spaces

// padState pads a state word to a fixed width so names align.
func padState(state string) string {
	if len(state) >= stateColumnWidth {
		return state
	}
	return state + strings.Repeat(" ", stateColumnWidth-len(state))
}

// ApplyTheme sets the app theme for the given config value ("light"/"dark").
func ApplyTheme(a fyne.App, themeName string) {
	if themeName == "light" {
		a.Settings().SetTheme(newAppTheme(theme.VariantLight))
		return
	}
	a.Settings().SetTheme(newAppTheme(theme.VariantDark))
}
