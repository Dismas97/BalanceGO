package criptografia

import (
	"sistema-balance/dto"
	"sistema-balance/config"
	"github.com/golang-jwt/jwt/v5"
)

func ParseToken(t string) (*dto.Credenciales, error) {
    token, err := jwt.ParseWithClaims(t, &dto.Credenciales{}, func(token *jwt.Token) (any, error) {
        return config.MainConfig.SecretJWT, nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(*dto.Credenciales); ok && token.Valid {
        return claims, nil
    }

    return nil, err
}

/*

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
*/
