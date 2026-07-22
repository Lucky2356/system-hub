package ui

import (
	"testing"

	"github.com/Lucky2356/system-hub/internal/config"
	"github.com/Lucky2356/system-hub/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// The window must fit a modest laptop screen. 1366x768 is still a common
// desktop resolution, and a window is not the whole screen, so a tab has to lay
// out inside this budget.
//
// Only width is asserted. Height is logged but not enforced, because Fyne's
// MinSize is not idempotent for text widgets — RichText caches a minimum on the
// first call and lays out differently afterwards, so the same Logs tab measures
// 473px standalone and 633px from inside AppTabs. Width does not suffer from
// this and matches what the running app does: the live window refused to go
// below 2015px, and that number is explained exactly by the untruncated labels
// this test now guards.
const maxTabWidth = 1000

// TestEveryTabFitsASmallWindow is the regression test for the defect that made
// the app unusable: the window could not be made narrower than 2015px — wider
// than a 1080p screen — so `w.Resize(900, 550)` in window.go did nothing.
//
// The cause is not any single tab. container.AppTabs takes its MinSize as the
// *maximum over every tab* (container/tabs.go), so the one widest tab sets the
// floor for the whole window, even while the user is looking at another one.
// That is why this test measures all of them: fixing the tab you happen to be
// on changes nothing until the widest one is fixed too.
// Both languages are measured: Russian is the default and its words are longer
// than the English keys, so testing only English would pass while the shipped
// UI overflowed.
func TestEveryTabFitsASmallWindow(t *testing.T) {
	for _, language := range i18n.Supported() {
		t.Run(language, func(t *testing.T) {
			defer i18n.SetLanguage(i18n.English)
			i18n.SetLanguage(language)

			for _, tc := range allTabs(t) {
				t.Run(tc.name, func(t *testing.T) {
					min := tc.build().MinSize()
					t.Logf("%s [%s] needs %.0fx%.0f", tc.name, language, min.Width, min.Height)

					if min.Width > maxTabWidth {
						t.Errorf("%s [%s] needs %.0fpx of width; the budget is %d. AppTabs takes the max over all tabs, so this alone forces the whole window wider.",
							tc.name, language, min.Width, maxTabWidth)
					}
				})
			}
		})
	}
}

// TestWindowFitsASmallScreen checks the assembled window rather than the tabs
// in isolation, so the tab bar and the header are included in the budget.
func TestWindowFitsASmallScreen(t *testing.T) {
	defer i18n.SetLanguage(i18n.English)
	i18n.SetLanguage(i18n.Russian) // the shipped default, and the longer words

	loadInitialData = false
	defer func() { loadInitialData = true }()

	a := newTestApp()
	defer func() {
		RunClosers()
		test.NewApp() // reset the global app for later tests
	}()

	w := a.NewWindow("")
	defer w.Close()

	content := buildWindowContent(a, w, testConfig())
	min := content.MinSize()

	t.Logf("the whole window needs %.0fx%.0f", min.Width, min.Height)

	if min.Width > maxTabWidth {
		t.Errorf("the window needs %.0fpx of width; the budget is %d", min.Width, maxTabWidth)
	}
}

// TestDataDoesNotWidenTheLayout is the test that actually pins the original
// defect, and the tab measurements above cannot catch it on their own: they
// measure a freshly built tab, which has no data in it yet.
//
// A widget.Label reports the full width of its text as its MinSize. So the
// window's minimum width grew with whatever the host happened to return — and on
// a machine without Docker it grew by the length of an error message. Measured:
// "Status: waiting" needs 137px, the real docker-unavailable error needs 1424px,
// and that is what forced the 2015px window.
//
// Truncation is what makes a label's MinSize independent of its content.
func TestDataDoesNotWidenTheLayout(t *testing.T) {
	newTestApp()
	defer test.NewApp()

	// The real message from a host with the Docker CLI installed but no daemon
	// running. This is data, not a caption: the user cannot make it shorter.
	const dockerError = "Status: load error — run docker ps: docker [ps -a --format {{json .}}]: " +
		"failed to connect to the docker API at npipe:////./pipe/dockerDesktopLinuxEngine; " +
		"check if the path is correct and if the daemon is running"

	label := newDataLabel("")
	empty := label.MinSize().Width

	label.SetText(dockerError)
	filled := label.MinSize().Width

	t.Logf("a data label is %.0fpx empty and %.0fpx holding a %d-character error",
		empty, filled, len(dockerError))

	if filled > empty {
		t.Errorf("text grew the label's minimum width from %.0f to %.0f; a data label must "+
			"truncate, otherwise the host's output decides how wide the window must be",
			empty, filled)
	}
}

// TestDashboardTilesFitTheirCells guards the trade the dashboard makes: a
// wrapping grid is what keeps the tab's minimum width at one cell, but the cell
// is fixed, so a tile whose content needs more than the cell does not shrink —
// it draws over its neighbour.
//
// This is not hypothetical. It is why an earlier attempt at a wrapping
// dashboard was reverted, and it happened again here: at 240x150 the uptime
// tile drew across the CPU tile's chart. A screenshot caught it both times;
// this test catches it before the screenshot.
func TestDashboardTilesFitTheirCells(t *testing.T) {
	newTestApp()
	defer test.NewApp()

	// Every composition the dashboard actually puts in a cell. Guessing which
	// one is "the richest" is how the Network tile slipped through the first
	// time: it carries two labels and two charts, which is taller than the CPU
	// tile that looked like the worst case.
	tiles := []struct {
		name string
		tile fyne.CanvasObject
	}{
		{"CPU (value + bar + chart)", NewTile("CPU", container.NewVBox(
			newDataLabel("CPU: 100.0%"),
			widget.NewProgressBar(),
			newSparkline(AccentColor, 100).object(),
		))},
		{"RAM (value + bar + details)", NewTile("RAM", container.NewVBox(
			newDataLabel("RAM: 100.0%"),
			widget.NewProgressBar(),
			newDataLabel("15.8 GB / 15.8 GB"),
		))},
		{"Network (2 values + 2 charts)", NewTile("Network", container.NewVBox(
			newDataLabel("RX: 4.1 GB  (13.3 KB/s)"),
			newSparkline(AccentColor, 0).object(),
			newDataLabel("TX: 5.5 GB  (750 B/s)"),
			newSparkline(WarningColor, 0).object(),
		))},
		{"Docker (three values)", NewTile("Docker", container.NewVBox(
			newDataLabel("Docker: unavailable"),
			newDataLabel("Containers: unavailable"),
			newDataLabel("Running: unavailable"),
		))},
	}

	for _, tc := range tiles {
		min := tc.tile.MinSize()
		t.Logf("%-30s needs %.0fx%.0f; the cell is %.0fx%.0f",
			tc.name, min.Width, min.Height, dashboardTileSize.Width, dashboardTileSize.Height)

		if min.Height > dashboardTileSize.Height {
			t.Errorf("%s needs %.0fpx of height but the cell gives %.0f; it will draw over the tile below it",
				tc.name, min.Height, dashboardTileSize.Height)
		}
		if min.Width > dashboardTileSize.Width {
			t.Errorf("%s needs %.0fpx of width but the cell gives %.0f",
				tc.name, min.Width, dashboardTileSize.Width)
		}
	}
}

type tabCase struct {
	name  string
	build func() fyne.CanvasObject
}

// allTabs builds every tab against a test app. The constructors start their own
// background refreshes; MinSize is read immediately, before those can report,
// so the measurement is of the layout and not of whatever the host happens to
// return.
func allTabs(t *testing.T) []tabCase {
	t.Helper()

	// Measure geometry only: a tab that starts fetching host data while being
	// laid out races Fyne's font cache through the test driver, a race the real
	// app avoids because fyne.Do marshals that work to the main goroutine.
	loadInitialData = false
	t.Cleanup(func() { loadInitialData = true })

	a := newTestApp()
	t.Cleanup(func() {
		RunClosers()
		test.NewApp()
	})

	w := a.NewWindow("")
	t.Cleanup(w.Close)

	cfg := testConfig()

	return []tabCase{
		{"Dashboard", func() fyne.CanvasObject { return buildDashboardTab(w, cfg) }},
		{"Services", func() fyne.CanvasObject { return buildServicesTab(w, cfg) }},
		{"Docker", func() fyne.CanvasObject { return buildDockerTab(w, cfg) }},
		{"Processes", func() fyne.CanvasObject { return buildProcessesTab(w) }},
		{"Files", func() fyne.CanvasObject { return buildFilesTab(w) }},
		{"Commands", func() fyne.CanvasObject { return buildCommandsTab(w) }},
		{"System Info", func() fyne.CanvasObject { return NewSystemInfoTab(w) }},
		{"Logs", func() fyne.CanvasObject { return buildLogsTab(cfg) }},
		{"Activity", func() fyne.CanvasObject { return buildActivityTab() }},
		{"Report", func() fyne.CanvasObject { return buildReportTab(w) }},
		{"Settings", func() fyne.CanvasObject { return buildSettingsTab(w, cfg) }},
	}
}

// newTestApp returns a test app carrying the real theme. The default test theme
// has its own paddings and text sizes, so measuring against it would report a
// layout the user never sees.
func newTestApp() fyne.App {
	a := test.NewApp()
	a.Settings().SetTheme(newAppTheme(theme.VariantDark))
	return a
}

// testConfig keeps auto-refresh off: a tab under test must not start polling the
// host while the layout is being measured.
func testConfig() config.Config {
	cfg := config.DefaultConfig()
	cfg.DashboardAutoRefresh = false
	cfg.ServicesAutoRefresh = false
	cfg.DockerAutoRefresh = false
	cfg.LogsAutoRefresh = false
	cfg.LogViewerAutoRefresh = false
	return cfg
}
