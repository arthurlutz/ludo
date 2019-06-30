package menu

import (
	"fmt"
	"os/user"
	"sort"

	"github.com/libretro/ludo/input"
	"github.com/libretro/ludo/libretro"
	"github.com/libretro/ludo/playlists"
	"github.com/libretro/ludo/scanner"
	"github.com/libretro/ludo/state"
	"github.com/libretro/ludo/utils"
	"github.com/libretro/ludo/video"
	colorful "github.com/lucasb-eyer/go-colorful"

	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
)

type sceneTabs struct {
	entry
}

func buildTabs() Scene {
	var list sceneTabs
	list.label = "Ludo"

	list.children = append(list.children, entry{
		label:    "Main Menu",
		subLabel: "Load cores and games manually",
		icon:     "main",
		callbackOK: func() {
			menu.Push(buildMainMenu())
		},
	})

	list.children = append(list.children, entry{
		label:    "Settings",
		subLabel: "Configure Ludo",
		icon:     "setting",
		callbackOK: func() {
			menu.Push(buildSettings())
		},
	})

	list.children = append(list.children, getPlaylists()...)

	list.children = append(list.children, entry{
		label:    "Add games",
		subLabel: "Scan your collection",
		icon:     "add",
		callbackOK: func() {
			usr, _ := user.Current()
			menu.Push(buildExplorer(usr.HomeDir, nil,
				func(path string) {
					scanner.ScanDir(path, refreshTabs)
				},
				&entry{
					label: "<Scan this directory>",
					icon:  "scan",
				}))
		},
	})

	list.segueMount()

	return &list
}

// refreshTabs is called after playlist scanning is complete. It inserts the new
// playlists in the tabs, and makes sure that all the icons are positioned and
// sized properly.
func refreshTabs() {
	e := menu.stack[0].Entry()
	l := len(e.children)
	pls := getPlaylists()

	// This assumes that the two first tabs are not playlists, and that the last
	// tab is the scanner.
	e.children = append(e.children[:2], append(pls, e.children[l-1:]...)...)

	// Update which tab is the active tab after the refresh
	if e.ptr >= 2 {
		e.ptr += len(pls) - (l - 3)
	}

	// Ensure new icons are styled properly
	for i := range e.children {
		if i == e.ptr {
			e.children[i].iconAlpha = 1
			e.children[i].scale = 0.75
			e.children[i].width = 500
		} else if i < e.ptr {
			e.children[i].iconAlpha = 1
			e.children[i].scale = 0.25
			e.children[i].width = 128
		} else if i > e.ptr {
			e.children[i].iconAlpha = 1
			e.children[i].scale = 0.25
			e.children[i].width = 128
		}
	}

	// Adapt the tabs scroll value
	if len(menu.stack) == 1 {
		menu.scroll = float32(e.ptr * 128)
	} else {
		e.children[e.ptr].margin = 1360
		menu.scroll = float32(e.ptr*128 + 680)
	}
}

