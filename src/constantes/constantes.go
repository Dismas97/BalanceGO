package constantes

const (
	CodOK = 0
	// Errores cliente (1000-2000)
	CodPeticionInvalida = 1001
	CodCamposFaltantes = 1002
	CodNoEncontrado = 1003
	CodCredencialesInvalidas = 1004

	// Errores de sesión
	CodSesionInvalida = 2001
	CodNoAutorizado = 2002

	// Errores de consistencia
	CodErrorConflicto = 3001
	// Errores server
	CodErrorInterno = 5001
	CodErrorSesion = 5002
)

const (
	//Msj respuesta
	MsjPeticionInvalida = "Error: Peticion inválida"
	MsjCamposFaltantes = "Error: Campos obligatorios faltantes"
	MsjNoEncontrado = "Error: No encontrado"
	MsjSesionInvalida = "Error: Sesión inválida"
	MsjNoAutorizado = "Error: No autorizado"
	MsjErrorInterno = "Error: Error inesperado"
)



type EstadoCuenta string
const (
	CuentaAbierta EstadoCuenta = "ABIERTA"
	CuentaCerrada EstadoCuenta = "CERRADA"
)
