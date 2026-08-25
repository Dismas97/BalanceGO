package control

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sistema-balance/constantes"
	"sistema-balance/dto"
	"sistema-balance/response"
	"slices"
	"strconv"
)

func credenciales(w http.ResponseWriter, r *http.Request) *dto.Credenciales {
	sesion := r.Context().Value("sesion")
	c, ok := sesion.(*dto.Credenciales)
	if !ok {
		response.ResponseError(w, http.StatusBadRequest, constantes.CodSesionInvalida, constantes.MsjSesionInvalida)
		return nil
	}
	return c
}

func ValidarPermisoRoot(permisoID constantes.PermisoID, c *dto.Credenciales) bool {
	if c == nil || c.EmpresaID != constantes.EmpresaRoot || c.UsuarioID != constantes.UsuarioRoot {
		return false
	}

	hasRol := slices.Contains(c.Roles, constantes.RolRoot)
	if !hasRol {
		return false
	}

	return slices.Contains(c.Permisos, int(permisoID))
}

func ValidarPermiso(permisoID constantes.PermisoID, c *dto.Credenciales) bool {
	return slices.Contains(c.Permisos, int(permisoID))
}

func EsPropietario(c *dto.Credenciales, empresaID int) bool {
    return c != nil && c.Propietario && c.EmpresaID == empresaID
}

func PuedeGestionarEmpresa(c *dto.Credenciales, empresaID int, permisoEmpresa, permisoRoot constantes.PermisoID) (bool, bool) {
    if ValidarPermisoRoot(permisoRoot,c) {
        return true, true
    }
	
    if c == nil || c.EmpresaID != empresaID {
        return false,false
    }
	
	if c.Propietario {
		log.Printf("SIII SOY PROPIETARIO")
        return true,false
    }
    return ValidarPermiso(permisoEmpresa, c), false
}


func requestDTO(w http.ResponseWriter, body io.ReadCloser, req any) error{
	if err := json.NewDecoder(body).Decode(req); err != nil {
        response.ResponseError(w, http.StatusBadRequest, constantes.CodPeticionInvalida, constantes.MsjPeticionInvalida)
        return err
	}
	return nil
}

func paginado(r *http.Request)(int,int){
	q := r.URL.Query()
	salto,_ := strconv.Atoi(q.Get("salto"))
	limite,_ := strconv.Atoi(q.Get("limite"))
	if salto < 0 {
		salto = 0
	}
	if limite <= 0 {
		limite = 10
	}
	return salto, limite
}

func ValorInt(v *int, def int) int {
    if v == nil {
        return def
    }
    return *v
}

func ValorBool(v *bool, def bool) bool {
    if v == nil {
        return def
    }
    return *v
}


func paginadoMeta(filas, paginas, salto, limite int) map[string]interface{} {
	return map[string]interface{}{
		"filas":   filas,
		"paginas": paginas,
		"salto":   salto,
		"limite":  limite,
	}
}

// ─── RESPONSE HELPERS ─────────────────────────────────────────────────────────
 
// respondePaginado es el cierre estándar de cualquier handler de lista:
// envía los datos con su metadata de paginación.
func respondePaginado(w http.ResponseWriter, data any, filas, paginas, salto, limite int) {
	response.ResponseSuccess(w, data, paginadoMeta(filas, paginas, salto, limite))
}
 
// respondeError registra el error en el log y escribe la respuesta HTTP.
func respondeError(w http.ResponseWriter, cod, status int, msj string, err error) {
	if err != nil {
		log.Printf("%s: %v", msj, err)
	}
	response.ResponseError(w, status, cod, msj)
}


