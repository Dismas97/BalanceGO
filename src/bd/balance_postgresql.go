package bd

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sistema-balance/constantes"
	"sistema-balance/dto"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

func AltaCuenta(c dto.Cuenta, db *sqlx.DB) (int, error) {
	var id int
	query := `INSERT INTO Cuenta (permite_deuda,usuario_id,empresa_id,nombre) values (:permite_deuda,:usuario_id,:empresa_id,:nombre) RETURNING id`

	stmt, err := db.PrepareNamed(query)
	if err != nil {
		log.Printf("Error: %v",err)
		return 0, err
	}
	defer stmt.Close()
	err = stmt.Get(&id, c)
	if err != nil {
		log.Printf("Error: %v",err)
		return 0, err
	}
	return id, nil
}

func BajaCuenta(id int, destruir bool, db *sqlx.DB) (bool, error) {
	var query string
	if !destruir {
		query = `UPDATE Cuenta SET estado = 'BAJA' WHERE id = $1`
	} else {
		query = `DELETE FROM Cuenta WHERE id = $1`
	}
	res, err := db.Exec(query,id)
	if err != nil {
		return false, err
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return (ra != 0), nil
}

func CambiarEstadoCuenta(cuenta_id, usuario_id int, estado_final constantes.EstadoCuenta, db *sqlx.DB) (*dto.HistorialCuenta, error){
	query := `INSERT INTO HistorialCuenta (cuenta_id,usuario_id,estado_final) VALUES ($1,$2,$3) RETURNING *`
	var hc dto.HistorialCuenta
	err := db.Get(&hc,query,cuenta_id,usuario_id,estado_final)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &hc, nil
}

func AbrirCuenta(cuenta_id, usuario_id int,db *sqlx.DB) (*dto.HistorialCuenta, error){
	query := `INSERT INTO HistorialCuenta (estado_final,cuenta_id,usuario_id) VALUES ('ABIERTA',$1,$2) RETURNING *`
	var hc dto.HistorialCuenta
	err := db.Get(&hc,query,cuenta_id,usuario_id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &hc, nil
}

func CerrarCuenta(cuenta_id, usuario_id int,db *sqlx.DB) (*dto.HistorialCuenta, error){
	query := `INSERT INTO HistorialCuenta (estado_final,cuenta_id,usuario_id) VALUES ('CERRADA',$1,$2)  RETURNING *`
	var hc dto.HistorialCuenta
	err := db.Get(&hc,query,cuenta_id,usuario_id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &hc, nil
}

func CuentasAbiertas(cuenta_id []int, db *sqlx.DB) (bool,error){
	var str_cuentas []string
	for _, id := range cuenta_id {
		str_cuentas = append(str_cuentas, strconv.Itoa(id))
	}
	cuentas := "("+strings.Join(str_cuentas, ",") + ")"
	var aux int
	query := `SELECT COUNT(DISTINCT c.id) FROM Cuenta c
LEFT JOIN (SELECT DISTINCT ON (cuenta_id) cuenta_id, estado_final FROM HistorialCuenta ORDER BY cuenta_id, id DESC) hc ON c.id = hc.cuenta_id WHERE hc.estado_final = 'ABIERTA' AND  c.id IN` + cuentas
	err := db.QueryRow(query).Scan(&aux)
	if err != nil{
		return false, err
	}
	return aux == len(cuenta_id),nil
}

func ActivosExistentes(activo_id []int, db *sqlx.DB) (bool,error){
	var str_cuentas []string
	for _, id := range activo_id {
		str_cuentas = append(str_cuentas, strconv.Itoa(id))
	}
	activos := "("+strings.Join(str_cuentas, ",") + ")"
	var aux int
	query := `SELECT COUNT(DISTINCT a.id) FROM Activo a WHERE a.estado = 'ALTA' AND  a.id IN` + activos
	err := db.QueryRow(query).Scan(&aux)
	if err != nil{
		return false, err
	}
	return aux == len(activo_id),nil
}

func UltimoHistorialCuenta(cuenta_id int, db *sqlx.DB)(*dto.HistorialCuenta,error){
	query := `SELECT * FROM HistorialCuenta WHERE cuenta_id=$1 ORDER BY reloj DESC LIMIT 1`
	var hc dto.HistorialCuenta
	err := db.Get(&hc,query,cuenta_id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &hc, nil
}

func AltaTransaccion(t dto.Transaccion, db *sqlx.DB) (int, error) {
	tx, err := db.Beginx()
	if err != nil{
		return 0, err
	}
    defer tx.Rollback()
	
	queryT := `INSERT INTO Transaccion (descripcion) 
	          VALUES (:descripcion) RETURNING id`


	stmt, err := tx.PrepareNamed(queryT)
	if err != nil {
		log.Printf("Error: %v",err)
		tx.Rollback()
		return 0, err
	}
	defer stmt.Close()
	var id_transaccion int
	err = stmt.Get(&id_transaccion, t)
	if err != nil {
		log.Printf("Error: %v",err)
		tx.Rollback()
		return 0, err
	}

	queryM := `INSERT INTO Movimiento (transaccion_id, cuenta_id, activo_id, monto) 
	          VALUES (:transaccion_id,:cuenta_id,:activo_id,:monto)`
	stmt, err = tx.PrepareNamed(queryM)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	defer stmt.Close()
	
	for _,m := range t.Movimientos {
		m.TransaccionID = int(id_transaccion)
		_, err := stmt.Exec(m)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
	}

	queryU := `UPDATE Transaccion SET estado_transaccion='FINALIZADA' WHERE id = $1`
	res,err := tx.Exec(queryU,db)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	ok, err := res.RowsAffected()
	if ok == 0 {
		tx.Rollback()
		return 0, fmt.Errorf("vamos a tener que empezar a logear bd en algun momento mostro...")
	}
	
	tx.Commit()

	return id_transaccion, nil
}
