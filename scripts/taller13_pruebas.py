from pathlib import Path
import subprocess, sys
root = Path(__file__).resolve().parents[1]
module = root/'Trabajo_autonomo_Lealcaf_hito2'
out = root/'docs/taller13'
out.mkdir(parents=True, exist_ok=True)
commands = {
 'servicio': ('prueba_servicio.txt', ['go','test','-count=1','./internal/fidelizacion','-run','Avisa','-v']),
 'websocket': ('prueba_websocket.txt', ['go','test','-count=1','./internal/fidelizacion','-run','WebSocket|Avisos','-v']),
 'reto': ('prueba_reto.txt', ['go','test','-count=1','./internal/fidelizacion','-run','Reto','-v']),
 'final': ('pruebas_finales.txt', ['go','test','-count=1','-v','./...']),
 'build': ('compilacion.txt', ['go','build','./...']),
 'vet': ('vet.txt', ['go','vet','./...']),
}
name, cmd = commands[sys.argv[1] if len(sys.argv)>1 else 'final']
p = subprocess.run(cmd, cwd=module, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
text = '$ '+' '.join(cmd)+'\n'+p.stdout.decode('utf-8', errors='replace')+f'\nCódigo de salida: {p.returncode}\n'
(out/name).write_text(text, encoding='utf-8')
print(text)
sys.exit(p.returncode)
