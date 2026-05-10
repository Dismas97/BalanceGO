package control

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"sistema-balance/bd"
	"sistema-balance/config"
	"sistema-balance/constantes"
	"sistema-balance/crypto"
	"sistema-balance/dto"
	"sistema-balance/response"
)
func AltaCuenta(w http.ResponseWriter, r *http.Request){
	sesion := r.Context().Value("sesion")
    claims, ok := sesion.(*crypto.Credenciales)
    if !ok {
        response.ResponseError(w, http.StatusBadRequest, constantes.CodSesionInvalida, constantes.MsjSesionInvalida)
        return
    }
	
	consistente := crypto.ValidarPermisoRoot(constantes.PermisoRootAltaCuenta,claims)
	
	if !consistente {
        response.ResponseError(w, http.StatusUnauthorized, constantes.CodNoAutorizado,constantes.MsjNoAutorizado)
        return
	}

	var req dto.AltaCuenta
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.ResponseError(w, http.StatusBadRequest, constantes.CodPeticionInvalida, constantes.MsjPeticionInvalida)
        return
	}
	
	err := bd.NewConnection(config.MainConfig);
	
	if err != nil {
        response.ResponseError(w, http.StatusInternalServerError, constantes.CodErrorInterno, constantes.MsjErrorInterno)
		log.Printf("Error: %v", err)
		return
	}
	ac := dto.Cuenta{
		Deuda: req.Deuda,
		UsuarioID: claims.UsuarioID,
		EmpresaID: claims.EmpresaID,
		Nombre: req.Nombre,
	}
	id, err := bd.AltaCuenta(ac,bd.DB)
	if err != nil {
        response.ResponseError(w, http.StatusInternalServerError, constantes.CodErrorInterno, constantes.MsjErrorInterno)
		log.Printf("Error: %v",err)
        return
    }

    response.ResponseSuccess(w, id, nil)
}

func AbrirCerrarCuenta(w http.ResponseWriter, r *http.Request) {
	sesion := r.Context().Value("sesion")
    claims, ok := sesion.(*crypto.Credenciales)
    if !ok {
        response.ResponseError(w, http.StatusBadRequest, constantes.CodSesionInvalida, constantes.MsjSesionInvalida)
        return
    }

    vars := mux.Vars(r)
    cuenta_str := vars["id"]
    cuenta_id, err := strconv.Atoi(cuenta_str)
    if err != nil {
        response.ResponseError(w, http.StatusBadRequest, constantes.CodPeticionInvalida, constantes.MsjPeticionInvalida)
        return
    }

	consistente := crypto.ValidarPermisoRoot(constantes.PermisoRootAbrirCerrarCuenta,claims)
	if !consistente {
        response.ResponseError(w, http.StatusUnauthorized, constantes.CodNoAutorizado,constantes.MsjNoAutorizado)
        return
	}

	err= bd.NewConnection(config.MainConfig);
	if err != nil {
        response.ResponseError(w, http.StatusInternalServerError, constantes.CodErrorInterno, constantes.MsjErrorInterno)
		log.Printf("Error: %v", err)
		return
	}
	hc, err := bd.UltimoHistorialCuenta(cuenta_id,bd.DB)
    if err != nil {
        response.ResponseError(w, http.StatusConflict, constantes.CodNoEncontrado, constantes.MsjNoEncontrado)
		log.Printf("Error: %v",err)
        return
    }
	if hc == nil {
		hc, err = bd.CerrarCuenta(cuenta_id,claims.UsuarioID,bd.DB)
	} else if hc.Estado == constantes.CuentaAbierta {
		hc, err = bd.CerrarCuenta(hc.CuentaID,claims.UsuarioID,bd.DB)
	} else {
		hc, err = bd.AbrirCuenta(hc.CuentaID,claims.UsuarioID,bd.DB)
	}
    if err != nil {
        response.ResponseError(w, http.StatusConflict, constantes.CodErrorInterno, constantes.MsjErrorInterno)
		log.Printf("Error: %v",err)
        return
    }
	
	res := dto.HistorialCuentaResponse {
		Reloj: hc.Reloj,
		Estado: hc.Estado,
		CuentaID: hc.CuentaID,
		UsuarioID: hc.UsuarioID,
	}
	response.ResponseSuccess(w, res, nil)
}
