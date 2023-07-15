### Workflows links

* [(1.0.0) Chatwoot Profile Update](https://raw.githubusercontent.com/nocodeleaks/quepasa/main/extra/n8n%2Bchatwoot/ChatwootProfileUpdate.json)
* [Chatwoot To Quepasa](https://raw.githubusercontent.com/nocodeleaks/quepasa/main/extra/n8n%2Bchatwoot/ChatwootToQuepasa.json)
* [Chatwoot To Quepasa Greetings](https://raw.githubusercontent.com/nocodeleaks/quepasa/main/extra/n8n%2Bchatwoot/ChatwootToQuepasaGreetings.json)
* [(1.0.0) Post To Chatwoot](https://raw.githubusercontent.com/nocodeleaks/quepasa/main/extra/n8n%2Bchatwoot/PostToChatwoot.json)
* [Quepasa Automatic](https://raw.githubusercontent.com/nocodeleaks/quepasa/main/extra/n8n%2Bchatwoot/QuepasaAutomatic.json)
* [Quepasa Chat Control](https://raw.githubusercontent.com/nocodeleaks/quepasa/main/extra/n8n%2Bchatwoot/QuepasaChatControl.json)
* [Quepasa Inbox Control](https://raw.githubusercontent.com/nocodeleaks/quepasa/main/extra/n8n%2Bchatwoot/QuepasaInboxControl.json)
* [Quepasa Qrcode](https://raw.githubusercontent.com/nocodeleaks/quepasa/main/extra/n8n%2Bchatwoot/QuepasaQrcode.json)
* [Quepasa To Chatwoot](https://raw.githubusercontent.com/nocodeleaks/quepasa/main/extra/n8n%2Bchatwoot/QuepasaToChatwoot.json)
* [(1.0.0) Get Chatwoot Contacts](https://raw.githubusercontent.com/nocodeleaks/quepasa/main/extra/n8n%2Bchatwoot/GetChatwootContacts.json)
* [(1.0.0) Post To WebCallBack](https://raw.githubusercontent.com/nocodeleaks/quepasa/main/extra/n8n%2Bchatwoot/PostToWebCallBack.json)

### Use N8N Environment File to set these variables:
> use your respectives ids

	# C8Q_QUEPASAINBOXCONTROL
	> (integer) Workflow Id - QuepasaInboxControl, used by ChatwootToQuepasa

	# C8Q_QUEPASACHATCONTROL
	> (integer) Workflow Id - QuepasaChatControl, used by ChatwootToQuepasa | QuepasaToChatwoot

	# C8Q_CHATWOOTPROFILEUPDATE
	> (integer) Workflow Id - ChatwootProfileUpdate, used by ChatwootToQuepasa

	# C8Q_POSTTOCHATWOOT
	> (integer) Workflow Id - PostToChatwoot, used by QuepasaToChatwoot

	# C8Q_POSTTOWEBCALLBACK
	> (integer) Workflow Id - PostToWebCallBack, used by QuepasaToChatwoot

	# C8Q_CHATWOOTTOQUEPASAGREETINGS
	> (integer) Workflow Id - ChatwootToQuepasaGreetings, used by QuepasaToChatwoot

	# C8Q_GETCHATWOOTCONTACTS
	> (integer) Workflow Id - GetChatwootContacts, (2023/06/19) used by QuepasaToChatwoot

	# C8Q_SINGLETHREAD
	> (boolean => true | false) Enable a single conversation per contact, for all life time, not just a ticket
	
	# C8Q_QP_DEFAULT_USER
	> (optional) (string => user@domain.com)

	# C8Q_CW_PUBLIC_URL
	> (optional) (string => https://chatwoot.com) get initial from $N8N_HOST, otherwise from here, used by ChatwootToQuepasa

	# C8Q_QP_BOTTITLE
	> (optional) (string => QuePasa [the best]) set title for sincrony bot for messages from whatsapp

	# C8Q_QP_CONTACT
	> (optional) (string => control@quepasa.io) set identifier for quepasa inbox control contact
	
	# C8Q_SUFFICIT_CONTEXTID
	> (optional) (string => 00000000-0000-0000-0000-000000000000) sufficit client identification
