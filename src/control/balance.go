package control

import (
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"sistema-balance/bd"
	"sistema-balance/config"
	"sistema-balance/constantes"
	"sistema-balance/dto"
	"sistema-balance/response"
)

func AltaCuenta(w http.ResponseWriter, r *http.Request){
	claims := credenciales(w, r)
	if claims == nil {
		return
	}

	consistente := validarPermisoRoot(w,constantes.PermisoRootAltaCuenta,claims)
	if !consistente {
		return
	}
	
	var req dto.AltaCuenta
	err := requestDTO(w,r.Body,&req)
	if err != nil {
		log.Printf("Error: %v",err)
		return
	}
	
	err = bd.NewConnection(config.MainConfig);
	
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
	claims := credenciales(w, r)
	if claims == nil {
		return
	}

	consistente := validarPermisoRoot(w,constantes.PermisoRootAbrirCerrarCuenta,claims)
	if !consistente {
		return
	}

    vars := mux.Vars(r)
    cuenta_str := vars["id"]
    cuenta_id, err := strconv.Atoi(cuenta_str)
    if err != nil {
        response.ResponseError(w, http.StatusBadRequest, constantes.CodPeticionInvalida, constantes.MsjPeticionInvalida)
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

func AltaTransaccion(w http.ResponseWriter, r *http.Request) {
	claims := credenciales(w, r)
	if claims == nil {
		return
	}
	if !validarPermisoRoot(w,constantes.PermisoRootAltatransaccion,claims) {
		return
	}

	var req dto.Transaccion
	err := requestDTO(w,r.Body,&req)
	if err != nil {
		log.Printf("Error: %v",err)
		return
	} 
	if !validarTransaccion(w,r,req) {
        return
	}
	id, err := bd.AltaTransaccion(req, bd.DB)
	if err != nil{
        response.ResponseError(w, http.StatusInternalServerError, constantes.CodErrorInterno, constantes.MsjErrorInterno)
		log.Printf("Error: %v",err)
        return
	}
	
	response.ResponseSuccess(w, id, nil)
}

func VerCuenta(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    cuenta_str := vars["id"]
    cuenta_id, err := strconv.Atoi(cuenta_str)
    if err != nil {
        response.ResponseError(w, http.StatusBadRequest, constantes.CodPeticionInvalida, constantes.MsjPeticionInvalida)
        return
    }
	
	claims := credenciales(w, r)
	if claims == nil {
		return
	}
	if !validarPermisoRoot(w, constantes.PermisoRootVerCuenta, claims) {
		return
	}

	cuenta, err := bd.VerCuenta(cuenta_id,bd.DB)
	
	if err != nil {
		log.Printf("Error al obtener cuenta: %v", err)
		response.ResponseError(w, http.StatusInternalServerError, constantes.CodErrorInterno, constantes.MsjErrorInterno)
		return
	}
	if cuenta == nil {
		log.Printf("Error al obtener cuenta: %v", err)
		response.ResponseError(w, http.StatusNotFound, constantes.CodNoEncontrado, constantes.MsjNoEncontrado)
		return
	}
	response.ResponseSuccess(w, cuenta, nil)
}

func VerCuentas(w http.ResponseWriter, r *http.Request) {
	claims := credenciales(w, r)
	if claims == nil {
		return
	}
	if !validarPermisoRoot(w, constantes.PermisoRootVerCuenta, claims) {
		return
	}

	query := r.URL.Query()	
	salto, err := strconv.Atoi(query.Get("salto"))
	if err != nil || salto < 0 {
		salto = 0
	}
	limite, err := strconv.Atoi(query.Get("limite"))
	if err != nil || limite <= 0 {
		limite = 10
	}

	filas, paginas, pagina, err := bd.VerCuentasPaginado(salto, limite, bd.DB)
	if err != nil {
		log.Printf("Error al obtener cuentas: %v", err)
		response.ResponseError(w, http.StatusInternalServerError, constantes.CodErrorInterno, constantes.MsjErrorInterno)
		return
	}

	metadata := map[string]interface{}{
		"filas": filas,
		"paginas": paginas,
		"salto": salto,
		"limite": limite,
	}

	response.ResponseSuccess(w, pagina, metadata)
}

func AltaActivo(w http.ResponseWriter, r *http.Request) {
	claims := credenciales(w, r)
	if claims == nil {
		return
	}

	if !validarPermisoRoot(w, constantes.PermisoRootAltaActivo, claims) {
		return
	}

	var req dto.AltaActivo
	err := requestDTO(w, r.Body, &req)
	if err != nil {
		log.Printf("Error al parsear request: %v", err)
		return
	}

	err = bd.NewConnection(config.MainConfig)
	if err != nil {
		response.ResponseError(w, http.StatusInternalServerError, constantes.CodErrorInterno, constantes.MsjErrorInterno)
		log.Printf("Error de conexión: %v", err)
		return
	}

	activo := dto.Activo{
		Nombre: req.Nombre,
		Unidad: req.Unidad,
	}

	id, err := bd.AltaActivo(activo, bd.DB)
	if err != nil {
		response.ResponseError(w, http.StatusInternalServerError, constantes.CodErrorInterno, constantes.MsjErrorInterno)
		log.Printf("Error al dar de alta activo: %v", err)
		return
	}

	response.ResponseSuccess(w, id, nil)
}

func BajaActivo(w http.ResponseWriter, r *http.Request) {
	claims := credenciales(w, r)
	if claims == nil {
		return
	}

	if !validarPermisoRoot(w, constantes.PermisoRootBajaActivo, claims) {
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.ResponseError(w, http.StatusBadRequest, constantes.CodPeticionInvalida, constantes.MsjPeticionInvalida)
		return
	}

	destruir := r.URL.Query().Get("destruir") == "true"

	err = bd.NewConnection(config.MainConfig)
	if err != nil {
		response.ResponseError(w, http.StatusInternalServerError, constantes.CodErrorInterno, constantes.MsjErrorInterno)
		log.Printf("Error de conexión: %v", err)
		return
	}

	ok, err := bd.BajaActivo(id, destruir, bd.DB)
	if err != nil {
		response.ResponseError(w, http.StatusInternalServerError, constantes.CodErrorInterno, constantes.MsjErrorInterno)
		log.Printf("Error al dar de baja activo: %v", err)
		return
	}
	if !ok {
		response.ResponseError(w, http.StatusNotFound, constantes.CodNoEncontrado, constantes.MsjNoEncontrado)
		return
	}

	response.ResponseSuccess(w, map[string]interface{}{"eliminado": true}, nil)
}

func VerActivos(w http.ResponseWriter, r *http.Request) {
	claims := credenciales(w, r)
	if claims == nil {
		return
	}

	if !validarPermisoRoot(w, constantes.PermisoRootVerActivo, claims) {
		return
	}

	query := r.URL.Query()
	salto, err := strconv.Atoi(query.Get("salto"))
	if err != nil || salto < 0 {
		salto = 0
	}
	limite, err := strconv.Atoi(query.Get("limite"))
	if err != nil || limite <= 0 {
		limite = 10
	}

	err = bd.NewConnection(config.MainConfig)
	if err != nil {
		response.ResponseError(w, http.StatusInternalServerError, constantes.CodErrorInterno, constantes.MsjErrorInterno)
		log.Printf("Error de conexión: %v", err)
		return
	}

	totalFilas, totalPaginas, activos, err := bd.VerActivosPaginado(salto, limite, bd.DB)
	if err != nil {
		response.ResponseError(w, http.StatusInternalServerError, constantes.CodErrorInterno, constantes.MsjErrorInterno)
		log.Printf("Error al listar activos: %v", err)
		return
	}

	metadata := map[string]interface{}{
		"filas":    totalFilas,
		"paginas":  totalPaginas,
		"salto":    salto,
		"limite":   limite,
	}

	response.ResponseSuccess(w, activos, metadata)
}


func validarTransaccion(w http.ResponseWriter, r *http.Request, t dto.Transaccion) (bool) {
	if len(t.Movimientos) == 0 {
		log.Print("la transacción no tiene movimientos")
		response.ResponseError(w,http.StatusBadRequest,constantes.CodErrorConflicto, constantes.MsjTransaccionInvalida+", la transacción no tiene movimientos")
		return false
	}
	const epsilon = 1e-9
	suma := make(map[int]float64)
	cuentas := make(map[int]struct{})
	
	for _, mov := range t.Movimientos {
		suma[mov.ActivoID] += mov.Monto
		cuentas[mov.CuentaID] = struct{}{}
	}
	for activoID, suma := range suma {
		if math.Abs(suma) > epsilon {
			log.Printf("los movimientos del activo %d no balancean: suma=%f",activoID,suma)
			response.ResponseError(w,http.StatusBadRequest,constantes.CodErrorConflicto, constantes.MsjTransaccionInvalida+", los movimientos no balancean")
			return false
		}
	}
	cuentas_id := make([]int, 0, len(cuentas))
	for id := range cuentas {
		cuentas_id = append(cuentas_id, id)
	}
	activos_id := make([]int, 0, len(suma))
	for id := range suma {
		activos_id = append(activos_id, id)
	}

	res, err := bd.ActivosExistentes(activos_id, bd.DB)
	if err != nil{
		response.ResponseError(w,http.StatusInternalServerError,constantes.CodErrorInterno, constantes.MsjErrorInterno)
		return false
	}
	if !res{
		log.Printf("activos no existentes")
		response.ResponseError(w,http.StatusBadRequest,constantes.CodErrorConflicto, constantes.MsjTransaccionInvalida+", activos no existentes")
		return false
	}
	
	res, err = bd.CuentasAbiertas(cuentas_id, bd.DB)
	if err != nil{
		response.ResponseError(w,http.StatusInternalServerError,constantes.CodErrorInterno, constantes.MsjErrorInterno)
		return false
	}
	if !res{
		log.Printf("Cuentas no abiertas")
		response.ResponseError(w,http.StatusBadRequest,constantes.CodErrorConflicto, constantes.MsjTransaccionInvalida+", cuentas no abiertas")
		return false
	}
	return true
}
