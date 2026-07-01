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
//CUENTAS
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

func AltaTasaIntercambio(t dto.TasaIntercambio, db *sqlx.DB) (int, error) {
	var id int
	query := `INSERT INTO TasaIntercambio (activo_a_id,activo_b_id, tasa, tasa_inversa, config, empresa_id) values (:activo_a_id,:activo_b_id,:tasa,:tasa_inversa,:config, :empresa_id) RETURNING id`

	stmt, err := db.PrepareNamed(query)
	if err != nil {
		log.Printf("Error: %v",err)
		return 0, err
	}
	defer stmt.Close()
	err = stmt.Get(&id, t)
	if err != nil {
		log.Printf("Error: %v",err)
		return 0, err
	}
	return id, nil
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


	mapaMontos := make(map[int][]dto.MontoCuenta)
	for _, m := range todosLosMontos {
		mapaMontos[m.CuentaID] = append(mapaMontos[m.CuentaID], m)
	}

	for i := range cuentas {
		cuentas[i].Monto = mapaMontos[cuentas[i].ID]
	}

	return filas, paginas, cuentas, nil
}

func VerCuentasEmpresa(empresaID,salto,limite int,busqueda *string,db *sqlx.DB) (int, int, []dto.Cuenta, error) {

	var cuentas []dto.Cuenta

	baseWhere := `
FROM Cuenta c
LEFT JOIN LATERAL (
	SELECT h.estado_final
	FROM HistorialCuenta h
	WHERE h.cuenta_id = c.id
	ORDER BY h.reloj DESC, h.id DESC
	LIMIT 1
) hc ON TRUE
WHERE c.empresa_id = $1
AND c.estado = 'ALTA'
`
	args := []any{empresaID}

	if busqueda != nil && *busqueda != "" {
		patron := "%" + *busqueda + "%"

		baseWhere += " AND " + buildBusquedaWhere(
			[]string{
				"c.id",
				"c.nombre",
			},
			2,
		)
		args = append(args, patron)
	}

	filas, err := contar(db,baseWhere,args...)
	if err != nil {
		log.Printf("Error count VerCuentasEmpresa: %v", err)
		return 0, 0, nil, err
	}

	paginas := calcularPaginas(filas, limite)

	query := `SELECT c.*, hc.estado_final AS estado_final` + baseWhere

	query = aplicarPaginacion(query,"c.id",len(args)+1,len(args)+2)

	args = append(args, limite, salto)

	err = db.Select(&cuentas, query, args...)
	if err != nil {
		log.Printf("Error select VerCuentasEmpresa: %v", err)
		return filas, paginas, nil, err
	}
	
	var cuentaIDs []int
	for _, c := range cuentas {
		cuentaIDs = append(cuentaIDs, c.ID)
	}

	queryIn, args, err := sqlx.In(`SELECT mc.cuenta_id,mc.activo_id,mc.monto,a.nombre FROM MontoCuenta mc JOIN Activo a ON a.id=mc.activo_id WHERE mc.cuenta_id IN (?)`, cuentaIDs)
	if err != nil {
		return filas, paginas, nil, err
	}
	queryIn = db.Rebind(queryIn)

	var todosLosMontos []dto.MontoCuenta
	if err := db.Select(&todosLosMontos, queryIn, args...); err != nil {
		return filas, paginas, nil, err
	}

	mapaMontos := make(map[int][]dto.MontoCuenta)
	for _, m := range todosLosMontos {
		mapaMontos[m.CuentaID] = append(mapaMontos[m.CuentaID], m)
	}

	for i := range cuentas {
		cuentas[i].Monto = mapaMontos[cuentas[i].ID]
	}
	
	return filas, paginas, cuentas, nil
}



