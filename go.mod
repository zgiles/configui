module github.com/zgiles/configui

go 1.13

require (
	fyne.io/fyne v1.2.2
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20190924025748-f65c72e2690d // indirect
	github.com/pjediny/mndp v0.0.0-20161125182440-29d461d9584f
	github.com/pkg/sftp v1.11.0
	golang.org/x/crypto v0.0.0-20200208060501-ecb85df21340
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
)

replace github.com/pjediny/mndp => github.com/zgiles/mndp v0.0.0-20200209053246-f5216b505b79
