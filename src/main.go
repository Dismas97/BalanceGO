package main

import (
	"log"
	"net/http"
	"os"
	"io"
	
	"crypto/tls"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	
	"sistema-balance/config"
	"sistema-balance/control"
	"sistema-balance/middleware"
	"sistema-balance/bd"
)

func main() {
	var err error
	
	logFile, err := os.OpenFile(config.MainConfig.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error abriendo archivo de log '%s': %v", config.MainConfig.LogPath, err)
	}
	defer logFile.Close()

	//log a consola
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Printf("Logging activado en: %s", config.MainConfig.LogPath)

	
	err = bd.NewConnection(config.MainConfig);
	if err != nil {
        log.Fatalf("Error conectando a bd: %v", err)
	}

    r := mux.NewRouter()
    // Rutas protegidas
    api := r.PathPrefix("/p").Subrouter()
	
	api.Use(middleware.AuthMiddleware)
	api.HandleFunc("/m/cuenta/{id:[0-9]+}", control.AbrirCerrarCuenta).Methods("PUT")
	api.HandleFunc("/a/empresa/{id:[0-9]+}/cuenta", control.AltaCuenta).Methods("POST")
	api.HandleFunc("/a/empresa/{id:[0-9]+}/activo", control.AltaActivo).Methods("POST")
	api.HandleFunc("/a/transaccion", control.AltaTransaccion).Methods("POST")
	api.HandleFunc("/v/cuenta/{id:[0-9]+}", control.VerCuenta).Methods("GET")
	api.HandleFunc("/v/cuenta/{id:[0-9]+}/transaccion", control.VerTransaccionesCuenta).Methods("GET")
	api.HandleFunc("/v/cuenta/{id:[0-9]+}/transaccion/detalle", control.VerTransaccionesCuentaDetalle).Methods("GET")
	api.HandleFunc("/v/transaccion/{id:[0-9]+}/movimiento", control.VerMovimientosTransaccion).Methods("GET")
	api.HandleFunc("/v/cuenta", control.VerCuentas).Methods("GET")
	api.HandleFunc("/v/activo", control.VerActivos).Methods("GET")
	api.HandleFunc("/v/empresa/{id:[0-9]+}/cuenta", control.VerCuentasEmpresa).Methods("GET")
	api.HandleFunc("/v/empresa/{id:[0-9]+}/cuenta/buscar", control.BuscarCuentas).Methods("GET")
	api.HandleFunc("/v/empresa/{id:[0-9]+}/activo/buscar", control.BuscarActivos).Methods("GET")
	api.HandleFunc("/v/empresa/{id:[0-9]+}/activo", control.VerActivosEmpresa).Methods("GET")
	api.HandleFunc("/v/empresa/{id:[0-9]+}/transaccion", control.VerTransaccionesEmpresa).Methods("GET")
	api.HandleFunc("/v/empresa/{id:[0-9]+}/transaccion/detalle", control.VerTransaccionesEmpresaDetalle).Methods("GET")
	api.HandleFunc("/v/transaccion",control.VerTransaccion).Methods("GET")
	api.HandleFunc("/v/unidad",control.VerUnidades).Methods("GET")
	api.HandleFunc("/v/unidad/buscar",control.BuscarUnidades).Methods("GET")


	handlerFinal := CORS(middleware.LoggingMiddleware(r))
	
	config_tls := &tls.Config{
        MinVersion: tls.VersionTLS12,
        PreferServerCipherSuites: true,
        CipherSuites: []uint16{
            tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
        },
    }

	srv := &http.Server{
        Addr:      ":" + config.MainConfig.ServerPort,
        Handler:   handlerFinal,
        TLSConfig: config_tls,
    }
	
    log.Printf("Servidor escuchando en :%s", config.MainConfig.ServerPort)
	log.Fatal(srv.ListenAndServeTLS("domain.cert","server.key"))
}

func CORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")
		
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }
        next.ServeHTTP(w, r)
    })
}
