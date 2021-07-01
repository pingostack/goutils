package goutils

import (
	"container/list"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/unrolled/secure"
)

type HttpListener interface {
	InitHttpRouter(http *HttpServer) error
}

type HttpOptions struct {
	ListenAddr    string
	ListenTlsAddr string
	CertFile      string
	CertKeyFile   string
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	ReleaseMode   bool
}

//Server  http服务
type ServerInfo struct {
	router *gin.Engine
	//	server   *http.Server
	groupMap map[string]*gin.RouterGroup
}

type HttpServer struct {
	http      *ServerInfo
	https     *ServerInfo
	event     chan string
	opt       HttpOptions
	listeners *list.List
}

func NewHttpServer(config HttpOptions) *HttpServer {
	hs := &HttpServer{
		event:     make(chan string),
		listeners: list.New(),
		opt:       config,
	}

	if config.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	defaultLogger.Debug("New http server by config %v", config)

	if config.ListenAddr != "" {
		//router := gin.New()

		// server := &http.Server{
		// 	Addr:         config.ListenAddr,
		// 	Handler:      router,
		// 	ReadTimeout:  time.Duration(config.ReadTimeout * time.Millisecond),
		// 	WriteTimeout: time.Duration(config.WriteTimeout * time.Millisecond),
		// }

		router := gin.Default()
		hs.http = &ServerInfo{
			router: router,
			//			server:   server,
			groupMap: make(map[string]*gin.RouterGroup),
		}
	}

	if config.ListenTlsAddr != "" {
		// router := gin.New()

		// server := &http.Server{
		// 	Addr:         config.ListenTlsAddr,
		// 	Handler:      router,
		// 	ReadTimeout:  time.Duration(config.ReadTimeout * time.Millisecond),
		// 	WriteTimeout: time.Duration(config.WriteTimeout * time.Millisecond),
		// }

		router := gin.Default()
		router.Use(TlsHandler(config.ListenTlsAddr))

		hs.https = &ServerInfo{
			router: router,
			//			server:   server,
			groupMap: make(map[string]*gin.RouterGroup),
		}
	}

	return hs
}

func TlsHandler(listenAddr string) gin.HandlerFunc {
	return func(c *gin.Context) {
		secureMiddleware := secure.New(secure.Options{
			SSLRedirect: true,
			SSLHost:     listenAddr,
		})
		err := secureMiddleware.Process(c.Writer, c.Request)

		// If there was an error, do not continue.
		if err != nil {
			return
		}

		c.Next()
	}
}

func (httpServer *HttpServer) AddListener(listener HttpListener) {
	httpServer.listeners.PushBack(listener)
}

func (httpServer *HttpServer) Start() error {

	for i := httpServer.listeners.Front(); i != nil; i = i.Next() {
		listener := i.Value.(HttpListener)
		listener.InitHttpRouter(httpServer)
	}

	if httpServer.http != nil {
		defaultLogger.Info("listen http server %s", httpServer.opt.ListenAddr)
		go httpServer.http.router.Run(httpServer.opt.ListenAddr)
	} else {
		defaultLogger.Debug("don't start http server")
	}

	if httpServer.https != nil {
		go func() {
			defaultLogger.Info("listen https server %s", httpServer.opt.ListenTlsAddr)
			err := httpServer.https.router.RunTLS(httpServer.opt.ListenTlsAddr, httpServer.opt.CertFile, httpServer.opt.CertKeyFile)
			if err != nil {
				defaultLogger.Error("Load tls file failed, %v", err)
				panic(err)
			}
		}()
	}

	// if httpServer.https != nil {
	// 	httpServer.https.router.RunTLS(httpServer.https.listenAddr)
	// }

	// go func() {
	// 	if httpServer.http != nil {
	// 		if err := httpServer.http.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
	// 			defaultLogger.Fatal("Http-Server start error: %s\n", err)
	// 			return
	// 		}
	// 	}

	// 	if httpServer.https != nil {
	// 		if err := httpServer.https.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
	// 			defaultLogger.Fatal("Https-Server start error: %s\n", err)
	// 			return
	// 		}
	// 	}

	// 	for event := range httpServer.event {
	// 		if event == "shutdown" {
	// 			if httpServer.http != nil {
	// 				httpServer.http.server.Shutdown(context.Background())
	// 			}
	// 			if httpServer.https != nil {
	// 				httpServer.https.server.Shutdown(context.Background())
	// 			}
	// 		}
	// 	}
	// }()

	return nil
}

func (httpServer *HttpServer) UseMiddleware(handler gin.HandlerFunc) {
	if httpServer.https != nil {
		httpServer.https.router.Use(handler)
	}

	if httpServer.http != nil {
		httpServer.http.router.Use(handler)
	}
}

func (httpServer *HttpServer) Shutdown() {
	httpServer.event <- "shutdown"
}

func (httpServer *HttpServer) findHttpGroup(si *ServerInfo, groupName string) *gin.RouterGroup {
	//	var group *gin.RouterGroup

	if si == nil {
		return nil
	}

	group, ok := si.groupMap[groupName]

	if !ok {
		group = si.router.Group(groupName)
		si.groupMap[groupName] = group
	}

	return group
}

func (httpServer *HttpServer) HttpGroup(groupName string) *gin.RouterGroup {
	return httpServer.findHttpGroup(httpServer.http, groupName)
}

func (httpServer *HttpServer) HttpsGroup(groupName string) *gin.RouterGroup {
	return httpServer.findHttpGroup(httpServer.https, groupName)
}

func (httpServer *HttpServer) HttpRouter() *gin.Engine {
	if httpServer.http != nil {
		return httpServer.http.router
	}

	return nil
}

func (httpServer *HttpServer) HttpsRouter() *gin.Engine {
	if httpServer.https != nil {
		return httpServer.https.router
	}

	return nil
}

func (httpServer *HttpServer) DefaultGroup() *gin.Engine {
	return gin.Default()
}
