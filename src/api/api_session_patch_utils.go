package api

import runtime "github.com/nocodeleaks/quepasa/runtime"

func buildSessionConfigurationPatch(request interface{}) *runtime.SessionConfigurationPatch {
	patch := &runtime.SessionConfigurationPatch{}

	switch req := request.(type) {
	case *InfoCreateRequest:
		if req != nil {
			patch.Groups = req.Groups
			patch.Broadcasts = req.Broadcasts
			patch.ReadReceipts = req.ReadReceipts
			patch.Calls = req.Calls
			patch.ReadUpdate = req.ReadUpdate
			patch.Devel = req.Devel
		}
	case *InfoPatchRequest:
		if req != nil {
			patch.Groups = req.Groups
			patch.Broadcasts = req.Broadcasts
			patch.ReadReceipts = req.ReadReceipts
			patch.Calls = req.Calls
			patch.ReadUpdate = req.ReadUpdate
			patch.Devel = req.Devel
		}
	}

	return patch
}
