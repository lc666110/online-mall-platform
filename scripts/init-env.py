"""Generate local Docker configuration without printing secrets or overwriting files."""
from pathlib import Path
import secrets
root=Path(__file__).resolve().parent.parent
p=root/'.env'
if p.exists():raise SystemExit('.env exists; retained unchanged')
values={'MYSQL_DATABASE':'mall','MYSQL_USER':'mall','MYSQL_PASSWORD':secrets.token_hex(24),'MYSQL_ROOT_PASSWORD':secrets.token_hex(24),'REDIS_PASSWORD':secrets.token_hex(24),'JWT_SECRET':secrets.token_hex(48),'ADMIN_EMAIL':'admin@example.com','ADMIN_PASSWORD':secrets.token_hex(16),'BIND_ADDRESS':'127.0.0.1','STOREFRONT_PORT':'8088','ADMIN_PORT':'8089'}
p.write_text('\n'.join(f'{k}={v}' for k,v in values.items())+'\n',encoding='utf8')
print('Created private .env. Review ADMIN_EMAIL; read ADMIN_PASSWORD locally.')
