package layout

import (
	"github.com/AkifhanIlgaz/tini/internal/shared/components/sidebar"
	"github.com/AkifhanIlgaz/tini/internal/shared/icons"
)

const AppName = "tını"

// NavItems is the app's sidebar link list, shared by every feature that
// renders inside the Dashboard shell (see Dashboard) so the sidebar stays
// identical no matter which page is active — moved here from
// internal/features/dashboard once a second feature (user) needed it too.
var NavItems = []sidebar.NavItem{
	{Label: "Anasayfa", Href: "/dashboard", Icon: icons.House},
	{Label: "Playlist", Href: "/playlist", Icon: icons.ListMusic},
	{Label: "Oylama", Href: "/oylama", Icon: icons.Vote},
	{Label: "Mekan bilgileri", Href: "/venue", Icon: icons.Building2},
	{Label: "Kullanıcılar", Href: "/users", Icon: icons.Users},
}

var AdminNavItems = []sidebar.NavItem{
	{Label: "Anasayfa", Href: "/dashboard", Icon: icons.House},
	{Label: "Playlist", Href: "/playlist", Icon: icons.ListMusic},
	{Label: "Oylama", Href: "/oylama", Icon: icons.Vote},
	{Label: "Mekan bilgileri", Href: "/venue", Icon: icons.Building2},
}

// FooterItems renders below NavItems, above the always-present Çıkış yap
// button (see Dashboard) — account-level pages rather than app navigation.
// Currently empty; kept as a slice (not removed) so DashboardProps.FooterItems
// stays a simple field every page can pass through unconditionally.
var FooterItems = []sidebar.NavItem{}