func VerCuentasEmpresaJerarquico(empresaID,salto,limite int, jerarquia, busqueda *string,db *sqlx.DB) (int, int, []dto.Cuenta, error) {

	var cuentas []dto.Cuenta

	baseWhere := `
FROM Cuenta c
LEFT JOIN LATERAL (
	SELECT h.estado_final
	FROM HistorialCuenta h
	WHERE h.cuenta_id = c.id
	ORDER BY h.reloj DESC, h.id DESC
	LIMIT 1
) hc ON TRUE
WHERE c.empresa_id = $1 AND c.nombre ILIKE $2
AND c.estado = 'ALTA'
`
	jerar := *jerarquia + ":%"
	args := []any{empresaID, jerar}

	if busqueda != nil && *busqueda != "" {
		patron := "%" + *busqueda + "%"

		baseWhere += " AND " + buildBusquedaWhere(
			[]string{
				"c.id",
				"c.nombre",
			},
			3,
		)
		args = append(args, patron)
	}

	filas, err := contar(db,baseWhere,args...)
	if err != nil {
		log.Printf("Error count VerCuentasEmpresaJerarquico: %v", err)
		return 0, 0, nil, err
	}

	paginas := calcularPaginas(filas, limite)

	query := `SELECT c.*, hc.estado_final AS estado_final` + baseWhere

	query = aplicarPaginacion(query,"c.id",len(args)+1,len(args)+2)
	args = append(args, limite, salto)

	log.Printf("Args: %v", args)
	err = db.Select(&cuentas, query, args...)
	if err != nil {
		log.Printf("Error select VerCuentasEmpresa: %v", err)
		return filas, paginas, nil, err
	}
	
	var cuentaIDs []int
	for _, c := range cuentas {
		cuentaIDs = append(cuentaIDs, c.ID)
	}

	if len(cuentaIDs) <= 0 {
		return filas, paginas, cuentas, nil
	}

	queryIn, args, err := sqlx.In(`SELECT mc.cuenta_id,mc.activo_id,mc.monto,a.nombre FROM MontoCuenta mc JOIN Activo a ON a.id=mc.activo_id WHERE mc.cuenta_id IN (?)`, cuentaIDs)
	if err != nil {
		return filas, paginas, nil, err
	}
	queryIn = db.Rebind(queryIn)

	var todosLosMontos []dto.MontoCuenta
	if err := db.Select(&todosLosMontos, queryIn, args...); err != nil {
		return filas, paginas, nil, err
	}

	mapaMontos := make(map[int][]dto.MontoCuenta)
	for _, m := range todosLosMontos {
		mapaMontos[m.CuentaID] = append(mapaMontos[m.CuentaID], m)
	}

	for i := range cuentas {
		cuentas[i].Monto = mapaMontos[cuentas[i].ID]
	}
	
	return filas, paginas, cuentas, nil
}








func VerTransaccionesCuenta(cuentaID, salto, limite int, db *sqlx.DB) (int, int, []dto.Transaccion, error) {
	
	var transacciones []dto.Transaccion
	var filas int
	if err := db.Get(&filas, `SELECT COUNT(DISTINCT t.id) FROM Transaccion t INNER JOIN Movimiento m ON m.transaccion_id = t.id WHERE m.cuenta_id = $1 AND t.estado='ALTA'`,cuentaID); err != nil {
		log.Printf("Error: %v", err)
		return 0, 0, nil, err
	}
	paginas := (filas+limite-1)/limite
	queryTransacciones := `SELECT DISTINCT (t.*) FROM Transaccion t INNER JOIN Movimiento m ON m.transaccion_id = t.id WHERE m.cuenta_id = $1 AND t.estado='ALTA' ORDER BY id LIMIT $2 OFFSET $3`
	if err := db.Select(&transacciones, queryTransacciones,cuentaID, limite, salto); err != nil {
		log.Printf("Error: %v", err)
		return filas, paginas, nil, err
	}
	return filas, paginas, transacciones, nil
}
//ACTIVOS
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

func VerActivosEmpresa(empresaID, salto, limite int, busqueda *string, db *sqlx.DB) (int, int, []dto.Activo, error) {
	aux := `%`+*busqueda+`%`
	var activos []dto.Activo
	var filas int

	if err := db.Get(&filas, `SELECT COUNT(*) FROM Activo WHERE (id::text ILIKE $1 OR nombre::text ILIKE $1) AND empresa_id=$2 AND estado='ALTA'`,aux,empresaID); err != nil {
		log.Printf("Error: %v", err)
		return 0, 0, nil, err
	}
	paginas := (filas+limite-1)/limite
	queryActivos := `SELECT a.*, u.simbolo as unidad_simbolo, u.nombre as unidad_nombre FROM Activo a INNER JOIN Unidad u ON a.unidad_id = u.id WHERE (a.id::text ILIKE $1 OR a.nombre::text ILIKE $1) AND a.empresa_id=$2 AND a.estado='ALTA' ORDER BY a.id LIMIT $3 OFFSET $4`
	if err := db.Select(&activos, queryActivos,aux,empresaID, limite, salto); err != nil {
		log.Printf("Error: %v", err)
		return filas, paginas, nil, err
	}
	return filas, paginas, activos, nil
}

