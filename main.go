package main

import (
	"log"
	"net/http"

	"github.com/sealor/PlayzApp/internal/player"
)

var mpv = player.Player{}

func servePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type:", "text/html")
	_, _ = w.Write([]byte(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>PlayzApp</title>
		</head>
		<body>
		<h1>MPV Remote</h1>
		<button onclick="send('set_property', 'pause', true)">Pause</button>
		<button onclick="send('set_property', 'pause', false)">Play</button>
		<button onclick="send('cycle', 'pause')">Toggle Pause</button>
		<script>
			function send(...cmd) {
				fetch('/api/command', {
					method: 'POST',
					headers: {'Content-Type': 'application/json'},
					body: JSON.stringify({ command: cmd }),
				}).then(res => res.text()).then(alert)
			}
		</script>
		</body>
		</html>`))
}

func handleCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	if err := mpv.Start(); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", servePage)
	http.HandleFunc("/api/command", handleCommand)

	log.Println("Server is running")
	_ = http.ListenAndServe(":8080", nil)

	if err := mpv.Stop(); err != nil {
		log.Fatal(err)
	}
}
