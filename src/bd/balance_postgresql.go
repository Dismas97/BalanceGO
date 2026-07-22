package bd

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sistema-balance/constantes"
	"sistema-balance/dto"
	sbe "sistema-balance/error"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
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

/*

   SELECT 
    m.*
FROM 
    Movimiento m
INNER JOIN 
    Transaccion t ON m.transaccion_id = t.id
WHERE 
    m.cuenta_id = ANY(:cuentas)
    AND t.creado BETWEEN :fecha_inicio AND :fecha_fin
    AND t.estado_transaccion = 'FINALIZADA' ORDER BY m.transaccion_id, t.creado;


   
SELECT
    m.cuenta_id,
    m.activo_id,
    a.nombre,
    c.nombre as cuenta,
    SUM(m.monto) as total
FROM 
    Movimiento m
INNER JOIN 
    Transaccion t ON m.transaccion_id = t.id
INNER JOIN
    Activo a ON m.activo_id = a.id
INNER JOIN
    Cuenta c ON m.cuenta_id = c.id
WHERE
    m.cuenta_id = ANY(:cuentas)
    AND t.creado BETWEEN :fecha_inicio AND :fecha_fin
    AND t.estado_transaccion = 'FINALIZADA'
GROUP BY 
    m.cuenta_id, m.activo_id, a.nombre, c.nombre
ORDER BY 
    m.cuenta_id, m.activo_id;
   
 */

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

func VerCuentaNombre(empresa_id int, cuenta_nombre *string, db *sqlx.DB)(*dto.Cuenta,error){
	query := `SELECT c.*, hc.estado_final AS estado_final
FROM Cuenta c
LEFT JOIN LATERAL (
	SELECT h.estado_final
	FROM HistorialCuenta h
	WHERE h.cuenta_id = c.id
	ORDER BY h.reloj DESC, h.id DESC
	LIMIT 1
) hc ON TRUE
WHERE c.empresa_id = $1 AND c.nombre = $2
AND c.estado = 'ALTA'`
	var cuenta dto.Cuenta
	err := db.Get(&cuenta,query,empresa_id,cuenta_nombre)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	query = `SELECT mc.cuenta_id,mc.activo_id,mc.monto,a.nombre FROM MontoCuenta mc JOIN Activo a ON a.id=mc.activo_id WHERE mc.cuenta_id=$1`
	var montos[] dto.MontoCuenta
	err = db.Select(&montos,query,cuenta.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	cuenta.Monto = montos
	return &cuenta, nil
}

func VerCuenta(cuenta_id int, db *sqlx.DB)(*dto.Cuenta,error){
	query := `SELECT c.*,  hc.estado_final AS estado_final FROM Cuenta c
LEFT JOIN LATERAL (
	SELECT h.estado_final
	FROM HistorialCuenta h
	WHERE h.cuenta_id = c.id
	ORDER BY h.reloj DESC, h.id DESC
	LIMIT 1
) hc ON TRUE WHERE c.id=$1`
	var cuenta dto.Cuenta
	err := db.Get(&cuenta,query,cuenta_id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	query = `SELECT mc.cuenta_id,mc.activo_id,mc.monto,a.nombre FROM MontoCuenta mc JOIN Activo a ON a.id=mc.activo_id WHERE mc.cuenta_id=$1`
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

func VerCuentaNombreMontoPaginado(empresa_id int, cuenta_nombre *string, salto, limite int, db *sqlx.DB) (*dto.Cuenta, int, error) {
	// Obtener la cuenta principal
	query := `SELECT c.*, hc.estado_final AS estado_final
FROM Cuenta c
LEFT JOIN LATERAL (
	SELECT h.estado_final
	FROM HistorialCuenta h
	WHERE h.cuenta_id = c.id
	ORDER BY h.reloj DESC, h.id DESC
	LIMIT 1
) hc ON TRUE
WHERE c.empresa_id = $1 AND c.nombre = $2
AND c.estado = 'ALTA'`

	var cuenta dto.Cuenta
	err := db.Get(&cuenta, query, empresa_id, cuenta_nombre)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, nil
		}
		return nil, 0, err
	}

	// Contar total de montos para esta cuenta
	var totalMontos int
	countQuery := `SELECT COUNT(*) FROM MontoCuenta WHERE cuenta_id = $1`
	err = db.Get(&totalMontos, countQuery, cuenta.ID)
	if err != nil {
		return nil, 0, err
	}

	// Calcular número de páginas
	paginas := 0
	if limite > 0 {
		paginas = (totalMontos + limite - 1) / limite
	}

	var montos []dto.MontoCuenta

	if limite > 0 {
		// Consulta paginada de montos
		montoQuery := `
SELECT mc.cuenta_id, mc.activo_id, mc.monto, a.nombre, u.simbolo
FROM MontoCuenta mc
JOIN Activo a ON a.id = mc.activo_id
JOIN Unidad u ON a.unidad_id = u.id
WHERE mc.cuenta_id = $1
ORDER BY mc.activo_id
LIMIT $2 OFFSET $3`
		err = db.Select(&montos, montoQuery, cuenta.ID, limite, salto)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return nil, 0, err
			}
			// Si no hay filas, dejamos slice vacío
			montos = []dto.MontoCuenta{}
		}
	} else {
		// Si límite no es positivo, devolvemos todos los montos (sin paginación)
		montoQuery := `
SELECT mc.cuenta_id, mc.activo_id, mc.monto, a.nombre, u.simbolo
FROM MontoCuenta mc
JOIN Activo a ON a.id = mc.activo_id
JOIN Unidad u ON a.unidad_id = u.id
WHERE mc.cuenta_id = $1`
		err = db.Select(&montos, montoQuery, cuenta.ID)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return nil, 0, err
			}
			montos = []dto.MontoCuenta{}
		}
	}

	cuenta.Monto = montos
	return &cuenta, paginas, nil
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
	jerar := *jerarquia + "%"
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
	query := `INSERT INTO Activo(nombre, unidad_id, empresa_id, alias_id)	VALUES (:nombre,:unidad_id,:empresa_id, :alias_id) RETURNING id`
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

