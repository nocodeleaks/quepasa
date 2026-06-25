## SIP Headers
* **From**: "{{whatsapp_contact_title}}" <sip:{{callerid_phone}}@{{public_ip}}:{{listener_port}}>;tag=*****
* **To**: <sip:{{receiver_whatsapp}}@{{remote_sip_server}}:{{remote_sip_server_port}}>
* **Via**: SIP/2.0/{{protocol}} {{public_ip}}:{{listener_port}};branch={{branch}};rport

* **protocol**: UDP|TCP|TLS
* **whatsapp_contact_title**: should be ommited if empty, also ommit `"" `, goes directly to <>
* **listener_port**: can be ommited if = 5060, :5060, it's good for reduce the total size of the invite
* **remote_sip_server_port**: can be ommited if = 5060, :5060, it's good for reduce the total size of the invite
* **branch**: z9hG4bK*****, this prefix indicates the RFC 3261 compliance
* **rport**: if using local random port, if not should ommit `;rport`