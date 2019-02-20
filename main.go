//go:generate fileb0x b0x.json
package main

import (
	"flag"
	"fmt"
	"github.com/GodlikePenguin/agogos-host/Applications"
	"github.com/GodlikePenguin/agogos-host/Containers"
	"github.com/GodlikePenguin/agogos-host/Datastore"
	"github.com/GodlikePenguin/agogos-host/Logger"
	"github.com/GodlikePenguin/agogos-host/Networks"
	"github.com/GodlikePenguin/agogos-host/Nodes"
	"github.com/GodlikePenguin/agogos-host/Overview"
	"github.com/GodlikePenguin/agogos-host/Storage"
	"github.com/GodlikePenguin/agogos-host/static"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"math"
	"net/http"
	"os"
	"runtime"
	"strings"
)

var getMethods = map[string]func(ctx *gin.Context){
	"/ping": func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	},
	"/overview":                    Overview.GetOverview,
	"/applications":                Applications.GetApplications,
	"/application/:name":           Applications.GetApplication,
	"/application/:name/:compName": Applications.GetApplicationComponentInformation,
	"/storage":                     Storage.ListStorage,
	"/networks":                    Networks.ListNetworks,
	"/nodes":                       Nodes.ListNodes,
	"/node/:name/stats":            Nodes.Stats,
}

var postMethods = map[string]func(ctx *gin.Context){
	"/applications":   Applications.CreateApplication,
	"/storage/:name":  Storage.CreateStorage,
	"/networks/:name": Networks.CreateNetwork,
	"/nodes/register": Nodes.Register,
}

var deleteMethods = map[string]func(ctx *gin.Context){
	"/applications/:name": Applications.DeleteApplication,
	"/all/applications":   Applications.DeleteAllApps,
	"/storage/:name":      Storage.DeleteStorage,
	"/all/storage":        Storage.DeleteAllStorage,
	"/networks/:name":     Networks.DeleteNetwork,
	"/all/networks":       Networks.DeleteAllNetworks,
	"/node/:name":         Nodes.DeleteNode,
	"/all/node/stats":     Nodes.DeleteAllStats,
	"/all/comp/stats":     Applications.DeleteAllComponentStats,
	"/everything":         Overview.Purge,
}

var putMethods = map[string]func(ctx *gin.Context){
	"/applications": Applications.UpdateApplication,
}

func main() {
	background := flag.Bool("background", false, "Whether the host is running as a background process")
	connectAddress := flag.String("connect", "", "Address of an existing Agogos primary to connect to (optional)")
	proxy := flag.Bool("proxy", false, "Whether to use the agogos proxy for routing container requests")
	advertiseAddr := flag.String("addr", "", "Address to use when other nodes connect to this node")
	flag.Parse()

	Logger.InitLogger(*background)
	agogosMode := "Primary"
	if *connectAddress != "" {
		agogosMode = "Secondary"
	}
	Logger.Printf("Starting Agogos in %s mode", agogosMode)
	//TODO Take from environment (Out of scope)
	containerRuntime := Containers.DockerRuntime{}
	Containers.SetupConfig(containerRuntime, *proxy)

	//Use multiple cores for efficiency
	runtime.GOMAXPROCS(int(math.Min(float64(runtime.NumCPU()), 4)))

	Datastore.SetupDatastore(agogosMode, *connectAddress, *advertiseAddr)
	Containers.StartProxies()
	r := setupRouter()
	for path, handler := range getMethods {
		r.GET(path, handler)
	}
	for path, handler := range postMethods {
		r.POST(path, handler)
	}
	for path, handler := range deleteMethods {
		r.DELETE(path, handler)
	}
	for path, handler := range putMethods {
		r.PUT(path, handler)
	}

	if agogosMode == "Secondary" {
		hostname, err := os.Hostname()
		if err != nil {
			panic("could not get hostname to register with primary: " + err.Error())
		}
		reqString := fmt.Sprintf(`{"name":"%s"}`, hostname)
		_, err = http.Post(fmt.Sprintf("http://%s:14440/nodes/register", *connectAddress),
			"application/json", strings.NewReader(reqString))
		if err != nil {
			Logger.ErrPrintf("Error sending register request to host %s: %s", *connectAddress, err.Error())
		}
	}
	Logger.Println("Starting front end")
	go http.ListenAndServe(":14441", http.FileServer(static.HTTP))
	Logger.Println("Starting reporter thread")
	Nodes.StartStatsReporter()
	Logger.Println("Ready to serve")
	log.Fatal(r.Run(":14440"))
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(Logger.Middleware())
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowMethods = append(corsConfig.AllowMethods, "DELETE")
	corsConfig.AllowAllOrigins = true
	r.Use(cors.New(corsConfig))
	return r
}
