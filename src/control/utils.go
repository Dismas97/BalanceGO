package control

import (
	"encoding/json"
	"io"
	"net/http"
	"sistema-balance/constantes"
	"sistema-balance/criptografia"
	"sistema-balance/response"
	"slices"
)

func credenciales(w http.ResponseWriter, r *http.Request) *criptografia.Credenciales {
	sesion := r.Context().Value("sesion")
	c, ok := sesion.(*criptografia.Credenciales)
	if !ok {
		response.ResponseError(w, http.StatusBadRequest, constantes.CodSesionInvalida, constantes.MsjSesionInvalida)
		return nil
	}
	return c
}

func validarPermisoRoot(w http.ResponseWriter, permiso_root constantes.PermisoID, c *criptografia.Credenciales) bool {
	if c == nil || c.EmpresaID != 1 || c.UsuarioID != 1 {
		return false
	}

	esroot := slices.Contains(c.Roles, 1)
	if !esroot {
		return false
	}
	if !criptografia.TienePermiso(c.Permisos, int(permiso_root)) {
        response.ResponseError(w, http.StatusUnauthorized, constantes.CodNoAutorizado,constantes.MsjNoAutorizado)
        return false
	}
	return true
}

func requestDTO(w http.ResponseWriter, body io.ReadCloser, req any) error{
	if err := json.NewDecoder(body).Decode(req); err != nil {
        response.ResponseError(w, http.StatusBadRequest, constantes.CodPeticionInvalida, constantes.MsjPeticionInvalida)
        return err
	}
	return nil
}
