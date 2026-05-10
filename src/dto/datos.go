package dto

import (
	"database/sql"
	"sistema-balance/constantes"
	"time"
)

type Activo struct {
    ID int `db:"id"`
    Creado time.Time `db:"creado"`
    UltMod time.Time `db:"ult_mod"`
    Estado string `db:"estado"`
	
    Nombre string `db:"nombre"`
    Unidad sql.NullString `db:"unidad"`
}

type Conversion struct {
    ID int `db:"id"`
    Creado time.Time `db:"creado"`
    UltMod time.Time `db:"ult_mod"`
    Estado string `db:"estado"`

    ActivoOrigen int `db:"id"`
    ActivoDestino int `db:"id"`	
    Recomendado float64 `db:"recomendado"`
}

type HistorialConversion struct {
    Reloj time.Time `db:"reloj"`
    RecAnterior float64 `db:"conversion_anterior"`
    ConversionID int `db:"conversion_id"`
}

type Cuenta struct {
    ID int `db:"id"`
    Creado time.Time `db:"creado"`
    UltMod time.Time `db:"ult_mod"`
    Estado string `db:"estado"`

	Deuda bool `db:"permite_deuda"`
    UsuarioID int `db:"usuario_id"`
    EmpresaID int `db:"empresa_id"`
    Nombre string `db:"nombre"`
	Monto []MontoCuenta
	UltTransacciones []Transaccion
}

type MontoCuenta struct {
    Creado time.Time `db:"creado"`
    UltMod time.Time `db:"ult_mod"`
	
    CuentaID int `db:"cuenta_id"`
    ActivoID int `db:"activo_id"`
    Monto float64 `db:"monto"`
}

type Transaccion struct {
    ID int `db:"id"`
    Creado time.Time `db:"creado"`
    Estado string `db:"estado"`
	
    EstadoTransaccion string `db:"estado_transaccion"`
	Descripcion string `db:"descripcion"`
	Movimientos []Movimiento
}


type Movimiento struct {
    ID int `db:"id"`
    TransaccionID int `db:"transaccion_id"`
    CuentaID int `db:"cuenta_id"`
    ActivoID int `db:"activo_id"`
    Monto float64 `db:"monto"`
}

type HistorialCuenta struct {
	Id int `db:"id"`
    Reloj time.Time `db:"reloj"`
    Estado constantes.EstadoCuenta `db:"estado_final"`
	
    CuentaID int `db:"cuenta_id"`
    UsuarioID int `db:"usuario_id"`
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
