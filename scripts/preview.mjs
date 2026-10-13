// Local production-build preview with same-origin API proxy. Public deployment uses Nginx.
import http from 'node:http'
import fs from 'node:fs'
import path from 'node:path'
import {fileURLToPath} from 'node:url'
const app=process.argv[2]??'frontend'
if(!['frontend','admin'].includes(app))throw new Error('app must be frontend or admin')
const root=path.resolve(path.dirname(fileURLToPath(import.meta.url)),'..',app,'dist')
const port=Number(process.argv[3]??(app==='admin'?8089:8088))
const types={'.html':'text/html; charset=utf-8','.js':'text/javascript; charset=utf-8','.css':'text/css; charset=utf-8','.svg':'image/svg+xml','.png':'image/png','.jpg':'image/jpeg','.json':'application/json'}
http.createServer((req,res)=>{
 res.setHeader('X-Content-Type-Options','nosniff');res.setHeader('X-Frame-Options','DENY')
 if(req.url?.startsWith('/api/')){
  const upstream=http.request({hostname:'127.0.0.1',port:8080,path:req.url,method:req.method,headers:{...req.headers,host:'127.0.0.1:8080'},timeout:30000},response=>{res.writeHead(response.statusCode??502,response.headers);response.pipe(res)})
  upstream.on('timeout',()=>upstream.destroy());upstream.on('error',()=>{if(!res.headersSent)res.writeHead(502,{'Content-Type':'application/json'});res.end(JSON.stringify({error:'后端暂不可用'}))});req.pipe(upstream);return
 }
 if(!['GET','HEAD'].includes(req.method??'')){res.writeHead(405);res.end();return}
 let relative
 try{relative=decodeURIComponent(new URL(req.url??'/', 'http://localhost').pathname)}catch{res.writeHead(400);res.end();return}
 let file=path.resolve(root,'.'+relative)
 if(file!==root&&!file.startsWith(root+path.sep)){res.writeHead(403);res.end();return}
 if(!fs.existsSync(file)||!fs.statSync(file).isFile()){
  if(relative.startsWith('/assets/')||relative.startsWith('/images/')){res.writeHead(404);res.end();return}
  file=path.join(root,'index.html')
 }
 if(!fs.existsSync(file)){res.writeHead(503);res.end('Build the application first');return}
 res.setHeader('Content-Type',types[path.extname(file)]??'application/octet-stream')
 res.setHeader('Cache-Control',relative.startsWith('/assets/')?'public,max-age=31536000,immutable':'no-cache')
 res.writeHead(200);if(req.method==='HEAD'){res.end();return};fs.createReadStream(file).pipe(res)
}).listen(port,'127.0.0.1',()=>console.log(`${app} production preview http://127.0.0.1:${port}`))
