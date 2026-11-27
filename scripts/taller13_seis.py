"""Complemento HTTP reproducible; no reemplaza las capturas de Bruno."""
from pathlib import Path
from urllib.request import Request, build_opener, ProxyHandler
from urllib.error import HTTPError
import argparse, json

root=Path(__file__).resolve().parents[1]
parser=argparse.ArgumentParser(description=__doc__)
parser.add_argument('--env',default='.env.taller13')
parser.add_argument('--etapa',required=True)
args=parser.parse_args()
config=dict(line.split('=',1) for line in Path(args.env).read_text(encoding='utf-8-sig').splitlines() if '=' in line and not line.startswith('#'))
base='http://127.0.0.1:'+config['PORT']
http=build_opener(ProxyHandler({}))
def call(method,path,body=None,token=None):
 headers={};data=None
 if body is not None:headers['Content-Type']='application/json';data=json.dumps(body).encode()
 if token:headers['Authorization']='Bearer '+token
 try:r=http.open(Request(base+path,data=data,headers=headers,method=method),timeout=10)
 except HTTPError as e:r=e
 with r:code=r.status;data=json.load(r)
 return code,data
def login(user):
 code,body=call('POST','/login',dict(usuario=user,password=config['DEMO_ADMIN_PASSWORD' if user=='admin' else 'DEMO_CLIENT_PASSWORD']))
 assert code==200,f'Login {user}: {code}; espere un minuto entre recorridos.'
 return code,body
_,admin=login('admin');_,client=login('cliente')
_,catalog=call('GET','/recompensas',token=client['token'])
reward=min(catalog['recompensas'],key=lambda r:r['puntos_necesarios'])
code,new=call('POST','/canjes',dict(recompensa_id=reward['id'],sucursal_id=2),client['token']);assert code==201
path='/canjes/'+str(new['id'])
codes=[call('GET',path)[0]]
code,client=login('cliente');codes.append(code)
codes += [call('GET',path,token=client['token'])[0],call('PATCH',path+'/estado',{'estado':'confirmado'},client['token'])[0],call('PATCH',path+'/estado',{'estado':'confirmado'},admin['token'])[0],call('PATCH',path+'/estado',{'estado':'cancelado'},admin['token'])[0]]
result=dict(herramienta='Python urllib; complemento de Bruno',url=base,canje_id=new['id'],codigos=codes)
out=root/'docs/taller13'/('http_'+args.etapa+'.json')
out.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
print(json.dumps(result,ensure_ascii=False))
assert codes==[401,200,200,403,200,422]