func ModificarActivoTx(a dto.ModificarActivo,tx *sqlx.Tx) (error) {
	query := `UPDATE Activo 
        SET 
            nombre = COALESCE(:nombre, nombre),
            unidad_id = COALESCE(:unidad_id, unidad_id),
            alias_id = COALESCE(:alias_id, alias_id)
        WHERE id = :id AND empresa_id = :empresa_id`

	res, err := tx.NamedExec(query,a)
	if err != nil {
        log.Printf("Error ejecutando actualización: %v", err)
        return err
    }
    rowsAffected, err := res.RowsAffected()
	if err != nil {
        log.Printf("Error ejecutando actualización: %v", err)
        return err
    }
    if rowsAffected == 0 {
        log.Printf("No se encontró el Activo con ID %d", a.ID)
		return &sbe.ENoEncontrado{}
    }
    return nil
}

func ModificarTasaTx(t dto.TasaIntercambio,tx *sqlx.Tx) (error) {
	query := `UPDATE TasaIntercambio 
        SET 
            tasa = :tasa,
            tasa_inversa = :tasa_inversa,
            config = :config
        WHERE activo_a_id = :activo_a_id AND activo_b_id = :activo_b_id`

	res, err := tx.NamedExec(query,t)
	if err != nil {
        log.Printf("Error ejecutando actualización: %v", err)
        return err
    }
    rowsAffected, _ := res.RowsAffected()
    if rowsAffected == 0 {
        log.Printf("No se encontro el par A:%d B:%d", t.ActivoA, t.ActivoB)
		return &sbe.ENoEncontrado{}
    }
    return nil
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

func VerResumenMovimientosEmpresa(
    fechaInicio *time.Time,
    fechaFin *time.Time,
    cuentas []int,
    activos []int,
    montoMin *float64,
    montoMax *float64,
	empresa int,
    salto int,
    limite int,
    db *sqlx.DB,
) (int, int, []dto.ResumenMov, error){
	var resumen []dto.ResumenMov
	query := `
	WITH movimientos AS (
    SELECT
        count(*) over() as total_movimientos,
        m.id,
        m.transaccion_id,
        m.cuenta_id,
        c.nombre AS cuenta_nombre,
        m.activo_id,
        a.nombre AS activo_nombre,
        u.simbolo AS unidad_simbolo,
        m.monto,

        t.creado,
        t.usuario_id,
        t.descripcion,
        t.tipo_transaccion_id
    FROM Movimiento m
    INNER JOIN Transaccion t ON t.id = m.transaccion_id
    INNER JOIN Cuenta c ON c.id = m.cuenta_id
    INNER JOIN Activo a ON a.id = m.activo_id
    INNER JOIN Unidad u ON u.id = a.unidad_id
    WHERE
        t.estado_transaccion = 'FINALIZADA' AND t.empresa_id=$9

        AND ($1::timestamptz IS NULL OR t.creado >= $1)
        AND ($2::timestamptz IS NULL OR t.creado <= $2)

        AND (
            $3::int[] IS NULL
            OR cardinality($3)=0
            OR m.cuenta_id = ANY($3)
        )

        AND (
            $4::int[] IS NULL
            OR cardinality($4)=0
            OR m.activo_id = ANY($4)
        )

        AND ($5::numeric IS NULL OR m.monto >= $5)
        AND ($6::numeric IS NULL OR m.monto <= $6)
),
estadisticas_activo AS (
SELECT

    activo_id,

    COUNT(*) cantidad,

    SUM(monto) total,

    AVG(monto) promedio,

    percentile_cont(0.5)
        WITHIN GROUP (ORDER BY monto) mediana

FROM movimientos

GROUP BY activo_id, id
)

SELECT
    m.*,
    ea.total,
    ea.promedio,
    ea.mediana

FROM movimientos m

JOIN estadisticas_activo ea
ON ea.activo_id = m.activo_id
ORDER BY creado DESC

LIMIT $7
OFFSET $8`
	
	err := db.Select(
		&resumen,
		query,
		fechaInicio,
		fechaFin,
		pq.Array(cuentas),
		pq.Array(activos),
		montoMin,
		montoMax,
		limite,
		salto,
		empresa,
	)
	if err != nil {
		return 0, 0, nil, err
	}
	var totalFilas, totalPaginas int
	if (len(resumen)>0){
		totalFilas = resumen[0].TotalMovimientos
		
		totalPaginas = max(0, resumen[0].TotalMovimientos-limite)
	}
	return totalFilas,totalPaginas, resumen, nil
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


func VerActivosEmpresaTipoComp(empresaID, tipoID, salto, limite int, busqueda *string, db *sqlx.DB) (int, int, []dto.Activo, error) {
	aux := `%`+*busqueda+`%`
	var activos []dto.Activo
	var filas int

	if err := db.Get(&filas, `SELECT COUNT(*) FROM Activo a
INNER JOIN 
    Unidad u ON a.unidad_id = u.id
INNER JOIN 
    TipoUnidad tu ON u.tipo_unidad_id = tu.id WHERE (a.nombre::text ILIKE $1) AND a.empresa_id=$2 AND tu.id <> $3 AND a.estado='ALTA'`,aux,empresaID,tipoID); err != nil {
		log.Printf("Error: %v", err)
		return 0, 0, nil, err
	}
	paginas := (filas+limite-1)/limite
	queryActivos := `SELECT a.*, u.simbolo as unidad_simbolo, u.nombre as unidad_nombre FROM Activo a
INNER JOIN 
    Unidad u ON a.unidad_id = u.id
INNER JOIN 
    TipoUnidad tu ON u.tipo_unidad_id = tu.id
WHERE (a.id::text ILIKE $1 OR a.nombre::text ILIKE $1)
    AND a.empresa_id = $2
    AND tu.id <> $3
    AND a.estado='ALTA' ORDER BY a.id LIMIT $4 OFFSET $5`
	if err := db.Select(&activos, queryActivos,aux,empresaID,tipoID, limite, salto); err != nil {
		log.Printf("Error: %v", err)
		return filas, paginas, nil, err
	}
	return filas, paginas, activos, nil
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

func AltaActivoTx(a dto.Activo, tx *sqlx.Tx) (int, error) {
	query := `INSERT INTO Activo(nombre, unidad_id, empresa_id, alias_id) VALUES (:nombre,:unidad_id,:empresa_id, :alias_id) RETURNING id`
	stmt, err := tx.PrepareNamed(query)
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

func AltaTasaIntercambioTx(t dto.TasaIntercambio, tx *sqlx.Tx) (int, error) {
	var id int
	query := `INSERT INTO TasaIntercambio (activo_a_id,activo_b_id, tasa, tasa_inversa, config, empresa_id) values (:activo_a_id,:activo_b_id,:tasa,:tasa_inversa,:config, :empresa_id) RETURNING id`

	stmt, err := tx.PrepareNamed(query)
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

func AltaTransaccionTx(t dto.Transaccion, tx *sqlx.Tx) (int, error) {
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

    return id, nil
}

func VerMontosCuentaPaginado(cuentaID, salto, limite int, db *sqlx.DB) (int, int, []dto.MontoCuenta, error) {
 	var totalFilas int
	countQuery := `SELECT COUNT(*) FROM MontoCuenta WHERE cuenta_id = $1`
	err := db.Get(&totalFilas, countQuery, cuentaID)
	if err != nil {
		log.Printf("Error contando montos de cuenta %d: %v", cuentaID, err)
		return 0, 0, nil, err
	}
	totalPaginas := 0
	if limite > 0 {
		totalPaginas = (totalFilas + limite - 1) / limite
	}
	if totalFilas == 0 {
		return 0, 0, []dto.MontoCuenta{}, nil
	}

	query := `
		SELECT 
			mc.cuenta_id,
			mc.activo_id,
			mc.monto,
			a.nombre AS nombre,
			u.simbolo AS simbolo
		FROM MontoCuenta mc
		JOIN Activo a ON a.id = mc.activo_id
		JOIN Unidad u ON u.id = a.unidad_id
		WHERE mc.cuenta_id = $1
		ORDER BY mc.activo_id
		LIMIT $2 OFFSET $3
	`
	var montos []dto.MontoCuenta
	err = db.Select(&montos, query, cuentaID, limite, salto)
	if err != nil {
		log.Printf("Error seleccionando montos paginados para cuenta %d: %v", cuentaID, err)
		return totalFilas, totalPaginas, nil, err
	}
	return totalFilas, totalPaginas, montos, nil
}
