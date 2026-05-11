package dto

import (
	"sistema-balance/constantes"
	"time"
)

type HistorialCuentaResponse struct {
    Reloj time.Time `json:"reloj"`
    Estado constantes.EstadoCuenta `json:"estado_final"`
    CuentaID int `json:"cuenta_id"`
    UsuarioID int `sjon:"usuario_id"`
}

/*
type UsuarioResponse struct {
    ID int `json:"id"`
    Usuario string  `json:"usuario"`
    Email *string `json:"email,omitempty"`
    Nombre *string `json:"nombre,omitempty"`
    Apellido *string `json:"apellido,omitempty"`
    Telefono *string `json:"telefono,omitempty"`
    Direccion *string `json:"direccion,omitempty"`
    EmpresaID int `json:"empresa_id,omitempty"`
    Creado time.Time `json:"creado"`
    UltMod time.Time `json:"ult_mod"`
    Estado string `json:"estado"`
    UltACc *time.Time `json:"uiltimo_acceso,omitempty"`
}

*/
