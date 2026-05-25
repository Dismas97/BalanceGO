package dto
import "github.com/golang-jwt/jwt/v5"

type AltaCuenta struct {
	Deuda bool `json:"permite_deuda,omitempty"`
	UsuarioID int `json:"usuario_id,omitempty"`
	EmpresaID  int `json:"empresa_id,omitempty"`
	Nombre string  `json:"nombre"`
}

type AltaActivo struct {
	Nombre string `json:"nombre"`
	Unidad string `json:"unidad"` 
}

type Credenciales struct {
	SesionID int `json:"sid"`
    UsuarioID int `json:"uid"`
    EmpresaID int `json:"emp"`
    Roles []int `json:"roles"`
    Permisos []int `json:"perms"`
	Propietario bool `json:"prop"`
    jwt.RegisteredClaims
}

type RequestPaginado struct {
	Limite int `schema:"limite"`
	Salto int `schema:"salto"`
	Filtros string `schema:"filtros,omitempty"`
}