func VerActivo(cuenta_id int, db *sqlx.DB)(*dto.Activo,error){
	query := `SELECT a.*, u.simbolo as unidad_simbolo, u.nombre as unidad_nombre FROM Activo a JOIN Unidad u ON a.unidad_id = u.id WHERE a.id=$1`
	var activo dto.Activo
	err := db.Get(&activo,query,cuenta_id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	query = `SELECT * FROM TasaIntercambio WHERE activo_a_id=$1 OR activo_b_id=$1`
	var tasas[] dto.TasaIntercambio
	err = db.Select(&tasas,query,cuenta_id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	activo.Tasas = tasas

	return &activo, nil
}

func VerTasasActivo(activoID int, db *sqlx.DB) ([]dto.TasaIntercambio, error) {
	var tasas []dto.TasaIntercambio

	queryActivos := `SELECT * FROM TasaIntercambio WHERE activo_a_id=$1 OR activo_b_id=$1`
	if err := db.Select(&tasas, queryActivos,activoID); err != nil {
		log.Printf("Error: %v", err)
		return nil, err
	}
	return tasas, nil
}

func VerTasasIntercambioEmpresa(empresaID int, db *sqlx.DB) ([]dto.TasaIntercambio, error) {
	var tasas []dto.TasaIntercambio

	queryActivos := `SELECT * FROM TasaIntercambio WHERE empresa_id=$1`
	if err := db.Select(&tasas, queryActivos,empresaID); err != nil {
		log.Printf("Error: %v", err)
		return nil, err
	}
	return tasas, nil
}

//TRANSACCIONES
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

func VerMovimientosTransaccion(transaccionID, salto, limite int, db *sqlx.DB) (int,int,[]dto.Movimiento, error) {
	var movimientos []dto.Movimiento
	var filas int

	if err := db.Get(&filas, `SELECT COUNT(*) FROM Movimiento WHERE transaccion_id=$1`,transaccionID); err != nil {
		log.Printf("Error: %v", err)
		return 0, 0, nil, err
	}
	paginas := (filas+limite-1)/limite
	queryTransacciones := `SELECT m.*, c.nombre as cuenta_nombre, a.nombre as activo_nombre FROM Movimiento m INNER JOIN Cuenta c ON c.id = m.cuenta_id INNER JOIN Activo a ON a.id=m.activo_id WHERE m.transaccion_id=$1 ORDER BY id LIMIT $2 OFFSET $3`
	if err := db.Select(&movimientos, queryTransacciones,transaccionID, limite, salto); err != nil {
		log.Printf("Error: %v", err)
		return filas, paginas, nil, err
	}
	return filas, paginas, movimientos, nil
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

func VerUnidades(salto, limite int, busqueda *string, db *sqlx.DB) (int, int, []dto.Unidad, error) {
	aux := `%`+*busqueda+`%`
	var unidades []dto.Unidad
	var filas int
	
	if err := db.Get(&filas, `SELECT COUNT(*) FROM Unidad WHERE (id::text ILIKE $1 OR nombre::text ILIKE $1 OR simbolo ILIKE $1) AND estado='ALTA'`,aux); err != nil {
		log.Printf("Error: %v", err)
		return 0, 0, nil, err
	}
	paginas := (filas+limite-1)/limite
	queryActivos := `SELECT u.*, tu.nombre as nombre_tipo FROM Unidad u INNER JOIN TipoUnidad tu on u.tipo_unidad_id = tu.id WHERE (u.id::text ILIKE $1 OR u.nombre::text ILIKE $1 OR simbolo ILIKE $1) AND u.estado='ALTA' ORDER BY u.id LIMIT $2 OFFSET $3`
	if err := db.Select(&unidades, queryActivos,aux, limite, salto); err != nil {
		log.Printf("Error: %v", err)
		return filas, paginas, nil, err
	}
	return filas, paginas, unidades, nil
}






































 



// BuscarCuentas busca cuentas de una empresa por nombre (ILIKE).
// nombre vacío devuelve todas (equivale a LIKE '%%').
func BuscarCuentas(nombre string, empresaID, salto, limite int, db *sqlx.DB) (int, int, []dto.Cuenta, error) {
	patron := "%" + nombre + "%"
	var filas int
 
	if err := db.Get(&filas,
		`SELECT COUNT(*) FROM Cuenta
		 WHERE empresa_id=$1 AND estado='ALTA' AND nombre ILIKE $2`,
		empresaID, patron,
	); err != nil {
		log.Printf("BuscarCuentas count: %v", err)
		return 0, 0, nil, err
	}
 
	paginas := (filas + limite - 1) / limite
 
	var cuentas []dto.Cuenta
	if err := db.Select(&cuentas,
		`SELECT * FROM Cuenta
		 WHERE empresa_id=$1 AND estado='ALTA' AND nombre ILIKE $2
		 ORDER BY nombre LIMIT $3 OFFSET $4`,
		empresaID, patron, limite, salto,
	); err != nil {
		log.Printf("BuscarCuentas select: %v", err)
		return filas, paginas, nil, err
	}
 
	return filas, paginas, cuentas, nil
}
 
// BuscarActivos busca activos de una empresa por nombre (ILIKE).
func BuscarActivos(nombre string, empresaID, salto, limite int, db *sqlx.DB) (int, int, []dto.Activo, error) {
	patron := "%" + nombre + "%"
	var filas int
 
	if err := db.Get(&filas,
		`SELECT COUNT(*) FROM Activo
		 WHERE empresa_id=$1 AND estado='ALTA' AND nombre ILIKE $2`,
		empresaID, patron,
	); err != nil {
		log.Printf("BuscarActivos count: %v", err)
		return 0, 0, nil, err
	}
 
	paginas := (filas + limite - 1) / limite
 
	var activos []dto.Activo
	if err := db.Select(&activos,
		`SELECT * FROM Activo
		 WHERE empresa_id=$1 AND estado='ALTA' AND nombre ILIKE $2
		 ORDER BY nombre LIMIT $3 OFFSET $4`,
		empresaID, patron, limite, salto,
	); err != nil {
		log.Printf("BuscarActivos select: %v", err)
		return filas, paginas, nil, err
	}
 
	return filas, paginas, activos, nil
}
 
// BuscarUnidades busca unidades por nombre o símbolo (ILIKE).
// Al ser un catálogo global no filtra por empresa.
func BuscarUnidades(nombre string, salto, limite int, db *sqlx.DB) (int, int, []dto.Unidad, error) {
	patron := "%" + nombre + "%"
	var filas int
 
	if err := db.Get(&filas,
		`SELECT COUNT(*) FROM Unidad
		 WHERE estado='ALTA' AND (nombre ILIKE $1 OR simbolo ILIKE $1)`,
		patron,
	); err != nil {
		log.Printf("BuscarUnidades count: %v", err)
		return 0, 0, nil, err
	}
 
	paginas := (filas + limite - 1) / limite
 
	var unidades []dto.Unidad
	if err := db.Select(&unidades,
		`SELECT * FROM Unidad
		 WHERE estado='ALTA' AND (nombre ILIKE $1 OR simbolo ILIKE $1)
		 ORDER BY tipo_unidad_id, nombre LIMIT $2 OFFSET $3`,
		patron, limite, salto,
	); err != nil {
		log.Printf("BuscarUnidades select: %v", err)
		return filas, paginas, nil, err
	}
 
	return filas, paginas, unidades, nil
}
 
// ─── DETALLE ANIDADO ─────────────────────────────────────────────────────────
 
// VerTransaccionDetalle devuelve una transacción con sus movimientos anidados.
// Filtra por empresaID para que un usuario no pueda consultar transacciones
// de otra empresa adivinando el ID.
func VerTransaccionDetalle(transaccionID, empresaID int, db *sqlx.DB) (*dto.Transaccion, error) {
	var t dto.Transaccion
	err := db.Get(&t,
		`SELECT * FROM Transaccion
		 WHERE id=$1 AND empresa_id=$2 AND estado='ALTA'`,
		transaccionID, empresaID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		log.Printf("VerTransaccionDetalle get: %v", err)
		return nil, err
	}
 
	var movs []dto.Movimiento
	if err := db.Select(&movs,
		`SELECT * FROM Movimiento WHERE transaccion_id=$1 ORDER BY id`,
		transaccionID,
	); err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("VerTransaccionDetalle movimientos: %v", err)
		return nil, err
	}
 
	t.Movimientos = movs
	return &t, nil
}
 
// VerTransaccionesEmpresaDetalle lista las transacciones de una empresa con
// sus movimientos anidados en un único batch (sin N+1).
func VerTransaccionesEmpresaDetalle(empresaID, salto, limite int, db *sqlx.DB) (int, int, []dto.Transaccion, error) {
	var filas int
	if err := db.Get(&filas,
		`SELECT COUNT(*) FROM Transaccion WHERE empresa_id=$1 AND estado='ALTA'`,
		empresaID,
	); err != nil {
		log.Printf("VerTransaccionesEmpresaDetalle count: %v", err)
		return 0, 0, nil, err
	}
 
	paginas := (filas + limite - 1) / limite
 
	var transacciones []dto.Transaccion
	if err := db.Select(&transacciones,
		`SELECT * FROM Transaccion
		 WHERE empresa_id=$1 AND estado='ALTA'
		 ORDER BY id DESC LIMIT $2 OFFSET $3`,
		empresaID, limite, salto,
	); err != nil {
		log.Printf("VerTransaccionesEmpresaDetalle select: %v", err)
		return filas, paginas, nil, err
	}
 
	if len(transacciones) == 0 {
		return filas, paginas, transacciones, nil
	}
 
	transacciones, err := anidaMovimientos(transacciones, db)
	if err != nil {
		return filas, paginas, nil, err
	}
 
	return filas, paginas, transacciones, nil
}
 
// VerTransaccionesCuentaDetalle lista las transacciones de una cuenta con sus
// movimientos anidados. Filtra por empresaID para que un usuario no pueda
// consultar cuentas de otra empresa adivinando el cuenta_id.
func VerTransaccionesCuentaDetalle(cuentaID, empresaID, salto, limite int, db *sqlx.DB) (int, int, []dto.Transaccion, error) {
	var filas int
	if err := db.Get(&filas,
		`SELECT COUNT(DISTINCT t.id)
		 FROM Transaccion t
		 JOIN Movimiento m ON m.transaccion_id = t.id
		 JOIN Cuenta     c ON c.id = m.cuenta_id
		 WHERE m.cuenta_id=$1 AND t.estado='ALTA' AND c.empresa_id=$2`,
		cuentaID, empresaID,
	); err != nil {
		log.Printf("VerTransaccionesCuentaDetalle count: %v", err)
		return 0, 0, nil, err
	}
 
	paginas := (filas + limite - 1) / limite
 
	var transacciones []dto.Transaccion
	if err := db.Select(&transacciones,
		`SELECT DISTINCT t.*
		 FROM Transaccion t
		 JOIN Movimiento m ON m.transaccion_id = t.id
		 JOIN Cuenta     c ON c.id = m.cuenta_id
		 WHERE m.cuenta_id=$1 AND t.estado='ALTA' AND c.empresa_id=$2
		 ORDER BY t.id DESC LIMIT $3 OFFSET $4`,
		cuentaID, empresaID, limite, salto,
	); err != nil {
		log.Printf("VerTransaccionesCuentaDetalle select: %v", err)
		return filas, paginas, nil, err
	}
 
	if len(transacciones) == 0 {
		return filas, paginas, transacciones, nil
	}
 
	transacciones, err := anidaMovimientos(transacciones, db)
	if err != nil {
		return filas, paginas, nil, err
	}
 
	return filas, paginas, transacciones, nil
}
 
// anidaMovimientos trae todos los movimientos de una lista de transacciones
// en un único query y los asigna a cada transacción (evita N+1).
func anidaMovimientos(transacciones []dto.Transaccion, db *sqlx.DB) ([]dto.Transaccion, error) {
	ids := make([]int, len(transacciones))
	for i, t := range transacciones {
		ids[i] = t.ID
	}
 
	query, args, err := sqlx.In(
		`SELECT * FROM Movimiento WHERE transaccion_id IN (?) ORDER BY id`,
		ids,
	)
	if err != nil {
		log.Printf("anidaMovimientos In: %v", err)
		return nil, err
	}
	query = db.Rebind(query)
 
	var movs []dto.Movimiento
	if err := db.Select(&movs, query, args...); err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("anidaMovimientos select: %v", err)
		return nil, err
	}
 
	mapa := make(map[int][]dto.Movimiento, len(transacciones))
	for _, m := range movs {
		mapa[m.TransaccionID] = append(mapa[m.TransaccionID], m)
	}
	for i, t := range transacciones {
		transacciones[i].Movimientos = mapa[t.ID]
	}
 
	return transacciones, nil
}
