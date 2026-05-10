package dto

type AltaCuenta struct {
	Deuda bool `json:"permite_deuda,omitempty"`
	UsuarioID int `json:"usuario_id,omitempty"`
	EmpresaID  int `json:"empresa_id,omitempty"`
	Nombre string  `json:"nombre"`
}

/*
type RequestRegistro struct {
	Usuario string `json:"usuario" form:"usuario"`
	Contra string `json:"contra" form:"contra"`
	Email string `json:"email" form:"email,omitempty"`
	Nombre string `json:"nombre,omitempty" form:"nombre"`
	Apellido string `json:"apellido,omitempty" form:"apellido"`
	Telefono string `json:"telefono,omitempty" form:"telefono"`
	Direccion string `json:"direccion,omitempty" form:"direccion"`
}

type RequestAcceso struct {
	Usuario string `json:"usuario"`
	Email string `json:"email,omitempty"`
	Contra  string `json:"contra,omitempty"`
}

type RequestPaginado struct {
	Limite int `schema:"limite"`
	Salto int `schema:"salto"`
	Filtros string `schema:"filtros,omitempty"`
}

type RequestAltaRol struct {
	Nombre string `json:"nombre"`
}

type RequestAltaPermiso struct {
	Nombre string `json:"nombre"`
}
*/
