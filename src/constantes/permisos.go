package constantes

type PermisoID int
const (
	PermisoRootAltaCuenta PermisoID = 27
	PermisoRootBajaCuenta PermisoID = 28
	PermisoRootVerCuenta PermisoID = 29
	PermisoRootModificarCuenta PermisoID = 30
	PermisoRootAbrirCerrarCuenta PermisoID = 31

	PermisoRootAltaTransaccion PermisoID = 32
	PermisoRootBajaTransaccion PermisoID = 33
	PermisoRootVerTransaccion PermisoID = 34

	PermisoRootVerMovimientos PermisoID = 35

	PermisoRootAltaActivo PermisoID = 36
	PermisoRootBajaActivo PermisoID = 37
	PermisoRootVerActivo PermisoID = 38
	PermisoRootModificarActivo PermisoID = 39
	
	PermisoRootVerMonto PermisoID = 40

	
	PermisoEmpresaAltaCuenta PermisoID = 41
	PermisoEmpresaBajaCuenta PermisoID = 42
	PermisoEmpresaVerCuenta PermisoID = 43
	PermisoEmpresaModificarCuenta PermisoID = 44
	PermisoEmpresaAbrirCerrarCuenta PermisoID = 45

	PermisoEmpresaAltaTransaccion PermisoID = 46
	PermisoEmpresaVerTransaccion PermisoID = 47
	PermisoEmpresaVerMovimientos PermisoID = 48

	PermisoEmpresaAltaActivo PermisoID = 49
	PermisoEmpresaBajaActivo PermisoID = 50
	PermisoEmpresaVerActivo PermisoID = 51
	PermisoEmpresaModificarActivo PermisoID = 52
	
	PermisoEmpresaVerMonto PermisoID = 53
)

var PermisosRoot = [14]PermisoID{
	PermisoRootAltaCuenta,
	PermisoRootBajaCuenta,
	PermisoRootVerCuenta,
	PermisoRootModificarCuenta,
	PermisoRootAbrirCerrarCuenta,
	PermisoRootAltaTransaccion,
	PermisoRootBajaTransaccion,
	PermisoRootVerTransaccion,
	PermisoRootVerMovimientos,
	PermisoRootAltaActivo,
	PermisoRootBajaActivo,
	PermisoRootVerActivo,
	PermisoRootModificarActivo,
	PermisoRootVerMonto,
}


const (
	UsuarioRoot int = 1
	RolRoot int = 1
	EmpresaRoot int = 1
)
