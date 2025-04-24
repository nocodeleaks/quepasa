[![Go Build](https://github.com/nocodeleaks/quepasa/actions/workflows/go.yml/badge.svg)](https://github.com/nocodeleaks/quepasa/actions/workflows/go.yml)

<p align="center">
	<img src="https://github.com/nocodeleaks/quepasa/raw/main/src/assets/favicon.png" alt="Quepasa-logo" width="100" />	
	<p align="center">Quepasa is a Open-source, all free license software to exchange messages with Whatsapp Platform</p>
</p>
<hr />
<p align="left">
	<img src="https://telegram.org/favicon.ico" alt="Telegram-logo" width="32" />
	<span>Chat with us on Telegram: </span>
	<a href="https://t.me/quepasa_api" target="_blank">Group</a>
	<span> || </span>
	<a href="https://t.me/quepasa_channel" target="_blank">Channel</a>
</p>
<p align="left">
	<span>Special thanks to <a target="_blank" href="https://agenciaoctos.com.br">Lukas Prais</a>, who developed this logo.</span>
</p>
<hr />
# QuePasa

> A (micro) web-application to make web-based [WhatsApp][0] bots easy to write.

[![Run in Postman](https://run.pstmn.io/button.svg)](https://god.gw.postman.com/run-collection/5047984-405506cf-59f5-479e-b512-4ba5b935411b?action=collection%2Ffork&source=rip_markdown&collection-url=entityId%3D5047984-405506cf-59f5-479e-b512-4ba5b935411b%26entityType%3Dcollection%26workspaceId%3Dbd72aaba-0c31-40ad-801c-d5ba19184aff#?env%5BQuepasa%5D=W3sia2V5IjoiYmFzZVVybCIsInZhbHVlIjoiIiwiZW5hYmxlZCI6dHJ1ZSwidHlwZSI6ImRlZmF1bHQiLCJzZXNzaW9uVmFsdWUiOiIiLCJjb21wbGV0ZVNlc3Npb25WYWx1ZSI6IiIsInNlc3Npb25JbmRleCI6MH0seyJrZXkiOiJ0b2tlbiIsInZhbHVlIjoiIiwiZW5hYmxlZCI6dHJ1ZSwidHlwZSI6ImRlZmF1bHQiLCJzZXNzaW9uVmFsdWUiOiIiLCJjb21wbGV0ZVNlc3Npb25WYWx1ZSI6IiIsInNlc3Npb25JbmRleCI6MX0seyJrZXkiOiJjaGF0SWQiLCJ2YWx1ZSI6IiIsImVuYWJsZWQiOnRydWUsInR5cGUiOiJkZWZhdWx0Iiwic2Vzc2lvblZhbHVlIjoiIiwiY29tcGxldGVTZXNzaW9uVmFsdWUiOiIiLCJzZXNzaW9uSW5kZXgiOjJ9LHsia2V5IjoiZmlsZU5hbWUiLCJ2YWx1ZSI6IiIsImVuYWJsZWQiOnRydWUsInR5cGUiOiJkZWZhdWx0Iiwic2Vzc2lvblZhbHVlIjoiIiwiY29tcGxldGVTZXNzaW9uVmFsdWUiOiIiLCJzZXNzaW9uSW5kZXgiOjN9LHsia2V5IjoidGV4dCIsInZhbHVlIjoiIiwiZW5hYmxlZCI6dHJ1ZSwidHlwZSI6ImRlZmF1bHQiLCJzZXNzaW9uVmFsdWUiOiIiLCJjb21wbGV0ZVNlc3Npb25WYWx1ZSI6IiIsInNlc3Npb25JbmRleCI6NH0seyJrZXkiOiJ0cmFja0lkIiwidmFsdWUiOiJwb3N0bWFuIiwiZW5hYmxlZCI6dHJ1ZSwidHlwZSI6ImRlZmF1bHQiLCJzZXNzaW9uVmFsdWUiOiJwb3N0bWFuIiwiY29tcGxldGVTZXNzaW9uVmFsdWUiOiJwb3N0bWFuIiwic2Vzc2lvbkluZGV4Ijo1fV0=)

**Features:**
  * Verify a number with a QR code
  * Persistence of account data and keys
  * Exposes HTTP endpoints for:
    * sending messages
    * receiving messages
    * download attachments
    * set webhook for receiving messages 

  **WARNING: This application has not been audited. It should not be regarded as
  secure, use at your own risk.**

  **This is a third-party effort, and is NOT in any affiliated with [WhatsApp][0].**

<details>
  <summary>Anything is section was not reviewed</summary>

  ### Why ?
  
  Angry, Angry ... WhatsApp keeps canceling our number.  
  
  When you need to communicate over WhatsApp from a different service, for example,
  [a help desk](http://zammad.org/) or other web-app, QuePasa provides a simple HTTP
  API to do so.

  QuePasa stores keys and WhatsApp account data in a postgres database. It does
  not come with HTTPS out of the box. Your QuePasa API tokens essentially give
  full access to your WhatsApp account (to the extent that QuePasa has
  implemented WhatsApp features). Use with caution.

  For HTTPS use Nginx.

  ## If are you looking for a NODE.JS Project

  Take a look at
  https://github.com/pedroslopez/whatsapp-web.js/pulls

  Its a lot more complete tool to whatsapp unofficial api

  ## Join our community 
  Matrix chat room #cdr-link-dev-support:matrix.org
  https://app.element.io/#/room/#cdr-link-dev-support:matrix.org

  ## Usage

  ## Prerequisites Local Deployment

  * Golang (Version go1.20 minimum version)

  ### *installing above golang version*

  ```bash
  cd /usr/src

  sudo wget https://go.dev/dl/go1.20.linux-amd64.tar.gz
  sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.20.linux-amd64.tar.gz

  #export the PATH
  export PATH=$PATH:/usr/local/go/bin

  ```

  ---


  ## Docker Implimentation

  ### Prerequisites

  For local development
  * docker
  * golang
  * postgresql

  ### Run using Docker

  * Add info about database migrations

  ```bash

  make docker_build
  # edit docker-compose.yml.sample to your hearts content
  docker-compose up
  ```
</details>

### Environment Variables

	# WEBAPIHOST
	> http server bind host (HOST:PORT). (default empty)	
	
	# WEBAPIPORT
	> http server bind port (HOST:PORT). (default 31000)
	
	# WEBSOCKETSSL
	> Should websocket for qrcode reads use ssl. (default false)	
		
	# APP_TITLE
	> Suffix for quepasa name on whatsapp devices list like (QuePasa Sufficit). (default empty)	
	
	# COMPATIBLE_MIME_AS_AUDIO
	> Should convert sending audio files to OGG codec and use as PTT. (default true)	
	
	# GOOS		
	> Operational System to Golang Extensions, "linux" | "windows". (default "linux")
		
	# REMOVEDIGIT9
	> Remove digit 9 from phones bigger than DDD 30. (default false)
	
	# GROUPS
	
	# BROADCASTS
	
	# READRECEIPTS
	> Trigger webhooks for read receipts events. (default false)

	# CALLS
	> defines if will be accepted calls

	# READUPDATE
	> Mark chat read when send any msg. (default true)

 	# CACHELENGTH
	> Defines the amount of messages that should be kept in cache (default empty(unlimited))

 	# MASTERKEY
  	> A master key for administration
   
	# SYNOPSISLENGTH
	> Length for synopsis msg at replies or reactions, (default 50)

  	# HISTORYSYNCDAYS
	> Defines the default amount of days of history that will be request on the QrCode Scan (default 0)
		
	# LOGLEVEL
	
	# PRESENCE
	> Sets defaults presence state (available|unavailable), (default unavailable)
	
	# HTTPLOGS
	> Log http requests. (default false)
	
	# WHATSMEOW_LOGLEVEL
	
	# WHATSMEOW_DBLOGLEVEL

  	# ACCOUNTSETUP
  	> enable or disable account creation setup. (default true)

	 
### License

[![License GNU AGPL v3.0](https://img.shields.io/badge/License-AGPL%203.0-lightgrey.svg)](https://github.com/nocodeleaks/quepasa-fork/blob/master/LICENSE.md)

QuePasa is a free software project licensed under the GNU Affero General Public License v3.0 (GNU AGPLv3) by "Someone Who Cares About You".

[0]: https://whatsapp.com
[1]: https://github.com/tulir/whatsmeow
