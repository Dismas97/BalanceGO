package constantes

type PermisoID int
const (
	PermisoRootAltaCuenta PermisoID = 18
	PermisoRootBajaCuenta PermisoID = 19
	PermisoRootVerCuenta PermisoID = 20
	PermisoRootModificarCuenta PermisoID = 21
	PermisoRootAbrirCerrarCuenta PermisoID = 22

	PermisoRootAltatransaccion PermisoID = 23
	PermisoRootBajatransaccion PermisoID = 24
	PermisoRootVertransaccion PermisoID = 25

	PermisoRootVerMovimientos PermisoID = 26

	PermisoRootAltaActivo PermisoID = 27
	PermisoRootBajaActivo PermisoID = 28
	PermisoRootVerActivo PermisoID = 29
	PermisoRootModificarActivo PermisoID = 30
	
	PermisoRootVerMonto PermisoID = 31
)
