package main

import (
	_ "embed"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"fyne.io/systray"
)

//go:embed warp_logo_connected.png
var iconConnected []byte

//go:embed warp_logo_disconnected.png
var iconDisconnected []byte

type connectionState int

const init_ connectionState = -1
const connected connectionState = 1
const disconnected connectionState = 0

var currentConnectionState connectionState = init_

var icon []byte = iconDisconnected

type NoopWriter struct{}

func (*NoopWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func main() {

	// log.SetOutput(new(NoopWriter))

	var start func()
	var stop func()

	loopTimer := time.NewTicker(2 * time.Second)
	for {
		<-loopTimer.C

		newState := getWarpStatus()

		if currentConnectionState != newState {
			log.Println("found state change to", newState)
			if stop != nil {
				stop()
				systray.ResetMenu()
				time.Sleep(500 * time.Millisecond)
			}

			currentConnectionState = newState
			if currentConnectionState == connected {
				icon = iconConnected
			} else {
				icon = iconDisconnected
			}
			start, stop = systray.RunWithExternalLoop(onReady, onExit)
			start()
		}

	}

}

func onReady() {
	systray.SetIcon(icon)
	systray.SetTitle("Cloudflare WARP Status UI")

	var btn *systray.MenuItem
	if currentConnectionState == connected {
		btn = systray.AddMenuItem("Disconnect", "Disconnect from warp")
	} else {
		btn = systray.AddMenuItem("Connect", "Connect to warp")
	}
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")

	go func() {
		<-mQuit.ClickedCh
		os.Exit(0)
	}()

	go func() {
		<-btn.ClickedCh
		if currentConnectionState == connected {
			disconnectWarp()
		} else {
			connectWarp()
		}
	}()

}

func onExit() {}

func connectWarp() {
	connectCmd := exec.Command("warp-cli", "connect")
	if err := connectCmd.Run(); err != nil {
		log.Println(err)
	}
}

func disconnectWarp() {
	connectCmd := exec.Command("warp-cli", "disconnect")
	if err := connectCmd.Run(); err != nil {
		log.Println(err)
	}
}

func getWarpStatus() connectionState {
	statusCmd := exec.Command("warp-cli", "status")
	out, err := statusCmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	if strings.Contains(string(out), "Status update: Connected") {
		return connected
	}
	return disconnected
}
