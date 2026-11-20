"""Comprobaciones locales del Taller 12. Ejecutar desde la raíz con Python 3."""

import argparse
import json
from pathlib import Path
from urllib.error import HTTPError
from urllib.parse import urlencode
from urllib.request import Request, build_opener, ProxyHandler


ROOT = Path(__file__).resolve().parents[1]
EVIDENCE = ROOT / "docs/taller12"
CONFIG = dict(line.split("=", 1) for line in (ROOT / ".env.taller12").read_text().splitlines() if "=" in line and not line.startswith("#"))
BASE = "http://127.0.0.1:" + CONFIG["PORT"]
HTTP = build_opener(ProxyHandler({}))


def request(method, path, body=None, token=None, headers=None):
    values = dict(headers or {})
    if token:
        values["Authorization"] = "Bearer " + token
    data = None
    if body is not None:
        values["Content-Type"] = "application/json"
        data = json.dumps(body).encode()
    req = Request(BASE + path, data=data, method=method, headers=values)
    try:
        response = HTTP.open(req, timeout=10)
    except HTTPError as error:
        response = error
    with response:
        raw = response.read().decode()
        status = response.status
        received = dict(response.headers)
    try:
        parsed = json.loads(raw)
    except ValueError:
        parsed = None
    # Solo las respuestas se registran. El login se maneja en memoria.
    text = f"HTTP {status}\n" + "".join(f"{key}: {value}\n" for key, value in received.items()) + "\n" + raw + "\n"
    return status, parsed, text, received


def login(user):
    key = "DEMO_ADMIN_PASSWORD" if user == "admin" else "DEMO_CLIENT_PASSWORD"
    code, result, _, _ = request("POST", "/login", {"usuario": user, "password": CONFIG[key]})
    if code != 200:
        raise RuntimeError(f"Login {user}: HTTP {code}. Espere la ventana o reinicie la API del taller.")
    return result["token"]


def prepare():
    admin, client = login("admin"), login("cliente")
    code, data, _, _ = request("GET", "/canjes", token=admin)
    assert code == 200
    confirmed = next(c["id"] for c in data["canjes"] if c["estado"] == "confirmado")
    code, catalog, _, _ = request("GET", "/recompensas", token=client)
    assert code == 200
    reward = min(catalog["recompensas"], key=lambda r: r["puntos_necesarios"])
    code, data, _, _ = request("POST", "/canjes", {"recompensa_id": reward["id"], "sucursal_id": 2}, client)
    assert code == 201
    return admin, client, confirmed, data["id"]


def six(stage):
    admin, client, confirmed, fresh = prepare()
    results = [request("GET", f"/canjes/{confirmed}")[0]]
    client = login("cliente")
    results.append(200)
    results.append(request("GET", f"/canjes/{confirmed}", token=client)[0])
    results.append(request("PATCH", f"/canjes/{fresh}/estado", {"estado": "confirmado"}, client)[0])
    results.append(request("PATCH", f"/canjes/{fresh}/estado", {"estado": "confirmado"}, admin)[0])
    results.append(request("PATCH", f"/canjes/{confirmed}/estado", {"estado": "cancelado"}, admin)[0])
    (EVIDENCE / f"peticiones_{stage}.json").write_text(json.dumps({"url": BASE, "canje_nuevo": fresh, "codigos": results}, indent=2) + "\n", encoding="utf-8")
    print(stage, results)
    assert results == [401, 200, 200, 403, 200, 422]


def injection():
    admin = login("admin")
    normal = request("GET", "/canjes?" + urlencode({"estado": "confirmado"}), token=admin)
    attack = request("GET", "/canjes?" + urlencode({"estado": "' OR 1=1 --"}), token=admin)
    invalid = request("GET", "/canjes/1%20OR%201=1", token=admin)
    (EVIDENCE / "inyeccion.txt").write_text(normal[2] + "\n" + attack[2] + "\n" + invalid[2], encoding="utf-8")
    assert normal[0] == 200 and normal[1]["canjes"]
    assert attack[0] in (400, 422) or (attack[0] == 200 and attack[1]["canjes"] == [])
    assert invalid[0] == 400
    print("Inyección", normal[0], attack[0], invalid[0])


def cors():
    for name, origin in [("permitido", "http://localhost:5173"), ("no_permitido", "http://malo.example")]:
        result = request("OPTIONS", "/canjes/1", headers={"Origin": origin, "Access-Control-Request-Method": "PATCH", "Access-Control-Request-Headers": "authorization,content-type"})
        (EVIDENCE / f"cors_{name}.txt").write_text(result[2], encoding="utf-8")
        allowed = {k.lower(): v for k, v in result[3].items()}.get("access-control-allow-origin")
        assert allowed == (origin if name == "permitido" else None)
    admin, _, _, fresh = prepare()
    body = {"estado": "confirmado"}
    (EVIDENCE / "cambio_estado.json").write_text(json.dumps(body) + "\n", encoding="utf-8")
    # curl ejecuta el PATCH real. Los secretos se pasan por stdin, no por argumentos.
    import subprocess
    configuration = f'url = "{BASE}/canjes/{fresh}/estado"\nrequest = "PATCH"\nheader = "Origin: http://malo.example"\nheader = "Authorization: Bearer {admin}"\nheader = "Content-Type: application/json"\ndata = "@{(EVIDENCE / "cambio_estado.json").as_posix()}"\nsilent\ninclude\n'
    response = subprocess.run(["curl.exe", "--config", "-"], input=configuration, text=True, capture_output=True, check=True)
    (EVIDENCE / "cors_curl_ejecuta.txt").write_text(response.stdout, encoding="utf-8")
    assert response.stdout.startswith("HTTP/1.1 200")
    assert "access-control-allow-origin:" not in response.stdout.lower()
    print("Preflight permitido/no permitido y PATCH de curl correctos")


def burst():
    body = {"usuario": "cliente", "password": "incorrecta-para-el-ensayo"}
    (EVIDENCE / "login_malo.json").write_text(json.dumps(body) + "\n", encoding="utf-8")
    results = [request("POST", "/login", body, headers={"X-Forwarded-For": f"10.0.0.{i}"}) for i in range(7)]
    (EVIDENCE / "rafaga.txt").write_text("\n".join(r[2] for r in results), encoding="utf-8")
    codes = [r[0] for r in results]
    print("Ráfaga", codes)
    assert codes == [401] * 5 + [429, 429]
    assert all(int({k.lower(): v for k, v in r[3].items()}["retry-after"]) > 0 for r in results[-2:])
    assert request("GET", "/canjes/1")[0] == 401


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("phase", choices=["antes", "defensas", "reto", "compose", "inyeccion", "cors", "rafaga"])
    phase = parser.parse_args().phase
    EVIDENCE.mkdir(parents=True, exist_ok=True)
    if phase == "inyeccion":
        injection()
    elif phase == "cors":
        cors()
    elif phase == "rafaga":
        burst()
    else:
        six(phase)
