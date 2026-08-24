package api

import (
	"fmt"
	"html/template"
	"net/http"
)

type pageDefinition struct {
	Title       string
	Subtitle    string
	Endpoint    string
	ActionLabel string
	ActionPath  string
	ActionBody  string
}

var pageDefinitions = map[string]pageDefinition{
	"lifts": {
		Title: "Lift Control", Subtitle: "Active container lifts and phase transitions",
		Endpoint: "/api/lifts", ActionLabel: "Start 40 ft lift", ActionPath: "/api/lifts",
		ActionBody: `{"container_id":"MSCU-729184-4","length_feet":40}`,
	},
	"spreader": {
		Title: "Spreader Console", Subtitle: "Geometry, engagement and four-corner lock proof",
		Endpoint: "/api/spreader", ActionLabel: "Extend to 40 ft", ActionPath: "/api/spreader/telescope",
		ActionBody: `{"length_feet":40,"mass_tonnes":12.8}`,
	},
	"motion": {
		Title: "Crane Motion", Subtitle: "Trolley trajectories and rail exclusion zones",
		Endpoint: "/api/motion?destination=320", ActionLabel: "Shift rail origin", ActionPath: "/api/motion/rail-origin",
		ActionBody: `{"origin":4,"destination":316}`,
	},
	"safety": {
		Title: "Safety Interlocks", Subtitle: "Brake, storm anchor and clearance state",
		Endpoint: "/api/safety", ActionLabel: "Trigger storm stop", ActionPath: "/api/safety/storm",
		ActionBody: `{"wind_mps":22.4}`,
	},
}

func (s *Server) page(name string) http.HandlerFunc {
	definition := pageDefinitions[name]
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := consoleTemplate.Execute(writer, definition); err != nil {
			http.Error(writer, fmt.Sprintf("render page: %v", err), http.StatusInternalServerError)
		}
	}
}

var consoleTemplate = template.Must(template.New("console").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>PortCrane | {{.Title}}</title>
  <style>
    :root { color-scheme: dark; --bg:#101418; --panel:#182027; --line:#34424c; --text:#eef4f5; --muted:#9cabb3; --accent:#3dd6a2; --warn:#ffcc66; }
    * { box-sizing:border-box; }
    body { margin:0; min-height:100vh; background:var(--bg); color:var(--text); font:15px/1.45 system-ui,sans-serif; }
    header { height:58px; border-bottom:1px solid var(--line); display:flex; align-items:center; justify-content:space-between; padding:0 24px; }
    .brand { font-weight:760; font-size:19px; }
    nav { display:flex; gap:4px; }
    nav a { color:var(--muted); text-decoration:none; padding:8px 11px; border-radius:4px; }
    nav a:hover { color:var(--text); background:#222d35; }
    main { max-width:1180px; margin:0 auto; padding:28px 24px; }
    .heading { display:flex; justify-content:space-between; align-items:flex-end; gap:20px; margin-bottom:20px; }
    h1 { margin:0; font-size:28px; letter-spacing:0; }
    p { margin:5px 0 0; color:var(--muted); }
    button { border:1px solid #52e9b7; background:var(--accent); color:#07120e; font-weight:720; padding:10px 14px; border-radius:5px; cursor:pointer; }
    button:disabled { opacity:.55; cursor:wait; }
    .status { border:1px solid var(--line); border-radius:6px; overflow:hidden; background:var(--panel); }
    .status-head { padding:12px 16px; border-bottom:1px solid var(--line); display:flex; justify-content:space-between; color:var(--muted); }
    .live { color:var(--accent); font-weight:700; }
    pre { min-height:420px; margin:0; padding:20px; overflow:auto; color:#d9e6e8; font:13px/1.6 ui-monospace,monospace; }
    .error { color:#ff8d8d; }
    @media (max-width:720px) { header { height:auto; padding:14px 16px; align-items:flex-start; } nav { display:grid; grid-template-columns:1fr 1fr; } .heading { align-items:flex-start; flex-direction:column; } main { padding:22px 16px; } button { width:100%; } }
  </style>
</head>
<body>
  <header><div class="brand">PORTCRANE / QC-07</div><nav><a href="/lifts">Lifts</a><a href="/spreader">Spreader</a><a href="/motion">Motion</a><a href="/safety">Safety</a></nav></header>
  <main>
    <div class="heading"><div><h1>{{.Title}}</h1><p>{{.Subtitle}}</p></div><button id="action">{{.ActionLabel}}</button></div>
    <section class="status"><div class="status-head"><span>Controller response</span><span class="live" id="indicator">LIVE</span></div><pre id="output">Loading controller state...</pre></section>
  </main>
  <script>
    const endpoint = {{printf "%q" .Endpoint}};
    const actionPath = {{printf "%q" .ActionPath}};
    const actionBody = {{printf "%q" .ActionBody}};
    const output = document.getElementById('output');
    const action = document.getElementById('action');
    async function load() {
      try { const response = await fetch(endpoint); const body = await response.json(); output.className=''; output.textContent=JSON.stringify(body,null,2); }
      catch (error) { output.className='error'; output.textContent=String(error); }
    }
    action.addEventListener('click', async () => {
      action.disabled=true;
      try { const response=await fetch(actionPath,{method:'POST',headers:{'Content-Type':'application/json'},body:actionBody}); const body=await response.json(); output.className=response.ok?'':'error'; output.textContent=JSON.stringify(body,null,2); }
      catch(error) { output.className='error'; output.textContent=String(error); }
      finally { action.disabled=false; }
    });
    load(); setInterval(load, 5000);
  </script>
</body>
</html>`))
