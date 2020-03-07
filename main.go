package main

import (
	"fmt"
	"fyne.io/fyne"
	fyneapp "fyne.io/fyne/app"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
	"runtime"

	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pjediny/mndp/mndplib"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"gopkg.in/alecthomas/kingpin.v2"
)

type peerlist []string

type rootConfig struct {
	rscfile    string
	npkfile    string
	debug      bool
	serverip   string
	username   string
	password   string
	statuschan chan string
}

var version string
var config rootConfig

func discoverdevices() (discovered []string) {
	tick := time.Tick(10 * time.Second)
	ch := make(chan *mndplib.MNDPMessage)

	listener := mndplib.NewMNDPListener()
	listener.Listen(ch)

	for {
		select {
		case msg := <-ch:
			// Mac address
			t, err := msg.TLV(1)
			if err != nil {
				continue
			}
			_, serial := t.Value()
			t, err = msg.TLV(5)
			if err != nil {
				continue
			}
			_, name := t.Value()
			address := msg.Source()
			discovered = append(discovered, fmt.Sprintf("%s-%s-%s", name, serial, address))
		case <-tick:
			return
		}
	}
}

func (config *rootConfig) performresetandrunconfig() error {

	// SSH to see if it works
	sshconfig := &ssh.ClientConfig{
		User: config.username,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.username),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	server := config.serverip + ":22"

	if config.rscfile == "" {
		err := fmt.Errorf("File name must not be empty")
		config.statuschan <- err.Error()
		log.Println(err)
		return err
	}

	_, err := os.Stat(config.rscfile)
	if os.IsNotExist(err) {
		config.statuschan <- err.Error()
		log.Println(err)
		return err
	}

	// Dialing
	config.statuschan <- "Dialing.."
	conn, err := ssh.Dial("tcp", server, sshconfig)
	if err != nil {
		config.statuschan <- "unable to connect: " + err.Error()
		fmt.Println(err)
		return err
	}
	defer conn.Close()

	config.statuschan <- "Uploading file"
	c, err := sftp.NewClient(conn, sftp.MaxPacket(1<<15))
	if err != nil {
		config.statuschan <- err.Error()
		log.Printf("unable to start sftp subsytem: %v", err)
		return err
	}

	w, err := c.Create(path.Join("flash", path.Base(config.rscfile)))
	if err != nil {
		config.statuschan <- err.Error()
		log.Println(err)
		return err
	}

	log.Println("Opening file")
	f, err := os.Open(config.rscfile)
	if err != nil {
		config.statuschan <- err.Error()
		log.Println(err)
		return err
	}

	_, err = io.Copy(w, f)
	if err != nil {
		config.statuschan <- err.Error()
		log.Println(err)
		return err
	}

	w.Close()
	f.Close()
	c.Close()

	config.statuschan <- "Running Command "

	// Connect and reset with file
	sess, err := conn.NewSession()
	if err != nil {
		config.statuschan <- err.Error()
		log.Println(err)
		return err
	}

	si, err := sess.StdinPipe()
	if err != nil {
		config.statuschan <- err.Error()
		log.Println(err)
		return err
	}

	err = sess.Shell()
	if err != nil {
		config.statuschan <- err.Error()
		log.Println(err)
		return err
	}

	_, err = si.Write([]byte("/system reset-configuration no-defaults=yes run-after-reset=" + path.Join("flash", path.Base(config.rscfile)) + "\n"))
	if err != nil {
		config.statuschan <- err.Error()
		log.Println(err)
		return err
	}

	_, err = si.Write([]byte("y"))
	if err != nil {
		config.statuschan <- err.Error()
		log.Println(err)
		return err
	}

	sess.Wait()

	config.statuschan <- "System is resetting. Done."

	return nil
}

