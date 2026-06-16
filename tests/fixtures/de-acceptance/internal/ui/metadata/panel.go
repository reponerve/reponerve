package metadata

import "example.com/deacceptance/internal/service/user"

// PanelRenderer renders metadata UI components.
type PanelRenderer interface {
	Render() error
}

// MetadataID is a stable metadata record identifier.
type MetadataID = string

// MetadataPanel is the primary metadata UI component.
type MetadataPanel struct {
	handler *user.Handler
}

// NewMetadataPanel constructs a metadata panel bound to the user service.
func NewMetadataPanel(h *user.Handler) *MetadataPanel {
	return &MetadataPanel{handler: h}
}

// Render displays the metadata panel.
func (p *MetadataPanel) Render() error {
	return p.handler.Handle()
}

func validateMetadataInput(id MetadataID) bool {
	return id != ""
}
