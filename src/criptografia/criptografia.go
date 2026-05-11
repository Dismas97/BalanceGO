package criptografia

import (
	"sistema-balance/config"
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


func TienePermiso(c []int, p int) bool {
	maxc := len(c)-1
	if p > c[maxc] || p < c[0] {
		return false
	}
	i := 0
	for p > c[i] {
		i++
	}
	return p == c[i]
}

func TienePermisos(c, p []int) bool {
	maxc := len(c)-1
	maxp := len(p)-1	
	if p[0] > c[maxc] || p[maxp] < c[0] {
		return false
	}
	var interseccion []int
	i, j := 0, 0
	for i <= maxc && j <= maxp {
		if c[i] == p[j] {
			interseccion = append(interseccion, p[i])
			i++
			j++
		} else if c[i] < p[j] {
			i++
		} else {
			j++
		}
	}
	return len(interseccion)-1 == maxp
}