func main() {
	config = rootConfig{}
	// var peersraw []string
	app := kingpin.New("configui", "Mesh Config UI")
	app.UsageTemplate(kingpin.CompactUsageTemplate)
	app.Flag("debug", "Debug on").BoolVar(&config.debug)
	app.Version(version)
	kingpin.MustParse(app.Parse(os.Args[1:]))
	// config.peers = peerlist(peersraw)

	if config.debug {
		log.Printf("%+v", config)
	}

	// defaults and allocations
	config.statuschan = make(chan string)
	config.username = "admin"
	config.serverip = "192.168.88.1"

	// Find all rsc and firmware files
	npkfiles := []string{}
	rscfiles := []string{}
	localdir := "."
	if runtime.GOOS == "darwin" {
		localdir = "~/Downloads"
	}
	localfiles, _ := ioutil.ReadDir(localdir)
	for _, file := range localfiles {
		if file.Mode().IsRegular() {
			if filepath.Ext(file.Name()) == ".rsc" {
				rscfiles = append(rscfiles, file.Name())
			}
			if filepath.Ext(file.Name()) == ".npk" {
				rscfiles = append(rscfiles, file.Name())
			}
		}
	}

	// GUI
	guiapp := fyneapp.New()
	w := guiapp.NewWindow("ConfigUI")
	w.Resize(fyne.Size{400, 200})

	statuslabel := widget.NewLabel("Searching for Devices")
	go func() {
		for s := range config.statuschan {
			statuslabel.SetText(s)
		}
	}()

	ipmanualentry := widget.NewEntry()
	ipmanualentry.SetPlaceHolder("Router IP")
	ipmanualentry.SetText(config.serverip)

	antennaip := widget.NewSelect([]string{}, nil)
	go func() {
		a := discoverdevices()
		antennaip.Options = a
		config.statuschan <- "Done searching. Select Below"
	}()

	ipauto := widget.NewCheck("Auto-discover routers", func(checked bool) {
		if !checked {
			ipmanualentry.Show()
			antennaip.Hide()
			return
		}
		ipmanualentry.Hide()
		antennaip.Show()
	})
	ipauto.SetChecked(true)

	rscentry := widget.NewSelect(rscfiles, nil)
	rscentry.Hide()
	npkentry := widget.NewSelect(npkfiles, nil)
	npkentry.Hide()

	rsccheck := widget.NewCheck("Flash Config", func(checked bool) {
		if !checked {
			rscentry.Hide()
			return
		}
		rscentry.Show()
	})
	rsccheck.SetChecked(false)

	npkcheck := widget.NewCheck("Flash Firmware", func(checked bool) {
		if !checked {
			npkentry.Hide()
			return
		}
		npkentry.Show()
	})
	npkcheck.SetChecked(false)

	form := &widget.Form{}
	form.Append("Status", statuslabel)
	form.Append("", ipauto)
	form.Append("", antennaip)
	form.Append("", ipmanualentry)
	form.Append("", npkcheck)
	form.Append("Firmware Files", npkentry)
	form.Append("", rsccheck)
	form.Append("Config Files", rscentry)

	startbutton := widget.NewButton("Start Configuration", func() {
		// antennaip.ReadOnly = true
		config.statuschan <- "Connecting.."
		if ipauto.Checked {
			chopped := strings.Split(antennaip.Selected, "-")[2]
			config.serverip = chopped
		} else {
			config.serverip = ipmanualentry.Text
		}
		config.serverip = antennaip.Selected
		config.rscfile = rscentry.Selected
		config.npkfile = npkentry.Selected
		config.performresetandrunconfig()
	})
	startbutton.Style = widget.PrimaryButton

	quitbutton := widget.NewButton("Quit", func() {
		guiapp.Quit()
	})

	buttonrow := fyne.NewContainerWithLayout(layout.NewGridLayout(3),
		startbutton,
		quitbutton,
	)

	w.SetContent(widget.NewVBox(
		widget.NewLabel("ConfigUI - Source: github.com/zgiles/configui\n"),
		widget.NewVBox(form),
		widget.NewLabel("WARNING. Clicking Start will wipe a device, take care it is the correct device."),
		buttonrow,
	))

	w.ShowAndRun()

	// end GUI

}
