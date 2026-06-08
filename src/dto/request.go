package dto

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AltaCuenta struct {
	Deuda bool `json:"permite_deuda,omitempty"`
	UsuarioID int `json:"usuario_id,omitempty"`
	EmpresaID  int `json:"empresa_id,omitempty"`
	Nombre string  `json:"nombre"`
}

type AltaActivo struct {
	Nombre string `json:"nombre"`
	UnidadID int    `json:"unidad_id"`
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
	Busqueda string `schema:"busqueda,omitempty"`
}


type AltaTransaccion struct {
	Creado time.Time `json:"creado"`
	EstadoTransaccion string `json:"estado_transaccion"`
	TipoTransaccionID int    `json:"tipo_transaccion_id"`
	EmpresaID         int `json:"empresa_id"`
	Descripcion string `json:"descripcion"`
	Movimientos []Movimiento  `json:"movimientos"`
}
