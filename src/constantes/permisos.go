package constantes

type PermisoID int
const (
	PermisoRootAltaCuenta PermisoID = 27
	PermisoRootBajaCuenta PermisoID = 28
	PermisoRootVerCuenta PermisoID = 29
	PermisoRootModificarCuenta PermisoID = 30
	PermisoRootAbrirCerrarCuenta PermisoID = 31

	PermisoRootAltatransaccion PermisoID = 32
	PermisoRootBajatransaccion PermisoID = 33
	PermisoRootVertransaccion PermisoID = 34

	PermisoRootVerMovimientos PermisoID = 35

	PermisoRootAltaActivo PermisoID = 36
	PermisoRootBajaActivo PermisoID = 37
	PermisoRootVerActivo PermisoID = 38
	PermisoRootModificarActivo PermisoID = 39
	
	PermisoRootVerMonto PermisoID = 40
)

var PermisosRoot = [14]PermisoID{
	PermisoRootAltaCuenta,
	PermisoRootBajaCuenta,
	PermisoRootVerCuenta,
	PermisoRootModificarCuenta,
	PermisoRootAbrirCerrarCuenta,
	PermisoRootAltatransaccion,
	PermisoRootBajatransaccion,
	PermisoRootVertransaccion,
	PermisoRootVerMovimientos,
	PermisoRootAltaActivo,
	PermisoRootBajaActivo,
	PermisoRootVerActivo,
	PermisoRootModificarActivo,
	PermisoRootVerMonto,
}
