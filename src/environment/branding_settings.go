package environment

// Branding environment variable names
const (
	ENV_BRANDING_TITLE           = "BRANDING_TITLE"           // Application title (default: "QuePasa")
	ENV_BRANDING_LOGO            = "BRANDING_LOGO"            // Logo URL (default: QuePasa logo)
	ENV_BRANDING_FAVICON         = "BRANDING_FAVICON"         // Favicon URL (default: QuePasa favicon)
	ENV_BRANDING_PRIMARY_COLOR   = "BRANDING_PRIMARY_COLOR"   // Primary color (default: #7C3AED - QuePasa purple)
	ENV_BRANDING_SECONDARY_COLOR = "BRANDING_SECONDARY_COLOR" // Secondary color (default: #5B21B6 - QuePasa dark purple)
	ENV_BRANDING_ACCENT_COLOR    = "BRANDING_ACCENT_COLOR"    // Accent color (default: #8B5CF6 - QuePasa light purple)
	ENV_BRANDING_COMPANY_NAME    = "BRANDING_COMPANY_NAME"    // Company name for footer
	ENV_BRANDING_COMPANY_URL     = "BRANDING_COMPANY_URL"     // Company URL for footer link
)

// Default QuePasa branding values
const (
	DefaultBrandingTitle          = "QuePasa"
	DefaultBrandingLogo           = "https://raw.githubusercontent.com/nocodeleaks/quepasa/main/src/assets/favicon.png"
	DefaultBrandingFavicon        = "https://raw.githubusercontent.com/nocodeleaks/quepasa/main/src/assets/favicon.png"
	DefaultBrandingPrimaryColor   = "#7C3AED" // QuePasa purple
	DefaultBrandingSecondaryColor = "#5B21B6" // QuePasa dark purple
	DefaultBrandingAccentColor    = "#8B5CF6" // QuePasa light purple
)

// BrandingSettings holds all branding configuration loaded from environment
type BrandingSettings struct {
	Title          string `json:"title"`
	Logo           string `json:"logo"`
	Favicon        string `json:"favicon"`
	PrimaryColor   string `json:"primaryColor"`
	SecondaryColor string `json:"secondaryColor"`
	AccentColor    string `json:"accentColor"`
	CompanyName    string `json:"companyName"`
	CompanyUrl     string `json:"companyUrl"`
}

// NewBrandingSettings creates new Branding settings by loading all values from environment
func NewBrandingSettings() BrandingSettings {
	// Use APP_TITLE as fallback for BRANDING_TITLE (for backwards compatibility)
	title := getEnvOrDefaultString(ENV_BRANDING_TITLE, "")
	if title == "" {
		title = getEnvOrDefaultString(ENV_TITLE, DefaultBrandingTitle)
	}

	return BrandingSettings{
		Title:          title,
		Logo:           getEnvOrDefaultString(ENV_BRANDING_LOGO, DefaultBrandingLogo),
		Favicon:        getEnvOrDefaultString(ENV_BRANDING_FAVICON, DefaultBrandingFavicon),
		PrimaryColor:   getEnvOrDefaultString(ENV_BRANDING_PRIMARY_COLOR, DefaultBrandingPrimaryColor),
		SecondaryColor: getEnvOrDefaultString(ENV_BRANDING_SECONDARY_COLOR, DefaultBrandingSecondaryColor),
		AccentColor:    getEnvOrDefaultString(ENV_BRANDING_ACCENT_COLOR, DefaultBrandingAccentColor),
		CompanyName:    getEnvOrDefaultString(ENV_BRANDING_COMPANY_NAME, ""),
		CompanyUrl:     getEnvOrDefaultString(ENV_BRANDING_COMPANY_URL, ""),
	}
}
