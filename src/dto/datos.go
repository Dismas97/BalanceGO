package dto

import (
	"database/sql"
	"sistema-balance/constantes"
	"time"
)

type BalanceCuentas struct {
	ActivoID int `db:"activo_id" json:"activo_id"`
	CuentaID int `db:"cuenta_id" json:"cuenta_id"`
	Nombre string `db:"nombre" json:"nombre"`
	Cuenta string `db:"cuenta" json:"cuenta"`
	Total string `db:"total" json:"total"`
}

type Unidad struct {
	ID            int       `db:"id" json:"id"`
	Creado        time.Time `db:"creado" json:"creado"`
	UltMod        time.Time `db:"ult_mod" json:"ult_mod"`
	Estado        string    `db:"estado" json:"estado"`
	Nombre        string `db:"nombre" json:"nombre"`
	Simbolo       string `db:"simbolo" json:"simbolo"`
	TipoUnidadID  int    `db:"tipo_unidad_id" json:"tipo_unidad_id"`
	NombreTipo  string `db:"nombre_tipo" json:"nombre_tipo"`
}

type Activo struct {
	ID        int       `db:"id" json:"id"`
	Creado    time.Time `db:"creado" json:"creado"`
	UltMod    time.Time `db:"ult_mod" json:"ult_mod"`
	Estado    string    `db:"estado" json:"estado"`

	AliasID  *int    `db:"alias_id" json:"alias_id"`
	Nombre    string `db:"nombre" json:"nombre"`
	UnidadID  int    `db:"unidad_id" json:"unidad_id"`
	UnidadNombre string  `db:"unidad_nombre" json:"unidad_nombre"`
	EmpresaID int    `db:"empresa_id" json:"empresa_id"`
	UnidadSimbolo string  `db:"unidad_simbolo" json:"unidad_simbolo"`
	
	Tasas []TasaIntercambio
}

type TasaIntercambio struct {
    ID int `db:"id" json:"id"`
    Creado time.Time `db:"creado" json:"creado"`
    UltMod time.Time `db:"ult_mod" json:"ult_mod"`
    Estado string `db:"estado" json:"estado"`

    ActivoA int `db:"activo_a_id" json:"activo_a_id"`
    ActivoB int `db:"activo_b_id" json:"activo_b_id"`
    Empresa int `db:"empresa_id" json:"empresa_id"`
    Tasa float64 `db:"tasa" json:"tasa"`
    TasaInversa float64 `db:"tasa_inversa" json:"tasa_inversa"`
	Config int `db:"config" json:"config"`
}

type HistorialConversion struct {
    Reloj time.Time `db:"reloj" json:"reloj"`
    RecAnterior float64 `db:"conversion_anterior" json:"conversion_anterior"`
    ConversionID int `db:"conversion_id" json:"conversion_id"`
}

type Cuenta struct {
	ID        int       `db:"id" json:"id"`
	Creado    time.Time `db:"creado" json:"creado"`
	UltMod    time.Time `db:"ult_mod" json:"ult_mod"`
	Estado    string    `db:"estado" json:"estado"`

	PermiteDeuda bool `db:"permite_deuda" json:"permite_deuda"`
	UsuarioID    int  `db:"usuario_id" json:"usuario_id"`
	EmpresaID    int  `db:"empresa_id" json:"empresa_id"`

	Nombre string `db:"nombre" json:"nombre"`
	EstadoFinal string `db:"estado_final" json:"estado_final"`

	Monto []MontoCuenta
}

type MontoCuenta struct {
    Creado time.Time `db:"creado" json:"creado"`
    UltMod time.Time `db:"ult_mod" json:"ult_mod"`
	
    CuentaID int `db:"cuenta_id" json:"cuenta_id"`
    ActivoID int `db:"activo_id" json:"activo_id"`
	ActivoNombre string `db:"nombre" json:"nombre"`
	UnidadSimbolo string `db:"simbolo" json:"simbolo"`
    Monto float64 `db:"monto" json:"monto"`
}

type Transaccion struct {
	ID        int       `db:"id" json:"id"`
	Creado    time.Time `db:"creado" json:"creado"`
	Estado    string    `db:"estado" json:"estado"`
	EstadoTransaccion string `db:"estado_transaccion" json:"estado_transaccion"`
	TipoTransaccionID int    `db:"tipo_transaccion_id" json:"tipo_transaccion_id"`
	EmpresaID         int    `db:"empresa_id" json:"empresa_id"`
	UsuarioID         int    `db:"usuario_id" json:"usuario_id"`
	Descripcion string `db:"descripcion" json:"descripcion"`
	Movimientos []Movimiento
}

type Movimiento struct {
    ID int `db:"id" json:"id"`
    TransaccionID int `db:"transaccion_id" json:"transaccion_id"`
    CuentaNombre string `db:"cuenta_nombre" json:"cuenta_nombre"`
    CuentaID int `db:"cuenta_id" json:"cuenta_id"`
    ActivoID int `db:"activo_id" json:"activo_id"`
    ActivoNombre string `db:"activo_nombre" json:"activo_nombre"`
    Monto float64 `db:"monto" json:"monto"`
}

type HistorialCuenta struct {
	Id int `db:"id" json:"id"`
    Reloj time.Time `db:"reloj" json:"reloj"`
    Estado constantes.EstadoCuenta `db:"estado_final" json:"estado_final"`
	
    CuentaID int `db:"cuenta_id" json:"cuenta_id"`
    UsuarioID int `db:"usuario_id" json:"usuario_id"`
}

type ModificarActivo struct {
	ID int `db:"id" json:"id"`
	AliasID *int `db:"alias_id"`
	Nombre string `db:"nombre"`
	UnidadID int `db:"unidad_id"`
	UnidadNombre string `db:"unidad_nombre"`
	EmpresaID int `db:"empresa_id"`
	UnidadSimbolo string `db:"unidad_simbolo"`
}
type ResumenMov struct {
	ID            int             `db:"id"`
	TransaccionID int             `db:"transaccion_id"`

	CuentaID     int    `db:"cuenta_id"`
	CuentaNombre string `db:"cuenta_nombre"`

	ActivoID      int    `db:"activo_id"`
	ActivoNombre  string `db:"activo_nombre"`
	UnidadSimbolo string `db:"unidad_simbolo"`

	Monto float64 `db:"monto"`

	Creado time.Time `db:"creado"`

	UsuarioID         int    `db:"usuario_id"`
	TipoTransaccionID int    `db:"tipo_transaccion_id"`
	Descripcion       string `db:"descripcion"`


	TotalMovimientos int `db:"total_movimientos"`
	TotalActivo     float64 `db:"total"`
	PromedioActivo  float64 `db:"promedio"`
	MedianaActivo   float64 `db:"mediana"`
	/*
	TotalGlobal     float64 `db:"total_global"`
	PromedioGlobal  float64 `db:"promedio_global"`
	MinimoGlobal    float64 `db:"minimo_global"`
	MaximoGlobal    float64 `db:"maximo_global"`
	MedianaGlobal   float64 `db:"mediana_global"`
	DesvioGlobal    float64 `db:"desvio_global"`
	*/
}

func NullStringToPtr(ns sql.NullString) *string {
    if ns.Valid {
        return &ns.String
    }
    return nil
}

func NullTimeToPtr(nt sql.NullTime) *time.Time {
    if nt.Valid {
        return &nt.Time
    }
    return nil
}

func StringToNullString(s string) sql.NullString {
	return sql.NullString{Valid: true, String: s}
}