// getPlaylists browse the filesystem for CSV files, parse them and returns
// a list of menu entries. It is used in the tabs, but could be used somewhere
// else too.
func getPlaylists() []entry {
	playlists.Load()

	// To store the keys in slice in sorted order
	var keys []string
	for k := range playlists.Playlists {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var pls []entry
	for _, path := range keys {
		path := path
		filename := utils.FileName(path)
		count := playlists.Count(path)
		pls = append(pls, entry{
			label:    playlists.ShortName(filename),
			subLabel: fmt.Sprintf("%d Games - 0 Favorites", count),
			icon:     filename,
			callbackOK: func() {
				menu.Push(buildPlaylist(path))
			},
		})
	}
	return pls
}

func (tabs *sceneTabs) Entry() *entry {
	return &tabs.entry
}

func (tabs *sceneTabs) segueMount() {
	for i := range tabs.children {
		e := &tabs.children[i]

		if i == tabs.ptr {
			e.labelAlpha = 1
			e.iconAlpha = 1
			e.scale = 0.75
			e.width = 500
		} else if i < tabs.ptr {
			e.labelAlpha = 0
			e.iconAlpha = 1
			e.scale = 0.25
			e.width = 128
		} else if i > tabs.ptr {
			e.labelAlpha = 0
			e.iconAlpha = 1
			e.scale = 0.25
			e.width = 128
		}
	}

	tabs.animate()
}

func (tabs *sceneTabs) segueBack() {
	tabs.animate()
}

func (tabs *sceneTabs) animate() {
	for i := range tabs.children {
		e := &tabs.children[i]

		var labelAlpha, iconAlpha, scale, width float32
		if i == tabs.ptr {
			labelAlpha = 1
			iconAlpha = 1
			scale = 0.75
			width = 500
		} else if i < tabs.ptr {
			labelAlpha = 0
			iconAlpha = 1
			scale = 0.25
			width = 128
		} else if i > tabs.ptr {
			labelAlpha = 0
			iconAlpha = 1
			scale = 0.25
			width = 128
		}

		menu.tweens[&e.labelAlpha] = gween.New(e.labelAlpha, labelAlpha, 0.15, ease.OutSine)
		menu.tweens[&e.iconAlpha] = gween.New(e.iconAlpha, iconAlpha, 0.15, ease.OutSine)
		menu.tweens[&e.scale] = gween.New(e.scale, scale, 0.15, ease.OutSine)
		menu.tweens[&e.width] = gween.New(e.width, width, 0.15, ease.OutSine)
		menu.tweens[&e.margin] = gween.New(e.margin, 0, 0.15, ease.OutSine)
	}
	menu.tweens[&menu.scroll] = gween.New(menu.scroll, float32(tabs.ptr*128), 0.15, ease.OutSine)
}

func (tabs *sceneTabs) segueNext() {
	cur := &tabs.children[tabs.ptr]
	menu.tweens[&cur.margin] = gween.New(cur.margin, 1360, 0.15, ease.OutSine)
	menu.tweens[&menu.scroll] = gween.New(menu.scroll, menu.scroll+680, 0.15, ease.OutSine)
	for i := range tabs.children {
		e := &tabs.children[i]
		if i != tabs.ptr {
			menu.tweens[&e.iconAlpha] = gween.New(e.iconAlpha, 0, 0.15, ease.OutSine)
		}
	}
}

func (tabs *sceneTabs) update(dt float32) {
	// Right
	repeatRight(dt, input.NewState[0][libretro.DeviceIDJoypadRight], func() {
		tabs.ptr++
		if tabs.ptr >= len(tabs.children) {
			tabs.ptr = 0
		}
		tabs.animate()
	})

	// Left
	repeatLeft(dt, input.NewState[0][libretro.DeviceIDJoypadLeft], func() {
		tabs.ptr--
		if tabs.ptr < 0 {
			tabs.ptr = len(tabs.children) - 1
		}
		tabs.animate()
	})

	// OK
	if input.Released[0][libretro.DeviceIDJoypadA] {
		if tabs.children[tabs.ptr].callbackOK != nil {
			tabs.segueNext()
			tabs.children[tabs.ptr].callbackOK()
		}
	}
}

// Tab is a widget that draws the homepage hexagon plus title
func Tab(props *Props, i int, e entry) func() {
	c := colorful.Hcl(float64(i)*20, 0.5, 0.5)
	return Box(props,
		Box(&Props{Width: e.width * menu.ratio},
			Image(&Props{
				X:      e.width/2*menu.ratio - 220*e.scale*menu.ratio,
				Y:      -220 * e.scale * menu.ratio,
				Width:  440 * menu.ratio,
				Height: 440 * menu.ratio,
				Scale:  e.scale,
				Color:  video.Color{R: float32(c.R), G: float32(c.B), B: float32(c.G), A: e.iconAlpha},
			}, menu.icons["hexagon"]),
			Image(&Props{
				X:      e.width/2*menu.ratio - 128*e.scale*menu.ratio,
				Y:      -128 * e.scale * menu.ratio,
				Width:  256 * menu.ratio,
				Height: 256 * menu.ratio,
				Scale:  e.scale,
				Color:  video.Color{R: 1, G: 1, B: 1, A: e.iconAlpha},
			}, menu.icons[e.icon]),
			Label(&Props{
				Y:         250 * menu.ratio,
				TextAlign: "center",
				Scale:     0.6 * menu.ratio,
				Color:     video.Color{R: float32(c.R), G: float32(c.B), B: float32(c.G), A: e.labelAlpha},
			}, e.label),
			Label(&Props{
				Y:         330 * menu.ratio,
				TextAlign: "center",
				Scale:     0.4 * menu.ratio,
				Color:     video.Color{R: float32(c.R), G: float32(c.B), B: float32(c.G), A: e.labelAlpha},
			}, e.subLabel),
		),
	)
}

func (tabs sceneTabs) render() {
	_, h := vid.Window.GetFramebufferSize()

	var children []func()
	for i, e := range tabs.children {
		children = append(children, Tab(&Props{
			Y:     float32(h) / 2,
			Width: e.width*menu.ratio + e.margin*menu.ratio,
		}, i, e))
	}

	HBox(&Props{
		X: 710*menu.ratio - menu.scroll*menu.ratio,
	}, children...)()
}

func (tabs sceneTabs) drawHintBar() {
	HintBar(&Props{},
		Hint(&Props{Hidden: !state.Global.CoreRunning}, "key-p", "RESUME"),
		Hint(&Props{}, "key-left-right", "NAVIGATE"),
		Hint(&Props{}, "key-x", "OPEN"),
	)()

	mkVBox(wProps{
		Color:        video.Color{0, 0, 0, 0.2},
		BorderRadius: 0.1,
		Padding:      20,
	},
		mkLabel(wProps{
			Scale:  0.8 * menu.ratio,
			Height: 50,
			Color:  video.Color{0.2, 0.2, 0, 1},
		}, "Bonjour"),
		mkLabel(wProps{
			Scale:  0.6 * menu.ratio,
			Height: 50,
			Color:  video.Color{0.2, 0.2, 0, 1},
		}, "This is a longer piece of text"),
		&hBox{
			Children: []Widget{
				mkButton("key-z", "Fuufuu", video.Color{0, 0.5, 0, 1}),
				mkButton("key-x", "Lehleh", video.Color{0.5, 0, 0.5, 1}),
			},
		},
	).Draw(0, 0)
}
