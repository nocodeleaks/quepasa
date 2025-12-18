package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

//region CONTROLLER - CONTACTS SEARCH

// ContactSearchController searches contacts based on criteria
//
//	@Summary		Search contacts
//	@Description	Search WhatsApp contacts by name, phone, and other filters
//	@Tags			Contacts
//	@Accept			json
//	@Produce		json
//	@Param			body	body		models.QpContactsSearchRequest	true	"Search criteria"
//	@Success		200		{object}	models.QpContactsResponse
//	@Failure		400		{object}	models.QpResponse
//	@Security		ApiKeyAuth
//	@Router			/contact/search [post]
func ContactSearchController(w http.ResponseWriter, r *http.Request) {

	// setting default response type as json
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpContactsResponse{}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Parse request body (accept empty body)
	var searchRequest models.QpContactsSearchRequest
	if r.Body != nil {
		err = json.NewDecoder(r.Body).Decode(&searchRequest)
		// Ignore EOF error (empty body is valid - means no filters)
		if err != nil && err != io.EOF {
			response.ParseError(err)
			RespondInterface(w, response)
			return
		}
	}

	// Validate phone field if provided
	if len(searchRequest.Phone) > 0 {
		if !isValidPhoneSearch(searchRequest.Phone) {
			response.ParseError(fmt.Errorf("invalid phone format: only numbers, '+', spaces, hyphens and parentheses are allowed"))
			RespondInterface(w, response)
			return
		}
	}

	// Get all contacts (works with active connection or cached data automatically)
	contacts, err := server.GetContacts()
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Apply filters
	filtered := filterContacts(contacts, searchRequest)

	// Sort contacts by ID to ensure consistent ordering
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Id < filtered[j].Id
	})

	response.Total = len(filtered)
	response.Contacts = filtered
	RespondSuccess(w, response)
}

// filterContacts applies search criteria to contact list
func filterContacts(contacts []whatsapp.WhatsappChat, request models.QpContactsSearchRequest) []whatsapp.WhatsappChat {
	var result []whatsapp.WhatsappChat

	for _, contact := range contacts {
		// Apply all filters - contact must match ALL criteria
		if !matchesSearchCriteria(contact, request) {
			continue
		}
		result = append(result, contact)
	}

	return result
}

// matchesSearchCriteria checks if contact matches all search criteria
func matchesSearchCriteria(contact whatsapp.WhatsappChat, request models.QpContactsSearchRequest) bool {
	// Filter by has_title
	if request.HasTitle != nil {
		hasTitle := len(strings.TrimSpace(contact.Title)) > 0
		if *request.HasTitle != hasTitle {
			return false
		}
	}

	// Filter by has_lid
	if request.HasLid != nil {
		// Contact has LID if: LId field is not empty OR Id ends with @lid
		hasLid := len(strings.TrimSpace(contact.LId)) > 0 || strings.HasSuffix(contact.Id, "@lid")
		if *request.HasLid != hasLid {
			return false
		}
	}

	// Filter by specific phone
	if len(request.Phone) > 0 {
		// If phone starts with "+", search by prefix (useful for region filtering, e.g., +55 for Brazil)
		// Otherwise, search by contains (anywhere in the phone number)
		if strings.HasPrefix(request.Phone, "+") {
			// Starts with search - for E.164 format regional filtering
			if !strings.HasPrefix(contact.Phone, request.Phone) {
				return false
			}
		} else {
			// Contains search - normalize and search anywhere
			requestPhone := normalizePhone(request.Phone)
			contactPhone := normalizePhone(contact.Phone)
			if !strings.Contains(contactPhone, requestPhone) {
				return false
			}
		}
	}

	// Filter by query (search in name and phone)
	if len(request.Query) > 0 {
		query := strings.ToLower(strings.TrimSpace(request.Query))

		// Search in title (name)
		title := strings.ToLower(contact.Title)
		if strings.Contains(title, query) {
			return true
		}

		// Search in phone
		phone := normalizePhone(contact.Phone)
		queryNormalized := normalizePhone(query)
		if strings.Contains(phone, queryNormalized) {
			return true
		}

		// Search in id (full WhatsApp ID)
		id := strings.ToLower(contact.Id)
		if strings.Contains(id, query) {
			return true
		}

		// If query was provided but no match found
		return false
	}

	return true
}

// normalizePhone removes non-numeric characters from phone number
func normalizePhone(phone string) string {
	// Remove common phone formatting characters
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")
	phone = strings.ReplaceAll(phone, "+", "")
	return strings.ToLower(phone)
}

// isValidPhoneSearch validates phone search string - allows only numbers, +, spaces, hyphens and parentheses
func isValidPhoneSearch(phone string) bool {
	for _, char := range phone {
		// Allow: digits (0-9), plus sign (+), space, hyphen (-), parentheses ( )
		if (char >= '0' && char <= '9') || char == '+' || char == ' ' || char == '-' || char == '(' || char == ')' {
			continue
		}
		// Any other character (letters, underscore, special chars) is invalid
		return false
	}
	return true
}

//endregion
