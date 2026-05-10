package crypto

import (
	"sistema-balance/config"
	"sistema-balance/constantes"
	"slices"
	"github.com/golang-jwt/jwt/v5"
)

type Credenciales struct {
	SesionID int `json:"sid"`
    UsuarioID int `json:"uid"`
    EmpresaID int `json:"emp"`
    Roles []int `json:"roles"`
    Permisos []int `json:"perms"`
    jwt.RegisteredClaims
}

func ParseToken(t string) (*Credenciales, error) {
    token, err := jwt.ParseWithClaims(t, &Credenciales{}, func(token *jwt.Token) (interface{}, error) {
        return config.MainConfig.SecretJWT, nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(*Credenciales); ok && token.Valid {
        return claims, nil
    }

    return nil, err
}


func ValidarPermisoRoot(permisoID constantes.PermisoID, c *Credenciales) bool {
	if c == nil || c.EmpresaID != 1 || c.UsuarioID != 1 {
		return false
	}

	hasRol := slices.Contains(c.Roles, 1)
	if !hasRol {
		return false
	}

	return slices.Contains(c.Permisos, int(permisoID))
}
