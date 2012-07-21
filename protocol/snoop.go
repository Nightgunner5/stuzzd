package protocol

import (
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"time"
)

func Snoop() {
	for {
		time.Sleep(10 * time.Minute)
		var memstats runtime.MemStats
		runtime.ReadMemStats(&memstats)
		http.PostForm("http://snoop.minecraft.net/server", url.Values{
			"version":         {SPECIFICATION_VERSION},
			"os_name":         {"todo"},
			"os_version":      {"todo"},
			"os_architecture": {"todo"},
			"memory_total":    {fmt.Sprint(memstats.TotalAlloc)},
			"memory_max":      {"-1"}, // Go does not give a maximum amount of memory per program.
			"memory_free":     {fmt.Sprint(memstats.TotalAlloc - memstats.Alloc)},
			"java_version":    {"Go " + runtime.Version()},
			"cpu_cores":       {fmt.Sprint(runtime.NumCPU())},
			"players_current": {fmt.Sprint("")},
			"players_max":     {fmt.Sprint("")},
			"players_seen":    {fmt.Sprint("")}, // simply counts the number of files in world/players
			"uses_auth":       {"1"},
			"server_brand":    {"StuzzD"},
		})
	}
}
