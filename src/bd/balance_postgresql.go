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

func VerCuenta(cuenta_id int, db *sqlx.DB)(*dto.Cuenta,error){
	query := `SELECT * FROM Cuenta WHERE id=$1`
	var cuenta dto.Cuenta
	err := db.Get(&cuenta,query,cuenta_id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	query = `SELECT * FROM MontoCuenta WHERE cuenta_id=$1`
	var montos[] dto.MontoCuenta
	err = db.Select(&montos,query,cuenta_id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	var transacciones []dto.Transaccion
	queryTransacciones := `
			SELECT t.*
			FROM Transaccion t
			JOIN Movimiento m ON t.id = m.transaccion_id
			WHERE m.cuenta_id = $1 GROUP BY t.id, m.cuenta_id ORDER BY id DESC LIMIT 10`
	
	err = db.Select(&transacciones,queryTransacciones,cuenta_id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var transaccionesIDs []int
	for _, t := range transacciones {
		transaccionesIDs = append(transaccionesIDs, t.ID)
	}

	mapaMovimientos := make(map[int][]dto.Movimiento)
	if len(transaccionesIDs) > 0 {
		queryMovs, argsM, err := sqlx.In(`SELECT * FROM Movimiento WHERE transaccion_id IN (?)`, transaccionesIDs)
		if err != nil {
			return nil, err
		}
		queryMovs = db.Rebind(queryMovs)

		var todosLosMovs []dto.Movimiento
		if err := db.Select(&todosLosMovs, queryMovs, argsM...); err != nil {
			return nil, err
		}
		for _, m := range todosLosMovs {
			mapaMovimientos[m.TransaccionID] = append(mapaMovimientos[m.TransaccionID], m)
		}
	}

	for i, t := range transacciones {
		transacciones[i].Movimientos = mapaMovimientos[t.ID]
	}

	cuenta.UltTransacciones = transacciones
	cuenta.Monto = montos

	return &cuenta, nil
}

func VerCuentasPaginado(salto, limite int, db *sqlx.DB) (int, int, []dto.Cuenta, error) {
	var cuentas []dto.Cuenta
	var filas int

	// 1. Total para paginación
	if err := db.Get(&filas, `SELECT COUNT(*) FROM Cuenta`); err != nil {
		return 0, 0, nil, err
	}
	paginas := (filas+limite-1)/limite

	queryCuentas := `SELECT * FROM Cuenta ORDER BY id LIMIT $2 OFFSET $1`
	if err := db.Select(&cuentas, queryCuentas, salto, limite); err != nil {
		return filas, paginas, nil, err
	}

	if len(cuentas) == 0 {
		return filas, paginas, cuentas, nil
	}

	var cuentaIDs []int
	for _, c := range cuentas {
		cuentaIDs = append(cuentaIDs, c.ID)
	}

	queryIn, args, err := sqlx.In(`SELECT * FROM MontoCuenta WHERE cuenta_id IN (?)`, cuentaIDs)
	if err != nil {
		return filas, paginas, nil, err
	}
	queryIn = db.Rebind(queryIn)

	var todosLosMontos []dto.MontoCuenta
	if err := db.Select(&todosLosMontos, queryIn, args...); err != nil {
		return filas, paginas, nil, err
	}

	queryTransacciones := `
		SELECT * FROM (
			SELECT t.*, m.cuenta_id as cuenta_asociada_id,
			ROW_NUMBER() OVER (PARTITION BY m.cuenta_id ORDER BY t.creado DESC) as rn
			FROM Transaccion t
			JOIN Movimiento m ON t.id = m.transaccion_id
			WHERE m.cuenta_id IN (?) GROUP BY t.id, m.cuenta_id
		) AS t_paginadas WHERE t_paginadas.rn <= 10`

	queryInT, argsT, err := sqlx.In(queryTransacciones, cuentaIDs)
	if err != nil {
		return filas, paginas, nil, err
	}
	queryInT = db.Rebind(queryInT)

	type TransaccionConCuenta struct {
		dto.Transaccion
		CuentaAsociadaID int `db:"cuenta_asociada_id"`
		RN int `db:"rn"`
	}
	var todasLasTrans []TransaccionConCuenta
	if err := db.Select(&todasLasTrans, queryInT, argsT...); err != nil {
		return filas, paginas, nil, err
	}

	var transaccionesIDs []int
	for _, t := range todasLasTrans {
		transaccionesIDs = append(transaccionesIDs, t.ID)
	}

	mapaMovimientos := make(map[int][]dto.Movimiento)
	if len(transaccionesIDs) > 0 {
		queryMovs, argsM, err := sqlx.In(`SELECT * FROM Movimiento WHERE transaccion_id IN (?)`, transaccionesIDs)
		if err != nil {
			return filas, paginas, nil, err
		}
		queryMovs = db.Rebind(queryMovs)

		var todosLosMovs []dto.Movimiento
		if err := db.Select(&todosLosMovs, queryMovs, argsM...); err != nil {
			return filas, paginas, nil, err
		}
		for _, m := range todosLosMovs {
			mapaMovimientos[m.TransaccionID] = append(mapaMovimientos[m.TransaccionID], m)
		}
	}

	mapaMontos := make(map[int][]dto.MontoCuenta)
	for _, m := range todosLosMontos {
		mapaMontos[m.CuentaID] = append(mapaMontos[m.CuentaID], m)
	}

	mapaTrans := make(map[int][]dto.Transaccion)
	for _, t := range todasLasTrans {
		trans := t.Transaccion
		trans.Movimientos = mapaMovimientos[t.ID]
		mapaTrans[t.CuentaAsociadaID] = append(mapaTrans[t.CuentaAsociadaID], trans)
	}

	for i := range cuentas {
		cuentas[i].Monto = mapaMontos[cuentas[i].ID]
		cuentas[i].UltTransacciones = mapaTrans[cuentas[i].ID]
	}

	return filas, paginas, cuentas, nil
}

func AltaTransaccion(t dto.Transaccion, db *sqlx.DB) (int, error) {
	tx, err := db.Beginx()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	queryT := `INSERT INTO Transaccion(tipo_transaccion_id,empresa_id,usuario_id,descripcion)
	VALUES (:tipo_transaccion_id,:empresa_id,:usuario_id,:descripcion)
	RETURNING id`

	stmt, err := tx.PrepareNamed(queryT)
	if err != nil {
		log.Printf("Error: %v",err)
		return 0, err
	}
	defer stmt.Close()

	var id int
	err = stmt.Get(&id, t)
	if err != nil {
		log.Printf("Error: %v",err)
		return 0, err
	}
	
	queryM := `INSERT INTO Movimiento (transaccion_id, cuenta_id, activo_id, monto)
	           VALUES (`+ strconv.Itoa(id)+ `, :cuenta_id, :activo_id, :monto)`
	_, err = tx.NamedExec(queryM, t.Movimientos)
	if err != nil {
		log.Printf("Error: %v",err)
		return 0, err
	}

	res, err := tx.Exec(`UPDATE Transaccion SET estado_transaccion='FINALIZADA' WHERE id=$1`, id)

	if err != nil {
		log.Printf("Error: %v",err)
		return 0, err
	}

	rowsAff, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if rowsAff == 0 {
		return 0, fmt.Errorf("no se pudo actualizar el estado de la transacción %d", id)
	}
	err = tx.Commit()
	return id, err
}

func AltaActivo(a dto.Activo, db *sqlx.DB) (int, error) {
	query := `INSERT INTO Activo(nombre, unidad_id, empresa_id)	VALUES (:nombre,:unidad_id,:empresa_id) RETURNING id`

	stmt, err := db.PrepareNamed(query)
	if err != nil {
		log.Printf("Error preparando consulta: %v", err)
		return 0, err
	}
	defer stmt.Close()

	var id int
	err = stmt.Get(&id, a)
	if err != nil {
		log.Printf("Error ejecutando inserción: %v", err)
		return 0, err
	}

	return id, err
}

func BajaActivo(id int, destruir bool, db *sqlx.DB) (bool, error) {
	var query string
	if !destruir {
		query = `UPDATE Activo SET estado = 'BAJA', ult_mod = CURRENT_TIMESTAMP WHERE id = $1`
	} else {
		query = `DELETE FROM Activo WHERE id = $1`
	}
	res, err := db.Exec(query, id)
	if err != nil {
		return false, err
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return ra != 0, nil
}

func VerActivosPaginado(salto, limite int, db *sqlx.DB) (int, int, []dto.Activo, error) {
	var activos []dto.Activo
	var totalFilas int

	err := db.Get(&totalFilas, `SELECT COUNT(*) FROM Activo WHERE estado = 'ALTA'`)
	if err != nil {
		return 0, 0, nil, err
	}

	totalPaginas := (totalFilas + limite - 1) / limite

	if salto >= totalFilas && totalFilas > 0 {
		salto = max(0, totalFilas-limite)
	}

	query := `SELECT * FROM Activo WHERE estado = 'ALTA' ORDER BY id LIMIT $2 OFFSET $1`
	err = db.Select(&activos, query, salto, limite)
	if err != nil {
		return totalFilas, totalPaginas, nil, err
	}

	return totalFilas, totalPaginas, activos, nil
}

func VerCuentasEmpresa(empresaID, salto, limite int, db *sqlx.DB) (int, int, []dto.Cuenta, error) {
	var cuentas []dto.Cuenta
	var filas int

	if err := db.Get(&filas, `SELECT COUNT(*) FROM Cuenta WHERE empresa_id=$1 AND estado='ALTA'`,empresaID); err != nil {
		log.Printf("Error: %v", err)
		return 0, 0, nil, err
	}
	paginas := (filas+limite-1)/limite

	queryCuentas := `SELECT * FROM Cuenta WHERE empresa_id=$1 AND estado='ALTA' ORDER BY id LIMIT $2 OFFSET $3`
	if err := db.Select(&cuentas, queryCuentas,empresaID, limite, salto); err != nil {
		log.Printf("Error: %v", err)
		return filas, paginas, nil, err
	}
	return filas, paginas, cuentas, nil
}

func VerActivosEmpresa(empresaID, salto, limite int, db *sqlx.DB) (int,int,[]dto.Activo, error) {
	var activos []dto.Activo
	var filas int

	if err := db.Get(&filas, `SELECT COUNT(*) FROM Activo WHERE empresa_id=$1 AND estado='ALTA'`,empresaID); err != nil {
		log.Printf("Error: %v", err)
		return 0, 0, nil, err
	}
	paginas := (filas+limite-1)/limite
	queryActivos := `SELECT * FROM Activo WHERE empresa_id=$1 AND estado='ALTA' ORDER BY id LIMIT $2 OFFSET $3`
	if err := db.Select(&activos, queryActivos,empresaID, limite, salto); err != nil {
		log.Printf("Error: %v", err)
		return filas, paginas, nil, err
	}
	return filas, paginas, activos, nil
}

func VerTransaccionesEmpresa(empresaID, salto, limite int, db *sqlx.DB) (int,int,[]dto.Transaccion, error) {
	var transacciones []dto.Transaccion
	var filas int

	if err := db.Get(&filas, `SELECT COUNT(*) FROM Transaccion WHERE empresa_id=$1 AND estado='ALTA'`,empresaID); err != nil {
		log.Printf("Error: %v", err)
		return 0, 0, nil, err
	}
	paginas := (filas+limite-1)/limite
	queryTransacciones := `SELECT * FROM Transaccion WHERE empresa_id=$1 AND estado='ALTA' ORDER BY id LIMIT $2 OFFSET $3`
	if err := db.Select(&transacciones, queryTransacciones,empresaID, limite, salto); err != nil {
		log.Printf("Error: %v", err)
		return filas, paginas, nil, err
	}
	return filas, paginas, transacciones, nil
}

func CuentasPertenecenEmpresa(cuentaIDs []int, empresaID int, db *sqlx.DB) (bool, error) {
	query, args, err := sqlx.In(`SELECT COUNT(*) FROM Cuenta WHERE id IN (?) AND empresa_id=? AND estado='ALTA'`, cuentaIDs, empresaID)

	if err != nil {
		log.Printf("Error: %v", err)
		return false, err
	}

	query = db.Rebind(query)

	var count int
	err = db.Get(&count, query, args...)
	if err != nil {
		log.Printf("Error: %v", err)
		return false, err
	}
	return count == len(cuentaIDs), err
}

func ActivosPertenecenEmpresa(activoIDs []int, empresaID int, db *sqlx.DB) (bool, error) {
	query, args, err := sqlx.In(`SELECT COUNT(*) FROM Activo WHERE id IN (?) AND empresa_id=? AND estado='ALTA'`, activoIDs, empresaID)

	if err != nil {
		log.Printf("Error: %v", err)
		return false, err
	}

	query = db.Rebind(query)

	var count int
	err = db.Get(&count, query, args...)
	return count == len(activoIDs), err
}

func VerUnidades(salto, limite int,	db *sqlx.DB) (int, int, []dto.Unidad, error) {
	var unidades []dto.Unidad
	var filas int

	err := db.Get(&filas,`SELECT COUNT(*) FROM Unidad WHERE estado='ALTA'`)

	if err != nil {
		log.Printf("Error: %v", err)
		return 0, 0, nil, err
	}

	paginas := (filas + limite - 1) / limite

	err = db.Select(&unidades,`SELECT *	FROM Unidad	WHERE estado='ALTA'	ORDER BY tipo_unidad_id, nombre LIMIT $1 OFFSET $2`, limite, salto)

	if err != nil {
		log.Printf("Error: %v", err)
		return filas, paginas, nil, err
	}
	return filas, paginas, unidades, nil
}
