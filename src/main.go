package main

import (
	"log"
	"net/http"
	"os"
	"io"

	"github.com/gorilla/mux"
	gh "github.com/gorilla/handlers"
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
	
	/*err = bd.InitDatabase(bd.DB)
    if err != nil {
		log.Printf("Error: %v", err)
    }*/

    r := mux.NewRouter()
    // Rutas protegidas
    api := r.PathPrefix("/p").Subrouter()
	
	api.Use(middleware.AuthMiddleware)
	api.HandleFunc("/m/cuenta/{id:[0-9]+}", control.AbrirCerrarCuenta).Methods("PUT")
	api.HandleFunc("/a/cuenta", control.AltaCuenta).Methods("POST")
	api.HandleFunc("/a/transaccion", control.AltaTransaccion).Methods("POST")
	//    api.HandleFunc("/a/empresa/{id:[0-9]+}/usuario", handlers.AltaUsuario).Methods("POST")
	//    api.HandleFunc("/a/permiso", handlers.AltaPermiso).Methods("POST")
	//    api.HandleFunc("/a/empresa/{id:[0-9]+}/rol", handlers.AltaRol).Methods("POST")
	


	cors_conf := gh.CORS(gh.AllowedOrigins([]string{"*"}),gh.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),gh.AllowedHeaders([]string{"Content-Type", "Authorization"}), gh.AllowCredentials(),gh.MaxAge(86400),)


	handlerFinal := cors_conf(middleware.LoggingMiddleware(r))
	
    log.Printf("Servidor escuchando en :%s", config.MainConfig.ServerPort)
	log.Fatal(http.ListenAndServe(":"+config.MainConfig.ServerPort,handlerFinal))
}
