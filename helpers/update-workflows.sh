#!/bin/bash

cd /root/.n8n
if [ -z $1 ]; then 
	/usr/bin/n8n import:workflow --input=/opt/quepasa-source/extra/n8n+chatwoot/ --separate 
else 
	/usr/bin/n8n import:workflow --input=/opt/quepasa-source/extra/n8n+chatwoot/ --separate --userId=$1
fi

/usr/bin/n8n update:workflow --id 1008 --active=true
/usr/bin/n8n update:workflow --id 1009 --active=true
/usr/bin/n8n update:workflow --id 1010 --active=true
/usr/bin/n8n update:workflow --id 1011 --active=true

exit 0