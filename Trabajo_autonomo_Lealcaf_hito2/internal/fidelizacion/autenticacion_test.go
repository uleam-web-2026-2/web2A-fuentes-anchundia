package fidelizacion

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestTodasLasRutasDelNegocioExigenToken(t *testing.T) {
	e := nuevoEntorno(t)
	for _, caso := range []struct{ metodo, ruta string }{
		{"GET", "/clientes/1/puntos"}, {"GET", "/recompensas"}, {"POST", "/compras"},
		{"POST", "/canjes"}, {"GET", "/canjes"}, {"GET", "/canjes/1"}, {"PATCH", "/canjes/1/estado"},
	} {
		t.Run(caso.metodo+caso.ruta, func(t *testing.T) {
			codigo, cuerpo := e.solicitar(caso.metodo, caso.ruta, "", `{}`)
			exigirCodigo(t, codigo, http.StatusUnauthorized, cuerpo)
		})
	}
}

func TestEncabezadosDelHito1NoSuplantanUnaSesion(t *testing.T) {
	e := nuevoEntorno(t)
	req := httptest.NewRequest("GET", "/canjes", nil)
	req.Header.Set("X-Rol", RolAdministrador)
	req.Header.Set("X-Cliente-ID", "1")
	w := httptest.NewRecorder()
	e.r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("encabezados editables autenticaron: HTTP %d %s", w.Code, w.Body.String())
	}
}

func TestLoginCorrectoYCredencialesInvalidas(t *testing.T) {
	e := nuevoEntorno(t)
	c := e.cliente(t, "0900000001", "a@example.com")
	e.cuenta("cliente-demo", "cliente-2026", RolCliente, c.ID)
	codigo, cuerpo := e.solicitar("POST", "/login", "", `{"usuario":"cliente-demo","password":"cliente-2026"}`)
	exigirCodigo(t, codigo, http.StatusOK, cuerpo)
	token, ok := cuerpo["token"].(string)
	if !ok || token == "" || cuerpo["tipo"] != "Bearer" || cuerpo["expira_en"] != float64(3600) {
		t.Fatalf("login incompleto: %v", cuerpo)
	}
	identidad, err := e.auth.ValidarToken(token)
	if err != nil || identidad.Rol != RolCliente || identidad.ClienteID != c.ID {
		t.Fatalf("identidad incorrecta: %+v %v", identidad, err)
	}
	var mensaje string
	for _, credenciales := range []string{`{"usuario":"cliente-demo","password":"incorrecta"}`, `{"usuario":"no-existe","password":"incorrecta"}`} {
		codigo, rechazo := e.solicitar("POST", "/login", "", credenciales)
		exigirCodigo(t, codigo, http.StatusUnauthorized, rechazo)
		actual, ok := rechazo["error"].(string)
		if !ok || actual == "" {
			t.Fatalf("falta error: %v", rechazo)
		}
		if mensaje != "" && mensaje != actual {
			t.Fatal("el error revela si existe el usuario")
		}
		mensaje = actual
		if _, existe := rechazo["token"]; existe {
			t.Fatal("el rechazo devolvió token")
		}
	}
	for _, invalido := range []string{`{`, `{}`, `{"usuario":"cliente-demo"}`} {
		codigo, rechazo := e.solicitar("POST", "/login", "", invalido)
		exigirCodigo(t, codigo, http.StatusBadRequest, rechazo)
	}
}

func TestTokenManipuladoNoElevaPermisos(t *testing.T) {
	e := nuevoEntorno(t)
	c := e.cliente(t, "0900000001", "a@example.com")
	token := e.token(RolCliente, c.ID)
	partes := strings.Split(token, ".")
	if len(partes) != 3 {
		t.Fatalf("no es JWT: %q", token)
	}
	payload, err := base64.RawURLEncoding.DecodeString(partes[1])
	if err != nil {
		t.Fatal(err)
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		t.Fatal(err)
	}
	claims["rol"] = RolAdministrador
	alterado, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}
	partes[1] = base64.RawURLEncoding.EncodeToString(alterado)
	codigo, cuerpo := e.solicitar("GET", "/canjes", strings.Join(partes, "."), "")
	exigirCodigo(t, codigo, http.StatusUnauthorized, cuerpo)
}

func TestTokensInvalidosYVencidosSeRechazan(t *testing.T) {
	e := nuevoEntorno(t)
	cuenta := e.cuenta("admin-expira", "admin-2026", RolAdministrador, 0)
	authCorta, err := NuevaAutenticacion(e.db, secretoPrueba, time.Nanosecond)
	if err != nil {
		t.Fatal(err)
	}
	vencido, err := authCorta.EmitirToken(cuenta)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Millisecond)
	for i, token := range []string{"no-es-un-token", "a.b.c", vencido} {
		t.Run(fmt.Sprintf("token_%d", i), func(t *testing.T) {
			codigo, cuerpo := e.solicitar("GET", "/canjes", token, "")
			exigirCodigo(t, codigo, http.StatusUnauthorized, cuerpo)
		})
	}
}

