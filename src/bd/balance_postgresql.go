package bd

import (
	"database/sql"
	"errors"
	"log"
	"sistema-balance/constantes"
	"sistema-balance/dto"

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

func AltaTransaccion(t dto.Transaccion, db *sqlx.DB) (int64, error) {
	tx, err := db.Beginx()
	if err != nil{
		return 0, err
	}
    defer tx.Rollback()
	
	queryT := `INSERT INTO Transaccion (descripcion) 
	          VALUES (:descripcion) RETURNING id`
	res, err := tx.NamedExec(queryT,t)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	id_transacccion,err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	queryM := `INSERT INTO Movimiento (transaccion_id, cuenta_id, activo_id, monto) 
	          VALUES (:transaccion_id,:cuenta_id,:activo_id,:monto)`
	stmt, err := tx.PrepareNamed(queryM)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	defer stmt.Close()
	
	for _,m := range t.Movimientos {
		m.TransaccionID = int(id_transacccion)
		_, err := stmt.Exec(m)
		if err != nil {
			tx.Rollback()
		}
	}
	tx.Commit()

	return id_transacccion, nil
}

