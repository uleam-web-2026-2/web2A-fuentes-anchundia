from pathlib import Path
from urllib.request import Request,build_opener,ProxyHandler
import json, subprocess
root=Path(__file__).resolve().parents[1]
values=dict(l.split('=',1) for l in (root/'.env.taller13').read_text().splitlines() if '=' in l)
base='http://127.0.0.1:'+values['PORT']
http=build_opener(ProxyHandler({}))
req=Request(base+'/login',data=json.dumps(dict(usuario='cliente',password=values['DEMO_CLIENT_PASSWORD'])).encode(),headers={'Content-Type':'application/json'})
with http.open(req) as r: token=json.load(r)['token']
for name,origin,credential,status in [('sin_token','http://localhost:5173','',401),('otro_origen','http://malo.example',token,403),('permitido','http://localhost:5173',token,101)]:
 url=base+'/ws'+('?token='+credential if credential else '')
 config='\n'.join(['silent','include','max-time = 2','url = "'+url+'"','header = "Connection: Upgrade"','header = "Upgrade: websocket"','header = "Sec-WebSocket-Version: 13"','header = "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ=="','header = "Origin: '+origin+'"'])+'\n'
 p=subprocess.run(['curl.exe','--config','-'],input=config.encode(),capture_output=True)
 output=p.stdout.decode('utf-8',errors='replace')
 (root/f'docs/taller13/ws_{name}.txt').write_text(output,encoding='utf-8')
 assert output.startswith('HTTP/1.1 '+str(status)),(name,output)
 assert p.returncode in ([0,28] if status==101 else [0])
 print(name,status,'curl exit',p.returncode)
