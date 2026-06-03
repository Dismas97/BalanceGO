package control

import (
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"sistema-balance/config"
	con "sistema-balance/constantes"
	"sistema-balance/bd"
	"sistema-balance/dto"
	"sistema-balance/response"
)

//CUENTAS

func AltaCuenta(w http.ResponseWriter, r *http.Request) {
	claims := credenciales(w, r)
	if claims == nil {
		return
	}
	vars := mux.Vars(r)
	empresa_str := vars["id"]
	empresa_id, err := strconv.Atoi(empresa_str)
	if err != nil {
		response.ResponseError(w, http.StatusBadRequest, con.CodPeticionInvalida, con.MsjPeticionInvalida)
		return
	}

	acceso, _ := PuedeGestionarEmpresa(claims, empresa_id,con.PermisoEmpresaAltaCuenta,con.PermisoRootAltaCuenta)
	if !acceso {
		response.ResponseError(w, http.StatusUnauthorized, con.CodNoAutorizado, con.MsjNoAutorizado)
		return
	}
	
	var req dto.AltaCuenta
	err = requestDTO(w,r.Body,&req)
	if err != nil {
		log.Printf("Error: %v",err)
		return
	}
	
	err = bd.NewConnection(config.MainConfig);
	
	if err != nil {
        response.ResponseError(w, http.StatusInternalServerError, con.CodErrorInterno, con.MsjErrorInterno)
		log.Printf("Error: %v", err)
		return
	}
	
	ac := dto.Cuenta{
		PermiteDeuda: req.Deuda,
		UsuarioID: claims.UsuarioID,
		EmpresaID: claims.EmpresaID,
		Nombre: req.Nombre,
	}
	
	id, err := bd.AltaCuenta(ac,bd.DB)
	if err != nil {
        response.ResponseError(w, http.StatusInternalServerError, con.CodErrorInterno, con.MsjErrorInterno)
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
	vars := mux.Vars(r)

	acceso, _ := PuedeGestionarEmpresa(claims, claims.EmpresaID,con.PermisoEmpresaAbrirCerrarCuenta,con.PermisoRootAbrirCerrarCuenta)
	if !acceso {
		response.ResponseError(w, http.StatusUnauthorized, con.CodNoAutorizado, con.MsjNoAutorizado)
		return
	}
	//Falta revisar que la cuenta sea de la misma empresa en bd..
    cuenta_str := vars["id"]
    cuenta_id, err := strconv.Atoi(cuenta_str)
    if err != nil {
        response.ResponseError(w, http.StatusBadRequest, con.CodPeticionInvalida, con.MsjPeticionInvalida)
        return
    }

	err= bd.NewConnection(config.MainConfig);
	if err != nil {
        response.ResponseError(w, http.StatusInternalServerError, con.CodErrorInterno, con.MsjErrorInterno)
		log.Printf("Error: %v", err)
		return
	}
	hc, err := bd.UltimoHistorialCuenta(cuenta_id,bd.DB)
    if err != nil {
        response.ResponseError(w, http.StatusConflict, con.CodNoEncontrado, con.MsjNoEncontrado)
		log.Printf("Error: %v",err)
        return
    }
	if hc == nil {
		hc, err = bd.CerrarCuenta(cuenta_id,claims.UsuarioID,bd.DB)
	} else if hc.Estado == con.CuentaAbierta {
		hc, err = bd.CerrarCuenta(hc.CuentaID,claims.UsuarioID,bd.DB)
	} else {
		hc, err = bd.AbrirCuenta(hc.CuentaID,claims.UsuarioID,bd.DB)
	}
    if err != nil {
        response.ResponseError(w, http.StatusConflict, con.CodErrorInterno, con.MsjErrorInterno)
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

func VerCuenta(w http.ResponseWriter, r *http.Request) {
	claims := credenciales(w, r)
	if claims == nil {
		return
	}

	acceso, _ := PuedeGestionarEmpresa(claims, claims.EmpresaID,con.PermisoEmpresaVerCuenta,con.PermisoRootVerCuenta)
	if !acceso {
		response.ResponseError(w, http.StatusUnauthorized, con.CodNoAutorizado, con.MsjNoAutorizado)
		return
	}
	
    vars := mux.Vars(r)
    cuenta_str := vars["id"]
    cuenta_id, err := strconv.Atoi(cuenta_str)
    if err != nil {
        response.ResponseError(w, http.StatusBadRequest, con.CodPeticionInvalida, con.MsjPeticionInvalida)
        return
    }
	
	cuenta, err := bd.VerCuenta(cuenta_id,bd.DB)
	
	if err != nil {
		log.Printf("Error al obtener cuenta: %v", err)
		response.ResponseError(w, http.StatusInternalServerError, con.CodErrorInterno, con.MsjErrorInterno)
		return
	}
	if cuenta == nil {
		log.Printf("Error al obtener cuenta: %v", err)
		response.ResponseError(w, http.StatusNotFound, con.CodNoEncontrado, con.MsjNoEncontrado)
		return
	}
	response.ResponseSuccess(w, cuenta, nil)
}

func VerCuentas(w http.ResponseWriter, r *http.Request) {
	claims := credenciales(w, r)
	if claims == nil {
		return
	}
	acceso := ValidarPermisoRoot(con.PermisoRootVerCuenta,claims)
	if !acceso {
		response.ResponseError(w, http.StatusUnauthorized, con.CodNoAutorizado, con.MsjNoAutorizado)
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
		response.ResponseError(w, http.StatusInternalServerError, con.CodErrorInterno, con.MsjErrorInterno)
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

func VerCuentasEmpresa(w http.ResponseWriter, r *http.Request) {
	claims := credenciales(w,r)
	if claims == nil {
		return
	}

	empresaID,_ := strconv.Atoi(mux.Vars(r)["id"])

	acceso,_ := PuedeGestionarEmpresa(claims,empresaID,con.PermisoEmpresaVerCuenta,con.PermisoRootVerCuenta)

	if !acceso {
		response.ResponseError(w,http.StatusUnauthorized,con.CodNoAutorizado,con.MsjNoAutorizado)
		return
	}
	
	salto,limite := paginado(r)

	filas,paginas,data,err := bd.VerCuentasEmpresa(empresaID,salto,limite,bd.DB)

	if err != nil {	response.ResponseError(w,http.StatusInternalServerError,con.CodErrorInterno,con.MsjErrorInterno)
		return
	}

	metadata := map[string]interface{}{
		"filas": filas,
		"paginas": paginas,
		"salto": salto,
		"limite": limite,
	}
	
	response.ResponseSuccess(w,data,metadata)
}

func VerTransaccionesCuenta(w http.ResponseWriter, r *http.Request) {
	claims := credenciales(w,r)
	if claims == nil {
		return
	}

	cuentaID,_ := strconv.Atoi(mux.Vars(r)["id"])

	acceso,_ := PuedeGestionarEmpresa(claims,claims.EmpresaID,con.PermisoEmpresaVerCuenta,con.PermisoRootVerCuenta)

	if !acceso {
		response.ResponseError(w,http.StatusUnauthorized,con.CodNoAutorizado,con.MsjNoAutorizado)
		return
	}
	
	salto,limite := paginado(r)

	filas,paginas,data,err := bd.VerTransaccionesCuenta(cuentaID,salto,limite,bd.DB)

	if err != nil {	response.ResponseError(w,http.StatusInternalServerError,con.CodErrorInterno,con.MsjErrorInterno)
		return
	}

	metadata := map[string]interface{}{
		"filas": filas,
		"paginas": paginas,
		"salto": salto,
		"limite": limite,
	}
	
	response.ResponseSuccess(w,data,metadata)
}

//ACTIVOS
func AltaActivo(w http.ResponseWriter, r *http.Request) {
	claims := credenciales(w, r)
	if claims == nil {
		return
	}
	
	vars := mux.Vars(r)
	empresa_str := vars["id"]
	empresa_id, err := strconv.Atoi(empresa_str)
	if err != nil {
		response.ResponseError(w, http.StatusBadRequest, con.CodPeticionInvalida, con.MsjPeticionInvalida)
		return
	}

	acceso, _ := PuedeGestionarEmpresa(claims,empresa_id,con.PermisoEmpresaAltaActivo,con.PermisoRootAltaActivo)
	if !acceso {
		response.ResponseError(w, http.StatusUnauthorized, con.CodNoAutorizado, con.MsjNoAutorizado)
		return
	}

	var req dto.AltaActivo
	err = requestDTO(w, r.Body, &req)
	if err != nil {
		log.Printf("Error al parsear request: %v", err)
		return
	}

	err = bd.NewConnection(config.MainConfig)
	if err != nil {
		response.ResponseError(w, http.StatusInternalServerError, con.CodErrorInterno, con.MsjErrorInterno)
		log.Printf("Error de conexión: %v", err)
		return
	}
	
	activo := dto.Activo{
		Nombre: req.Nombre,
		UnidadID: req.UnidadID,
		EmpresaID: empresa_id,
	}

	id, err := bd.AltaActivo(activo, bd.DB)
	if err != nil {
		response.ResponseError(w, http.StatusInternalServerError, con.CodErrorInterno, con.MsjErrorInterno)
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
	
	vars := mux.Vars(r)

	acceso, _ := PuedeGestionarEmpresa(claims, claims.EmpresaID,con.PermisoEmpresaBajaActivo,con.PermisoRootBajaActivo)
	if !acceso {
		response.ResponseError(w, http.StatusUnauthorized, con.CodNoAutorizado, con.MsjNoAutorizado)
		return
	}
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.ResponseError(w, http.StatusBadRequest, con.CodPeticionInvalida, con.MsjPeticionInvalida)
		return
	}

	err = bd.NewConnection(config.MainConfig)
	if err != nil {
		response.ResponseError(w, http.StatusInternalServerError, con.CodErrorInterno, con.MsjErrorInterno)
		log.Printf("Error de conexión: %v", err)
		return
	}

	ok, err := bd.BajaActivo(id, false, bd.DB)
	if err != nil {
		response.ResponseError(w, http.StatusInternalServerError, con.CodErrorInterno, con.MsjErrorInterno)
		log.Printf("Error al dar de baja activo: %v", err)
		return
	}
	if !ok {
		response.ResponseError(w, http.StatusNotFound, con.CodNoEncontrado, con.MsjNoEncontrado)
		return
	}

	response.ResponseSuccess(w, map[string]interface{}{"eliminado": true}, nil)
}

func VerActivos(w http.ResponseWriter, r *http.Request) {
	claims := credenciales(w, r)
	if claims == nil {
		return
	}
	acceso := ValidarPermisoRoot(con.PermisoRootVerActivo,claims)
	if !acceso {
		response.ResponseError(w, http.StatusUnauthorized, con.CodNoAutorizado, con.MsjNoAutorizado)
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
		response.ResponseError(w, http.StatusInternalServerError, con.CodErrorInterno, con.MsjErrorInterno)
		log.Printf("Error de conexión: %v", err)
		return
	}

	totalFilas, totalPaginas, activos, err := bd.VerActivosPaginado(salto, limite, bd.DB)
	if err != nil {
		response.ResponseError(w, http.StatusInternalServerError, con.CodErrorInterno, con.MsjErrorInterno)
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

func VerActivosEmpresa(w http.ResponseWriter, r *http.Request) {
	claims := credenciales(w,r)
	if claims == nil {
		return
	}

	empresaID,_ := strconv.Atoi(mux.Vars(r)["id"])

	acceso,_ := PuedeGestionarEmpresa(claims,empresaID,con.PermisoEmpresaVerActivo,con.PermisoRootVerActivo)

	if !acceso {
		response.ResponseError(w,http.StatusUnauthorized,con.CodNoAutorizado,con.MsjNoAutorizado)
		return
	}

	salto,limite := paginado(r)

	filas,paginas,data,err := bd.VerActivosEmpresa(empresaID,salto,	limite,	bd.DB)

	if err != nil {	response.ResponseError(w,http.StatusInternalServerError,con.CodErrorInterno,con.MsjErrorInterno)
		return
	}

	metadata := map[string]interface{}{
		"filas": filas,
		"paginas": paginas,
		"salto": salto,
		"limite": limite,
	}
	
	response.ResponseSuccess(w,data,metadata)
}

//TRANSACCIONES
func AltaTransaccion(w http.ResponseWriter, r *http.Request) {
	claims := credenciales(w, r)
	if claims == nil {
		return
	}

	acceso, _ := PuedeGestionarEmpresa(claims, claims.EmpresaID,con.PermisoEmpresaAltaTransaccion,con.PermisoRootAltaTransaccion)
	if !acceso {
		response.ResponseError(w, http.StatusUnauthorized, con.CodNoAutorizado, con.MsjNoAutorizado)
		return
	}

	var req dto.AltaTransaccion
	if err := requestDTO(w,r.Body,&req); err != nil {
		log.Printf("Error: %v",err)
		return
	}
	tran := dto.Transaccion {
		UsuarioID: claims.UsuarioID,
			TipoTransaccionID: req.TipoTransaccionID,
			EmpresaID: req.EmpresaID,
			Descripcion: req.Descripcion,
			Movimientos: req.Movimientos,
		}
	if !validarTransaccion(w,r,tran) {
        return
	}
	id, err := bd.AltaTransaccion(tran, bd.DB)
	if err != nil{
        response.ResponseError(w, http.StatusInternalServerError, con.CodErrorInterno, con.MsjErrorInterno)
		log.Printf("Error: %v",err)
        return
	}	
	response.ResponseSuccess(w, id, nil)
}

func VerTransaccionesEmpresa(w http.ResponseWriter, r *http.Request) {
	claims := credenciales(w,r)
	if claims == nil {
		return
	}

	empresaID,_ := strconv.Atoi(mux.Vars(r)["id"])

	acceso,_ := PuedeGestionarEmpresa(claims,empresaID,con.PermisoEmpresaVerTransaccion,con.PermisoRootVerTransaccion)

	if !acceso {
		response.ResponseError(w,http.StatusUnauthorized,con.CodNoAutorizado,con.MsjNoAutorizado)
		return
	}

	salto,limite := paginado(r)

	filas,paginas,data,err := bd.VerTransaccionesEmpresa(empresaID,salto,limite,bd.DB)

	if err != nil {	response.ResponseError(w,http.StatusInternalServerError,con.CodErrorInterno,con.MsjErrorInterno)
		return
	}

	metadata := map[string]interface{}{
		"filas": filas,
		"paginas": paginas,
		"salto": salto,
		"limite": limite,
	}
	
	response.ResponseSuccess(w,data,metadata)
}

func VerMovimientosTransaccion(w http.ResponseWriter, r *http.Request) {
	claims := credenciales(w,r)
	if claims == nil {
		return
	}

	transaccionID,_ := strconv.Atoi(mux.Vars(r)["id"])

	acceso,_ := PuedeGestionarEmpresa(claims,transaccionID,con.PermisoEmpresaVerTransaccion,con.PermisoRootVerTransaccion)

	if !acceso {
		response.ResponseError(w,http.StatusUnauthorized,con.CodNoAutorizado,con.MsjNoAutorizado)
		return
	}
	
	salto,limite := paginado(r)

	filas,paginas,data,err := bd.VerMovimientosTransaccion(transaccionID,salto,limite,bd.DB)

	if err != nil {	response.ResponseError(w,http.StatusInternalServerError,con.CodErrorInterno,con.MsjErrorInterno)
		return
	}

	metadata := map[string]interface{}{
		"filas": filas,
		"paginas": paginas,
		"salto": salto,
		"limite": limite,
	}
	
	response.ResponseSuccess(w,data,metadata)
}

func validarTransaccion(w http.ResponseWriter, r *http.Request, t dto.Transaccion) (bool) {
	if len(t.Movimientos) == 0 {
		log.Print("la transacción no tiene movimientos")
		response.ResponseError(w,http.StatusBadRequest,con.CodErrorConflicto, con.MsjTransaccionInvalida+", la transacción no tiene movimientos")
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
			response.ResponseError(w,http.StatusBadRequest,con.CodErrorConflicto, con.MsjTransaccionInvalida+", los movimientos no balancean")
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

	ok, err := bd.CuentasPertenecenEmpresa(cuentas_id,t.EmpresaID,bd.DB)
	if err != nil || !ok {
		response.ResponseError(w,http.StatusBadRequest,con.CodErrorConflicto,"Cuentas Invalidas",)
		return false
	}

	ok, err = bd.ActivosPertenecenEmpresa(activos_id,t.EmpresaID,bd.DB)
	if err != nil || !ok {
		response.ResponseError(w,http.StatusBadRequest,con.CodErrorConflicto,"Activos Invalidos")
		return false
	}

	res, err := bd.ActivosExistentes(activos_id, bd.DB)
	if err != nil{
		response.ResponseError(w,http.StatusInternalServerError,con.CodErrorInterno, con.MsjErrorInterno)
		return false
	}
	if !res{
		log.Printf("Activos no existentes")
		response.ResponseError(w,http.StatusBadRequest,con.CodErrorConflicto, con.MsjTransaccionInvalida+", activos no existentes")
		return false
	}
	
	res, err = bd.CuentasAbiertas(cuentas_id, bd.DB)
	if err != nil{
		response.ResponseError(w,http.StatusInternalServerError,con.CodErrorInterno, con.MsjErrorInterno)
		return false
	}
	if !res{
		log.Printf("Cuentas no abiertas")
		response.ResponseError(w,http.StatusBadRequest,con.CodErrorConflicto, con.MsjTransaccionInvalida+", cuentas no abiertas")
		return false
	}
	return true
}

func VerUnidades(w http.ResponseWriter, r *http.Request) {
	claims := credenciales(w, r)
	if claims == nil {
		return
	}

	salto, limite := paginado(r)

	filas, paginas, unidades, err := bd.VerUnidades(salto,limite,bd.DB)

	if err != nil {	response.ResponseError(w,http.StatusInternalServerError,con.CodErrorInterno,con.MsjErrorInterno)
		return
	}
	metadata := map[string]interface{}{
		"filas": filas,
		"paginas": paginas,
		"salto": salto,
		"limite": limite,
	}

	response.ResponseSuccess(w, unidades,metadata)
}

















































// ─── BÚSQUEDA POR NOMBRE ──────────────────────────────────────────────────────
 
// GET /empresa/{id}/cuentas/buscar?nombre=&salto=&limite=
// Busca cuentas de la empresa por nombre (ILIKE). empresa_id viene de {id} en la ruta;
// el token confirma que el usuario tiene acceso a esa empresa.
func BuscarCuentas(w http.ResponseWriter, r *http.Request) {
	claims := credenciales(w, r)
	if claims == nil {
		return
	}
 
	empresaID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		response.ResponseError(w, http.StatusBadRequest, con.CodPeticionInvalida, con.MsjPeticionInvalida)
		return
	}
 
	acceso, _ := PuedeGestionarEmpresa(claims, empresaID, con.PermisoEmpresaVerCuenta, con.PermisoRootVerCuenta)
	if !acceso {
		response.ResponseError(w, http.StatusUnauthorized, con.CodNoAutorizado, con.MsjNoAutorizado)
		return
	}
 
	nombre := r.URL.Query().Get("nombre")
	salto, limite := paginado(r)
 
	filas, paginas, data, err := bd.BuscarCuentas(nombre, empresaID, salto, limite, bd.DB)
	if err != nil {
		log.Printf("BuscarCuentas: %v", err)
		response.ResponseError(w, http.StatusInternalServerError, con.CodErrorInterno, con.MsjErrorInterno)
		return
	}
 
	response.ResponseSuccess(w, data, map[string]interface{}{
		"filas": filas, "paginas": paginas, "salto": salto, "limite": limite,
	})
}
 
// GET /empresa/{id}/activos/buscar?nombre=&salto=&limite=
// Busca activos de la empresa por nombre (ILIKE).
func BuscarActivos(w http.ResponseWriter, r *http.Request) {
	claims := credenciales(w, r)
	if claims == nil {
		return
	}
 
	empresaID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		response.ResponseError(w, http.StatusBadRequest, con.CodPeticionInvalida, con.MsjPeticionInvalida)
		return
	}
 
	acceso, _ := PuedeGestionarEmpresa(claims, empresaID, con.PermisoEmpresaVerActivo, con.PermisoRootVerActivo)
	if !acceso {
		response.ResponseError(w, http.StatusUnauthorized, con.CodNoAutorizado, con.MsjNoAutorizado)
		return
	}
 
	nombre := r.URL.Query().Get("nombre")
	salto, limite := paginado(r)
 
	filas, paginas, data, err := bd.BuscarActivos(nombre, empresaID, salto, limite, bd.DB)
	if err != nil {
		log.Printf("BuscarActivos: %v", err)
		response.ResponseError(w, http.StatusInternalServerError, con.CodErrorInterno, con.MsjErrorInterno)
		return
	}
 
	response.ResponseSuccess(w, data, map[string]interface{}{
		"filas": filas, "paginas": paginas, "salto": salto, "limite": limite,
	})
}
 
// GET /unidades/buscar?nombre=&salto=&limite=
// Busca unidades por nombre o símbolo (ILIKE). No requiere empresa;
// cualquier usuario autenticado puede consultar este recurso compartido.
func BuscarUnidades(w http.ResponseWriter, r *http.Request) {
	claims := credenciales(w, r)
	if claims == nil {
		return
	}
	_ = claims // autenticación es suficiente para este recurso
 
	nombre := r.URL.Query().Get("nombre")
	salto, limite := paginado(r)
 
	filas, paginas, data, err := bd.BuscarUnidades(nombre, salto, limite, bd.DB)
	if err != nil {
		log.Printf("BuscarUnidades: %v", err)
		response.ResponseError(w, http.StatusInternalServerError, con.CodErrorInterno, con.MsjErrorInterno)
		return
	}
 
	response.ResponseSuccess(w, data, map[string]interface{}{
		"filas": filas, "paginas": paginas, "salto": salto, "limite": limite,
	})
}
 
// ─── RESPUESTAS ANIDADAS ──────────────────────────────────────────────────────
 
// GET /transaccion/{id}
// Devuelve una transacción con sus movimientos anidados.
// Valida que la transacción pertenezca a la empresa del token.
func VerTransaccion(w http.ResponseWriter, r *http.Request) {
	claims := credenciales(w, r)
	if claims == nil {
		return
	}
 
	acceso, _ := PuedeGestionarEmpresa(claims, claims.EmpresaID, con.PermisoEmpresaVerTransaccion, con.PermisoRootVerTransaccion)
	if !acceso {
		response.ResponseError(w, http.StatusUnauthorized, con.CodNoAutorizado, con.MsjNoAutorizado)
		return
	}
 
	transaccionID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		response.ResponseError(w, http.StatusBadRequest, con.CodPeticionInvalida, con.MsjPeticionInvalida)
		return
	}
 
	t, err := bd.VerTransaccionDetalle(transaccionID, claims.EmpresaID, bd.DB)
	if err != nil {
		log.Printf("VerTransaccion: %v", err)
		response.ResponseError(w, http.StatusInternalServerError, con.CodErrorInterno, con.MsjErrorInterno)
		return
	}
	if t == nil {
		response.ResponseError(w, http.StatusNotFound, con.CodNoEncontrado, con.MsjNoEncontrado)
		return
	}
 
	response.ResponseSuccess(w, t, nil)
}
 
// GET /empresa/{id}/transacciones/detalle?salto=&limite=
// Igual que VerTransaccionesEmpresa pero con los movimientos anidados en cada transacción.
func VerTransaccionesEmpresaDetalle(w http.ResponseWriter, r *http.Request) {
	claims := credenciales(w, r)
	if claims == nil {
		return
	}
 
	empresaID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		response.ResponseError(w, http.StatusBadRequest, con.CodPeticionInvalida, con.MsjPeticionInvalida)
		return
	}
 
	acceso, _ := PuedeGestionarEmpresa(claims, empresaID, con.PermisoEmpresaVerTransaccion, con.PermisoRootVerTransaccion)
	if !acceso {
		response.ResponseError(w, http.StatusUnauthorized, con.CodNoAutorizado, con.MsjNoAutorizado)
		return
	}
 
	salto, limite := paginado(r)
 
	filas, paginas, data, err := bd.VerTransaccionesEmpresaDetalle(empresaID, salto, limite, bd.DB)
	if err != nil {
		log.Printf("VerTransaccionesEmpresaDetalle: %v", err)
		response.ResponseError(w, http.StatusInternalServerError, con.CodErrorInterno, con.MsjErrorInterno)
		return
	}
 
	response.ResponseSuccess(w, data, map[string]interface{}{
		"filas": filas, "paginas": paginas, "salto": salto, "limite": limite,
	})
}
 
// GET /cuenta/{id}/transacciones/detalle?salto=&limite=
// Igual que VerTransaccionesCuenta pero con los movimientos anidados en cada transacción.
// La BD valida que la cuenta pertenezca a la empresa del token para evitar
// que un usuario consulte cuentas de otra empresa con solo adivinar un ID.
func VerTransaccionesCuentaDetalle(w http.ResponseWriter, r *http.Request) {
	claims := credenciales(w, r)
	if claims == nil {
		return
	}
 
	acceso, _ := PuedeGestionarEmpresa(claims, claims.EmpresaID, con.PermisoEmpresaVerCuenta, con.PermisoRootVerCuenta)
	if !acceso {
		response.ResponseError(w, http.StatusUnauthorized, con.CodNoAutorizado, con.MsjNoAutorizado)
		return
	}
 
	cuentaID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		response.ResponseError(w, http.StatusBadRequest, con.CodPeticionInvalida, con.MsjPeticionInvalida)
		return
	}
 
	salto, limite := paginado(r)
 
	filas, paginas, data, err := bd.VerTransaccionesCuentaDetalle(cuentaID, claims.EmpresaID, salto, limite, bd.DB)
	if err != nil {
		log.Printf("VerTransaccionesCuentaDetalle: %v", err)
		response.ResponseError(w, http.StatusInternalServerError, con.CodErrorInterno, con.MsjErrorInterno)
		return
	}
 
	response.ResponseSuccess(w, data, map[string]interface{}{
		"filas": filas, "paginas": paginas, "salto": salto, "limite": limite,
	})
}
