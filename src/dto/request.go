package dto

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AltaTasaIntercambio struct {
    ActivoA int `json:"activo_a_id"`
    ActivoB int `json:"activo_b_id"`
    Tasa float64 `json:"tasa"`
    TasaInversa float64 `json:"tasa_inversa"`
	Config int `json:"config"`
}

type AltaCuenta struct {
	Deuda *bool `json:"permite_deuda,omitempty"`
	UsuarioID *int `json:"usuario_id,omitempty"`
	EmpresaID *int `json:"empresa_id,omitempty"`
	Nombre string  `json:"nombre"`
}

type AltaActivo struct {
	Nombre   string `json:"nombre"`
	UnidadID int    `json:"unidad_id"`
	AliasID *int    `json:"alias_id,omitempty"`
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

type VerResumenMovimientos struct {
	FechaInicio *time.Time `schema:"fecha_inicio,omitempty"`
	FechaFin *time.Time `schema:"fecha_fin,omitempty"`
	Activos []int `schema:"activos,omitempty"`
	Cuentas []int `schema:"cuentas,omitempty"`
	
	MontoMin *float64 `schema:"monto_min,omitempty"`
	MontoMax *float64 `schema:"monto_max,omitempty"`
	
	Limite int `schema:"limite"`
	Salto int `schema:"salto"`
}
 
type AltaProductoRequest struct {
    Nombre        string  `json:"nombre"`
    UnidadID      int     `json:"unidad_id"`
    EmpresaID     int     `json:"empresa_id"`
    ValorUnitario float64 `json:"valor_unitario"`
    UsuarioID     int     `json:"usuario_id"`
	AliasID       *int    `json:"alias_id"`
    
    TipoTransaccionID  int    `json:"tipo_transaccion_id"`
    Descripcion        string `json:"descripcion"`
    Cuenta int   `json:"cuenta"`
    CuentaContrapartida int   `json:"cuenta_contrapartida"`
    ActivoBaseID int `json:"activo_base_id"`
	Monto float64  `json:"monto"`
}

type ModificarProductoRequest struct {
    Nombre        string  `json:"nombre"`
    UnidadID      int     `json:"unidad_id"`
    ValorUnitario float64 `json:"valor_unitario"`
	AliasID       int    `json:"alias_id"`
    ActivoBaseID int `json:"activo_base_id"`
}
