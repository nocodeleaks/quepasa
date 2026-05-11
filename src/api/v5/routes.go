package v5

import "github.com/go-chi/chi/v5"

type Groups struct {
	Public    func(chi.Router)
	Protected func(chi.Router)
}

func RegisterControllers(r chi.Router, aliases []string, groups Groups) {
	for _, alias := range aliases {
		mountAlias(r, alias, func(router chi.Router) {
			if groups.Public != nil {
				groups.Public(router)
			}
			if groups.Protected != nil {
				router.Group(groups.Protected)
			}
		})
	}
}

func mountAlias(r chi.Router, alias string, register func(chi.Router)) {
	if alias == "" {
		r.Group(register)
		return
	}

	r.Route(alias, register)
}
