#!/bin/bash

if [ "$EUID" -ne 0 ]; then
  echo "Este script precisa ser executado como root. Digite sua senha para obter privilégios:"
  exec sudo "$0" "$@"
fi

mkdir -p /usr/local/bin/nex

LATEST_RELEASE=$(curl -s https://api.github.com/repos/XDukeHD/nex-server/releases/latest | grep "browser_download_url.*nex-server" | cut -d '"' -f 4)

curl -L -o /usr/local/bin/nex/nex-server "$LATEST_RELEASE"

chmod u+x /usr/local/bin/nex/nex-server

cat > /etc/systemd/system/nex.service <<EOF
[Unit]
Description=Nex Server Daemon
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/nex/nex-server
Restart=on-failure
User=root

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable nex
systemctl start nex

echo "Instalado! O Nex Server está rodando na porta 9384!"