func TestConfiguracionYCuentaInvalidasNoGeneranTokens(t *testing.T) {
	e := nuevoEntorno(t)
	for _, secreto := range []string{"", "corto"} {
		if _, err := NuevaAutenticacion(e.db, secreto, time.Hour); err == nil {
			t.Fatal("se aceptó secreto insuficiente")
		}
	}
	if _, err := NuevaAutenticacion(e.db, secretoPrueba, 0); err == nil {
		t.Fatal("se aceptó duración sin vigencia")
	}
	for _, cuenta := range []Cuenta{{}, {ID: 1, Rol: "desconocido"}, {ID: 1, Rol: RolCliente}} {
		if _, err := e.auth.EmitirToken(cuenta); err == nil {
			t.Fatalf("token para identidad inválida: %+v", cuenta)
		}
	}
}

func TestListadoDeCanjesSoloParaAdministrador(t *testing.T) {
	e := nuevoEntorno(t)
	c := e.cliente(t, "0900000001", "a@example.com")
	codigo, cuerpo := e.llamar("GET", "/canjes", RolCliente, int(c.ID), "")
	exigirCodigo(t, codigo, http.StatusForbidden, cuerpo)
	codigo, cuerpo = e.llamar("GET", "/canjes", RolAdministrador, 0, "")
	exigirCodigo(t, codigo, http.StatusOK, cuerpo)
}

func TestCuentaEliminadaRevocaAcceso(t *testing.T) {
	e := nuevoEntorno(t)
	cuenta := e.cuenta("admin-revocado", "admin-2026", RolAdministrador, 0)
	token, err := e.auth.EmitirToken(cuenta)
	if err != nil {
		t.Fatal(err)
	}
	if err := e.db.Delete(&cuenta).Error; err != nil {
		t.Fatal(err)
	}
	codigo, cuerpo := e.solicitar("GET", "/canjes", token, "")
	exigirCodigo(t, codigo, http.StatusUnauthorized, cuerpo)
}

func TestTokenRequiereExpiracionYClaimsDeLaAplicacion(t *testing.T) {
	e := nuevoEntorno(t)
	cuenta := e.cuenta("admin-claims", "admin-2026", RolAdministrador, 0)
	for _, caso := range []struct {
		nombre, campo string
		valor         any
	}{
		{"sin_expiracion", "exp", nil},
		{"sin_fecha_emision", "iat", nil},
		{"otro_emisor", "iss", "otra-aplicacion"},
		{"otra_audiencia", "aud", "otra-api"},
		{"emision_futura", "iat", time.Now().Add(time.Hour).Unix()},
		{"cuenta_invalida", "sub", "0"},
	} {
		t.Run(caso.nombre, func(t *testing.T) {
			claims := jwt.MapClaims{"sub": fmt.Sprint(cuenta.ID), "rol": RolAdministrador, "cliente_id": 0,
				"iss": "lealcaf", "aud": "lealcaf-api", "iat": time.Now().Add(-time.Second).Unix(), "exp": time.Now().Add(2 * time.Hour).Unix()}
			if caso.valor == nil {
				delete(claims, caso.campo)
			} else {
				claims[caso.campo] = caso.valor
			}
			token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secretoPrueba))
			if err != nil {
				t.Fatal(err)
			}
			codigo, cuerpo := e.solicitar("GET", "/canjes", token, "")
			exigirCodigo(t, codigo, http.StatusUnauthorized, cuerpo)
		})
	}
}

func TestTokenNoAceptaAlgoritmoNoneNiOtroAlgoritmo(t *testing.T) {
	e := nuevoEntorno(t)
	cuenta := e.cuenta("admin-algoritmo", "admin-2026", RolAdministrador, 0)
	claims := jwt.MapClaims{"sub": fmt.Sprint(cuenta.ID), "rol": RolAdministrador, "cliente_id": 0,
		"iss": "lealcaf", "aud": "lealcaf-api", "iat": time.Now().Add(-time.Second).Unix(), "exp": time.Now().Add(time.Hour).Unix()}
	for _, caso := range []struct {
		metodo jwt.SigningMethod
		clave  any
	}{
		{jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType},
		{jwt.SigningMethodHS384, []byte(secretoPrueba)},
	} {
		t.Run(caso.metodo.Alg(), func(t *testing.T) {
			token, err := jwt.NewWithClaims(caso.metodo, claims).SignedString(caso.clave)
			if err != nil {
				t.Fatal(err)
			}
			codigo, cuerpo := e.solicitar("GET", "/canjes", token, "")
			exigirCodigo(t, codigo, http.StatusUnauthorized, cuerpo)
		})
	}
}
