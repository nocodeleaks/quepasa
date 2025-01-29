package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	models "github.com/nocodeleaks/quepasa/models"
	log "github.com/sirupsen/logrus"
	"github.com/skip2/go-qrcode"
)

func ScannerController(w http.ResponseWriter, r *http.Request) {
	// setting default response type as json
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpResponse{}

	token := GetToken(r)
	if len(token) == 0 {
		err := fmt.Errorf("scanner controller, missing token")
		RespondBadRequest(w, err)
		return
	}

	username, ex := ValidateUsername(r)
	if ex != nil {
		ex.Prepend("scanner controller, username validation")
		response.ParseError(ex)
		RespondInterface(w, response)
		return
	}

	HSDString := models.GetRequestParameter(r, "historysyncdays")
	historysyncdays, _ := strconv.ParseUint(HSDString, 10, 32)

	pairing := &models.QpWhatsappPairing{
		Token:           token,
		Username:        username,
		HistorySyncDays: uint32(historysyncdays),
	}

	con, err := pairing.GetConnection()
	if err != nil {
		err := fmt.Errorf("scanner controller, cant get connection: %s", err.Error())
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	log.Infof("scanner controller, requesting qrcode for token %s", token)
	result := con.GetWhatsAppQRCode()

	var png []byte
	png, err = qrcode.Encode(result, qrcode.Medium, 256)
	if err != nil {
		err := fmt.Errorf("scanner controller, cant get qrcode: %s", err.Error())
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=qrcode.png")
	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(png))
}

func ValidateUsername(r *http.Request) (username string, ex ApiException) {
	username, err := GetUsername(r)
	if err != nil {
		ex = &BadRequestException{ApiExceptionBase{Inner: err}}
		return
	}

	if username == models.DEFAULTEMAIL {
		ex = &BadRequestException{}
		ex.Prepend("really ? are you dumb ? or am I ?")
		return
	}

	return
}

func PairCodeController(w http.ResponseWriter, r *http.Request) {
	// setting default response type as json
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpResponse{}

	token := GetToken(r)
	if len(token) == 0 {
		err := fmt.Errorf("pair code controller, missing token")
		RespondBadRequest(w, err)
		return
	}

	username, ex := ValidateUsername(r)
	if ex != nil {
		ex.Prepend("pair code controller, username validation")
		response.ParseError(ex)
		RespondInterface(w, response)
		return
	}

	pairing := &models.QpWhatsappPairing{Token: token, Username: username}
	con, err := pairing.GetConnection()
	if err != nil {
		err := fmt.Errorf("pair code controller, can't get connection: %s", err.Error())
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	phone := models.GetRequestParameter(r, "phone")
	if len(phone) == 0 {
		err := errors.New("pair code controller, missing phone number")
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	code, err := con.PairPhone(phone)
	if err != nil {
		err := fmt.Errorf("pair code controller, pair phone error: %s", err.Error())
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	response.Success = true
	response.Status = code
	RespondSuccess(w, response)
}
