package models

const DEFAULTEMAIL string = "default@quepasa.io"

func InitialSeed() (err error) {
	exists, err := WhatsappService.DB.Users.Exists(DEFAULTEMAIL)
	if err != nil {
		return
	}

	if !exists {
		_, err = WhatsappService.DB.Users.Create(DEFAULTEMAIL, "")
		if err != nil {
			return
		}
	}
	return
}
