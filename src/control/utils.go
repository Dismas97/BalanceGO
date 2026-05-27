package control

import (
	"encoding/json"
	"io"
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
