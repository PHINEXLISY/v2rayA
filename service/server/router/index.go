package router

import (
	"crypto/md5"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/pkg/server/reqCache"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"github.com/v2rayA/v2rayA/server/controller"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

////go:embed web
//var webRoot embed.FS

// relativeFS implements fs.FS
type relativeFS struct {
	root        fs.FS
	relativeDir string
}

func (c relativeFS) Open(name string) (fs.File, error) {
	return c.root.Open(path.Join(c.relativeDir, name))
}

func cachedHTML(html []byte) func(ctx *gin.Context) {
	etag := fmt.Sprintf("W/%x", md5.Sum(html))
	h := string(html)
	return func(ctx *gin.Context) {
		if ctx.IsAborted() {
			return
		}
		ctx.Header("Content-Type", "text/html; charset=utf-8")
		ctx.Header("Cache-Control", "public, must-revalidate")
		ctx.Header("ETag", etag)
		if match := ctx.GetHeader("If-None-Match"); match != "" {
			if strings.Contains(match, etag) {
				ctx.Status(http.StatusNotModified)
				return
			}
		}
		ctx.String(http.StatusOK, h)
	}
}

//func ServeGUI(r *gin.Engine) {
//	webDir := conf.GetEnvironmentConfig().WebDir
//	if webDir == "" {
//		webFS, err := fs.Sub(webRoot, "web")
//		if err != nil {
//			log.Fatal("fs.Sub: %v", err)
//		}
//		ss := http.StripPrefix("/static", statigz.FileServer(webFS.(fs.ReadDirFS)))
//		r.GET("/static/*w", func(c *gin.Context) {
//			ss.ServeHTTP(c.Writer, c.Request)
//		})
//		f, err := webFS.Open("index.html")
//		if err != nil {
//			log.Fatal("webFS.Open index.html:", err)
//		}
//		defer f.Close()
//		html, err := io.ReadAll(f)
//		if err != nil {
//			log.Fatal("ReadAll index.html: %v", err)
//		}
//		r.GET("/", cachedHTML(html))
//	} else {
//		if _, err := os.Stat(webDir); os.IsNotExist(err) {
//			log.Warn("web files cannot be found at %v. web UI cannot be served", webDir)
//		} else {
//			filepath.Walk(webDir, func(path string, info os.FileInfo, err error) error {
//				if path == webDir {
//					return nil
//				}
//				if info.IsDir() {
//					r.Static("/static/"+info.Name(), path)
//					return filepath.SkipDir
//				}
//				r.StaticFile("/static/"+info.Name(), path)
//				return nil
//			})
//
//			f, err := os.Open(path.Join(webDir, "index.html"))
//			if err != nil {
//				log.Fatal("Open index.html: %v", err)
//			}
//			defer f.Close()
//			html, err := io.ReadAll(f)
//			if err != nil {
//				log.Fatal("ReadAll index.html: %v", err)
//			}
//			r.GET("/", cachedHTML(html))
//		}
//	}
//
//	app := conf.GetEnvironmentConfig()
//
//	ip, port, _ := net.SplitHostPort(app.Address)
//	addrs, err := net.InterfaceAddrs()
//	if net.ParseIP(ip).IsUnspecified() && err == nil {
//		for _, addr := range addrs {
//			if ipnet, ok := addr.(*net.IPNet); ok {
//				printRunningAt("http://" + net.JoinHostPort(ipnet.IP.String(), port))
//			}
//		}
//	} else {
//		printRunningAt("http://" + app.Address)
//	}
//}

func nocache(c *gin.Context) {
	c.Header("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
}

func Run() error {
	engine := gin.New()
	//ginpprof.Wrap(engine)
	engine.Use(gin.Recovery())
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{
		"GET", "POST", "DELETE", "PUT", "PATCH", "OPTIONS", "HEAD",
	}
	corsConfig.AllowWebSockets = true
	corsConfig.AllowCredentials = true
	corsConfig.AddAllowHeaders("Authorization", common.RequestIdHeader, "F-Token")
	engine.Use(cors.New(corsConfig))
	noAuth := engine.Group("api",
		nocache,
		reqCache.ReqCache,
	)
	{
		noAuth.GET("version", controller.GetVersion)
		noAuth.POST("login", controller.PostLogin)
		noAuth.POST("account", controller.PostAccount)

		noAuth.POST("import", controller.PostImport)
		noAuth.GET("touch", controller.GetTouch)
		noAuth.DELETE("touch", controller.DeleteTouch)
		noAuth.POST("connection", controller.PostConnection)
		noAuth.DELETE("connection", controller.DeleteConnection)
		noAuth.POST("v2ray", controller.PostV2ray)
		noAuth.DELETE("v2ray", controller.DeleteV2ray)
		noAuth.GET("pingLatency", controller.GetPingLatency)
		noAuth.GET("httpLatency", controller.GetHttpLatency)
		noAuth.GET("sharingAddress", controller.GetSharingAddress)
		noAuth.GET("remoteGFWListVersion", controller.GetRemoteGFWListVersion)
		noAuth.GET("setting", controller.GetSetting)
		noAuth.PUT("setting", controller.PutSetting)
		noAuth.PUT("gfwList", controller.PutGFWList)
		noAuth.PUT("subscription", controller.PutSubscription)
		noAuth.PATCH("subscription", controller.PatchSubscription)
		noAuth.GET("ports", controller.GetPorts)
		noAuth.PUT("ports", controller.PutPorts)
		//noAuth.PUT("account", controller.PutAccount)
		noAuth.GET("dnsList", controller.GetDnsList)
		noAuth.PUT("dnsList", controller.PutDnsList)
		noAuth.GET("routingA", controller.GetRoutingA)
		noAuth.PUT("routingA", controller.PutRoutingA)
		noAuth.GET("outbounds", controller.GetOutbounds)
		noAuth.POST("outbound", controller.PostOutbound)
		noAuth.DELETE("outbound", controller.DeleteOutbound)
		noAuth.GET("message", controller.WsMessage)
		noAuth.GET("logger", controller.GetLogger)
	}
	//auth := engine.Group("api",
	//	nocache,
	//	func(ctx *gin.Context) {
	//		if !configure.HasAnyAccounts() {
	//			common.Response(ctx, common.UNAUTHORIZED, gin.H{
	//				"first": true,
	//			})
	//			ctx.Abort()
	//			return
	//		}
	//	},
	//	jwt.JWTAuth(false),
	//	reqCache.ReqCache,
	//)
	//{
	//	auth.POST("import", controller.PostImport)
	//	auth.GET("touch", controller.GetTouch)
	//	auth.DELETE("touch", controller.DeleteTouch)
	//	auth.POST("connection", controller.PostConnection)
	//	auth.DELETE("connection", controller.DeleteConnection)
	//	auth.POST("v2ray", controller.PostV2ray)
	//	auth.DELETE("v2ray", controller.DeleteV2ray)
	//	auth.GET("pingLatency", controller.GetPingLatency)
	//	auth.GET("httpLatency", controller.GetHttpLatency)
	//	auth.GET("sharingAddress", controller.GetSharingAddress)
	//	auth.GET("remoteGFWListVersion", controller.GetRemoteGFWListVersion)
	//	auth.GET("setting", controller.GetSetting)
	//	auth.PUT("setting", controller.PutSetting)
	//	auth.PUT("gfwList", controller.PutGFWList)
	//	auth.PUT("subscription", controller.PutSubscription)
	//	auth.PATCH("subscription", controller.PatchSubscription)
	//	auth.GET("ports", controller.GetPorts)
	//	auth.PUT("ports", controller.PutPorts)
	//	//auth.PUT("account", controller.PutAccount)
	//	auth.GET("dnsList", controller.GetDnsList)
	//	auth.PUT("dnsList", controller.PutDnsList)
	//	auth.GET("routingA", controller.GetRoutingA)
	//	auth.PUT("routingA", controller.PutRoutingA)
	//	auth.GET("outbounds", controller.GetOutbounds)
	//	auth.POST("outbound", controller.PostOutbound)
	//	auth.DELETE("outbound", controller.DeleteOutbound)
	//	auth.GET("message", controller.WsMessage)
	//	auth.GET("logger", controller.GetLogger)
	//}

	//ServeGUI(engine)

	return engine.Run(conf.GetEnvironmentConfig().Address)
}

func printRunningAt(address string) {
	log.Alert("v2rayA is listening at %v", address)
}
